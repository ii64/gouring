package main

import (
	"log"

	"github.com/ii64/gouring"
	"golang.org/x/sys/unix"
)

func main() {

	h, err := gouring.New(256, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fd, err := unix.Open("/tmp/test", unix.O_RDWR, 0677)

	sqe := h.GetSqe()
	b := make([]byte, 20)
	gouring.PrepRead(sqe, fd, &b[0], len(b), 0)
	log.Println("Buffer: ", b)

	submitted, err := h.SubmitAndWait(1)
	if err != nil {
		log.Fatal(err)
	}
	println(submitted) // 1

	var cqe *gouring.IoUringCqe
	err = h.WaitCqe(&cqe)
	if err != nil {
		log.Fatal(err)
	} // check also EINTR
	if cqe == nil {
		log.Fatal("CQE is NULL!")
	}
	h.SeenCqe(cqe)

	log.Println("CQE: ", cqe)
	log.Println("Buffer: ", b)
	log.Println("Buffer: ", string(b))

	_ = cqe.UserData
	_ = cqe.Res
	_ = cqe.Flags
}
