package gouring

import "unsafe"

const (
	SizeofUnsigned     = unsafe.Sizeof(uint32(0))
	SizeofUint32       = unsafe.Sizeof(uint32(0))
	SizeofIoUringSqe   = unsafe.Sizeof(IoUringSqe{})
	Align128IoUringSqe = 64
	SizeofIoUringCqe   = unsafe.Sizeof(IoUringCqe{})
	Align32IoUringCqe  = SizeofIoUringCqe

	SizeofIoUringProbe   = unsafe.Sizeof(IoUringProbe{})
	SizeofIoUringProbeOp = unsafe.Sizeof(IoUringProbeOp{})
	SizeofIoUringBufRing = unsafe.Sizeof(IoUringBufRing{})
	SizeofIoUringBuf     = unsafe.Sizeof(IoUringBuf{})
)

func _SizeChecker() {
	var x [1]struct{}
	_ = x[SizeofIoUringSqe-64]
	_ = x[SizeofIoUringCqe-16]
	_ = x[SizeofIoUringProbe-16]
	_ = x[SizeofIoUringProbeOp-8]
	_ = x[SizeofIoUringBufRing-16]
	_ = x[SizeofIoUringBuf-16]
}

type IoUring struct {
	Sq     IoUringSq
	Cq     IoUringCq
	Flags  uint32
	RingFd int

	Features    uint32
	EnterRingFd int
	IntFlags    uint8

	pad  [3]uint8
	pad2 uint32
}

type IoUringSq struct {
	khead        unsafe.Pointer // *uint32
	ktail        unsafe.Pointer // *uint32
	kringMask    unsafe.Pointer // *uint32
	kringEntries unsafe.Pointer // *uint32
	kflags       unsafe.Pointer // *uint32
	kdropped     unsafe.Pointer // *uint32

	Array uint32Array     //ptr arith
	Sqes  ioUringSqeArray //ptr arith

	SqeHead uint32
	SqeTail uint32

	RingSz  uint32
	RingPtr unsafe.Pointer

	RingMask, RingEntries uint32

	pad [2]uint32
}

func (sq *IoUringSq) _KHead() *uint32        { return (*uint32)(sq.khead) }
func (sq *IoUringSq) _KTail() *uint32        { return (*uint32)(sq.ktail) }
func (sq *IoUringSq) _KRingMask() *uint32    { return (*uint32)(sq.kringMask) }
func (sq *IoUringSq) _KRingEntries() *uint32 { return (*uint32)(sq.kringEntries) }
func (sq *IoUringSq) _KFlags() *uint32       { return (*uint32)(sq.kflags) }
func (sq *IoUringSq) _KDropped() *uint32     { return (*uint32)(sq.kdropped) }

type IoUringCq struct {
	khead        unsafe.Pointer // *uint32
	ktail        unsafe.Pointer // *uint32
	kringMask    unsafe.Pointer // *uint32
	kringEntries unsafe.Pointer // *uint32
	kflags       unsafe.Pointer // *uint32
	koverflow    unsafe.Pointer // *uint32

	Cqes ioUringCqeArray //ptr arith

	RingSz  uint32
	RingPtr unsafe.Pointer

	RingMask, RingEntries uint32

	pad [2]uint32
}

func (cq *IoUringCq) _KHead() *uint32        { return (*uint32)(cq.khead) }
func (cq *IoUringCq) _KTail() *uint32        { return (*uint32)(cq.ktail) }
func (cq *IoUringCq) _KRingMask() *uint32    { return (*uint32)(cq.kringMask) }
func (cq *IoUringCq) _KRingEntries() *uint32 { return (*uint32)(cq.kringEntries) }
func (cq *IoUringCq) _KFlags() *uint32       { return (*uint32)(cq.kflags) }
func (cq *IoUringCq) _KOverflow() *uint32    { return (*uint32)(cq.koverflow) }
