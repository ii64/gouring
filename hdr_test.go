package gouring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSize(t *testing.T) {

	assert.Equal(t, 64, int(SizeofIoUringSqe), "sqe data size mismatch")
	assert.Equal(t, 16, int(SizeofIoUringCqe), "cqe data size mismatch")
}
