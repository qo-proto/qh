package qh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerMinCompressionSize(t *testing.T) {
	tests := []struct {
		name               string
		minCompressionSize int
		responseBody       string
		shouldCompress     bool
	}{
		{
			name:               "small response below default threshold (1KB)",
			minCompressionSize: 1024,
			responseBody:       strings.Repeat("a", 500),
			shouldCompress:     false,
		},
		{
			name:               "large response above default threshold",
			minCompressionSize: 1024,
			responseBody:       strings.Repeat("a", 2000),
			shouldCompress:     true,
		},
		{
			name:               "custom threshold - small",
			minCompressionSize: 100,
			responseBody:       strings.Repeat("a", 200),
			shouldCompress:     true,
		},
		{
			name:               "custom threshold - below",
			minCompressionSize: 500,
			responseBody:       strings.Repeat("a", 400),
			shouldCompress:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(WithMinCompressionSize(tt.minCompressionSize))

			req := &Request{
				Method:  GET,
				Host:    "localhost",
				Path:    "/",
				Version: Version,
				Headers: map[string]string{
					"accept-encoding": "zstd, br, gzip",
				},
			}

			resp := TextResponse(200, tt.responseBody)
			server.applyCompression(req, resp)

			if tt.shouldCompress {
				_, hasEncoding := resp.Headers["content-encoding"]
				assert.True(t, hasEncoding, "response should be compressed")
				assert.Less(t, len(resp.Body), len(tt.responseBody), "compressed body should be smaller")
			} else {
				_, hasEncoding := resp.Headers["content-encoding"]
				assert.False(t, hasEncoding, "response should not be compressed")
				assert.Equal(t, tt.responseBody, string(resp.Body), "body should be unchanged")
			}
		})
	}
}

func TestServerSupportedEncodings(t *testing.T) {
	tests := []struct {
		name               string
		supportedEncodings []Encoding
		clientAccepts      string
		expectedEncoding   Encoding
	}{
		{
			name:               "default encodings - client prefers zstd",
			supportedEncodings: []Encoding{Zstd, Brotli, Gzip},
			clientAccepts:      "zstd, br, gzip",
			expectedEncoding:   Zstd,
		},
		{
			name:               "server only supports gzip",
			supportedEncodings: []Encoding{Gzip},
			clientAccepts:      "zstd, br, gzip",
			expectedEncoding:   Gzip,
		},
		{
			name:               "server only supports zstd",
			supportedEncodings: []Encoding{Zstd},
			clientAccepts:      "br, gzip",
			expectedEncoding:   "",
		},
		{
			name:               "custom priority order",
			supportedEncodings: []Encoding{Brotli, Gzip},
			clientAccepts:      "gzip, br, zstd",
			expectedEncoding:   Gzip, // client prefers gzip first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(
				WithSupportedEncodings(tt.supportedEncodings),
				WithMinCompressionSize(100),
			)

			req := &Request{
				Method:  GET,
				Host:    "localhost",
				Path:    "/",
				Version: Version,
				Headers: map[string]string{
					"accept-encoding": tt.clientAccepts,
				},
			}

			// Large enough to compress
			resp := TextResponse(200, strings.Repeat("test data ", 100))
			server.applyCompression(req, resp)

			if tt.expectedEncoding == "" {
				_, hasEncoding := resp.Headers["content-encoding"]
				assert.False(t, hasEncoding, "no compression should occur when no common encoding")
			} else {
				encoding, hasEncoding := resp.Headers["content-encoding"]
				assert.True(t, hasEncoding, "response should be compressed")
				assert.Equal(t, string(tt.expectedEncoding), encoding)
			}
		})
	}
}

func TestServerNoCompressionForBinary(t *testing.T) {
	server := NewServer(WithMinCompressionSize(100))

	req := &Request{
		Method:  GET,
		Host:    "localhost",
		Path:    "/",
		Version: Version,
		Headers: map[string]string{
			"accept-encoding": "zstd, br, gzip",
		},
	}

	resp := NewResponse(200, []byte(strings.Repeat("binary", 200)), map[string]string{
		"content-type": "4", // OctetStream = 4
	})

	server.applyCompression(req, resp)

	_, hasEncoding := resp.Headers["content-encoding"]
	assert.False(t, hasEncoding, "binary content should not be compressed")
}
