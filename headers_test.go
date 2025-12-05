package qh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderEncodingFormat1(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"SecChUAMobile", "sec-ch-ua-mobile", "?0"},
		{"ContentTypeJSON", "content-type", "application/json; charset=UTF-8"},
		{"AcceptAll", "accept", "*/*"},
		{"Connection", "connection", "keep-alive"},
		{"CacheControl", "cache-control", "no-cache"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				tt.key: tt.value,
			}

			encoded := encodeHeaders(headers, requestHeaderCompletePairs, requestHeaderNameOnly)

			// Verify it's in the complete pairs table
			lookupKey := tt.key + ":" + tt.value
			expectedID, exists := requestHeaderCompletePairs[lookupKey]
			require.True(t, exists, "%s should be in request complete pairs table", lookupKey)

			require.Len(t, encoded, 1, "Complete pair should encode to single byte")
			assert.Equal(t, expectedID, encoded[0])
		})
	}
}

func TestHeaderEncodingFormat2(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"UserAgent", "user-agent", "Mozilla/5.0"},
		{"Accept", "accept", "3,2,1"},
		{"ContentType", "content-type", "2"},
		{"Host", "host", "example.com"},
		{"Authorization", "authorization", "Bearer token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				tt.key: tt.value,
			}

			encoded := encodeHeaders(headers, requestHeaderCompletePairs, requestHeaderNameOnly)

			// Verify it's in the name-only table
			expectedID, exists := requestHeaderNameOnly[tt.key]
			require.True(t, exists, "%s should be in request name-only table", tt.key)

			require.Greater(t, len(encoded), 1, "Name-only header should have ID + length + value")
			assert.Equal(t, expectedID, encoded[0], "First byte should be header ID")

			valueLen, n, err := ReadUvarint(encoded, 1)
			require.NoError(t, err)
			assert.Equal(t, uint64(len(tt.value)), valueLen)

			actualValue := string(encoded[1+n : 1+n+int(valueLen)])
			assert.Equal(t, tt.value, actualValue)
		})
	}
}

func TestHeaderEncodingFormat3(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"CustomHeader1", "x-custom-header", "custom-value"},
		{"CustomHeader2", "x-request-id", "abc123"},
		{"CustomHeader3", "x-trace-id", "trace-xyz-789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				tt.key: tt.value,
			}

			encoded := encodeHeaders(headers, requestHeaderCompletePairs, requestHeaderNameOnly)

			require.Greater(t, len(encoded), 2, "Custom header should have 0x00 + key length + key + value length + value")
			assert.Equal(t, CustomHeader, encoded[0], "First byte should be 0x00 for custom header")

			offset := 1

			keyLen, n, err := ReadUvarint(encoded, offset)
			require.NoError(t, err)
			offset += n

			actualKey := string(encoded[offset : offset+int(keyLen)])
			assert.Equal(t, tt.key, actualKey)
			offset += int(keyLen)

			valueLen, n, err := ReadUvarint(encoded, offset)
			require.NoError(t, err)
			offset += n

			actualValue := string(encoded[offset : offset+int(valueLen)])
			assert.Equal(t, tt.value, actualValue)
		})
	}
}

func TestHeaderCaseInsensitivity(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
	}{
		{
			"MixedCase",
			map[string]string{
				"Content-Type": "application/json",
				"Accept":       "text/html",
			},
		},
		{
			"UpperCase",
			map[string]string{
				"CONTENT-TYPE": "text/plain",
				"USER-AGENT":   "Test/1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/",
				Version: Version,
				Headers: tt.input,
			}

			data := req.Format()
			parsed, err := ParseRequest(data)
			require.NoError(t, err)
			for key := range parsed.Headers {
				assert.Equal(t, strings.ToLower(key), key, "Header names should be lowercase")
			}
		})
	}
}

func TestHeaderEncodingIntegration(t *testing.T) {
	headers := map[string]string{
		"content-type":  "application/json; charset=UTF-8", // Format 1
		"accept":        "text/html",                       // Format 2
		"x-custom":      "value",                           // Format 3
		"user-agent":    "QH/1.0",                          // Format 2
		"cache-control": "no-cache",                        // Format 1
	}

	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/test",
		Version: Version,
		Headers: headers,
	}

	data := req.Format()
	parsed, err := ParseRequest(data)
	require.NoError(t, err)

	assert.Equal(t, "application/json; charset=UTF-8", parsed.Headers["content-type"])
	assert.Equal(t, "text/html", parsed.Headers["accept"])
	assert.Equal(t, "value", parsed.Headers["x-custom"])
	assert.Equal(t, "QH/1.0", parsed.Headers["user-agent"])
	assert.Equal(t, "no-cache", parsed.Headers["cache-control"])
}

