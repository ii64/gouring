package gouring

import _ "unsafe"

var io_uring_smp_mb = io_uring_smp_mb_fallback

func io_uring_smp_mb_fallback()
func io_uring_smp_mb_mfence()

func init() {
	// temporary
	io_uring_smp_mb = io_uring_smp_mb_mfence
}
