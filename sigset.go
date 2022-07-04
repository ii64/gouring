package gouring

import (
	"unsafe"
)

const (
	SizeofUint64  = unsafe.Sizeof(uint64(0))
	SIGSET_NWORDS = (1024 / (8 * SizeofUint64))
	SIGTMIN       = 32
	SIGTMAX       = SIGTMIN
	NSIG          = (SIGTMAX + 1)
)

type Sigset_t struct {
	Val [SIGSET_NWORDS]uint64
}

// https://baike.baidu.com/item/sigset_t/4481187
