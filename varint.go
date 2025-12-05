package qh

import (
	"encoding/binary"
	"errors"
)

var (
	errVarintOverflow   = errors.New("varint overflows uint64")
	errVarintIncomplete = errors.New("varint is incomplete (buffer too short)")
)

func ReadUvarint(buf []byte, offset int) (uint64, int, error) {
	if offset >= len(buf) {
		return 0, 0, errVarintIncomplete
	}

	val, n := binary.Uvarint(buf[offset:])
	switch {
	case n == 0:
		return 0, 0, errVarintIncomplete
	case n < 0:
		return 0, 0, errVarintOverflow
	default:
		return val, n, nil
	}
}

func appendUvarint(buf []byte, v uint64) []byte {
	start := len(buf)
	buf = append(buf, make([]byte, binary.MaxVarintLen64)...)
	n := binary.PutUvarint(buf[start:], v)
	return buf[:start+n]
}
