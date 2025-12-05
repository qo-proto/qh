package qh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAddr(t *testing.T) {
	t.Run("LiteralIPv4", func(t *testing.T) {
		ip, err := resolveAddr("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", ip.String())
	})

	t.Run("LiteralIPv6", func(t *testing.T) {
		ip, err := resolveAddr("::1")
		require.NoError(t, err)
		assert.Equal(t, "::1", ip.String())
	})

	t.Run("LiteralIPv6Full", func(t *testing.T) {
		ip, err := resolveAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		require.NoError(t, err)
		assert.NotNil(t, ip)
	})

	t.Run("Localhost", func(t *testing.T) {
		ip, err := resolveAddr("localhost")
		require.NoError(t, err)
		assert.NotNil(t, ip)
		assert.True(t, ip.IsLoopback(), "Should resolve to a loopback address")
	})

	t.Run("InvalidHostname", func(t *testing.T) {
		_, err := resolveAddr("not-real-12345.invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve hostname")
	})
}

func TestLookupPubKey(t *testing.T) {
	t.Run("NoTXTRecord", func(t *testing.T) {
		key := lookupPubKey("example.com")
		assert.Empty(t, key, "Should return empty string when no record exists")
	})

	t.Run("InvalidDomain", func(t *testing.T) {
		key := lookupPubKey("fake-54321.invalid")
		assert.Empty(t, key, "Should return empty string on DNS lookup failure")
	})
}

func TestClientDecompression(t *testing.T) {
	client := NewClient()

	t.Run("NoCompression", func(t *testing.T) {
		resp := &Response{
			Headers: map[string]string{},
			Body:    []byte("plain text data"),
		}
		originalBody := string(resp.Body)

		err := client.decompressResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, originalBody, string(resp.Body))
		assert.NotContains(t, resp.Headers, "content-encoding")
	})

	t.Run("ZstdDecompression", func(t *testing.T) {
		original := []byte(strings.Repeat("test data ", 100))
		compressed, err := Compress(original, Zstd)
		require.NoError(t, err)

		resp := &Response{
			Headers: map[string]string{"content-encoding": "zstd"},
			Body:    compressed,
		}

		err = client.decompressResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, original, resp.Body)
		assert.NotContains(t, resp.Headers, "content-encoding", "Should remove content-encoding header")
	})

	t.Run("InvalidCompressedData", func(t *testing.T) {
		resp := &Response{
			Headers: map[string]string{"content-encoding": "zstd"},
			Body:    []byte("this is not actually compressed data"),
		}

		err := client.decompressResponse(resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decompress")
	})

	t.Run("UnknownEncoding", func(t *testing.T) {
		resp := &Response{
			Headers: map[string]string{"content-encoding": "unknown-codec"},
			Body:    []byte("some data"),
		}

		err := client.decompressResponse(resp)
		require.Error(t, err)
	})
}

func TestClientRedirectHandling(t *testing.T) {
	t.Run("MaxRedirectsReached", func(t *testing.T) {
		client := NewClient(WithMaxRedirects(3))

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
				"location": "http://example.com/redirect",
			},
		}

		// Simulate already having done 3 redirects
		_, err := client.handleRedirect(req, resp, 3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many redirects")
	})

	t.Run("MissingLocationHeader", func(t *testing.T) {
		client := NewClient()

		req := &Request{
			Method:  GET,
			Host:    "example.com",
			Path:    "/",
			Version: Version,
			Headers: map[string]string{},
		}

		resp := &Response{
			StatusCode: StatusMovedPermanently,
			Headers:    map[string]string{}, // No location header
		}

		_, err := client.handleRedirect(req, resp, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing location")
	})

	t.Run("InvalidLocationURL", func(t *testing.T) {
		client := NewClient()

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
				"location": "://invalid-url",
			},
		}

		_, err := client.handleRedirect(req, resp, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid location")
	})
}

func TestClientRequestNotConnected(t *testing.T) {
	client := NewClient()

	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/",
		Version: Version,
		Headers: map[string]string{},
	}

	_, err := client.Request(req, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}
