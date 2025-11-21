package benchmark

import (
	"log/slog"

	"github.com/qo-proto/qh"
)

func EncodeQH(tc TestCase) EncodedResult {
	method := parseMethod(tc.Request.Method)
	req := &qh.Request{
		Method:  method,
		Host:    tc.Request.Host,
		Path:    tc.Request.Path,
		Version: qh.Version,
		Headers: tc.Request.Headers,
		Body:    nil,
	}
	reqBytes := req.Format()

	resp := &qh.Response{
		Version:    qh.Version,
		StatusCode: tc.Response.StatusCode,
		Headers:    tc.Response.Headers,
		Body:       nil,
	}
	respBytes := resp.Format()

	return EncodedResult{
		RequestBytes:       reqBytes,
		ResponseBytes:      respBytes,
		RequestSize:        len(reqBytes),
		ResponseSize:       len(respBytes),
		TotalSize:          len(reqBytes) + len(respBytes),
		RequestHeaderSize:  len(reqBytes),
		ResponseHeaderSize: len(respBytes),
	}
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
