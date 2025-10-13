// Package compression provides compression/decompression utilities
package compression

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

type Encoding string

const (
	Gzip    Encoding = "gzip"
	Deflate Encoding = "deflate"
	Brotli  Encoding = "br"
	Zstd    Encoding = "zstd"
)

// TODO: add quality values (e.g., `gzip;q=0.8, br;q=1.0`)
// TODO: Wildcard encodings (`*`)
// TODO: add `identity` encoding back?
// TODO: behavior for multiple encodings? (currently just uses first one), encoding selection based on server preferences

// parse Accept-Encoding header, example: "gzip, br, zstd" -> [gzip, br, zstd]
func ParseAcceptEncoding(acceptEncoding string) []Encoding {
	if acceptEncoding == "" {
		return []Encoding{}
	}

	parts := strings.Split(acceptEncoding, ",")
	encodings := make([]Encoding, 0, len(parts))

	for _, part := range parts {
		enc := Encoding(strings.TrimSpace(part))
		switch enc {
		case Gzip, Deflate, Brotli, Zstd:
			encodings = append(encodings, enc)
		}
		// Unknown encodings (including "identity") are ignored
	}

	return encodings
}

func SelectEncoding(acceptedEncodings []Encoding) Encoding {
	if len(acceptedEncodings) > 0 {
		return acceptedEncodings[0]
	}
	return "" // No compression
}

func Compress(data []byte, encoding Encoding) ([]byte, error) {
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

	case Deflate:
		var buf bytes.Buffer
		w, err := flate.NewWriter(&buf, flate.DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("deflate writer error: %w", err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, fmt.Errorf("deflate write error: %w", err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("deflate close error: %w", err)
		}
		return buf.Bytes(), nil

	case Brotli:
		var buf bytes.Buffer
		// use quality 4 for now
		w := brotli.NewWriterLevel(&buf, 4)
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

func Decompress(data []byte, encoding Encoding) ([]byte, error) {
	if len(data) == 0 || encoding == "" {
		return data, nil
	}

	switch encoding {
	case Gzip:
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %w", err)
		}
		defer r.Close()
		decompressed, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("gzip read error: %w", err)
		}
		return decompressed, nil

	case Deflate:
		r := flate.NewReader(bytes.NewReader(data))
		defer r.Close()
		decompressed, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("deflate read error: %w", err)
		}
		return decompressed, nil

	case Brotli:
		r := brotli.NewReader(bytes.NewReader(data))
		decompressed, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("brotli read error: %w", err)
		}
		return decompressed, nil

	case Zstd:
		decoder, err := zstd.NewReader(nil)
		if err != nil {
			return nil, fmt.Errorf("zstd decoder error: %w", err)
		}
		decompressed, err := decoder.DecodeAll(data, nil)
		if err != nil {
			return nil, fmt.Errorf("zstd decode error: %w", err)
		}
		return decompressed, nil

	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}
