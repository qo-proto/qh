package qh

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOversizedHost(t *testing.T) {
	tests := []struct {
		name      string
		hostLen   int
		shouldErr bool
	}{
		{"ValidHost253", 253, false},
		{"OversizedHost254", 254, true},
		{"OversizedHost300", 300, true},
		{"OversizedHost1000", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := strings.Repeat("a", tt.hostLen)
			req := &Request{
				Method:  GET,
				Host:    host,
				Path:    "/test",
				Version: Version,
				Headers: map[string]string{},
			}

			data := req.Format()
			parsed, err := ParseRequest(data)

			if tt.shouldErr {
				require.Error(t, err, "Should reject oversized host")
				assert.Contains(t, err.Error(), "host exceeds maximum length")
			} else {
				require.NoError(t, err)
				assert.Equal(t, host, parsed.Host)
			}
		})
	}
}

func TestInvalidMethod(t *testing.T) {
	// Method field is 3 bits, so values 0-7 are possible
	// Valid methods are 0-6, so 7 is invalid
	invalidMethod := byte(7)

	firstByte := (Version << versionBitShift) | (invalidMethod << methodBitShift)
	data := []byte{firstByte}
	data = AppendUvarint(data, 11)
	data = append(data, []byte("example.com")...)
	data = AppendUvarint(data, 1)
	data = append(data, '/')
	data = AppendUvarint(data, 0)
	data = AppendUvarint(data, 0)

	_, err := ParseRequest(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid method")
}

func TestUnknownHeaderID(t *testing.T) {
	// Create response with unknown header ID (0xFF is reserved/unknown)
	data := []byte{0x00}      // Status 200
	data = append(data, 0x02) // Headers length: 2 bytes
	data = append(data, 0xFF) // Unknown header ID
	data = append(data, 0x00) // Some data
	data = append(data, 0x00) // Body length: 0

	_, err := ParseResponse(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown header ID")
}

func TestInfiniteRedirectLoop(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	// Setup redirect loop: /a -> /b -> /a
	srv.HandleFunc("/a", GET, func(_ *Request) *Response {
		headers := map[string]string{
			"location": "qh://127.0.0.1/b",
		}
		return NewResponse(302, nil, headers)
	})

	srv.HandleFunc("/b", GET, func(_ *Request) *Response {
		headers := map[string]string{
			"location": "qh://127.0.0.1/a",
		}
		return NewResponse(302, nil, headers)
	})

	client := NewClient(WithMaxRedirects(5))
	defer client.Close()
	require.NoError(t, client.Connect(addr))

	_, err := client.GET("127.0.0.1", "/a", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many redirects")
}

func TestZeroMaxRedirects(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/redirect", GET, func(_ *Request) *Response {
		return NewResponse(301, nil, map[string]string{
			"location": "qh://127.0.0.1/target",
		})
	})

	client := NewClient(WithMaxRedirects(0))
	defer client.Close()
	require.NoError(t, client.Connect(addr))

	_, err := client.GET("127.0.0.1", "/redirect", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many redirects")
}

func TestIntegerOverflowLengthFields(t *testing.T) {
	t.Run("OverflowInHostLength", func(t *testing.T) {
		data := []byte{0x00}                       // GET
		data = AppendUvarint(data, math.MaxUint64) // Max uint64 host length
		data = append(data, []byte("x")...)

		_, err := ParseRequest(data)
		require.Error(t, err, "Should reject impossibly large host length")
		assert.Contains(t, err.Error(), "exceeds buffer")
	})

	t.Run("OverflowInPathLength", func(t *testing.T) {
		data := []byte{0x00}           // GET
		data = AppendUvarint(data, 11) // host length
		data = append(data, []byte("example.com")...)
		data = AppendUvarint(data, math.MaxUint64) // Max uint64 path length
		data = append(data, []byte("/")...)

		_, err := ParseRequest(data)
		require.Error(t, err, "Should reject impossibly large path length")
		assert.Contains(t, err.Error(), "exceeds buffer")
	})

	t.Run("OverflowInHeadersLength", func(t *testing.T) {
		data := []byte{0x00}           // GET
		data = AppendUvarint(data, 11) // host length
		data = append(data, []byte("example.com")...)
		data = AppendUvarint(data, 1) // path length
		data = append(data, '/')
		data = AppendUvarint(data, math.MaxUint64) // Max uint64 headers length

		_, err := ParseRequest(data)
		require.Error(t, err, "Should reject impossibly large headers length")
		assert.Contains(t, err.Error(), "exceeds buffer")
	})

	t.Run("OverflowInBodyLength", func(t *testing.T) {
		data := []byte{0x00}           // GET
		data = AppendUvarint(data, 11) // host length
		data = append(data, []byte("example.com")...)
		data = AppendUvarint(data, 1) // path length
		data = append(data, '/')
		data = AppendUvarint(data, 0)              // headers length
		data = AppendUvarint(data, math.MaxUint64) // Max uint64 body length

		_, err := ParseRequest(data)
		require.Error(t, err, "Should reject impossibly large body length")
		assert.Contains(t, err.Error(), "exceeds buffer")
	})

	t.Run("OverflowInResponseBodyLength", func(t *testing.T) {
		data := []byte{0x00}                       // Status 200
		data = AppendUvarint(data, 0)              // headers length
		data = AppendUvarint(data, math.MaxUint64) // Max uint64 body length

		_, err := ParseResponse(data)
		require.Error(t, err, "Should reject impossibly large response body length")
		assert.Contains(t, err.Error(), "exceeds buffer")
	})
}
