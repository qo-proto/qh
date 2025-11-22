package qh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientConfiguration(t *testing.T) {
	// Verify default
	cDefault := NewClient()
	assert.Equal(t, 50*1024*1024, cDefault.maxResponseSize, "Default MaxResponseSize should be 50MB")

	// Verify custom option
	customSize := 100 * 1024 * 1024 // 100MB
	cCustom := NewClient(WithMaxResponseSize(customSize))
	assert.Equal(t, customSize, cCustom.maxResponseSize, "Should respect WithMaxResponseSize option")
}
