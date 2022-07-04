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

func io_uring_setup(entries uint32, params *IoUringParams) (ret int, err error) {
	r1, _, e1 := syscall.RawSyscall(SYS_IO_URING_SETUP, uintptr(entries), uintptr(unsafe.Pointer(params)), 0)
	ret = int(r1)
	if e1 != 0 {
		err = e1
	}
	return
}

func io_uring_enter(fd int32, toSubmit uint32, minComplete uint32, flags uint32, sig *Sigset_t) (ret int, err error) {
	return io_uring_enter2(fd, toSubmit, minComplete, flags, sig, NSIG/8)
}

func io_uring_enter2(fd int32, toSubmit uint32, minComplete uint32, flags uint32, sig *Sigset_t, sz int32) (ret int, err error) {
	r1, _, e1 := syscall.RawSyscall6(SYS_IO_URING_ENTER,
		uintptr(fd),
		uintptr(toSubmit), uintptr(minComplete),
		uintptr(flags), uintptr(unsafe.Pointer(sig)), uintptr(sz))
	ret = int(r1)
	if e1 != 0 {
		err = e1
	}
	return
}

func io_uring_register(fd int32, opcode uint32, arg unsafe.Pointer, nrArgs uintptr) (ret int, err error) {
	r1, _, e1 := syscall.RawSyscall6(SYS_IO_URING_REGISTER, uintptr(fd), uintptr(opcode), uintptr(arg), uintptr(nrArgs), 0, 0)
	ret = int(r1)
	if e1 != 0 {
		err = e1
	}
	return
}

//go:linkname mmap syscall.mmap
func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (xaddr uintptr, err error)

//go:linkname munmap syscall.munmap
func munmap(addr uintptr, length uintptr) (err error)

//

func increase_rlimit_nofile(nr uint64) error {
	var rlim syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		return err
	}
	if rlim.Cur < nr {
		rlim.Cur += nr
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	}
	return err
}
