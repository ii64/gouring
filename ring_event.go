package gouring

import "unsafe"

var (
	_sqe    = SQEvent{}
	_sqe_mm = make([]byte, _sz_sqe)
	_sz_sqe = unsafe.Sizeof(_sqe)
	_cqe    = CQEvent{}
	_cqe_mm = make([]byte, _sz_cqe)
	_sz_cqe = unsafe.Sizeof(_cqe)
)

type SQEvent struct {
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

func (sqe *SQEvent) Offset() *uint64 {
	return &sqe.off__addr2
}
func (sqe *SQEvent) Addr2() *uint64 {
	return &sqe.off__addr2
}

func (sqe *SQEvent) Addr() *uint64 {
	return &sqe.addr__splice_off_in
}
func (sqe *SQEvent) SpliceOffIn() *uint64 {
	return &sqe.addr__splice_off_in
}

func (sqe *SQEvent) OpcodeFlags() *uint32 {
	return &sqe.opcode__flags_events
}
func (sqe *SQEvent) OpodeEvents() *uint32 {
	return &sqe.opcode__flags_events
}

func (sqe *SQEvent) BufIndex() *uint16 {
	return &sqe.buf__index_group
}
func (sqe *SQEvent) BufGroup() *uint16 {
	return &sqe.buf__index_group
}

func (sqe *SQEvent) SpliceFdIn() *int32 {
	return &sqe.splice_fd_in__file_index
}
func (sqe *SQEvent) FileIndex() *uint32 {
	return (*uint32)(unsafe.Pointer(&sqe.splice_fd_in__file_index))
}

//

type CQEvent struct {
	UserData uint64 /* sqe->data submission passed back */
	Res      int32  /* result code for this event */
	Flags    UringCQEFlag
}
