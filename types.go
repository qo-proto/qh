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
	GET Method = iota
	POST
	PUT
	PATCH
	DELETE
	HEAD
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
	Custom ContentType = iota // Allows for a custom string in the body if needed
	TextPlain
	JSON
	HTML
	OctetStream
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
	// CustomHeader is a special header ID (0) used to indicate custom headers
	CustomHeader byte = 0
)

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

// encodeHeaders implements the three-format header encoding:
// Format 1 (complete key-value pairs): <headerID>
// Format 2 (known header name with value): <headerID><varint:valueLen><value>
// Format 3 (custom header): <0x00><varint:keyLen><key><varint:valueLen><value>
func encodeHeaders(
	headers map[string]string,
	completePairs map[string]byte,
	nameOnly map[string]byte,
) []byte {
	var result []byte

	for key, value := range headers {
		// Try Format 1: exact match for complete key-value pair, just send header ID
		lookupKey := key + ":" + value
		if headerID, exists := completePairs[lookupKey]; exists {
			result = append(result, headerID)
			continue
		}

		// Try Format 2: name-only match with custom value, encode ID
		if headerID, exists := nameOnly[key]; exists {
			result = append(result, headerID)
			result = AppendUvarint(result, uint64(len(value)))
			result = append(result, []byte(value)...)
			continue
		}

		// Format 3: Custom header not in static table
		result = append(result, CustomHeader)
		result = AppendUvarint(result, uint64(len(key)))
		result = append(result, []byte(key)...)
		result = AppendUvarint(result, uint64(len(value)))
		result = append(result, []byte(value)...)
	}

	return result
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

	result = append(
		result,
		encodeHeaders(r.Headers, requestHeaderCompletePairs, requestHeaderNameOnly)...)

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

	result = append(
		result,
		encodeHeaders(r.Headers, responseHeaderCompletePairs, responseHeaderNameOnly)...)

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
	staticTable map[byte]headerEntry,
) (string, string, int, error) {
	if headerID == CustomHeader {
		// Format 3: Custom header <0x00><varint:keyLen><key><varint:valueLen><value>
		return parseCustomHeader(data, offset)
	}

	if entry, exists := staticTable[headerID]; exists {
		if entry.value != "" {
			// Format 1: Complete key-value pair, just return the entry
			return entry.name, entry.value, offset, nil
		}
		// Format 2: Known header name with custom value <headerID><varint:valueLen><value>
		value, newOffset, err := parseKnownHeader(data, offset)
		return entry.name, value, newOffset, err
	}

	// Unknown header ID
	return "", "", offset, fmt.Errorf(
		"unknown header ID 0x%02X - protocol version mismatch or corrupted message",
		headerID,
	)
}

func parseHeaders(
	data []byte,
	offset int,
	numHeaders uint64,
	staticTable map[byte]headerEntry,
) (map[string]string, int, error) {
	headers := make(map[string]string)

	for range numHeaders {
		if offset >= len(data) {
			return nil, offset, errors.New("unexpected end while reading headers")
		}

		headerID := data[offset]
		offset++

		key, value, newOffset, err := parseHeaderEntry(data, offset, headerID, staticTable)
		if err != nil {
			return nil, offset, err
		}
		offset = newOffset

		headers[key] = value
	}

	return headers, offset, nil
}

// validate and skip over a length-prefixed field
func checkField(data []byte, offset *int, fieldName string) (bool, error) {
	length, n, err := ReadUvarint(data, *offset)
	if errors.Is(err, ErrVarintIncomplete) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s length: %w", fieldName, err)
	}

	const maxInt = int(^uint(0) >> 1)
	if length > uint64(maxInt) {
		return false, fmt.Errorf("%s length too large: %d", fieldName, length)
	}

	lengthInt := int(length)

	if *offset > maxInt-lengthInt {
		return false, fmt.Errorf("%s offset overflow", fieldName)
	}

	*offset += n
	if *offset+lengthInt > len(data) {
		return false, nil // Need more data
	}

	*offset += lengthInt
	return true, nil
}

func IsRequestComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	offset := 1 // Skip first byte (version + method)

	if complete, err := checkField(data, &offset, "host"); !complete {
		return false, err
	}

	if complete, err := checkField(data, &offset, "path"); !complete {
		return false, err
	}

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
		if headerID == CustomHeader {
			if complete, err := checkField(data, &offset, "header key"); !complete {
				return false, err
			}
		}

		if complete, err := checkField(data, &offset, "header value"); !complete {
			return false, err
		}
	}

	if complete, err := checkField(data, &offset, "body"); !complete {
		return false, err
	}

	return true, nil
}

func IsResponseComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	offset := 1 // Skip first byte (version + status)

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
		if headerID == CustomHeader {
			if complete, err := checkField(data, &offset, "header key"); !complete {
				return false, err
			}
		}

		if complete, err := checkField(data, &offset, "header value"); !complete {
			return false, err
		}
	}

	if complete, err := checkField(data, &offset, "body"); !complete {
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

	if version > 3 {
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	httpStatusCode := DecodeStatusCode(compactStatus)

	numHeaders, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid response: failed to read header count: %w", err)
	}
	offset += n

	headers, newOffset, err := parseHeaders(data, offset, numHeaders, responseHeaderStaticTable)
	if err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	offset = newOffset

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

	if version > 3 {
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	if method < GET || method > HEAD { // valid methods are 0-5
		return nil, fmt.Errorf("invalid method value: %d", method)
	}

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

	numHeaders, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read header count: %w", err)
	}
	offset += n

	headers, newOffset, err := parseHeaders(data, offset, numHeaders, requestHeaderStaticTable)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	offset = newOffset

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
