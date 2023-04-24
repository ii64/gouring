package main

import (
	"fmt"
	"reflect"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	uring "github.com/ii64/gouring"
	"golang.org/x/sys/unix"
)

func main() {
	ring, err := uring.New(64, 0)
	if err != nil {
		panic(err)
	}
	defer ring.Close()

	fd, err := unix.MemfdCreate("mymemfd", unix.MFD_CLOEXEC)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	const BSIZE = 512
	unix.Ftruncate(fd, BSIZE)

	addr, err := mmap(nil, BSIZE, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, syscall.MAP_SHARED, fd, 0)
	if err != nil {
		panic(err)
	}
	defer munmap(addr, BSIZE)

	var rbuf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&rbuf))
	sh.Data = uintptr(addr)
	sh.Cap = BSIZE
	sh.Len = BSIZE

	tnow := func() string { return fmt.Sprintf("CLOCK:%d\n", time.Now().UnixMilli()) }

	go func() {
		flen := len(tnow())
		// monitor written bytes
		for {
			// copy
			payload := string(rbuf[:flen])
			fmt.Printf("> %q\n", payload)
			time.Sleep(time.Millisecond * 50)
		}
	}()

	var buf [BSIZE]byte
	refresh := func() int {
		b := []byte(tnow())
		copy(buf[:], b)
		return len(b)
	}

	qWrite := func() {
		sqe := ring.GetSqe()
		uring.PrepWrite(sqe, fd, &buf[0], refresh(), 0)
		sqe.UserData.SetUint64(0xaaaaaaaa)
	}
	qRead := func() {
		sqe := ring.GetSqe()
		uring.PrepRead(sqe, fd, &buf[0], len(buf), 0)
		sqe.UserData.SetUint64(0xbbbbbbbb)
	}

	qWrite()

	submitted, err := ring.SubmitAndWait(1)
	if err != nil {
		panic(err)
	}
	println("submitted:", submitted)

	var cqe *uring.IoUringCqe
	for {
		err = ring.WaitCqe(&cqe)
		switch err {
		case syscall.EINTR, syscall.EAGAIN, syscall.ETIME:
			runtime.Gosched()
			continue
		case nil:
			goto cont
		default:
			panic(err)
		}
	cont:
		switch cqe.UserData {
		case 0xaaaaaaaa:
			qRead()
		case 0xbbbbbbbb:
			qWrite()
		}

		ring.SeenCqe(cqe)
		submitted, err := ring.Submit()
		if err != nil {
			panic(err)
		} else {
			_ = submitted
			// println("submitted:", submitted)
		}
	}

}

//go:linkname mmap syscall.mmap
func mmap(addr unsafe.Pointer, length uintptr, prot int, flags int, fd int, offset int64) (xaddr unsafe.Pointer, err error)

//go:linkname munmap syscall.munmap
func munmap(addr unsafe.Pointer, length uintptr) (err error)

func msync(addr unsafe.Pointer, length uintptr, flags uintptr) error {
	r1, _, e1 := syscall.Syscall(syscall.SYS_MSYNC, uintptr(addr), length, flags)
	if e1 != 0 {
		return syscall.Errno(e1)
	}
	_ = r1
	return nil
}
