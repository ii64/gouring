package gouring

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	SizeofUint64  = unsafe.Sizeof(uint64(0))
	SIGSET_NWORDS = (1024 / (8 * SizeofUint64))
	SIGTMIN       = 32
	SIGTMAX       = SIGTMIN
	NSIG          = (SIGTMAX + 1)
)

type Sigset_t = unix.Sigset_t
