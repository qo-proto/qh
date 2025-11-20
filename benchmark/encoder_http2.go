package benchmark

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func EncodeHTTP2(tc TestCase) EncodedResult {
	// Encode request
	reqBuf := &bytes.Buffer{}
	reqFramer := http2.NewFramer(reqBuf, nil)
	reqHeaderBuf := &bytes.Buffer{}
	reqEncoder := hpack.NewEncoder(reqHeaderBuf)

	// HTTP/2 pseudo-headers
	if err := reqEncoder.WriteField(hpack.HeaderField{Name: ":method", Value: tc.Request.Method}); err != nil {
		panic("failed to write :method: " + err.Error())
	}
	if err := reqEncoder.WriteField(hpack.HeaderField{Name: ":path", Value: tc.Request.Path}); err != nil {
		panic("failed to write :path: " + err.Error())
	}
	if err := reqEncoder.WriteField(hpack.HeaderField{Name: ":scheme", Value: "https"}); err != nil {
		panic("failed to write :scheme: " + err.Error())
	}
	if err := reqEncoder.WriteField(hpack.HeaderField{Name: ":authority", Value: tc.Request.Host}); err != nil {
		panic("failed to write :authority: " + err.Error())
	}

	for k, v := range tc.Request.Headers {
		if err := reqEncoder.WriteField(hpack.HeaderField{Name: strings.ToLower(k), Value: v}); err != nil {
			panic("failed to write header " + k + ": " + err.Error())
		}
	}

	if err := reqFramer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: reqHeaderBuf.Bytes(),
		EndHeaders:    true,
		EndStream:     true,
	}); err != nil {
		panic("failed to write HEADERS frame: " + err.Error())
	}

	// Encode response
	respBuf := &bytes.Buffer{}
	respFramer := http2.NewFramer(respBuf, nil)
	respHeaderBuf := &bytes.Buffer{}
	respEncoder := hpack.NewEncoder(respHeaderBuf)

	// :status pseudo-header
	if err := respEncoder.WriteField(hpack.HeaderField{
		Name:  ":status",
		Value: strconv.Itoa(tc.Response.StatusCode),
	}); err != nil {
		panic("failed to write :status: " + err.Error())
	}

	for k, v := range tc.Response.Headers {
		if err := respEncoder.WriteField(hpack.HeaderField{Name: strings.ToLower(k), Value: v}); err != nil {
			panic("failed to write response header " + k + ": " + err.Error())
		}
	}

	if err := respFramer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: respHeaderBuf.Bytes(),
		EndHeaders:    true,
		EndStream:     true,
	}); err != nil {
		panic("failed to write response HEADERS frame: " + err.Error())
	}

	reqBytes := reqBuf.Bytes()
	respBytes := respBuf.Bytes()

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
