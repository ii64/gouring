package syscall

import (
	"syscall"
	_ "unsafe"
)

var _zero uintptr

//go:linkname anyToSockaddr syscall.anyToSockaddr
func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error)
