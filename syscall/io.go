package syscall

import (
	"syscall"
	"unsafe"

	"github.com/ii64/gouring"
)

var (
	_ = syscall.Accept

	_ = syscall.Read
	_ = syscall.Write

	_ = syscall.Close
)

// Accept
func Accept(sqe *gouring.SQEntry, lisFd int, raw *syscall.RawSockaddrAny) {
	sqe.Opcode = gouring.IORING_OP_ACCEPT
	sqe.Fd = int32(lisFd)
	var len uintptr = syscall.SizeofSockaddrAny
	*sqe.Addr2() = uint64(uintptr(unsafe.Pointer(&len)))
	*sqe.Addr() = uint64(uintptr(unsafe.Pointer(raw)))
}

// Read
func Read(sqe *gouring.SQEntry, fd int, b []byte) {
	sqe.Opcode = gouring.IORING_OP_READ
	sqe.Fd = int32(fd)
	sqe.Len = uint32(len(b))
	*sqe.Addr() = uint64(uintptr(unsafe.Pointer(&b[0])))
}

// Readv
func Readv(sqe *gouring.SQEntry, fd int, iovs []syscall.Iovec) {
	sqe.Opcode = gouring.IORING_OP_READV
	sqe.Fd = int32(fd)
	sqe.Len = uint32(len(iovs))
	*sqe.Addr() = uint64(uintptr(unsafe.Pointer(&iovs[0])))
}

// Write
func Write(sqe *gouring.SQEntry, fd int, b []byte) {
	sqe.Opcode = gouring.IORING_OP_WRITE
	sqe.Fd = int32(fd)
	sqe.Len = uint32(len(b))
	*sqe.Addr() = uint64(uintptr(unsafe.Pointer(&b[0])))
}

// Writev
func Writev(sqe *gouring.SQEntry, fd int, iovs []syscall.Iovec) {
	sqe.Opcode = gouring.IORING_OP_WRITEV
	sqe.Fd = int32(fd)
	sqe.Len = uint32(len(iovs))
	*sqe.Addr() = uint64(uintptr(unsafe.Pointer(&iovs[0])))
}

// Close
func Close(sqe *gouring.SQEntry, fd int) {
	sqe.Opcode = gouring.IORING_OP_CLOSE
	sqe.Fd = int32(fd)
}
