package gouring

import (
	"unsafe"
)

/*
 * GetBigCqe
 *
 * If the ring is initialized with IORING_SETUP_CQE32, then this field
 * contains 16-bytes of padding, doubling the size of the CQE.
 */
func (cqe *IoUringCqe) GetBigCqe() unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(cqe), SizeofIoUringCqe)
}

/*
 * GetOps
 *
 * Get io_uring probe ops
 */
func (probe *IoUringProbe) GetOps() unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(probe), SizeofIoUringProbe)
}
func (probe *IoUringProbe) GetOpAt(index int) *IoUringProbeOp {
	return (*IoUringProbeOp)(unsafe.Add(probe.GetOps(), SizeofIoUringProbeOp*uintptr(index)))
}

/*
 * GetBufs
 *
 * Get io_uring buf_ring bufs
 */
func (bring *IoUringBufRing) GetBufs() unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(bring), SizeofIoUringBufRing)
}
func (bring *IoUringBufRing) GetBufAt(index int) *IoUringBuf {
	return (*IoUringBuf)(unsafe.Add(bring.GetBufs(), SizeofIoUringBuf*uintptr(index)))
}
