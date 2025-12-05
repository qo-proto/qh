package qh

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

const (
	defaultBrotliQuality = 4
)

// Encoding represents a compression encoding type used in Content-Encoding
// and Accept-Encoding headers.
type Encoding string

// Supported compression encoding constants.
const (
	Gzip   Encoding = "gzip"
	Brotli Encoding = "br"
	Zstd   Encoding = "zstd"
)

// parse Accept-Encoding header, example: "gzip, br, zstd" -> [gzip, br, zstd]
func parseAcceptEncoding(acceptEncoding string) []Encoding {
	if acceptEncoding == "" {
		return []Encoding{}
	}

	parts := strings.Split(acceptEncoding, ",")
	encodings := make([]Encoding, 0, len(parts))

	for _, part := range parts {
		enc := Encoding(strings.TrimSpace(part))
		switch enc {
		case Gzip, Brotli, Zstd:
			encodings = append(encodings, enc)
		}
		// Unknown encodings (including "identity") are ignored
	}

	return encodings
}

func selectEncoding(acceptedEncodings []Encoding, serverSupported []Encoding) Encoding {
	for _, clientEnc := range acceptedEncodings {
		if slices.Contains(serverSupported, clientEnc) {
			return clientEnc // First match
		}
	}
	return "" // No common encoding, don't compress
}

func compress(data []byte, encoding Encoding) ([]byte, error) {
	if len(data) == 0 || encoding == "" {
		return data, nil
	}

	switch encoding {
	case Gzip:
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		if _, err := w.Write(data); err != nil {
			return nil, fmt.Errorf("gzip write error: %w", err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("gzip close error: %w", err)
		}
		return buf.Bytes(), nil

	case Brotli:
		var buf bytes.Buffer
		w := brotli.NewWriterLevel(&buf, defaultBrotliQuality)
		if _, err := w.Write(data); err != nil {
			return nil, fmt.Errorf("brotli write error: %w", err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("brotli close error: %w", err)
		}
		return buf.Bytes(), nil

	case Zstd:
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			return nil, fmt.Errorf("zstd encoder error: %w", err)
		}
		defer encoder.Close()
		compressed := encoder.EncodeAll(data, make([]byte, 0, len(data)))
		return compressed, nil

	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

func decompress(data []byte, encoding Encoding, maxSize int) ([]byte, error) {
	if len(data) == 0 || encoding == "" {
		return data, nil
	}

	var r io.Reader
	switch encoding {
	case Gzip:
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %w", err)
		}
		defer gz.Close()
		r = gz

	case Brotli:
		r = brotli.NewReader(bytes.NewReader(data))

	case Zstd:
		zs, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("zstd decoder error: %w", err)
		}
		defer zs.Close()
		r = zs

	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}

	limitR := io.LimitReader(r, int64(maxSize)+1)
	decompressed, err := io.ReadAll(limitR)
	if err != nil {
		return nil, fmt.Errorf("%s read error: %w", encoding, err)
	}

	if len(decompressed) > maxSize {
		return nil, fmt.Errorf("decompressed size exceeds limit of %d bytes", maxSize)
	}

	return decompressed, nil
}
