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
	// type of operation for this sqe
	Opcode UringOpcode

	// IOSQE_ flags
	Flags UringSQEFlag

	// ioprio for the request
	Ioprio uint16

	// file descriptor to do IO on
	Fd int32

	/* union {
	  off,            // offset into file
	  addr2;
	} */
	off__addr2 uint64

	/* union {
	  addr,           // pointer to buffer or iovecs
	  splice_off_in
	} */
	addr__splice_off_in uint64

	// buffer size or number iovecs
	Len uint32

	/* union of events and flags for Opcode
	union {
		__kernel_rwf_t	rw_flags;
		__u32		fsync_flags;
		__u16		poll_events;	  // compatibility
		__u32		poll32_events;	  // word-reversed for BE
		__u32		sync_range_flags;
		__u32		msg_flags;
		__u32		timeout_flags;
		__u32		accept_flags;
		__u32		cancel_flags;
		__u32		open_flags;
		__u32		statx_flags;
		__u32		fadvise_advice;
		__u32		splice_flags;
		__u32		rename_flags;
		__u32		unlink_flags;
		__u32		hardlink_flags;
		__u32		xattr_flags;
	} */
	opcode__flags_events uint32

	// data to be passed back at completion time
	UserData uint64

	/* pack this to avoid bogus arm OABI complaints
	union {
		// index into fixed buffers, if used
		__u16	buf_index;
		// for grouped buffer selection
		__u16	buf_group;
	} __attribute__((packed)); */
	buf__index_group uint16

	// personality to use, if used
	Personality uint16

	/* -
	union {
		__s32	splice_fd_in;
		__u32	file_index;
	} */
	splice_fd_in__file_index int32

	addr3 uint64
	pad2  [1]uint64
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

func (sqe *SQEntry) Addr3() *uint64 {
	return (*uint64)(unsafe.Pointer(&sqe.addr3))
}
func (sqe *SQEntry) SetAddr3(v interface{}) {
	*sqe.Addr3() = (uint64)(reflect.ValueOf(v).Pointer())
}

//

func (sqe *SQEntry) Reset() {
	*sqe = SQEntry{}
}

//-- CQEntry

type CQEntry struct {
	UserData uint64 /* sqe->data submission passed back */
	Res      int32  /* result code for this event */
	Flags    UringCQEFlag
}

func (cqe *CQEntry) Reset() {
	*cqe = CQEntry{}
}
