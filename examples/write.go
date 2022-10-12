package main

import (
	"github.com/ii64/gouring"
	"golang.org/x/sys/unix"
	"log"
)

func main() {

	h, err := gouring.New(256, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fd, err := unix.Open("/tmp/test", unix.O_RDWR, 0677)

	sqe := h.GetSqe()
	b := []byte("hello from io_uring!\n")
	gouring.PrepWrite(sqe, fd, &b[0], len(b), 0)

	submitted, err := h.SubmitAndWait(1)
	if err != nil { /*...*/
	}
	println(submitted) // 1

	var cqe *gouring.IoUringCqe
	err = h.WaitCqe(&cqe)
	if err != nil {
		log.Fatal(err)
	} // check also EINTR

	_ = cqe.UserData
	_ = cqe.Res
	_ = cqe.Flags
}
