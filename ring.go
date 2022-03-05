package gouring

import (
	"sync/atomic"
	"unsafe"
)

type Ring struct {
	fd     int
	params IOUringParams
	sq     SQRing
	cq     CQRing

	// cached ringn value
	sqRingPtr, cqRingPtr, sqesPtr uintptr
	sringSz, cringSz              uint32
}

//
//-- SQ

type SQRing struct {
	Head        *uint32
	Tail        *uint32
	RingMask    *uint32
	RingEntries *uint32
	Flags       *uint32
	Array       uint32Array
	Event       []SQEntry
}

func (sq SQRing) IsCQOverflow() bool {
	return atomic.LoadUint32(sq.Flags)&IORING_SQ_CQ_OVERFLOW > 0
}
func (sq SQRing) IsNeedWakeup() bool {
	return atomic.LoadUint32(sq.Flags)&IORING_SQ_NEED_WAKEUP > 0
}

//
type uint32Array uintptr

func (a uint32Array) Get(idx uint32) *uint32 {
	return (*uint32)(unsafe.Pointer(uintptr(a) + uintptr(idx)*_sz_uint32))
}

func (a uint32Array) Set(idx uint32, v uint32) {
	atomic.StoreUint32(a.Get(idx), v)
}

//
//-- CQ

type CQRing struct {
	Head        *uint32
	Tail        *uint32
	RingMask    *uint32
	RingEntries *uint32
	Event       []CQEntry
}
