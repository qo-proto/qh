package qh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAcceptEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Encoding
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []Encoding{},
		},
		{
			name:     "single encoding",
			input:    "gzip",
			expected: []Encoding{Gzip},
		},
		{
			name:     "multiple encodings",
			input:    "gzip, br, zstd",
			expected: []Encoding{Gzip, Brotli, Zstd},
		},
		{
			name:     "with spaces",
			input:    "gzip,  br,   zstd",
			expected: []Encoding{Gzip, Brotli, Zstd},
		},
		{
			name:     "unsupported encoding ignored",
			input:    "gzip, unknown, br",
			expected: []Encoding{Gzip, Brotli},
		},
		{
			name:     "all supported encodings",
			input:    "gzip, br, zstd",
			expected: []Encoding{Gzip, Brotli, Zstd},
		},
		{
			name:     "identity encoding ignored",
			input:    "gzip, identity, br",
			expected: []Encoding{Gzip, Brotli},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAcceptEncoding(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSelectEncoding(t *testing.T) {
	tests := []struct {
		name           string
		clientAccepts  []Encoding
		serverSupports []Encoding
		expected       Encoding
		description    string
	}{
		{
			name:           "client and server both prefer zstd",
			clientAccepts:  []Encoding{Zstd, Gzip, Brotli},
			serverSupports: []Encoding{Zstd, Brotli, Gzip},
			expected:       Zstd,
			description:    "should select zstd when both agree",
		},
		{
			name:           "client preference wins over server preference",
			clientAccepts:  []Encoding{Gzip, Brotli, Zstd},
			serverSupports: []Encoding{Zstd, Brotli, Gzip},
			expected:       Gzip,
			description:    "should use client preference",
		},
		{
			name:           "no common encoding",
			clientAccepts:  []Encoding{Gzip},
			serverSupports: []Encoding{Zstd, Brotli},
			expected:       "",
			description:    "should return empty string when no match",
		},
		{
			name:           "server supports subset of client",
			clientAccepts:  []Encoding{Zstd, Brotli, Gzip},
			serverSupports: []Encoding{Brotli},
			expected:       Brotli,
			description:    "should find the only common encoding",
		},
		{
			name:           "client accepts only one encoding",
			clientAccepts:  []Encoding{Gzip},
			serverSupports: []Encoding{Zstd, Brotli, Gzip},
			expected:       Gzip,
			description:    "should select the only client-accepted encoding",
		},
		{
			name:           "empty lists return no encoding",
			clientAccepts:  []Encoding{},
			serverSupports: []Encoding{Zstd, Gzip},
			expected:       "",
			description:    "should return empty when lists are empty",
		},
		{
			name:           "first client preference wins when multiple matches",
			clientAccepts:  []Encoding{Brotli, Gzip, Zstd},
			serverSupports: []Encoding{Zstd, Gzip, Brotli},
			expected:       Brotli,
			description:    "should select first client preference that server supports",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectEncoding(tt.clientAccepts, tt.serverSupports)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestCompressDecompress(t *testing.T) {
	testData := []byte(strings.Repeat("Hello, QH Protocol! This is some test data that should compress. ", 100))
	encodings := []Encoding{Gzip, Brotli, Zstd}

	for _, encoding := range encodings {
		t.Run(string(encoding), func(t *testing.T) {
			compressed, err := Compress(testData, encoding)
			require.NoError(t, err, "compression should succeed")

			// Verify compression actually happened
			if len(compressed) >= len(testData) {
				t.Logf("warning: compressed size (%d) >= original size (%d) for %s",
					len(compressed), len(testData), encoding)
			} else {
				savings := float64(len(testData)-len(compressed)) / float64(len(testData)) * 100
				t.Logf("%s saved %.1f%% (%d -> %d bytes)",
					encoding, savings, len(testData), len(compressed))
			}

			decompressed, err := Decompress(compressed, encoding, 10*1024*1024)
			require.NoError(t, err, "decompression should succeed")
			assert.Equal(t, testData, decompressed, "decompressed data should match original")
		})
	}
}

func TestCompressDecompressEmptyEncoding(t *testing.T) {
	testData := []byte("Hello, World!")

	compressed, err := Compress(testData, "")
	require.NoError(t, err)
	assert.Equal(t, testData, compressed, "empty encoding compression should return same data")

	decompressed, err := Decompress(compressed, "", 10*1024*1024)
	require.NoError(t, err)
	assert.Equal(t, testData, decompressed, "empty encoding decompression should return same data")
}

func TestCompressDecompressEmpty(t *testing.T) {
	testData := []byte{}
	encodings := []Encoding{Gzip, Brotli, Zstd}

	for _, encoding := range encodings {
		t.Run(string(encoding), func(t *testing.T) {
			compressed, err := Compress(testData, encoding)
			require.NoError(t, err, "compression of empty data should succeed")

			decompressed, err := Decompress(compressed, encoding, 10*1024*1024)
			require.NoError(t, err, "decompression of empty data should succeed")

			assert.Empty(t, decompressed, "decompressed empty data should remain empty")
		})
	}
}

func TestCompressInvalidEncoding(t *testing.T) {
	testData := []byte("test")
	_, err := Compress(testData, Encoding("invalid"))
	assert.Error(t, err, "should return error for invalid encoding")
}

func TestDecompressInvalidEncoding(t *testing.T) {
	testData := []byte("test")
	_, err := Decompress(testData, Encoding("invalid"), 10*1024*1024)
	assert.Error(t, err, "should return error for invalid encoding")
}

func BenchmarkCompressions(b *testing.B) {
	testData := []byte(strings.Repeat("Hello, QH Protocol! This is benchmark data. ", 1000))
	encodings := []Encoding{Gzip, Brotli, Zstd}

	for _, encoding := range encodings {
		b.Run(string(encoding), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_, _ = Compress(testData, encoding)
			}
		})
	}
}

func TestDecompressSizeLimit(t *testing.T) {
	original := []byte(strings.Repeat("test ", 1000))
	compressed, _ := Compress(original, Zstd)
	sizeLimit := 1000

	_, err := Decompress(compressed, Zstd, sizeLimit)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds limit")
}
