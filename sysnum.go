package gouring

import (
	"syscall"
	"unsafe"
)

const (
	// uring syscall no.

	SYS_IO_URING_SETUP    = 425
	SYS_IO_URING_ENTER    = 426
	SYS_IO_URING_REGISTER = 427
)

// SQOffsets represents submission queue
type SQOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Flags       uint32
	Dropped     uint32
	Array       uint32
	Resv1       uint32
	Resv2       uint64
}

// CQOffsets represents completion queue
type CQOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Overflow    uint32
	CQEs        uint32
	Flags       uint32
	Resv1       uint32
	Resv2       uint64
}

// IOUringParams io_uring_setup params
type IOUringParams struct {
	SQEntries    uint32                // sq_entries
	CQEntries    uint32                // cq_entries
	Flags        UringSetupFlag        // flags
	SQThreadCPU  uint32                // sq_thread_cpu
	SQThreadIdle uint32                // sq_threead_idle
	Features     UringParamFeatureFlag // features
	WQFd         uint32                // wq_fd

	resv [3]uint32 // resv

	SQOff SQOffsets // sq_off
	CQOff CQOffsets // cq_off
}

//go:inline
func io_uring_setup(entries uint, params *IOUringParams) (ret int, err error) {
	r1, _, e1 := syscall.Syscall(SYS_IO_URING_SETUP, uintptr(entries), uintptr(unsafe.Pointer(params)), 0)
	ret = int(r1)
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}

//go:inline
func io_uring_enter(ringFd int, toSubmit uint, minComplete uint, flags uint, sig *Sigset_t) (ret int, err error) {
	return io_uring_enter2(ringFd, toSubmit, minComplete, flags, sig, NSIG/8)
}

//go:inline
func io_uring_enter2(ringFd int, toSubmit uint, minComplete uint, flags uint, sig *Sigset_t, sz int) (ret int, err error) {
	r1, _, e1 := syscall.Syscall6(SYS_IO_URING_ENTER, uintptr(ringFd), uintptr(toSubmit), uintptr(minComplete), uintptr(flags), uintptr(unsafe.Pointer(sig)), uintptr(sz))
	ret = int(r1)
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}

//go:inline
func io_uring_register(ringFd int, opcode uint /*const*/, arg uintptr, nrArgs uint) (ret int, err error) {
	r1, _, e1 := syscall.Syscall6(SYS_IO_URING_REGISTER, uintptr(ringFd), uintptr(opcode), arg, uintptr(nrArgs), 0, 0)
	ret = int(r1)
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}
