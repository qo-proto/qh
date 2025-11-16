package benchmark

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

// If there's a body, it's in a DATA frame (9-byte header + body)
// Header size = total - DATA frame if present
func calculateHTTP2HeaderSize(data []byte, bodySize int) int {
	if bodySize == 0 {
		return len(data)
	}
	// DATA frame overhead: 9 bytes (frame header) + body size
	dataFrameSize := 9 + bodySize
	return len(data) - dataFrameSize
}

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

	endStream := len(tc.Request.Body) == 0
	if err := reqFramer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: reqHeaderBuf.Bytes(),
		EndHeaders:    true,
		EndStream:     endStream,
	}); err != nil {
		panic("failed to write HEADERS frame: " + err.Error())
	}

	if len(tc.Request.Body) > 0 {
		if err := reqFramer.WriteData(1, true, []byte(tc.Request.Body)); err != nil {
			panic("failed to write DATA frame: " + err.Error())
		}
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

	endStream = len(tc.Response.Body) == 0
	if err := respFramer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: respHeaderBuf.Bytes(),
		EndHeaders:    true,
		EndStream:     endStream,
	}); err != nil {
		panic("failed to write response HEADERS frame: " + err.Error())
	}

	if len(tc.Response.Body) > 0 {
		if err := respFramer.WriteData(1, true, []byte(tc.Response.Body)); err != nil {
			panic("failed to write response DATA frame: " + err.Error())
		}
	}

	reqBytes := reqBuf.Bytes()
	respBytes := respBuf.Bytes()

	reqBodySize := len(tc.Request.Body)
	reqHeaderSize := calculateHTTP2HeaderSize(reqBytes, reqBodySize)

	respBodySize := len(tc.Response.Body)
	respHeaderSize := calculateHTTP2HeaderSize(respBytes, respBodySize)

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
