package gouring

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	_uint32    uint32 = 0
	_sz_uint32        = unsafe.Sizeof(_uint32)
)

//go:linkname mmap syscall.mmap
func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (xaddr uintptr, err error)

//go:linkname munmap syscall.munmap
func munmap(addr uintptr, length uintptr) (err error)

// io uring setup
func setup(r *Ring, entries uint, parmas *IOUringParams) (ringFd int, err error) {
	var sq = &r.sq
	var cq = &r.cq
	var p = &r.params

	if ringFd, err = io_uring_setup(entries, p); err != nil {
		err = errors.Wrap(err, "io_uring_setup")
		return
	}
	if ringFd < 0 {
		err = syscall.EAGAIN
		return
	}

	featSingleMap := p.Features&IORING_FEAT_SINGLE_MMAP > 0

	r.sringSz = p.SQOff.Array + p.SQEntries*uint32(_sz_uint32)
	r.cringSz = p.CQOff.CQEs + p.CQEntries*uint32(_sz_cqe)
	if featSingleMap {
		if r.cringSz > r.sringSz {
			r.sringSz = r.cringSz
		}
	}

	// allocate ring mem
	var sqRingPtr uintptr
	var cqRingPtr uintptr
	sqRingPtr, err = mmap(0, uintptr(r.sringSz),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE,
		ringFd, IORING_OFF_SQ_RING)
	if err != nil {
		err = errors.Wrap(err, "mmap sqring")
		return
	}
	if featSingleMap {
		cqRingPtr = sqRingPtr
	} else {
		cqRingPtr, err = mmap(0, uintptr(r.cringSz),
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED|syscall.MAP_POPULATE,
			ringFd, IORING_OFF_CQ_RING)
		if err != nil {
			err = errors.Wrap(err, "mmap cqring")
			return
		}
	}
	r.sqRingPtr = sqRingPtr
	r.cqRingPtr = cqRingPtr

	// SQ

	sq.Head = (*uint32)(unsafe.Pointer(sqRingPtr + uintptr(p.SQOff.Head)))
	sq.Tail = (*uint32)(unsafe.Pointer(sqRingPtr + uintptr(p.SQOff.Tail)))
	sq.RingMask = (*uint32)(unsafe.Pointer(sqRingPtr + uintptr(p.SQOff.RingMask)))
	sq.RingEntries = (*uint32)(unsafe.Pointer(sqRingPtr + uintptr(p.SQOff.RingEntries)))
	sq.Flags = (*uint32)(unsafe.Pointer(sqRingPtr + uintptr(p.SQOff.Flags)))
	sq.Array = uint32Array(sqRingPtr + uintptr(p.SQOff.Array)) // non fixed array size, controlled by ring mask
	r.sqesPtr, err = mmap(0, uintptr(p.SQEntries*uint32(_sz_sqe)),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE,
		ringFd, IORING_OFF_SQES)
	if err != nil {
		err = errors.Wrap(err, "mmap sqes")
		return
	}

	sq.Event = *(*[]SQEntry)(unsafe.Pointer(&reflect.SliceHeader{
		Data: r.sqesPtr,
		Len:  int(p.SQEntries),
		Cap:  int(p.SQEntries),
	}))

	fmt.Printf("insp %+#v %+#v\n", len(sq.Event), cap(sq.Event))

	// CQ

	cq.Head = (*uint32)(unsafe.Pointer(cqRingPtr + uintptr(p.CQOff.Head)))
	cq.Tail = (*uint32)(unsafe.Pointer(cqRingPtr + uintptr(p.CQOff.Tail)))
	cq.RingMask = (*uint32)(unsafe.Pointer(cqRingPtr + uintptr(p.CQOff.RingMask)))
	cq.RingEntries = (*uint32)(unsafe.Pointer(cqRingPtr + uintptr(p.CQOff.RingEntries)))
	cqesPtr := cqRingPtr + uintptr(p.CQOff.CQEs)

	cq.Event = *(*[]CQEntry)(unsafe.Pointer(&reflect.SliceHeader{
		Data: cqesPtr,
		Len:  int(p.CQEntries),
		Cap:  int(p.CQEntries),
	}))

	fmt.Printf("insp %+#v %+#v\n", len(cq.Event), cap(cq.Event))

	return
}

// io uring unsetup
func unsetup(r *Ring) (err error) {
	if r.sqesPtr != 0 {
		if err = munmap(r.sqesPtr, uintptr(r.params.SQEntries)); err != nil {
			err = errors.Wrap(err, "munmap sqes")
			return
		}
	}

	featSingleMap := r.params.Features&IORING_FEAT_SINGLE_MMAP > 0
	if err = munmap(r.sqRingPtr, uintptr(r.sringSz)); err != nil {
		err = errors.Wrap(err, "munmap sq")
	}
	if !featSingleMap || r.sqRingPtr != r.cqRingPtr { // not a single map
		if err = munmap(r.cqRingPtr, uintptr(r.cringSz)); err != nil {
			err = errors.Wrap(err, "munmap cq")
			return
		}
	}
	return
}

// io uring register fd
func register(r *Ring, opcode UringRegisterOpcode, arg uintptr, nrArg uint) (ret int, err error) {
	if ret, err = io_uring_register(r.fd, opcode, arg, nrArg); err != nil {
		err = errors.Wrap(err, "io_uring_register")
		return
	}
	return
}

// io uirng enter
func enter(r *Ring, toSubmit, minComplete uint, flags UringEnterFlag, sig *Sigset_t) (ret int, err error) {
	if ret, err = io_uring_enter(r.fd, toSubmit, minComplete, uint(flags), sig); err != nil {
		err = errors.Wrap(err, "io_uring_enter")
		return
	}
	return
}
