package gouring

import (
	"unsafe"
)

const (
	_uint64       uint64 = 0
	_sz_uint64           = unsafe.Sizeof(_uint64)
	SIGSET_NWORDS        = (1024 / (8 * _sz_uint64))
	SIGTMIN              = 32
	SIGTMAX              = SIGTMIN
	NSIG                 = (SIGTMAX + 1)
)

type Sigset_t struct {
	Val [SIGSET_NWORDS]uint64
}

// https://baike.baidu.com/item/sigset_t/4481187
