package qh

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

const Version = 0

type Method int

const (
	GET  Method = 0
	POST Method = 1
)

func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	default:
		return "UNKNOWN"
	}
}

// TODO: maybe add a helper method protocol.AcceptHeader() to use types directly in the client
// e.g. protocol.AcceptHeader(protocol.JSON, protocol.TextPlain) and not "1,2"
type ContentType int

const (
	// 4 bits for content type (16 types)
	Custom      ContentType = 0 // Allows for a custom string in the body if needed
	TextPlain   ContentType = 1
	JSON        ContentType = 2
	HTML        ContentType = 3
	OctetStream ContentType = 4
	// ... up to 15
)

func IsValidContentType(code int) bool {
	return code >= 0 && code <= 15
}

func (ct ContentType) String() string {
	switch ct {
	case Custom:
		return "custom"
	case TextPlain:
		return "text/plain"
	case JSON:
		return "application/json"
	case HTML:
		return "text/html"
	case OctetStream:
		return "application/octet-stream"
	default:
		return "unknown"
	}
}

const (
	HeaderCustom byte = 0

	HeaderReqAccept         byte = 1
	HeaderReqAcceptEncoding byte = 2
	// ID 3 is reserved (conflicts with ETX separator \x03)
	HeaderReqAcceptLanguage  byte = 4
	HeaderReqContentType     byte = 5
	HeaderReqContentLength   byte = 6
	HeaderReqAuthorization   byte = 7
	HeaderReqCookie          byte = 8
	HeaderReqUserAgent       byte = 9
	HeaderReqReferer         byte = 10
	HeaderReqOrigin          byte = 11
	HeaderReqIfNoneMatch     byte = 12
	HeaderReqIfModifiedSince byte = 13
	HeaderReqRange           byte = 14
	HeaderReqXPayment        byte = 15

	HeaderRespContentType   byte = 1
	HeaderRespContentLength byte = 2
	// ID 3 is reserved (conflicts with ETX separator \x03)
	HeaderRespCacheControl              byte = 4
	HeaderRespContentEncoding           byte = 5
	HeaderRespDate                      byte = 6
	HeaderRespETag                      byte = 7
	HeaderRespExpires                   byte = 8
	HeaderRespLastModified              byte = 9
	HeaderRespAccessControlAllowOrigin  byte = 10
	HeaderRespAccessControlAllowMethods byte = 11
	HeaderRespAccessControlAllowHeaders byte = 12
	HeaderRespSetCookie                 byte = 13
	HeaderRespLocation                  byte = 14
	HeaderRespContentSecurityPolicy     byte = 15
	HeaderRespXContentTypeOptions       byte = 16
	HeaderRespXFrameOptions             byte = 17
	HeaderRespVary                      byte = 18
	HeaderRespXPaymentResponse          byte = 19
)

var requestHeaderTable = map[string]byte{
	"Accept":            HeaderReqAccept,
	"Accept-Encoding":   HeaderReqAcceptEncoding,
	"Accept-Language":   HeaderReqAcceptLanguage,
	"Content-Type":      HeaderReqContentType,
	"Content-Length":    HeaderReqContentLength,
	"Authorization":     HeaderReqAuthorization,
	"Cookie":            HeaderReqCookie,
	"User-Agent":        HeaderReqUserAgent,
	"Referer":           HeaderReqReferer,
	"Origin":            HeaderReqOrigin,
	"If-None-Match":     HeaderReqIfNoneMatch,
	"If-Modified-Since": HeaderReqIfModifiedSince,
	"Range":             HeaderReqRange,
	"X-Payment":         HeaderReqXPayment,
}

var requestHeaderNames map[byte]string // filled in init()

var responseHeaderTable = map[string]byte{
	"Content-Type":                 HeaderRespContentType,
	"Content-Length":               HeaderRespContentLength,
	"Cache-Control":                HeaderRespCacheControl,
	"Content-Encoding":             HeaderRespContentEncoding,
	"Date":                         HeaderRespDate,
	"ETag":                         HeaderRespETag,
	"Expires":                      HeaderRespExpires,
	"Last-Modified":                HeaderRespLastModified,
	"Access-Control-Allow-Origin":  HeaderRespAccessControlAllowOrigin,
	"Access-Control-Allow-Methods": HeaderRespAccessControlAllowMethods,
	"Access-Control-Allow-Headers": HeaderRespAccessControlAllowHeaders,
	"Set-Cookie":                   HeaderRespSetCookie,
	"Location":                     HeaderRespLocation,
	"Content-Security-Policy":      HeaderRespContentSecurityPolicy,
	"X-Content-Type-Options":       HeaderRespXContentTypeOptions,
	"X-Frame-Options":              HeaderRespXFrameOptions,
	"Vary":                         HeaderRespVary,
	"X-Payment-Response":           HeaderRespXPaymentResponse,
}

var responseHeaderNames map[byte]string // filled in init()

func init() {
	requestHeaderNames = make(map[byte]string, len(requestHeaderTable))
	for name, id := range requestHeaderTable {
		requestHeaderNames[id] = name
	}

	responseHeaderNames = make(map[byte]string, len(responseHeaderTable))
	for name, id := range responseHeaderTable {
		responseHeaderNames[id] = name
	}
}

