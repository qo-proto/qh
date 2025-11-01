//nolint:gosec // G115: Ignore integer overflow warnings
package qh

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const Version = 0

type Method int

const (
	GET    Method = 0
	POST   Method = 1
	PUT    Method = 2
	PATCH  Method = 3
	DELETE Method = 4
	HEAD   Method = 5
)

func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case PATCH:
		return "PATCH"
	case DELETE:
		return "DELETE"
	case HEAD:
		return "HEAD"
	default:
		return "UNKNOWN"
	}
}

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

func (ct ContentType) HeaderValue() string {
	return strconv.Itoa(int(ct))
}

func AcceptHeader(types ...ContentType) string {
	if len(types) == 0 {
		return ""
	}
	parts := make([]string, len(types))
	for i, ct := range types {
		parts[i] = strconv.Itoa(int(ct))
	}
	return strings.Join(parts, ",")
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

// Format encodes a QH request into the wire format using varint length prefixes.
// Wire format: <1-byte-method><varint:hostLen><host><varint:pathLen><path><varint:numHeaders>[headers]<varint:bodyLen><body>
func (r *Request) Format() []byte {
	// The first byte contains: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	// Bit layout: [Version (2 bits) | Method (3 bits) | Reserved (3 bits)]
	firstByte := (r.Version << 6) | (byte(r.Method) << 3)
	result := []byte{firstByte}
	result = AppendUvarint(result, uint64(len(r.Host)))
	result = append(result, []byte(r.Host)...)
	result = AppendUvarint(result, uint64(len(r.Path)))
	result = append(result, []byte(r.Path)...)
	result = AppendUvarint(result, uint64(len(r.Headers)))

	// Headers: each header has format based on header ID
	// TODO: Format 1 not implemented yet
	// Format 1 (complete key-value pairs): <headerID>
	// Format 2 (header name): <headerID><varint:valueLen><value>
	// Format 3 (custom header 0x00): <0x00><varint:keyLen><key><varint:valueLen><value>
	for key, value := range r.Headers {
		if headerID, exists := requestHeaderTable[key]; exists {
			// Format 2: Known header name, encode ID + value
			result = append(result, headerID)
			result = AppendUvarint(result, uint64(len(value)))
			result = append(result, []byte(value)...)
		} else {
			// Format 3: Custom header, encode 0x00 + key + value
			result = append(result, HeaderCustom)
			result = AppendUvarint(result, uint64(len(key)))
			result = append(result, []byte(key)...)
			result = AppendUvarint(result, uint64(len(value)))
			result = append(result, []byte(value)...)
		}
	}

	result = AppendUvarint(result, uint64(len(r.Body)))
	result = append(result, r.Body...)

	return result
}

// Format encodes a QH response into the wire format using varint length prefixes.
// Wire format: <1-byte-status><varint:numHeaders>[headers]<varint:bodyLen><body>
func (r *Response) Format() []byte {
	compactStatus := EncodeStatusCode(r.StatusCode)
	// First byte: Version (upper 2 bits) + Status Code (lower 6 bits)
	firstByte := (r.Version << 6) | compactStatus
	result := []byte{firstByte}
	result = AppendUvarint(result, uint64(len(r.Headers)))

	// Headers: each header has format based on header ID
	// TODO: Format 1 is not implemented yet
	// Format 1 (complete key-value pairs): <headerID>
	// Format 2 (header name): <headerID><varint:valueLen><value>
	// Format 3 (custom header 0x00): <0x00><varint:keyLen><key><varint:valueLen><value>
	for key, value := range r.Headers {
		if headerID, exists := responseHeaderTable[key]; exists {
			// Format 2: Known header name, encode ID + value
			result = append(result, headerID)
			result = AppendUvarint(result, uint64(len(value)))
			result = append(result, []byte(value)...)
		} else {
			// Format 3: Custom header, encode 0x00 + key + value
			result = append(result, HeaderCustom)
			result = AppendUvarint(result, uint64(len(key)))
			result = append(result, []byte(key)...)
			result = AppendUvarint(result, uint64(len(value)))
			result = append(result, []byte(value)...)
		}
	}
	result = AppendUvarint(result, uint64(len(r.Body)))
	result = append(result, r.Body...)

	return result
}

func parseCustomHeader(data []byte, offset int) (string, string, int, error) {
	keyLen, n, readErr := ReadUvarint(data, offset)
	if readErr != nil {
		return "", "", offset, fmt.Errorf("failed to read custom header key length: %w", readErr)
	}
	offset += n

	keyLenInt := int(keyLen)
	if offset+keyLenInt > len(data) {
		return "", "", offset, errors.New("custom header key length exceeds buffer")
	}
	key := string(data[offset : offset+keyLenInt])
	offset += keyLenInt

	valueLen, n, readErr := ReadUvarint(data, offset)
	if readErr != nil {
		return "", "", offset, fmt.Errorf("failed to read custom header value length: %w", readErr)
	}
	offset += n

	valueLenInt := int(valueLen)
	if offset+valueLenInt > len(data) {
		return "", "", offset, errors.New("custom header value length exceeds buffer")
	}
	value := string(data[offset : offset+valueLenInt])
	offset += valueLenInt

	return key, value, offset, nil
}

func parseKnownHeader(data []byte, offset int) (string, int, error) {
	valueLen, n, readErr := ReadUvarint(data, offset)
	if readErr != nil {
		return "", offset, fmt.Errorf("failed to read header value length: %w", readErr)
	}
	offset += n

	valueLenInt := int(valueLen)
	if offset+valueLenInt > len(data) {
		return "", offset, errors.New("header value length exceeds buffer")
	}
	value := string(data[offset : offset+valueLenInt])
	offset += valueLenInt

	return value, offset, nil
}

func parseHeaderEntry(
	data []byte,
	offset int,
	headerID byte,
	headerNames map[byte]string,
) (string, string, int, error) {
	if headerID == HeaderCustom {
		// Format 3: Custom header <0x00><varint:keyLen><key><varint:valueLen><value>
		return parseCustomHeader(data, offset)
	}

	if name, exists := headerNames[headerID]; exists {
		// Format 2: Known header <headerID><varint:valueLen><value>
		value, newOffset, err := parseKnownHeader(data, offset)
		return name, value, newOffset, err
	}

	// Unknown header ID: skip value (forward compatibility), return empty key
	_, newOffset, err := parseKnownHeader(data, offset)
	if err != nil {
		return "", "", offset, fmt.Errorf("failed to read unknown header value length: %w", err)
	}
	return "", "", newOffset, nil
}

func parseHeaders(
	data []byte,
	offset int,
	numHeaders uint64,
	headerNames map[byte]string,
) (map[string]string, int, error) {
	headers := make(map[string]string)

	for range numHeaders {
		if offset >= len(data) {
			return nil, offset, errors.New("unexpected end while reading headers")
		}

		headerID := data[offset]
		offset++

		key, value, newOffset, err := parseHeaderEntry(data, offset, headerID, headerNames)
		if err != nil {
			return nil, offset, err
		}
		offset = newOffset

		// Skip unknown headers
		if key == "" {
			continue
		}

		headers[key] = value
	}

	return headers, offset, nil
}

func IsRequestComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	offset := 1 // Skip first byte (version + method)

	// Helper function to reduce duplication
	checkField := func(fieldName string) (bool, error) {
		length, n, err := ReadUvarint(data, offset)
		if errors.Is(err, ErrVarintIncomplete) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("reading %s length: %w", fieldName, err)
		}

		// Prevent integer overflow and DoS
		if length > uint64(len(data)) {
			return false, fmt.Errorf("%s length too large: %d", fieldName, length)
		}

		offset += n
		lengthInt := int(length)
		if offset+lengthInt > len(data) {
			return false, nil // Need more data
		}

		offset += lengthInt
		return true, nil
	}

	// Read host
	if complete, err := checkField("host"); !complete {
		return false, err
	}

	// Read path
	if complete, err := checkField("path"); !complete {
		return false, err
	}

	// Read headers count
	numHeaders, n, err := ReadUvarint(data, offset)
	if errors.Is(err, ErrVarintIncomplete) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading header count: %w", err)
	}
	offset += n

	// Process headers
	for range numHeaders {
		if offset >= len(data) {
			return false, nil
		}

		headerID := data[offset]
		offset++

		// Custom header has key + value, others just value
		if headerID == HeaderCustom {
			if complete, err := checkField("header key"); !complete {
				return false, err
			}
		}

		if complete, err := checkField("header value"); !complete {
			return false, err
		}
	}

	// Read body
	if complete, err := checkField("body"); !complete {
		return false, err
	}

	return true, nil
}

func IsResponseComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	offset := 1 // Skip first byte (version + status)

	// Helper function to reduce duplication
	checkField := func(fieldName string) (bool, error) {
		length, n, err := ReadUvarint(data, offset)
		if errors.Is(err, ErrVarintIncomplete) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("reading %s length: %w", fieldName, err)
		}

		// Prevent integer overflow and DoS
		if length > uint64(len(data)) {
			return false, fmt.Errorf("%s length too large: %d", fieldName, length)
		}

		offset += n
		lengthInt := int(length)
		if offset+lengthInt > len(data) {
			return false, nil // Need more data
		}

		offset += lengthInt
		return true, nil
	}

	// Read headers count
	numHeaders, n, err := ReadUvarint(data, offset)
	if errors.Is(err, ErrVarintIncomplete) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading header count: %w", err)
	}
	offset += n

	// Process headers
	for range numHeaders {
		if offset >= len(data) {
			return false, nil
		}

		headerID := data[offset]
		offset++

		// Custom header has key + value, others just value
		if headerID == HeaderCustom {
			if complete, err := checkField("header key"); !complete {
				return false, err
			}
		}

		if complete, err := checkField("header value"); !complete {
			return false, err
		}
	}

	// Read body
	if complete, err := checkField("body"); !complete {
		return false, err
	}

	return true, nil
}

