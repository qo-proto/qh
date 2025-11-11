package benchmark

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/quic-go/qpack"
	"github.com/quic-go/quic-go/quicvarint"
)

func EncodeHTTP3(tc TestCase) EncodedResult {
	// Encode request
	var reqQpackBuf bytes.Buffer
	reqEncoder := qpack.NewEncoder(&reqQpackBuf)

	if err := reqEncoder.WriteField(qpack.HeaderField{Name: ":method", Value: tc.Request.Method}); err != nil {
		panic("failed to write :method: " + err.Error())
	}
	if err := reqEncoder.WriteField(qpack.HeaderField{Name: ":path", Value: tc.Request.Path}); err != nil {
		panic("failed to write :path: " + err.Error())
	}
	if err := reqEncoder.WriteField(qpack.HeaderField{Name: ":scheme", Value: "https"}); err != nil {
		panic("failed to write :scheme: " + err.Error())
	}
	if err := reqEncoder.WriteField(qpack.HeaderField{Name: ":authority", Value: tc.Request.Host}); err != nil {
		panic("failed to write :authority: " + err.Error())
	}

	for k, v := range tc.Request.Headers {
		if err := reqEncoder.WriteField(qpack.HeaderField{
			Name:  strings.ToLower(k),
			Value: v,
		}); err != nil {
			panic("failed to write header " + k + ": " + err.Error())
		}
	}

	reqEncodedHeaders := reqQpackBuf.Bytes()

	// Build request HEADERS frame
	var reqResult bytes.Buffer
	reqResult.Write(quicvarint.Append(nil, 0x01))                           // Frame type: HEADERS
	reqResult.Write(quicvarint.Append(nil, uint64(len(reqEncodedHeaders)))) // Length
	reqResult.Write(reqEncodedHeaders)                                      // Payload

	// Add request DATA frame if there's a body
	if len(tc.Request.Body) > 0 {
		reqResult.Write(quicvarint.Append(nil, 0x00))                         // Frame type: DATA
		reqResult.Write(quicvarint.Append(nil, uint64(len(tc.Request.Body)))) // Length
		reqResult.Write(tc.Request.Body)                                      // Payload
	}

	// Encode response
	var respQpackBuf bytes.Buffer
	respEncoder := qpack.NewEncoder(&respQpackBuf)

	// :status pseudo-header
	if err := respEncoder.WriteField(qpack.HeaderField{
		Name:  ":status",
		Value: strconv.Itoa(tc.Response.StatusCode),
	}); err != nil {
		panic("failed to write :status: " + err.Error())
	}

	for k, v := range tc.Response.Headers {
		if err := respEncoder.WriteField(qpack.HeaderField{
			Name:  strings.ToLower(k),
			Value: v,
		}); err != nil {
			panic("failed to write response header " + k + ": " + err.Error())
		}
	}

	respEncodedHeaders := respQpackBuf.Bytes()

	// Build response HEADERS frame
	var respResult bytes.Buffer
	respResult.Write(quicvarint.Append(nil, 0x01))                            // Frame type: HEADERS
	respResult.Write(quicvarint.Append(nil, uint64(len(respEncodedHeaders)))) // Length
	respResult.Write(respEncodedHeaders)                                      // Payload

	// Add response DATA frame if there's a body
	if len(tc.Response.Body) > 0 {
		respResult.Write(quicvarint.Append(nil, 0x00))                          // Frame type: DATA
		respResult.Write(quicvarint.Append(nil, uint64(len(tc.Response.Body)))) // Length
		respResult.Write(tc.Response.Body)                                      // Payload
	}

	reqBytes := reqResult.Bytes()
	respBytes := respResult.Bytes()

	return EncodedResult{
		RequestBytes:  reqBytes,
		ResponseBytes: respBytes,
		RequestSize:   len(reqBytes),
		ResponseSize:  len(respBytes),
		TotalSize:     len(reqBytes) + len(respBytes),
	}
}
