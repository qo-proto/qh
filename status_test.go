package qh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeStatusCode(t *testing.T) {
	require.Equal(t, uint8(0), EncodeStatusCode(200))  // OK
	require.Equal(t, uint8(1), EncodeStatusCode(404))  // Not Found
	require.Equal(t, uint8(2), EncodeStatusCode(500))  // Internal Server Error
	require.Equal(t, uint8(3), EncodeStatusCode(302))  // Found
	require.Equal(t, uint8(12), EncodeStatusCode(204)) // No Content
}

func TestDecodeStatusCode(t *testing.T) {
	require.Equal(t, 200, DecodeStatusCode(0))  // OK
	require.Equal(t, 404, DecodeStatusCode(1))  // Not Found
	require.Equal(t, 500, DecodeStatusCode(2))  // Internal Server Error
	require.Equal(t, 302, DecodeStatusCode(3))  // Found
	require.Equal(t, 204, DecodeStatusCode(12)) // No Content
}

func TestEncodeStatusCodeFallback(t *testing.T) {
	// all unmapped codes should return compact code for 500 -> 3
	require.Equal(t, uint8(2), EncodeStatusCode(419))
	require.Equal(t, uint8(2), EncodeStatusCode(451))
}

func TestDecodeStatusCodeFallback(t *testing.T) {
	require.Equal(t, 500, DecodeStatusCode(219))
	require.Equal(t, 500, DecodeStatusCode(195))
	require.Equal(t, 500, DecodeStatusCode(100))
}

func TestRoundTripStatusCodes(t *testing.T) { // catches mapping mistakes
	for httpCode := range StatusToCompact {
		t.Run(string(rune(httpCode)), func(t *testing.T) {
			compact := EncodeStatusCode(httpCode)
			decoded := DecodeStatusCode(compact)
			require.Equal(t, httpCode, decoded, "Round-trip failed for HTTP code %d", httpCode)
		})
	}
}
