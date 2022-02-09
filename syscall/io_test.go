package syscall

import (
	"syscall"
	"testing"
	"unsafe"
)

func TestAccept(t *testing.T) {

	var raw syscall.RawSockaddrAny

	t.Logf("%d\n", unsafe.Sizeof(raw))

	t.Fail()
}
