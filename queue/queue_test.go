package queue

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ii64/gouring"
	"github.com/stretchr/testify/assert"
)

func write(sqe *gouring.SQEntry, fd int, b []byte) {
	sqe.Opcode = gouring.IORING_OP_WRITE
	sqe.Fd = int32(fd)
	sqe.Len = uint32(len(b))
	sqe.SetOffset(0)
	// *sqe.Addr() = (uint64)(uintptr(unsafe.Pointer(&b[0])))
	sqe.SetAddr(&b[0])
}

func mkdata(i int) []byte {
	return []byte("queue pls" + strings.Repeat("!", i) + fmt.Sprintf("%d", i) + "\n")
}

func TestQueue(t *testing.T) {
	ring, err := gouring.New(256, nil)
	assert.NoError(t, err, "create ring")
	defer func() {
		err := ring.Close()
		assert.NoError(t, err, "close ring")
	}()

	N := 64 + 64
	var wg sync.WaitGroup
	btests := [][]byte{}
	for i := 0; i < N; i++ {
		btests = append(btests, mkdata(i))
	}
	wg.Add(N)

	// create new queue
	q := New(ring)
	defer q.Close()
	go func() {
		for i, b := range btests {
			sqe := q.GetSQEntry()
			sqe.UserData = uint64(i)
			// sqe.Flags = gouring.IOSQE_IO_DRAIN
			write(sqe, syscall.Stdout, b)
			if (i+1)%2 == 0 {
				n, err := q.Submit()
				assert.NoError(t, err, "queue submit")
				assert.Equal(t, n, 2, "submit count mismatch")
				fmt.Printf("submitted %d\n", n)
			}
		}
	}()
	go func() {
		q.Run(true, func(cqe *gouring.CQEntry) (err error) {
			defer wg.Done()
			fmt.Printf("cqe: %+#v\n", cqe)
			assert.Condition(t, func() (success bool) {
				return cqe.UserData < uint64(len(btests))
			}, "userdata is set with the btest index")
			assert.Conditionf(t, func() (success bool) {
				return len(btests[cqe.UserData]) == int(cqe.Res)
			}, "OP_WRITE result mismatch: %+#v", cqe)

			return nil
		})
	}()

	wg.Wait()
}

func TestQueuePolling(t *testing.T) {
	ring, err := gouring.New(64, &gouring.IOUringParams{})
	assert.NoError(t, err, "create ring")
	defer func() {
		err := ring.Close()
		assert.NoError(t, err, "close ring")
	}()

	q := New(ring)
	defer q.Close()

	var tb = []byte("write me on stdout\n")
	var tu uint64 = 0xfafa
	chDone := make(chan struct{}, 1)

	go q.RunPoll(true, 1, func(cqe *gouring.CQEntry) (err error) {
		if cqe.Res < 0 {
			t.Error(syscall.Errno(cqe.Res * -1))
		}
		assert.Equal(t, uint64(tu), uint64(cqe.UserData), "mismatch userdata")
		assert.Equal(t, uint64(len(tb)), uint64(cqe.Res), "mismatch written size")
		chDone <- struct{}{}
		return nil
	})

	t.Log("wait 3 second...")
	sqe := q.GetSQEntry()
	sqe.UserData = tu
	write(sqe, syscall.Stdout, tb)

	n, err := q.Submit()
	assert.NoError(t, err, "submit")
	assert.Equal(t, 1, n, "submitted count")

	select {
	case <-chDone:
	case <-time.After(time.Second * 5):
		t.Error("timeout wait")
		t.Fail()
	}
}
