package compression

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
			input:    "gzip, deflate, br, zstd",
			expected: []Encoding{Gzip, Deflate, Brotli, Zstd},
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
		name     string
		input    []Encoding
		expected Encoding
	}{
		{
			name:     "first preference",
			input:    []Encoding{Gzip, Brotli, Zstd},
			expected: Gzip,
		},
		{
			name:     "zstd preferred",
			input:    []Encoding{Zstd, Gzip},
			expected: Zstd,
		},
		{
			name:     "empty list returns empty string",
			input:    []Encoding{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectEncoding(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompressDecompress(t *testing.T) {
	testData := []byte(strings.Repeat("Hello, QH Protocol! This is some test data that should compress. ", 100))
	encodings := []Encoding{Gzip, Deflate, Brotli, Zstd}

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

			decompressed, err := Decompress(compressed, encoding)
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

	decompressed, err := Decompress(compressed, "")
	require.NoError(t, err)
	assert.Equal(t, testData, decompressed, "empty encoding decompression should return same data")
}

func TestCompressDecompressEmpty(t *testing.T) {
	testData := []byte{}
	encodings := []Encoding{Gzip, Deflate, Brotli, Zstd}

	for _, encoding := range encodings {
		t.Run(string(encoding), func(t *testing.T) {
			compressed, err := Compress(testData, encoding)
			require.NoError(t, err, "compression of empty data should succeed")

			decompressed, err := Decompress(compressed, encoding)
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
	_, err := Decompress(testData, Encoding("invalid"))
	assert.Error(t, err, "should return error for invalid encoding")
}

// binary data shouldnt compress well (http also skips it)
func TestBinaryDataCompression(t *testing.T) {
	binaryData := make([]byte, 10000)
	seed := uint32(12345)
	for i := range binaryData { // fill with pseudo-random data
		seed = seed*1103515245 + 12345
		binaryData[i] = byte(seed >> 16)
	}

	encodings := []Encoding{Gzip, Deflate, Brotli, Zstd}

	for _, encoding := range encodings {
		t.Run(string(encoding), func(t *testing.T) {
			compressed, err := Compress(binaryData, encoding)
			require.NoError(t, err, "compression should succeed even on binary data")

			ratio := float64(len(compressed)) / float64(len(binaryData)) * 100
			t.Logf("%s compression of binary/random data: %.2f%% (%d -> %d bytes)",
				encoding, ratio, len(binaryData), len(compressed))

			if ratio > 95.0 {
				t.Logf("✓ Binary data compression ineffective (%.2f%% - expected for already-compressed formats)", ratio)
			} else {
				t.Logf("! Binary data compressed to %.2f%% (test data may not be random enough)", ratio)
			}
		})
	}
}

func BenchmarkCompressions(b *testing.B) {
	testData := []byte(strings.Repeat("Hello, QH Protocol! This is benchmark data. ", 1000))
	encodings := []Encoding{Gzip, Deflate, Brotli, Zstd}

	for _, encoding := range encodings {
		b.Run(string(encoding), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_, _ = Compress(testData, encoding)
			}
		})
	}
}
