package gouring

import (
	"strings"
	"syscall"
	"testing"

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
		sqTail := *sq.Tail
		sqIdx := sqTail & *sq.RingMask

		sqe := &sq.Event[sqIdx]
		sqe.Reset()

		m := mkdata(i)

		sqe.Opcode = IORING_OP_WRITE
		sqe.Fd = int32(syscall.Stdout)
		sqe.UserData = uint64(i)
		sqe.Len = uint32(len(m))
		sqe.SetOffset(0)
		sqe.SetAddr(&m[0])

		*sq.Array.Get(sqIdx) = *sq.Head & *sq.RingMask
		*sq.Tail++

		done, err := ring.Enter(1, 1, IORING_ENTER_GETEVENTS, nil)
		assert.NoError(t, err, "ring enter")
		t.Logf("done %d", done)
	}

	// get cq
	cq := ring.CQ()
	for i := 0; i < int(*cq.Tail); i++ {
		cqHead := *cq.Head
		cqIdx := cqHead & *cq.RingMask

		cqe := cq.Event[cqIdx]

		*cq.Head++
		t.Logf("CQE %+#v", cqe)
	}

}
