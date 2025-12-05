// Package qh implements the QH (Quite Ok HTTP) protocol.
//
//nolint:gosec // G115: false warnings
package qh

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// Version is the current QH protocol version number.
	// This value is encoded in the first byte of both requests and responses.
	Version = 0

	versionBitShift = 6          // Version is stored in upper 2 bits (bits 7-6)
	methodBitShift  = 3          // Method is stored in middle 3 bits (bits 5-3)
	statusCodeMask  = 0b00111111 // Status code uses lower 6 bits
	methodMask      = 0b00000111 // Method uses 3 bits
	maxVersionValue = 3          // Maximum version (2 bits: 0-3)
	maxHostLength   = 253        // Maximum host length (DNS label length limit)
	firstByteOffset = 1          // Offset to skip the first byte in wire format
)

// Method represents a QH method encoded as an integer for compact wire format.
// Methods are encoded in 3 bits.
type Method int

// QH method constants for use in QH requests.
// These are encoded as 3-bit values in the wire format.
const (
	GET     Method = iota // GET retrieves a resource
	POST                  // POST submits data to be processed
	PUT                   // PUT replaces a resource
	PATCH                 // PATCH partially modifies a resource
	DELETE                // DELETE removes a resource
	HEAD                  // HEAD retrieves headers only
	OPTIONS               // OPTIONS describes communication options
)

// String returns the QH method name as a string (e.g., "GET", "POST").
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
	case OPTIONS:
		return "OPTIONS"
	default:
		return "UNKNOWN"
	}
}

const (
	// customHeader is a special header ID (0) used to indicate custom headers
	customHeader byte = 0
)

// Request represents a QH protocol request message.
// It contains the QH method, target host and path, protocol version,
// headers as key-value pairs, and an optional body.
type Request struct {
	Method  Method            // QH method (GET, POST, etc.)
	Host    string            // Target hostname
	Path    string            // Request path (e.g., "/api/users")
	Version uint8             // Protocol version number
	Headers map[string]string // Request headers as key-value pairs
	Body    []byte            // Optional request body
}

// Response represents a QH protocol response message.
// It contains the protocol version, QH status code, headers, and body.
type Response struct {
	Version    uint8             // Protocol version number
	StatusCode int               // QH status code
	Headers    map[string]string // Response headers as key-value pairs
	Body       []byte            // Response body content
}

// encodeHeaders implements the three-format header encoding:
// Format 1 (complete key-value pairs): <headerID>
// Format 2 (known header name with value): <headerID><varint:valueLen><value>
// Format 3 (custom header): <0x00><varint:keyLen><key><varint:valueLen><value>
// NOTE: All header names MUST be normalized (converted to lowercase)
func encodeHeaders(
	headers map[string]string,
	completePairs map[string]byte,
	nameOnly map[string]byte,
) []byte {
	var result []byte

	for key, value := range headers {
		key = strings.ToLower(key)

		// Try Format 1: exact match for complete key-value pair, just send header ID
		lookupKey := key + ":" + value
		if headerID, exists := completePairs[lookupKey]; exists {
			result = append(result, headerID)
			continue
		}

		// Try Format 2: name-only match with custom value, encode ID
		if headerID, exists := nameOnly[key]; exists {
			result = append(result, headerID)
			result = appendUvarint(result, uint64(len(value)))
			result = append(result, []byte(value)...)
			continue
		}

		// Format 3: Custom header not in static table
		result = append(result, customHeader)
		result = appendUvarint(result, uint64(len(key)))
		result = append(result, []byte(key)...)
		result = appendUvarint(result, uint64(len(value)))
		result = append(result, []byte(value)...)
	}

	return result
}

