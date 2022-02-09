package syscall

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/ii64/gouring"
	"github.com/ii64/gouring/queue"
	"github.com/stretchr/testify/assert"
)

func initRing(t *testing.T) (ring *gouring.Ring, q *queue.Queue, cq chan *gouring.CQEntry) {
	ring, err := gouring.New(128, nil)
	if err != nil {
		panic(err)
	}
	q = queue.New(ring)

	cq = make(chan *gouring.CQEntry)

	go q.RunPoll(true, 1, func(cqe *gouring.CQEntry) (err error) {
		fmt.Printf("got cqe: %+#v\n", cqe)
		cq <- cqe
		return nil
	})

	go func() {
		<-time.After(5 * time.Second)
		t.Logf("timeout")
		t.Fail()
	}()
	return
}

// func TestAccept(t *testing.T) {
// 	sqe := q.GetSQEntry()
// }

func TestRead(t *testing.T) {
	ring, q, cq := initRing(t)
	defer q.Close()
	defer ring.Close()

	var f *os.File
	var err error
	f, err = os.Open("/dev/urandom")
	assert.NoError(t, err, "urandom")
	fd := f.Fd()
	defer f.Close()

	b := make([]byte, 25)
	ud := uint64(gouring.IORING_OP_READ)
	sqe := q.GetSQEntry()
	Read(sqe, int(fd), b)
	sqe.UserData = ud

	ret, err := q.Submit()
	assert.NoError(t, err)
	assert.Equal(t, 1, ret, "mismatch submit return value")

	cqe := <-cq
	assert.Equal(t, ud, cqe.UserData)
	assert.Equal(t, len(b), int(cqe.Res))
}

func TestReadv(t *testing.T) {
	ring, q, cq := initRing(t)
	defer q.Close()
	defer ring.Close()

	var f *os.File
	var err error
	f, err = os.Open("/dev/urandom")
	assert.NoError(t, err, "urandom")
	fd := f.Fd()
	defer f.Close()

	bs := [][]byte{}
	bN := 25
	iovs := []syscall.Iovec{}
	iovN := 5
	for i := 0; i < iovN; i++ {
		b := make([]byte, bN)
		bs = append(bs, b)
		iovs = append(iovs, syscall.Iovec{
			Base: &b[0],
			Len:  uint64(len(b)),
		})
	}

	ud := uint64(gouring.IORING_OP_READV)
	sqe := q.GetSQEntry()
	Readv(sqe, int(fd), iovs)
	sqe.UserData = ud

	ret, err := q.Submit()
	assert.NoError(t, err)
	assert.Equal(t, 1, ret, "mismatch submit return value")

	cqe := <-cq
	assert.Equal(t, ud, cqe.UserData)
	assert.Equal(t, bN*iovN, int(cqe.Res))

	eb := make([]byte, bN)
	for i := 0; i < iovN; i++ {
		if bytes.Compare(bs[i], eb) == 0 {
			assert.NotEqual(t, eb, bs[i], "read urandom")
		}
	}
}

func TestWrite(t *testing.T) {
	ring, q, cq := initRing(t)
	defer q.Close()
	defer ring.Close()

	wr := "hello"
	ud := uint64(gouring.IORING_OP_WRITE)
	sqe := q.GetSQEntry()
	Write(sqe, syscall.Stdout, []byte(wr))
	sqe.UserData = ud

	ret, err := q.Submit()
	assert.NoError(t, err)
	assert.Equal(t, 1, ret, "mismatch submit return value")

	cqe := <-cq
	assert.Equal(t, ud, cqe.UserData)
	assert.Equal(t, len(wr), int(cqe.Res))
}

func TestWritev(t *testing.T) {
	ring, q, cq := initRing(t)
	defer q.Close()
	defer ring.Close()

	wr := "hello\n"
	bs := [][]byte{}
	iovs := []syscall.Iovec{}
	iovN := 5
	for i := 0; i < iovN; i++ {
		b := []byte(wr)
		bs = append(bs, b)
		iovs = append(iovs, syscall.Iovec{
			Base: &b[0],
			Len:  uint64(len(b)),
		})
	}

	ud := uint64(gouring.IORING_OP_WRITEV)
	sqe := q.GetSQEntry()
	Writev(sqe, syscall.Stdout, iovs)
	sqe.UserData = ud

	ret, err := q.Submit()
	assert.NoError(t, err)
	assert.Equal(t, 1, ret, "mismatch submit retrun value")

	cqe := <-cq
	assert.Equal(t, ud, cqe.UserData)
	assert.Equal(t, len(wr)*iovN, int(cqe.Res))
}
