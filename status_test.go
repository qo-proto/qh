package qh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeStatusCode(t *testing.T) {
	require.Equal(t, uint8(20), encodeStatusCode(200)) // OK
	require.Equal(t, uint8(44), encodeStatusCode(404)) // Not Found
	require.Equal(t, uint8(50), encodeStatusCode(500)) // Internal Server Error
	require.Equal(t, uint8(32), encodeStatusCode(302)) // Found
	require.Equal(t, uint8(24), encodeStatusCode(204)) // No Content
}

func TestDecodeStatusCode(t *testing.T) {
	require.Equal(t, 200, DecodeStatusCode(20)) // OK
	require.Equal(t, 404, DecodeStatusCode(44)) // Not Found
	require.Equal(t, 500, DecodeStatusCode(50)) // Internal Server Error
	require.Equal(t, 302, DecodeStatusCode(32)) // Found
	require.Equal(t, 204, DecodeStatusCode(24)) // No Content
}

func TestEncodeStatusCodeFallback(t *testing.T) {
	// all unmapped codes should return compact code for 500 -> 50
	require.Equal(t, uint8(50), encodeStatusCode(419))
	require.Equal(t, uint8(50), encodeStatusCode(451))
}

func TestDecodeStatusCodeFallback(t *testing.T) {
	require.Equal(t, 500, DecodeStatusCode(219))
	require.Equal(t, 500, DecodeStatusCode(195))
	require.Equal(t, 500, DecodeStatusCode(100))
}

func TestRoundTripStatusCodes(t *testing.T) { // catches mapping mistakes
	for httpCode := range statusToCompact {
		t.Run(string(rune(httpCode)), func(t *testing.T) {
			compact := encodeStatusCode(httpCode)
			decoded := DecodeStatusCode(compact)
			require.Equal(t, httpCode, decoded, "Round-trip failed for HTTP code %d", httpCode)
		})
	}
}
