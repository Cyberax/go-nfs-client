package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCleanup(t *testing.T) {
	cleaned := false
	cl := NewCleanup(func() {
		cleaned = true
	})
	cl.Cleanup()
	assert.True(t, cleaned)

	cleaned2 := false
	cl2 := NewCleanup(func() {
		cleaned2 = true
	})
	cl2.Disarm()
	cl2.Cleanup()
	assert.False(t, cleaned2)
}

func TestCleanupErr(t *testing.T) {
	cleaned := false
	cl := NewCleanupErr(func() error {
		cleaned = true
		return nil
	})
	cl.Cleanup()
	assert.True(t, cleaned)

	cleaned2 := false
	cl2 := NewCleanupErr(func() error {
		cleaned2 = true
		return nil
	})
	cl2.Disarm()
	cl2.Cleanup()
	assert.False(t, cleaned2)
}
