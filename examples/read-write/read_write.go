package main

import (
	"bytes"
	"fmt"
	"log"
	"syscall"

	uring "github.com/ii64/gouring"
	"golang.org/x/sys/unix"
)

func main() {
	ring, err := uring.New(256, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer ring.Close()

	fd, err := unix.Open("/tmp/test", unix.O_RDWR|unix.O_CREAT, 0677)
	if err != nil {
		panic(err)
	}

	sqe := ring.GetSqe()
	b := []byte("hello from io_uring!\n")
	uring.PrepWrite(sqe, fd, &b[0], len(b), 0)
	sqe.UserData.SetUint64(0x0001)
	sqe.Flags |= uring.IOSQE_IO_LINK

	sqe = ring.GetSqe()
	uring.PrepWrite(sqe, syscall.Stdout, &b[0], len(b), 0)
	sqe.UserData.SetUint64(0x0002)
	sqe.Flags |= uring.IOSQE_IO_LINK

	sqe = ring.GetSqe()
	var buf = make([]byte, len(b))
	uring.PrepRead(sqe, fd, &buf[0], len(buf), 0)
	sqe.UserData.SetUint64(0x0003)
	sqe.Flags |= uring.IOSQE_IO_LINK

	sqe = ring.GetSqe()
	uring.PrepClose(sqe, fd)
	sqe.UserData.SetUint64(0x0004)

	const N = 4
	submitted, err := ring.SubmitAndWait(N)
	if err != nil { /*...*/
		log.Fatal(err)
	}
	println(submitted) // 1

	var cqe *uring.IoUringCqe
	for i := 1; i <= N; i++ {
		err = ring.WaitCqe(&cqe)
		if err != nil {
			log.Fatal(err)
		} // check also EINTR
		if cqe == nil {
			log.Fatal("CQE is NULL!")
		}
		log.Printf("CQE: %+#v\n", cqe)
		if uring.UserData(i) != cqe.UserData {
			panic("UNORDERED")
		}

		if cqe.Res < 0 {
			panic(syscall.Errno(-cqe.Res))
		}

		if i == 0x0003 {
			retvb := buf[:cqe.Res]
			fmt.Printf("retv buf %+#v\n", retvb)
			if !bytes.Equal(b, retvb) {
				panic("RET BUF NOT EQ")
			}
		}

		ring.SeenCqe(cqe)
	}
	_ = cqe.UserData
	_ = cqe.Res
	_ = cqe.Flags
}
