package syscall

import (
	"syscall"
	"unsafe"

	"github.com/ii64/gouring"
)

var _ = syscall.EpollCtl

// EpollCtl
func EpollCtl(sqe *gouring.SQEntry, epfd int, op int, fd int, event *syscall.EpollEvent) {
	sqe.Opcode = gouring.IORING_OP_EPOLL_CTL
	sqe.Fd = int32(epfd)
	*sqe.Addr() = uint64(uintptr(unsafe.Pointer(event)))
	sqe.Len = uint32(op)
	*sqe.Offset() = uint64(fd)
}
