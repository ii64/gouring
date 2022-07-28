package gouring

import "unsafe"

const (
	SizeofUnsigned   = unsafe.Sizeof(uint32(0))
	SizeofUint32     = unsafe.Sizeof(uint32(0))
	SizeofIoUringSqe = unsafe.Sizeof(IoUringSqe{})
	SizeofIoUringCqe = unsafe.Sizeof(IoUringCqe{})
)

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
	head        unsafe.Pointer // *uint32
	tail        unsafe.Pointer // *uint32
	ringMask    unsafe.Pointer // *uint32
	ringEntries unsafe.Pointer // *uint32
	flags       unsafe.Pointer // *uint32
	dropped     unsafe.Pointer // *uint32

	Array uint32Array     //ptr arith
	Sqes  ioUringSqeArray //ptr arith

	SqeHead uint32
	SqeTail uint32

	RingSz  uint32
	RingPtr unsafe.Pointer

	pad [4]uint32
}

func (sq *IoUringSq) _Head() *uint32        { return (*uint32)(sq.head) }
func (sq *IoUringSq) _Tail() *uint32        { return (*uint32)(sq.tail) }
func (sq *IoUringSq) _RingMask() *uint32    { return (*uint32)(sq.ringMask) }
func (sq *IoUringSq) _RingEntries() *uint32 { return (*uint32)(sq.ringEntries) }
func (sq *IoUringSq) _Flags() *uint32       { return (*uint32)(sq.flags) }
func (sq *IoUringSq) _Dropped() *uint32     { return (*uint32)(sq.dropped) }

type IoUringCq struct {
	head        unsafe.Pointer // *uint32
	tail        unsafe.Pointer // *uint32
	ringMask    unsafe.Pointer // *uint32
	ringEntries unsafe.Pointer // *uint32
	flags       unsafe.Pointer // *uint32
	overflow    unsafe.Pointer // *uint32

	Cqes ioUringCqeArray //ptr arith

	RingSz  uint32
	RingPtr unsafe.Pointer

	pad [4]uint32
}

func (cq *IoUringCq) _Head() *uint32        { return (*uint32)(cq.head) }
func (cq *IoUringCq) _Tail() *uint32        { return (*uint32)(cq.tail) }
func (cq *IoUringCq) _RingMask() *uint32    { return (*uint32)(cq.ringMask) }
func (cq *IoUringCq) _RingEntries() *uint32 { return (*uint32)(cq.ringEntries) }
func (cq *IoUringCq) _Flags() *uint32       { return (*uint32)(cq.flags) }
func (cq *IoUringCq) _Overflow() *uint32    { return (*uint32)(cq.overflow) }
