package benchmark

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

func EncodeHTTP1(tc TestCase) EncodedResult {
	reqBuf := &bytes.Buffer{}
	req := &http.Request{
		Method: tc.Request.Method,
		URL: &url.URL{
			Scheme: "https",
			Host:   tc.Request.Host,
			Path:   tc.Request.Path,
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader([]byte(tc.Request.Body))),
		Host:       tc.Request.Host,
	}

	for k, v := range tc.Request.Headers {
		req.Header.Set(k, v)
	}

	if len(tc.Request.Body) > 0 {
		req.ContentLength = int64(len(tc.Request.Body))
	}

	if err := req.Write(reqBuf); err != nil {
		panic("failed to write HTTP/1.1 request: " + err.Error())
	}

	respBuf := &bytes.Buffer{}
	resp := &http.Response{
		Status:     http.StatusText(tc.Response.StatusCode),
		StatusCode: tc.Response.StatusCode,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader([]byte(tc.Response.Body))),
	}

	for k, v := range tc.Response.Headers {
		resp.Header.Set(k, v)
	}

	if len(tc.Response.Body) > 0 {
		resp.ContentLength = int64(len(tc.Response.Body))
	}

	if err := resp.Write(respBuf); err != nil {
		panic("failed to write HTTP/1.1 response: " + err.Error())
	}

	reqBytes := reqBuf.Bytes()
	respBytes := respBuf.Bytes()

	return EncodedResult{
		RequestBytes:  reqBytes,
		ResponseBytes: respBytes,
		RequestSize:   len(reqBytes),
		ResponseSize:  len(respBytes),
		TotalSize:     len(reqBytes) + len(respBytes),
	}
}
