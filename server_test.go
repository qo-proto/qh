package qh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxRequestSize(t *testing.T) {
	// Verify default
	sDefault := NewServer()
	assert.Equal(t, MaxRequestSize, sDefault.maxRequestSize, "Default MaxRequestSize should be 10MB")

	// Verify custom option
	customSize := 5 * 1024 * 1024 // 5MB
	sCustom := NewServer(WithMaxRequestSize(customSize))
	assert.Equal(t, customSize, sCustom.maxRequestSize, "Should respect WithMaxRequestSize option")
}
