package queue

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"

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

func TestQueueSQPoll(t *testing.T) {
	ring, err := gouring.New(256, &gouring.IOUringParams{
		Flags:        gouring.IORING_SETUP_SQPOLL,
		SQThreadIdle: 70 * 1000,
	})
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
	defer func() {
		err := q.Close()
		assert.NoError(t, err, "close queue")
	}()

	//

	go func() {
		for i, b := range btests {
			sqe := q.GetSQEntry()
			sqe.UserData = uint64(i)
			// sqe.Flags = gouring.IOSQE_IO_DRAIN
			write(sqe, syscall.Stdout, b)
			if (i+1)%2 == 0 {
				n, err := q.Submit()
				assert.NoError(t, err, "queue submit")
				// assert.Equal(t, 2, n, "submit count mismatch") // may varies due to kernel thread consumng submission queue process time
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
	defer func() {
		err := q.Close()
		assert.NoError(t, err, "close queue")
	}()

	//

	go func() {
		for i, b := range btests {
			sqe := q.GetSQEntry()
			sqe.UserData = uint64(i)
			// sqe.Flags = gouring.IOSQE_IO_DRAIN
			write(sqe, syscall.Stdout, b)
			if (i+1)%2 == 0 {
				n, err := q.Submit()
				assert.NoError(t, err, "queue submit")
				assert.Equal(t, 2, n, "submit count mismatch")
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
	defer func() {
		err := q.Close()
		assert.NoError(t, err, "close queue")
	}()

	// test data

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

//

var (
	Entries = 512
	ring    *gouring.Ring
	q       *Queue
)

func init() {
	var err error
	ring, err = gouring.New(uint(Entries), nil)
	if err != nil {
		panic(err)
	}
	q = New(ring)
}

func BenchmarkQueueBatchingNOP(b *testing.B) {
	var sqe *gouring.SQEntry
	for j := 0; j < b.N; j++ {
		for i := 0; i < Entries; i++ {
			sqe = q.GetSQEntry()
			sqe.Opcode = gouring.IORING_OP_NOP
			sqe.UserData = uint64(i)
		}
		n, err := q.SubmitAndWait(uint(Entries))
		assert.NoError(b, err, "submit")
		assert.Equal(b, Entries, n, "submit result entries")
		for i := 0; i < Entries; i++ {
			v := uint64(i)
			cqe, err := q.GetCQEntry(true)
			assert.NoError(b, err, "cqe wait error")
			assert.Equal(b, int32(0), cqe.Res)
			assert.Equal(b, v, cqe.UserData)
		}
	}
}

//

var (
	_sqe    StructTest
	_sz_sqe = unsafe.Sizeof(_sqe)
	_sqe_mm = make([]byte, _sz_sqe)

	m = &StructTest{}
)

type StructTest = gouring.SQEntry

func BenchmarkSetPtrVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		*m = _sqe
	}
}

func BenchmarkSetPtrValEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		*m = StructTest{}
	}
}

func BenchmarkSetPtrCpy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		copy(*(*[]byte)(unsafe.Pointer(m)), _sqe_mm)
	}
}