type Request struct {
	Method  Method
	Host    string
	Path    string
	Version uint8
	Headers map[string]string
	Body    []byte
}

type Response struct {
	Version    uint8
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// TODO: pre-allocate capacity in both format methods

// format QH request into wire format
func (r *Request) Format() []byte {
	// The first byte contains: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	// Bit layout: [Version (2 bits) | Method (3 bits) | Reserved (3 bits)]
	firstByte := (r.Version << 6) | (byte(r.Method) << 3)

	// Build message: first byte + headers + ETX + body
	result := []byte{firstByte}
	result = append(result, []byte(r.Host)...)
	result = append(result, '\x00')
	result = append(result, []byte(r.Path)...)
	result = append(result, '\x00')

	for key, value := range r.Headers {
		if headerID, exists := requestHeaderTable[key]; exists {
			result = append(result, headerID)
			result = append(result, '\x00')
			result = append(result, []byte(value)...)
			result = append(result, '\x00')
		} else {
			result = append(result, HeaderCustom)
			result = append(result, '\x00')
			result = append(result, []byte(key)...)
			result = append(result, '\x00')
			result = append(result, []byte(value)...)
			result = append(result, '\x00')
		}
	}

	result = append(result, '\x03')
	result = append(result, r.Body...)
	return result
}

// format QH response into wire format
func (r *Response) Format() []byte {
	compactStatus := EncodeStatusCode(r.StatusCode)
	// First byte: Version (upper 2 bits) + Status Code (lower 6 bits)
	firstByte := (r.Version << 6) | compactStatus

	// Build message: first byte + headers + ETX + body
	result := []byte{firstByte}

	for key, value := range r.Headers {
		if headerID, exists := responseHeaderTable[key]; exists {
			result = append(result, headerID)
			result = append(result, '\x00')
			result = append(result, []byte(value)...)
			result = append(result, '\x00')
		} else {
			result = append(result, HeaderCustom)
			result = append(result, '\x00')
			result = append(result, []byte(key)...)
			result = append(result, '\x00')
			result = append(result, []byte(value)...)
			result = append(result, '\x00')
		}
	}

	result = append(result, '\x03')
	result = append(result, r.Body...)
	return result
}

func IsRequestComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	allHeaders, bodyBytes, found := bytes.Cut(data, []byte{'\x03'})
	if !found {
		return false, nil
	}

	headerBytes := allHeaders[1:]

	// ensure we have at least host\x00path\x00 (minimum 2 null bytes)
	// NOTE: This is faster than parsing the full header structure as before
	offset := 0
	nullCount := 0
	for offset < len(headerBytes) {
		if headerBytes[offset] == '\x00' {
			nullCount++
			if nullCount >= 2 {
				break
			}
		}
		offset++
	}
	if nullCount < 2 {
		return false, nil
	}

	headers := parseRequestHeaders(headerBytes)

	if contentLengthStr, ok := headers["Content-Length"]; ok {
		expectedLen, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return false, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
		}
		return len(bodyBytes) >= expectedLen, nil
	}

	return true, nil
}

func parseRequestHeaders(headerBytes []byte) map[string]string {
	headers := make(map[string]string)
	offset := 0

	// Skip host field (already parsed by ParseRequest)
	for offset < len(headerBytes) && headerBytes[offset] != 0 {
		offset++
	}
	offset++ // Skip \x00 separator

	// Skip path field (already parsed by ParseRequest)
	for offset < len(headerBytes) && headerBytes[offset] != 0 {
		offset++
	}
	offset++ // Skip \x00 separator

	// Parse headers: each is <header-id>\x00<value>\x00
	// Custom headers: \x00\x00<key-name>\x00<value>\x00
	for offset < len(headerBytes) {
		headerID := headerBytes[offset]
		offset++

		// Expect separator after header ID
		if offset >= len(headerBytes) || headerBytes[offset] != 0 {
			break
		}
		offset++ // Skip \x00 separator

		var key string
		if headerID == HeaderCustom {
			// Custom header: read key name
			keyStart := offset
			for offset < len(headerBytes) && headerBytes[offset] != 0 {
				offset++
			}
			key = string(headerBytes[keyStart:offset])
			offset++ // Skip \x00 separator
		} else if name, exists := requestHeaderNames[headerID]; exists {
			// Known header: look up name from ID
			key = name
		} else {
			// Unknown header ID: skip value and continue
			for offset < len(headerBytes) && headerBytes[offset] != 0 {
				offset++
			}
			offset++ // Skip \x00 separator
			continue
		}

		// Read header value
		valueStart := offset
		for offset < len(headerBytes) && headerBytes[offset] != 0 {
			offset++
		}
		value := string(headerBytes[valueStart:offset])
		offset++ // Skip \x00 separator

		headers[key] = value
	}

	return headers
}

func IsResponseComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	allHeaders, bodyBytes, found := bytes.Cut(data, []byte{'\x03'})
	if !found {
		return false, nil
	}

	headerBytes := allHeaders[1:]

	headers := parseResponseHeaders(headerBytes)

	if contentLengthStr, ok := headers["Content-Length"]; ok {
		expectedLen, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return false, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
		}
		return len(bodyBytes) >= expectedLen, nil
	}

	return true, nil
}

func parseResponseHeaders(headerBytes []byte) map[string]string {
	headers := make(map[string]string)
	offset := 0

	// Parse headers: each is <header-id>\x00<value>\x00
	// Custom headers: \x00\x00<key-name>\x00<value>\x00
	for offset < len(headerBytes) {
		headerID := headerBytes[offset]
		offset++

		// Expect separator after header ID
		if offset >= len(headerBytes) || headerBytes[offset] != 0 {
			break
		}
		offset++ // Skip \x00 separator

		var key string
		if headerID == HeaderCustom {
			// Custom header: read key name
			keyStart := offset
			for offset < len(headerBytes) && headerBytes[offset] != 0 {
				offset++
			}
			key = string(headerBytes[keyStart:offset])
			offset++ // Skip \x00 separator
		} else if name, exists := responseHeaderNames[headerID]; exists {
			// Known header: look up name from ID
			key = name
		} else {
			// Unknown header ID: skip value and continue
			for offset < len(headerBytes) && headerBytes[offset] != 0 {
				offset++
			}
			offset++ // Skip \x00 separator
			continue
		}

		// Read header value
		valueStart := offset
		for offset < len(headerBytes) && headerBytes[offset] != 0 {
			offset++
		}
		value := string(headerBytes[valueStart:offset])
		offset++ // Skip \x00 separator

		headers[key] = value
	}

	return headers
}

func ParseResponse(data []byte) (*Response, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid response: empty data")
	}

	// Split at ETX (\x03) separator between headers and body
	allHeaders, body, found := bytes.Cut(data, []byte{'\x03'})
	if !found {
		return nil, errors.New("invalid response: missing body separator")
	}

	// Parse first byte: Version (2 bits, bits 7-6) | Status Code (6 bits, bits 5-0)
	// Wire format: <first-byte>[<header-id>\x00<value>\x00...]\x03<body>
	firstByte := allHeaders[0]
	version := firstByte >> 6               // Extract upper 2 bits
	compactStatus := firstByte & 0b00111111 // Extract lower 6 bits

	if version > 3 { // 2 bits can hold values 0-3
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	httpStatusCode := DecodeStatusCode(compactStatus)

	headerBytes := allHeaders[1:]
	headers := parseResponseHeaders(headerBytes)

	resp := &Response{
		Version:    version,
		StatusCode: httpStatusCode,
		Headers:    headers,
		Body:       body,
	}

	if contentLengthStr, ok := headers["Content-Length"]; ok {
		expectedLen, err := strconv.Atoi(contentLengthStr)
		if err == nil && len(body) != expectedLen {
			return nil, errors.New("invalid response: body length does not match Content-Length")
		}
	}

	return resp, nil
}

func ParseRequest(data []byte) (*Request, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid request: empty data")
	}

	// Split at ETX (\x03) separator between headers and body
	allHeaders, body, found := bytes.Cut(data, []byte{'\x03'})
	if !found {
		return nil, errors.New("invalid request: missing body separator")
	}

	// Parse first byte: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	firstByte := allHeaders[0]
	version := firstByte >> 6                       // Extract upper 2 bits
	method := Method((firstByte >> 3) & 0b00000111) // Extract middle 3 bits

	if method != GET && method != POST {
		return nil, fmt.Errorf("invalid method value: %d", method)
	}

	// Parse host and path from header section
	// Wire format: <first-byte><host>\x00<path>\x00[<header-id>\x00<value>\x00...]\x03<body>
	headerBytes := allHeaders[1:]
	offset := 0

	// Extract host (required)
	hostStart := offset
	for offset < len(headerBytes) && headerBytes[offset] != 0 {
		offset++
	}
	host := string(headerBytes[hostStart:offset])

	// Check if we found a null terminator or reached the end
	if offset >= len(headerBytes) {
		return nil, errors.New("invalid request: missing null terminator after host")
	}
	offset++ // Skip \x00 separator

	if host == "" {
		return nil, errors.New("invalid request: empty host")
	}

	// Extract path (defaults to "/" if empty)
	// Path may be empty or missing, both default to "/"
	var path string
	if offset < len(headerBytes) {
		pathStart := offset
		for offset < len(headerBytes) && headerBytes[offset] != 0 {
			offset++
		}
		path = string(headerBytes[pathStart:offset])
	}

	if path == "" {
		path = "/"
	}

	headers := parseRequestHeaders(headerBytes)

	req := &Request{
		Method:  method,
		Host:    host,
		Path:    path,
		Version: version,
		Headers: headers,
		Body:    body,
	}

	if contentLengthStr, ok := headers["Content-Length"]; ok {
		expectedLen, err := strconv.Atoi(contentLengthStr)
		if err == nil && len(body) != expectedLen {
			return nil, errors.New("invalid request: body length does not match Content-Length")
		}
	}

	return req, nil
}
