package syscall

import (
	"syscall"

	"github.com/ii64/gouring"
)

var _ = syscall.EpollCtl

// EpollCtl
func EpollCtl(sqe *gouring.SQEntry, epfd int, op int, fd int, event *syscall.EpollEvent) (err error) {
	sqe.Opcode = gouring.IORING_OP_EPOLL_CTL
	return nil
}
