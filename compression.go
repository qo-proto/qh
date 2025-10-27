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

type Encoding string

const (
	Gzip   Encoding = "gzip"
	Brotli Encoding = "br"
	Zstd   Encoding = "zstd"
)

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
		case Gzip, Brotli, Zstd:
			encodings = append(encodings, enc)
		}
		// Unknown encodings (including "identity") are ignored
	}

	return encodings
}

func SelectEncoding(acceptedEncodings []Encoding, serverSupported []Encoding) Encoding {
	for _, clientEnc := range acceptedEncodings {
		if slices.Contains(serverSupported, clientEnc) {
			return clientEnc // First match
		}
	}
	return "" // No common encoding, don't compress
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
