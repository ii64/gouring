package queue

// Modified form
// https://github.com/iceber/io_uring-go types.go

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ii64/gouring"
)

type Queue struct {
	ring *gouring.Ring
	sq   *gouring.SQRing
	cq   *gouring.CQRing

	sqeHead uint32
	sqeTail uint32

	cqMx sync.RWMutex // tbd...
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
	}
}

//

func (q *Queue) _getSQEntry() *gouring.SQEntry {
	head := atomic.LoadUint32(q.sq.Head())
	next := q.sqeTail + 1
	if (next - head) <= atomic.LoadUint32(q.sq.RingEntries()) {
		sqe := q.sq.Get(q.sqeTail & atomic.LoadUint32(q.sq.RingMask()))
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
		runtime.Gosched()
	}
}

func (q *Queue) sqFallback(d uint32) {
	q.sqeTail -= d
}

func (q *Queue) sqFlush() uint32 {
	ktail := atomic.LoadUint32(q.sq.Tail())
	if q.sqeHead == q.sqeTail {
		return ktail - atomic.LoadUint32(q.sq.Head())
	}

	for toSubmit := q.sqeTail; toSubmit > 0; toSubmit-- {
		kmask := *q.sq.RingMask()
		*q.sq.Array().Get(ktail & kmask) = q.sqeHead & kmask

		ktail++
		q.sqeHead++
	}
	atomic.StoreUint32(q.sq.Tail(), ktail)
	return ktail - *q.sq.Head()
}

func (q *Queue) isNeedEnter(flags *uint32) bool {
	if (q.ring.Params().Features & gouring.IORING_SETUP_SQPOLL) > 0 {
		return true
	}
	if q.sq.IsNeedWakeup() {
		*flags |= gouring.IORING_SQ_NEED_WAKEUP
		return true
	}
	return false
}

func (q *Queue) Submit() (ret int, err error) {
	submitted := q.sqFlush()

	var flags uint32
	if !q.isNeedEnter(&flags) || submitted == 0 {
		return
	}

	if q.ring.Params().Flags&gouring.IORING_SETUP_IOPOLL > 0 {
		flags |= gouring.IORING_ENTER_GETEVENTS
	}

	ret, err = q.ring.Enter(uint(submitted), 0, flags, nil)
	return
}

//

func (q *Queue) cqPeek() (cqe *gouring.CQEntry) {
	if atomic.LoadUint32(q.cq.Head()) != atomic.LoadUint32(q.cq.Tail()) {
		cqe = q.cq.Get(atomic.LoadUint32(q.cq.Head()) & atomic.LoadUint32(q.cq.RingMask()))
	}
	return
}

func (q *Queue) cqAdvance(d uint32) {
	if d != 0 {
		atomic.AddUint32(q.cq.Head(), d)
	}
}

func (q *Queue) getCQEvent(wait bool) (cqe *gouring.CQEntry, err error) {
	var tryPeeks int
	for {
		if cqe = q.cqPeek(); cqe != nil {
			q.cqAdvance(1)
			return
		}

		if !wait && !q.sq.IsCQOverflow() {
			err = syscall.EAGAIN
			return
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

func (q *Queue) Run(f func(cqe *gouring.CQEntry)) {
	for {
		cqe, err := q.getCQEvent(true)
		if cqe == nil || err != nil {
			fmt.Printf("run error: %+#v\n", err)
			continue
		}

		f(cqe)
	}
}
