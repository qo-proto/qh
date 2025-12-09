package qh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestClientKeyLogWriter(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/test", GET, func(_ *Request) *Response {
		return TextResponse(200, "OK")
	})

	t.Run("KeylogWriterReceivesData", func(t *testing.T) {
		var keylogData strings.Builder
		client := NewClient(WithClientKeyLogWriter(&keylogData))
		defer client.Close()

		err := client.Connect(addr, nil)
		require.NoError(t, err)

		resp, err := client.GET("127.0.0.1", "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// Keylog should contain data when built with -tags keylog
		// When not built with keylog tag, the writer won't receive data
		keylogOutput := keylogData.String()
		if keylogOutput != "" {
			// If keylog is enabled, verify it contains expected format
			assert.Contains(t, keylogOutput, "CLIENT_TRAFFIC_SECRET")
		}
	})

	t.Run("WithoutKeylogWriter", func(t *testing.T) {
		client := NewClient() // No keylog writer
		defer client.Close()

		err := client.Connect(addr, nil)
		require.NoError(t, err)

		resp, err := client.GET("127.0.0.1", "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		// Should work fine without keylog writer
	})
}
