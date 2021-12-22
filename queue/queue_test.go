package queue

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"testing"

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

func TestQueue(t *testing.T) {
	ring, err := gouring.New(256, nil)
	assert.NoError(t, err, "create ring")
	defer func() {
		err := ring.Close()
		assert.NoError(t, err, "close ring")
	}()

	mkdata := func(i int) []byte {
		return []byte("queue pls" + strings.Repeat("!", i) + fmt.Sprintf("%d", i) + "\n")
	}

	N := 64 + 64
	var wg sync.WaitGroup
	btests := [][]byte{}
	for i := 0; i < N; i++ {
		btests = append(btests, mkdata(i))
	}
	wg.Add(N)

	// create new queue
	q := New(ring)
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
		q.Run(func(cqe *gouring.CQEntry) {
			defer wg.Done()
			assert.Condition(t, func() (success bool) {
				return cqe.UserData < uint64(len(btests))
			}, "userdata is set with the btest index")
			assert.Conditionf(t, func() (success bool) {
				return len(btests[cqe.UserData]) == int(cqe.Res)
			}, "OP_WRITE result mismatch: %+#v", cqe)
		})
	}()

	wg.Wait()
}