func TestEmptyHeaderValue(t *testing.T) {
	headers := map[string]string{
		"x-empty": "",
	}

	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/",
		Version: Version,
		Headers: headers,
	}

	data := req.Format()
	parsed, err := ParseRequest(data)
	require.NoError(t, err)

	assert.Empty(t, parsed.Headers["x-empty"])
}

func TestHeaderValuesAreBinarySafe(t *testing.T) {
	headers := map[string]string{
		"x-test": "value with spaces",
		"x-json": `{"key":"value"}`,
		"x-url":  "https://example.com/path?query=1",
	}

	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/",
		Version: Version,
		Headers: headers,
	}

	data := req.Format()
	parsed, err := ParseRequest(data)
	require.NoError(t, err)

	assert.Equal(t, headers["x-test"], parsed.Headers["x-test"])
	assert.Equal(t, headers["x-json"], parsed.Headers["x-json"])
	assert.Equal(t, headers["x-url"], parsed.Headers["x-url"])
}

func TestResponseHeaderIntegration(t *testing.T) {
	headers := map[string]string{
		"content-type":  "application/json",
		"cache-control": "max-age=3600",
		"etag":          `"abc123"`,
	}

	resp := &Response{
		Version:    Version,
		StatusCode: 200,
		Headers:    headers,
		Body:       []byte(`{"message":"ok"}`),
	}

	data := resp.Format()
	parsed, err := ParseResponse(data)
	require.NoError(t, err)

	assert.Equal(t, "application/json", parsed.Headers["content-type"])
	assert.Equal(t, "max-age=3600", parsed.Headers["cache-control"])
	assert.Equal(t, `"abc123"`, parsed.Headers["etag"])
}

func TestResponseHeaderEncoding(t *testing.T) {
	t.Run("Format1_CompletePair", func(t *testing.T) {
		headers := map[string]string{
			"content-encoding": "gzip",
		}

		encoded := encodeHeaders(headers, responseHeaderCompletePairs, responseHeaderNameOnly)

		expectedID, exists := responseHeaderCompletePairs["content-encoding:gzip"]
		require.True(t, exists, "content-encoding:gzip should be in response complete pairs table")

		require.Len(t, encoded, 1, "Complete pair should encode to single byte")
		assert.Equal(t, expectedID, encoded[0])
	})

	t.Run("Format2_NameOnly", func(t *testing.T) {
		headers := map[string]string{
			"content-type": "2",
		}

		encoded := encodeHeaders(headers, responseHeaderCompletePairs, responseHeaderNameOnly)

		expectedID, exists := responseHeaderNameOnly["content-type"]
		require.True(t, exists, "content-type should be in response name-only table")

		require.Greater(t, len(encoded), 1, "Name-only header should have ID + length + value")
		assert.Equal(t, expectedID, encoded[0])

		// decode
		valueLen, n, err := ReadUvarint(encoded, 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), valueLen)

		actualValue := string(encoded[1+n : 1+n+int(valueLen)])
		assert.Equal(t, "2", actualValue)
	})

	t.Run("Format3_Custom", func(t *testing.T) {
		headers := map[string]string{
			"x-custom-response": "value",
		}

		encoded := encodeHeaders(headers, responseHeaderCompletePairs, responseHeaderNameOnly)

		require.Greater(t, len(encoded), 2, "Custom header should have marker + key + value")
		assert.Equal(t, CustomHeader, encoded[0], "First byte should be 0x00 for custom header")

		// Verify key and value are encoded
		offset := 1
		keyLen, n, err := ReadUvarint(encoded, offset)
		require.NoError(t, err)
		offset += n

		actualKey := string(encoded[offset : offset+int(keyLen)])
		assert.Equal(t, "x-custom-response", actualKey)
	})
}

func TestLargeHeaderValue(t *testing.T) {
	largeValue := strings.Repeat("a", 10000)
	headers := map[string]string{
		"x-large": largeValue,
	}

	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/",
		Version: Version,
		Headers: headers,
	}

	data := req.Format()
	parsed, err := ParseRequest(data)
	require.NoError(t, err)

	assert.Equal(t, largeValue, parsed.Headers["x-large"])
}

func TestHeaderEncodingPriority(t *testing.T) {
	// When a header exists in both complete pairs and name-only,
	// Format 1 (complete pair) should take priority
	headers := map[string]string{
		"content-type": "application/json", // Exists in both tables
	}

	encoded := encodeHeaders(headers, requestHeaderCompletePairs, requestHeaderNameOnly)

	// Should use Format 1 (complete pair) - single byte
	require.Len(t, encoded, 1)

	expectedID := requestHeaderCompletePairs["content-type:application/json"]
	assert.Equal(t, expectedID, encoded[0])
}
