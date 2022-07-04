package gouring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingSetup(t *testing.T) {
	h, err := New(256, 0)
	assert.NoError(t, err)
	assert.NotNil(t, h)
	assert.NotEqual(t, 0, h.RingFd)
	h.Close()
}
