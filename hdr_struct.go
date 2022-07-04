package gouring

import "unsafe"

var (
	SizeofUnsigned   = unsafe.Sizeof(uint32(0))
	SizeofUint32     = unsafe.Sizeof(uint32(0))
	SizeofIoUringSqe = unsafe.Sizeof(IoUringSqe{})
	SizeofIoUringCqe = unsafe.Sizeof(IoUringCqe{})
)

type IoUring struct {
	Sq     IoUringSq
	Cq     IoUringCq
	Flags  uint32
	RingFd int32

	Features    uint32
	EnterRingFd int32
	IntFlags    uint8

	pad  [3]uint8
	pad2 uint32
}

type IoUringSq struct {
	Head        *uint32
	Tail        *uint32
	RingMask    *uint32
	RingEntries *uint32
	Flags       *uint32
	Dropped     *uint32

	Array uint32Array     //ptr arith
	Sqes  ioUringSqeArray //ptr arith

	SqeHead uint32
	SqeTail uint32

	RingSz  uint32
	RingPtr uintptr

	pad [4]uint32
}

type IoUringCq struct {
	Head        *uint32
	Tail        *uint32
	RingMask    *uint32
	RingEntries *uint32
	Flags       *uint32
	Overflow    *uint32

	Cqes ioUringCqeArray //ptr arith

	RingSz  uint32
	RingPtr uintptr

	pad [4]uint32
}
