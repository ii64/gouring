package gouring

import (
	"reflect"
	"unsafe"
)

var (
	_sqe    = SQEntry{}
	_sqe_mm = make([]byte, _sz_sqe)
	_sz_sqe = unsafe.Sizeof(_sqe)
	_cqe    = CQEntry{}
	_cqe_mm = make([]byte, _sz_cqe)
	_sz_cqe = unsafe.Sizeof(_cqe)
)

//-- SQEntry

type SQEntry struct {
	Opcode UringOpcode
	Flags  UringSQEFlag
	Ioprio uint16
	Fd     int32

	off__addr2          uint64 // union { off, addr2 }
	addr__splice_off_in uint64 // union { addr, splice_off_in }

	Len uint32

	opcode__flags_events uint32 // union of events and flags for opcode

	UserData uint64

	buf__index_group uint16 // union {buf_index, buf_group}

	Personality uint16

	splice_fd_in__file_index int32 // union { __s32 splice_fd_in, __u32 file_index }

	pad2 [2]uint64
}

func (sqe *SQEntry) Offset() *uint64 {
	return &sqe.off__addr2
}
func (sqe *SQEntry) SetOffset(v uint64) {
	*sqe.Offset() = v
}
func (sqe *SQEntry) Addr2() *uint64 {
	return &sqe.off__addr2
}
func (sqe *SQEntry) SetAddr2(v interface{}) {
	*sqe.Addr2() = (uint64)(reflect.ValueOf(v).Pointer())
}

func (sqe *SQEntry) Addr() *uint64 {
	return &sqe.addr__splice_off_in
}
func (sqe *SQEntry) SetAddr(v interface{}) {
	*sqe.Addr() = (uint64)(reflect.ValueOf(v).Pointer())
}
func (sqe *SQEntry) SpliceOffIn() *uint64 {
	return &sqe.addr__splice_off_in
}

func (sqe *SQEntry) OpcodeFlags() *uint32 {
	return &sqe.opcode__flags_events
}
func (sqe *SQEntry) OpodeEvents() *uint32 {
	return &sqe.opcode__flags_events
}

func (sqe *SQEntry) BufIndex() *uint16 {
	return &sqe.buf__index_group
}
func (sqe *SQEntry) BufGroup() *uint16 {
	return &sqe.buf__index_group
}

func (sqe *SQEntry) SpliceFdIn() *int32 {
	return &sqe.splice_fd_in__file_index
}
func (sqe *SQEntry) FileIndex() *uint32 {
	return (*uint32)(unsafe.Pointer(&sqe.splice_fd_in__file_index))
}

//

func (sqe *SQEntry) Reset() {
	*sqe = _sqe
}

//-- CQEntry

type CQEntry struct {
	UserData uint64 /* sqe->data submission passed back */
	Res      int32  /* result code for this event */
	Flags    UringCQEFlag
}

func (cqe *CQEntry) Reset() {
	*cqe = _cqe
}
