package queue

// Modified form
// https://github.com/iceber/io_uring-go types.go

import (
	"errors"
	"runtime"
	"sync/atomic"
	"syscall"

	"github.com/ii64/gouring"
)

var (
	ErrQueueClosed = errors.New("queue closed")
)

type QueueCQEHandler func(cqe *gouring.CQEntry) (err error)

type Queue struct {
	ring *gouring.Ring
	sq   *gouring.SQRing
	cq   *gouring.CQRing

	sqeHead uint32
	sqeTail uint32

	err error

	clq uint32
}

func New(ring *gouring.Ring) *Queue {
	if ring == nil {
		return nil
	}
	sq := ring.SQ()
	cq := ring.CQ()
	return &Queue{
		ring: ring,
		sq:   sq,
		cq:   cq,
		clq:  0,
	}
}

func (q *Queue) Close() error {
	atomic.StoreUint32(&q.clq, 1)
	return nil
}
func (q *Queue) precheck() error {
	if clq := atomic.LoadUint32(&q.clq); clq == 1 {
		q.err = ErrQueueClosed
		return q.err
	}
	return nil
}

//

func (q *Queue) _getSQEntry() *gouring.SQEntry {
	head := atomic.LoadUint32(q.sq.Head())
	next := q.sqeTail + 1
	if (next - head) <= atomic.LoadUint32(q.sq.RingEntries()) {
		sqe := q.sq.Get(q.sqeTail & (*q.sq.RingMask()))
		q.sqeTail = next
		sqe.Reset()
		return sqe
	}
	return nil
}

func (q *Queue) GetSQEntry() (sqe *gouring.SQEntry) {
	for {
		sqe = q._getSQEntry()
		if sqe != nil {
			return
		}
		// runtime.Gosched()
	}
}

func (q *Queue) sqFallback(d uint32) {
	q.sqeTail -= d
}

func (q *Queue) sqFlush() uint32 {
	khead := atomic.LoadUint32(q.sq.Head())
	ktail := atomic.LoadUint32(q.sq.Tail())
	if q.sqeHead == q.sqeTail {
		return ktail - khead
	}

	for toSubmit := q.sqeTail - q.sqeHead; toSubmit > 0; toSubmit-- {
		*q.sq.Array().Get(ktail & (*q.sq.RingMask())) = q.sqeHead & (*q.sq.RingMask())
		ktail++
		q.sqeHead++
	}
	atomic.StoreUint32(q.sq.Tail(), ktail)
	return ktail - khead
}

func (q *Queue) isNeedEnter(flags *uint32) bool {
	if (q.ring.Params().Flags & gouring.IORING_SETUP_SQPOLL) == 0 {
		return true
	}
	if q.sq.IsNeedWakeup() {
		*flags |= gouring.IORING_SQ_NEED_WAKEUP
		return true
	}
	return false
}

func (q *Queue) Submit() (ret int, err error) {
	// q.sMx.Lock()
	// defer q.sMx.Unlock()
	return q.SubmitAndWait(0)
}

func (q *Queue) SubmitAndWait(waitNr uint) (ret int, err error) {
	// q.sMx.Lock()
	// defer q.sMx.Unlock()
	submitted := q.sqFlush()

	var flags uint32
	if !q.isNeedEnter(&flags) || submitted == 0 {
		return
	}

	if (q.ring.Params().Flags & gouring.IORING_SETUP_IOPOLL) > 0 {
		flags |= gouring.IORING_ENTER_GETEVENTS
	}

	ret, err = q.ring.Enter(uint(submitted), waitNr, flags, nil)
	return
}

//

func (q *Queue) cqPeek() (cqe *gouring.CQEntry) {
	khead := atomic.LoadUint32(q.cq.Head())
	if khead != atomic.LoadUint32(q.cq.Tail()) {
		cqe = q.cq.Get(khead & atomic.LoadUint32(q.cq.RingMask()))
	}
	return
}

func (q *Queue) cqAdvance(d uint32) {
	if d != 0 {
		atomic.AddUint32(q.cq.Head(), d) // mark readed
	}
}

func (q *Queue) GetCQEntry(wait bool) (cqe *gouring.CQEntry, err error) {
	return q.GetCQEntryWait(wait, 0)
}

func (q *Queue) GetCQEntryWait(wait bool, waitNr uint) (cqe *gouring.CQEntry, err error) {
	// q.cqMx.Lock()
	// defer q.cqMx.Unlock()
	if err = q.precheck(); err != nil {
		return
	}
	var tryPeeks int
	for {
		if cqe = q.cqPeek(); cqe != nil {
			q.cqAdvance(1)
			return
		}

		if !wait && !q.sq.IsCQOverflow() {
			err = syscall.EAGAIN
			return
		} else if waitNr > 0 {
			_, err = q.ring.Enter(0, waitNr, gouring.IORING_ENTER_GETEVENTS, nil)
			if err != nil {
				return
			}
		}

		if q.sq.IsCQOverflow() {
			_, err = q.ring.Enter(0, 0, gouring.IORING_ENTER_GETEVENTS, nil)
			if err != nil {
				return
			}
			continue
		}

		if tryPeeks++; tryPeeks < 3 {
			runtime.Gosched()
			continue
		}
		// implement interrupt
	}
}

func (q *Queue) Err() error {
	return q.err
}

func (q *Queue) RunPoll(wait bool, waitNr uint, f QueueCQEHandler) (err error) {
	for q.precheck() == nil {
		cqe, err := q.GetCQEntryWait(wait, waitNr)
		if cqe == nil || err != nil {
			q.err = err
			if err == ErrQueueClosed {
				return err
			}
			continue
		}

		err = f(cqe)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) Run(wait bool, f QueueCQEHandler) (err error) {
	for q.precheck() == nil {
		cqe, err := q.GetCQEntry(wait)
		if cqe == nil || err != nil {
			q.err = err
			if err == ErrQueueClosed {
				return err
			}
			continue
		}

		err = f(cqe)
		if err != nil {
			return err
		}
	}
	return nil
}
