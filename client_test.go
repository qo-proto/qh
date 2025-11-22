package qh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientZeroRedirects(t *testing.T) {
	client := NewClient(WithMaxRedirects(0))

	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/",
		Version: Version,
		Headers: map[string]string{},
	}

	resp := &Response{
		StatusCode: StatusMovedPermanently,
		Headers: map[string]string{
			"location": "http://example.com/new",
		},
	}

	_, err := client.handleRedirect(req, resp, 0)
	require.Error(t, err, "should immediately error when maxRedirects is 0")
	assert.Contains(t, err.Error(), "too many redirects")
}
