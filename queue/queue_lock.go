package queue

import (
	"sync"

	"github.com/ii64/gouring"
)

type QueueLocks struct {
	*Queue

	sMx  sync.Mutex
	cqMx sync.Mutex
}

func NewWithLocks(ring *gouring.Ring) *QueueLocks {
	q := &QueueLocks{
		Queue: New(ring),
	}
	return q
}

func (q *QueueLocks) Submit() (ret int, err error) {
	q.sMx.Lock()
	defer q.sMx.Unlock()
	return q.Queue.Submit()
}

func (q *QueueLocks) SubmitAndWait(waitNr uint) (ret int, err error) {
	q.sMx.Lock()
	defer q.sMx.Unlock()
	return q.Queue.SubmitAndWait(waitNr)
}

//

func (q *QueueLocks) GetCQEntry(wait bool) (cqe *gouring.CQEntry, err error) {
	q.cqMx.Lock()
	defer q.cqMx.Unlock()
	return q.Queue.GetCQEntry(wait)
}

func (q *QueueLocks) GetCQEntryWait(wait bool, waitNr uint) (cqe *gouring.CQEntry, err error) {
	q.cqMx.Lock()
	defer q.cqMx.Unlock()
	return q.Queue.GetCQEntryWait(wait, waitNr)
}

func (q *QueueLocks) RunPoll(wait bool, waitNr uint, f QueueCQEHandler) (err error) {
	for q.precheck() == nil {
		cqe, err := q.GetCQEntryWait(wait, waitNr)
		if cqe == nil || err != nil {
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
	return
}

func (q *QueueLocks) Run(wait bool, f QueueCQEHandler) (err error) {
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
