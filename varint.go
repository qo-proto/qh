package qh

import (
	"encoding/binary"
	"errors"
)

var (
	ErrVarintOverflow   = errors.New("varint overflows uint64")
	ErrVarintIncomplete = errors.New("varint is incomplete (buffer too short)")
)

func ReadUvarint(buf []byte, offset int) (uint64, int, error) {
	if offset >= len(buf) {
		return 0, 0, ErrVarintIncomplete
	}

	value, n := binary.Uvarint(buf[offset:])
	if n == 0 {
		return 0, 0, ErrVarintIncomplete
	}
	if n < 0 {
		return 0, 0, ErrVarintOverflow
	}

	return value, n, nil
}

func AppendUvarint(buf []byte, v uint64) []byte {
	start := len(buf)
	buf = append(buf, make([]byte, binary.MaxVarintLen64)...)
	n := binary.PutUvarint(buf[start:], v)
	return buf[:start+n]
}