func ParseResponse(data []byte) (*Response, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid response: empty data")
	}

	offset := 0

	// Parse first byte: Version (2 bits, bits 7-6) | Status Code (6 bits, bits 5-0)
	firstByte := data[offset]
	offset++

	version := firstByte >> 6               // Extract upper 2 bits
	compactStatus := firstByte & 0b00111111 // Extract lower 6 bits

	if version > 3 { // 2 bits can hold values 0-3
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	httpStatusCode := DecodeStatusCode(compactStatus)

	// Read number of headers
	numHeaders, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid response: failed to read header count: %w", err)
	}
	offset += n

	// Parse headers
	headers, newOffset, err := parseHeaders(data, offset, numHeaders, responseHeaderNames)
	if err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	offset = newOffset

	// Read body length and body
	bodyLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid response: failed to read body length: %w", err)
	}
	offset += n

	bodyLenInt := int(bodyLen)
	if offset+bodyLenInt > len(data) {
		return nil, errors.New("invalid response: body length exceeds buffer")
	}
	body := data[offset : offset+bodyLenInt]

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

	offset := 0

	// Parse first byte: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	firstByte := data[offset]
	offset++

	version := firstByte >> 6                       // Extract upper 2 bits
	method := Method((firstByte >> 3) & 0b00000111) // Extract middle 3 bits

	if method < GET || method > HEAD { // valid methods are 0-5
		return nil, fmt.Errorf("invalid method value: %d", method)
	}

	// Read host length and host
	hostLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read host length: %w", err)
	}
	offset += n

	hostLenInt := int(hostLen)
	if offset+hostLenInt > len(data) {
		return nil, errors.New("invalid request: host length exceeds buffer")
	}
	host := string(data[offset : offset+hostLenInt])
	offset += hostLenInt

	if host == "" {
		return nil, errors.New("invalid request: empty host")
	}

	// Read path length and path
	pathLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read path length: %w", err)
	}
	offset += n

	pathLenInt := int(pathLen)
	if offset+pathLenInt > len(data) {
		return nil, errors.New("invalid request: path length exceeds buffer")
	}
	path := string(data[offset : offset+pathLenInt])
	offset += pathLenInt

	if path == "" {
		path = "/"
	}

	// Read number of headers
	numHeaders, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read header count: %w", err)
	}
	offset += n

	// Parse headers
	headers, newOffset, err := parseHeaders(data, offset, numHeaders, requestHeaderNames)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	offset = newOffset

	// Read body length and body
	bodyLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read body length: %w", err)
	}
	offset += n

	bodyLenInt := int(bodyLen)
	if offset+bodyLenInt > len(data) {
		return nil, errors.New("invalid request: body length exceeds buffer")
	}
	body := data[offset : offset+bodyLenInt]

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
