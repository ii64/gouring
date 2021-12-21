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
	head        uintptr
	tail        uintptr
	ringMask    uintptr
	ringEntries uintptr
	flags       uintptr
	array       uint32Array
	sqes        sqeArray
}

func (sq SQRing) Get(idx uint32) *SQEntry {
	if uintptr(idx) >= uintptr(*sq.RingEntries()) {
		return nil
	}
	return sq.sqes.Get(uintptr(idx))
}
func (sq SQRing) Head() *uint32 {
	return (*uint32)(unsafe.Pointer(sq.head))
}
func (sq SQRing) Tail() *uint32 {
	return (*uint32)(unsafe.Pointer(sq.tail))
}
func (sq SQRing) RingMask() *uint32 {
	return (*uint32)(unsafe.Pointer(sq.ringMask))
}
func (sq SQRing) RingEntries() *uint32 {
	return (*uint32)(unsafe.Pointer(sq.ringEntries))
}
func (sq SQRing) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(sq.flags))
}
func (sq SQRing) Array() uint32Array {
	return sq.array
}
func (sq SQRing) Event() sqeArray {
	return sq.sqes
}

func (sq SQRing) IsCQOverflow() bool {
	return atomic.LoadUint32(sq.Flags())&IORING_SQ_CQ_OVERFLOW > 0
}
func (sq SQRing) IsNeedWakeup() bool {
	return atomic.LoadUint32(sq.Flags())&IORING_SQ_NEED_WAKEUP > 0
}

//
type uint32Array uintptr

func (a uint32Array) Get(idx uint32) *uint32 {
	return (*uint32)(unsafe.Pointer(uintptr(a) + uintptr(idx)*_sz_uint32))
}

func (a uint32Array) Set(idx uint32, v uint32) {
	atomic.StoreUint32(a.Get(idx), v)
}

type sqeArray uintptr

func (sa sqeArray) Get(idx uintptr) *SQEntry {
	return (*SQEntry)(unsafe.Pointer(uintptr(sa) + idx*_sz_sqe))
}

func (sa sqeArray) Set(idx uintptr, v SQEntry) {
	*sa.Get(idx) = v
}

//
//-- CQ

type CQRing struct {
	head        uintptr
	tail        uintptr
	ringMask    uintptr
	ringEntries uintptr
	cqes        cqeArray
}

func (cq CQRing) Get(idx uint32) *CQEntry {
	if uintptr(idx) >= uintptr(*cq.RingEntries()) { // avoid lookup overflow
		return nil
	}
	return cq.cqes.Get(uintptr(idx))
}
func (cq CQRing) Head() *uint32 {
	return (*uint32)(unsafe.Pointer(cq.head))
}
func (cq CQRing) Tail() *uint32 {
	return (*uint32)(unsafe.Pointer(cq.tail))
}
func (cq CQRing) RingMask() *uint32 {
	return (*uint32)(unsafe.Pointer(cq.ringMask))
}
func (cq CQRing) RingEntries() *uint32 {
	return (*uint32)(unsafe.Pointer(cq.ringEntries))
}
func (cq CQRing) Event() cqeArray {
	return cq.cqes
}

//

type cqeArray uintptr

func (ca cqeArray) Get(idx uintptr) *CQEntry {
	return (*CQEntry)(unsafe.Pointer(uintptr(ca) + idx*_sz_cqe))
}

func (ca cqeArray) Set(idx uintptr, v CQEntry) {
	*ca.Get(idx) = v
}
