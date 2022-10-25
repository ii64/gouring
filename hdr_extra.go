package gouring

import (
	"unsafe"
)

/*
 * If the ring is initialized with IORING_SETUP_CQE32, then this field
 * contains 16-bytes of padding, doubling the size of the CQE.
 */
func (cqe *IoUringCqe) GetBigCqe() *[2]uint64 {
	return (*[2]uint64)(unsafe.Pointer(uintptr(unsafe.Pointer(cqe)) + SizeofIoUringSqe))
}