// Format encodes a QH request into wire format bytes using varint length prefixes.
//
// Wire format structure:
//   - 1 byte: Version (2 bits) | Method (3 bits) | Reserved (3 bits)
//   - varint: host length, followed by host bytes
//   - varint: path length, followed by path bytes
//   - varint: headers length, followed by encoded headers
//   - varint: body length, followed by body bytes
func (r *Request) Format() []byte {
	// The first byte contains: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	// Bit layout: [Version (2 bits) | Method (3 bits) | Reserved (3 bits)]
	firstByte := (r.Version << versionBitShift) | (byte(r.Method) << methodBitShift)
	result := []byte{firstByte}
	result = appendUvarint(result, uint64(len(r.Host)))
	result = append(result, []byte(r.Host)...)
	result = appendUvarint(result, uint64(len(r.Path)))
	result = append(result, []byte(r.Path)...)

	// Encode headers first to get total length
	encodedHeaders := encodeHeaders(r.Headers, requestHeaderCompletePairs, requestHeaderNameOnly)
	result = appendUvarint(result, uint64(len(encodedHeaders)))
	result = append(result, encodedHeaders...)

	result = appendUvarint(result, uint64(len(r.Body)))
	result = append(result, r.Body...)

	return result
}

// Format encodes a QH response into wire format bytes using varint length prefixes.
//
// Wire format structure:
//   - 1 byte: Version (2 bits) | Compact status code (6 bits)
//   - varint: headers length, followed by encoded headers
//   - varint: body length, followed by body bytes
func (r *Response) Format() []byte {
	compactStatus := encodeStatusCode(r.StatusCode)
	// First byte: Version (upper 2 bits) + Status Code (lower 6 bits)
	firstByte := (r.Version << versionBitShift) | compactStatus
	result := []byte{firstByte}

	// Encode headers first to get total length
	encodedHeaders := encodeHeaders(r.Headers, responseHeaderCompletePairs, responseHeaderNameOnly)
	result = appendUvarint(result, uint64(len(encodedHeaders)))
	result = append(result, encodedHeaders...)

	result = appendUvarint(result, uint64(len(r.Body)))
	result = append(result, r.Body...)

	return result
}

func parseCustomHeader(data []byte, offset int) (string, string, int, error) {
	keyLen, n, readErr := ReadUvarint(data, offset)
	if readErr != nil {
		return "", "", offset, fmt.Errorf("failed to read custom header key length: %w", readErr)
	}
	offset += n

	if keyLen > uint64(len(data)-offset) {
		return "", "", offset, errors.New("custom header key length exceeds buffer")
	}
	keyLenInt := int(keyLen)
	key := string(data[offset : offset+keyLenInt])
	offset += keyLenInt

	valueLen, n, readErr := ReadUvarint(data, offset)
	if readErr != nil {
		return "", "", offset, fmt.Errorf("failed to read custom header value length: %w", readErr)
	}
	offset += n

	if valueLen > uint64(len(data)-offset) {
		return "", "", offset, errors.New("custom header value length exceeds buffer")
	}
	valueLenInt := int(valueLen)
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

	if valueLen > uint64(len(data)-offset) {
		return "", offset, errors.New("header value length exceeds buffer")
	}
	valueLenInt := int(valueLen)
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
	if headerID == customHeader {
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
	headersLen uint64,
	staticTable map[byte]headerEntry,
) (map[string]string, int, error) {
	headers := make(map[string]string)
	if headersLen > uint64(len(data)-offset) {
		return nil, offset, errors.New("headers length exceeds buffer")
	}
	endOffset := offset + int(headersLen)

	for offset < endOffset {
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

		key = strings.ToLower(key)
		headers[key] = value
	}

	return headers, offset, nil
}

// validate and skip over a length-prefixed field
func checkField(data []byte, offset *int, fieldName string) (bool, error) {
	length, n, err := ReadUvarint(data, *offset)
	if errors.Is(err, errVarintIncomplete) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s length: %w", fieldName, err)
	}

	*offset += n
	remaining := len(data) - *offset
	if remaining < 0 || length > uint64(remaining) {
		return false, nil // Need more data
	}

	*offset += int(length)
	return true, nil
}

func isRequestComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	offset := firstByteOffset // Skip first byte (version + method)

	if complete, err := checkField(data, &offset, "host"); !complete {
		return false, err
	}

	if complete, err := checkField(data, &offset, "path"); !complete {
		return false, err
	}

	// Check headers length field and skip headers section
	if complete, err := checkField(data, &offset, "headers"); !complete {
		return false, err
	}

	if complete, err := checkField(data, &offset, "body"); !complete {
		return false, err
	}

	return true, nil
}

func isResponseComplete(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	offset := firstByteOffset // Skip first byte (version + status)

	// Check headers length field and skip headers section
	if complete, err := checkField(data, &offset, "headers"); !complete {
		return false, err
	}

	if complete, err := checkField(data, &offset, "body"); !complete {
		return false, err
	}

	return true, nil
}

func parseResponse(data []byte) (*Response, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid response: empty data")
	}

	offset := 0

	// Parse first byte: Version (2 bits, bits 7-6) | Status Code (6 bits, bits 5-0)
	firstByte := data[offset]
	offset++

	version := firstByte >> versionBitShift     // Extract upper 2 bits
	compactStatus := firstByte & statusCodeMask // Extract lower 6 bits

	if version > maxVersionValue {
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	httpStatusCode := decodeStatusCode(compactStatus)

	headersLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid response: failed to read headers length: %w", err)
	}
	offset += n

	headers, newOffset, err := parseHeaders(data, offset, headersLen, responseHeaderStaticTable)
	if err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	offset = newOffset

	bodyLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid response: failed to read body length: %w", err)
	}
	offset += n

	if bodyLen > uint64(len(data)-offset) {
		return nil, errors.New("invalid response: body length exceeds buffer")
	}
	bodyLenInt := int(bodyLen)
	body := data[offset : offset+bodyLenInt]

	resp := &Response{
		Version:    version,
		StatusCode: httpStatusCode,
		Headers:    headers,
		Body:       body,
	}

	return resp, nil
}

func parseRequest(data []byte) (*Request, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid request: empty data")
	}

	offset := 0

	// Parse first byte: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	firstByte := data[offset]
	offset++

	version := firstByte >> versionBitShift                      // Extract upper 2 bits
	method := Method((firstByte >> methodBitShift) & methodMask) // Extract middle 3 bits

	if version > maxVersionValue {
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	if method < GET || method > OPTIONS { // valid methods are 0-6
		return nil, fmt.Errorf("invalid method value: %d", method)
	}

	hostLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read host length: %w", err)
	}
	offset += n

	if hostLen > uint64(len(data)-offset) {
		return nil, errors.New("invalid request: host length exceeds buffer")
	}
	hostLenInt := int(hostLen)
	host := string(data[offset : offset+hostLenInt])
	offset += hostLenInt

	if host == "" {
		return nil, errors.New("invalid request: empty host")
	}

	if len(host) > maxHostLength {
		return nil, fmt.Errorf("invalid request: host exceeds maximum length of %d characters", maxHostLength)
	}

	pathLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read path length: %w", err)
	}
	offset += n

	if pathLen > uint64(len(data)-offset) {
		return nil, errors.New("invalid request: path length exceeds buffer")
	}
	pathLenInt := int(pathLen)
	path := string(data[offset : offset+pathLenInt])
	offset += pathLenInt

	if path == "" {
		path = "/"
	}

	headersLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read headers length: %w", err)
	}
	offset += n

	headers, newOffset, err := parseHeaders(data, offset, headersLen, requestHeaderStaticTable)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	offset = newOffset

	bodyLen, n, err := ReadUvarint(data, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to read body length: %w", err)
	}
	offset += n

	if bodyLen > uint64(len(data)-offset) {
		return nil, errors.New("invalid request: body length exceeds buffer")
	}
	bodyLenInt := int(bodyLen)
	body := data[offset : offset+bodyLenInt]

	req := &Request{
		Method:  method,
		Host:    host,
		Path:    path,
		Version: version,
		Headers: headers,
		Body:    body,
	}

	return req, nil
}
