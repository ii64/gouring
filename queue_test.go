package gouring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingQueue(t *testing.T) {
	h, err := New(256, 0)
	assert.NoError(t, err)
	defer h.Close()

	sqe := h.io_uring_get_sqe()
	fmt.Printf("%+#v\n", sqe)

}
