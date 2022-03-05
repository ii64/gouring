package syscall

import (
	"syscall"
	"unsafe"

	"github.com/ii64/gouring"
)

var _ = syscall.Madvise

// Madvise
func Madvise(sqe *gouring.SQEntry, b []byte, advice int) {
	var ptr unsafe.Pointer
	if len(b) > 0 {
		ptr = unsafe.Pointer(&b[0])
	} else {
		ptr = unsafe.Pointer(&_zero)
	}
	sqe.Opcode = gouring.IORING_OP_MADVISE
	sqe.Fd = -1
	*sqe.Addr() = uint64(uintptr(ptr))
	sqe.Len = uint32(len(b))
	*sqe.Offset() = 0
}
