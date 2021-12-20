package gouring

import (
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCore(t *testing.T) {
	ring, err := New(256, nil)
	assert.NoError(t, err, "create ring")
	defer func() {
		err := ring.Close()
		assert.NoError(t, err, "close ring")
	}()

	mkdata := func(i int) []byte {
		return []byte("print me to stdout please" + strings.Repeat("!", i) + "\n")
	}

	sq := ring.SQ()
	n := 5
	for i := 0; i < n; i++ {
		sqTail := *sq.Tail()
		sqIdx := sqTail & *sq.RingMask()

		sqe := sq.Get(sqIdx)

		m := mkdata(i)

		sqe.Opcode = IORING_OP_WRITE
		sqe.Fd = int32(syscall.Stdout)
		sqe.UserData = uint64(i)
		*sqe.Addr() = (uint64)(uintptr(unsafe.Pointer(&m[0])))

		*sq.Array().Get(sqIdx) = *sq.Head() & *sq.RingMask()
		*sq.Tail()++

		t.Logf("Queued %d: %+#v", i, sqe)
	}

	done, err := ring.Enter(uint(n), uint(n), IORING_ENTER_GETEVENTS, nil)
	assert.NoError(t, err, "ring enter")
	t.Logf("done %d", done)

	// get cq
	cq := ring.CQ()
	for i := 0; i < int(*cq.Tail()); i++ {
		cqHead := *cq.Head()
		cqIdx := cqHead & *cq.RingMask()

		cqe := cq.Get(cqIdx)

		*cq.Head()++
		t.Logf("CQE %+#v", cqe)
	}

}
