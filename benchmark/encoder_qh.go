package benchmark

import (
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
		Body:    tc.Request.Body,
	}
	reqBytes := req.Format()

	resp := &qh.Response{
		Version:    qh.Version,
		StatusCode: tc.Response.StatusCode,
		Headers:    tc.Response.Headers,
		Body:       tc.Response.Body,
	}
	respBytes := resp.Format()

	return EncodedResult{
		RequestBytes:  reqBytes,
		ResponseBytes: respBytes,
		RequestSize:   len(reqBytes),
		ResponseSize:  len(respBytes),
		TotalSize:     len(reqBytes) + len(respBytes),
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
	default:
		return qh.GET
	}
}
