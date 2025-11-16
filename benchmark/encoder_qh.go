package benchmark

import (
	"encoding/binary"
	"log/slog"

	"github.com/qh-project/qh"
)

func EncodeQH(tc TestCase) EncodedResult {
	method := parseMethod(tc.Request.Method)
	req := &qh.Request{
		Method:  method,
		Host:    tc.Request.Host,
		Path:    tc.Request.Path,
		Version: qh.Version,
		Headers: tc.Request.Headers,
		Body:    []byte(tc.Request.Body),
	}
	reqBytes := req.Format()

	resp := &qh.Response{
		Version:    qh.Version,
		StatusCode: tc.Response.StatusCode,
		Headers:    tc.Response.Headers,
		Body:       []byte(tc.Response.Body),
	}
	respBytes := resp.Format()

	reqBodySize := len(tc.Request.Body)
	reqHeaderSize := calculateQHHeaderSize(reqBytes, reqBodySize)

	respBodySize := len(tc.Response.Body)
	respHeaderSize := calculateQHHeaderSize(respBytes, respBodySize)

	return EncodedResult{
		RequestBytes:       reqBytes,
		ResponseBytes:      respBytes,
		RequestSize:        len(reqBytes),
		ResponseSize:       len(respBytes),
		TotalSize:          len(reqBytes) + len(respBytes),
		RequestHeaderSize:  reqHeaderSize,
		RequestBodySize:    reqBodySize,
		ResponseHeaderSize: respHeaderSize,
		ResponseBodySize:   respBodySize,
	}
}

func calculateQHHeaderSize(messageBytes []byte, bodySize int) int {
	totalSize := len(messageBytes)
	var buf [binary.MaxVarintLen64]byte
	bodyLenVarintSize := binary.PutUvarint(buf[:], uint64(bodySize))

	return totalSize - bodyLenVarintSize - bodySize
}

func parseMethod(method string) qh.Method {
	switch method {
	case "GET":
		return qh.GET
	case "POST":
		return qh.POST
	case "PUT":
		return qh.PUT
	case "PATCH":
		return qh.PATCH
	case "DELETE":
		return qh.DELETE
	case "HEAD":
		return qh.HEAD
	case "OPTIONS":
		return qh.OPTIONS
	default:
		slog.Warn("unsupported HTTP method, defaulting to GET", "method", method)
		return qh.GET
	}
}
