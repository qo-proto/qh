package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

type Request struct {
	Method  Method
	Host    string
	Path    string
	Version uint8
	Headers []string // Ordered headers by position
	Body    []byte
}

type Response struct {
	Version    uint8
	StatusCode int
	Headers    []string // Ordered headers by position
	Body       []byte
}

// format QH request into wire format
func (r *Request) Format() []byte {
	// The first byte contains: Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
	// Bit layout: [Version (2 bits) | Method (3 bits) | Reserved (3 bits)]
	firstByte := (r.Version << 6) | (byte(r.Method) << 3)

	otherParts := []string{r.Host, r.Path}
	otherParts = append(otherParts, r.Headers...)
	headerPart := string(firstByte) + strings.Join(otherParts, "\x00")

	// Build message: headers + ETX + body
	result := []byte(headerPart)
	result = append(result, '\x03')
	result = append(result, r.Body...)
	return result
}

// format QH response into wire format
func (r *Response) Format() []byte {
	compactStatus := EncodeStatusCode(r.StatusCode)
	// First byte: Version (upper 2 bits) + Status Code (lower 6 bits)
	firstByte := (r.Version << 6) | compactStatus

	headerPart := string(firstByte) + strings.Join(r.Headers, "\x00")

	// Build message: headers + ETX + body
	result := []byte(headerPart)
	result = append(result, '\x03')
	result = append(result, r.Body...)
	return result
}

// IsRequestComplete checks if we have received a complete request
func IsRequestComplete(data []byte) (bool, error) {
	dataStr := string(data)
	headerPart, bodyPart, found := strings.Cut(dataStr, "\x03")
	if !found {
		return false, nil
	}

	if len(headerPart) == 0 {
		return false, nil
	}

	// Skip first byte (version + method), then split remaining header fields
	stringHeaderPart := headerPart[1:]
	if stringHeaderPart == "" {
		// Need at least host and path, which can't be present if header part after first byte is empty
		return false, nil
	}

	parts := strings.Split(stringHeaderPart, "\x00")
	// Expect at least host and path
	if len(parts) < 2 {
		return false, nil
	}

	// Headers follow host and path
	var headers []string
	if len(parts) > 2 {
		headers = parts[2:]
	}

	// If a Content-Length header is present (index 1 in ordered headers), enforce it
	if len(headers) > 1 && headers[1] != "" {
		expectedLen, err := strconv.Atoi(headers[1])
		if err != nil {
			return false, fmt.Errorf("invalid Content-Length: %s", headers[1])
		}
		return len(bodyPart) >= expectedLen, nil
	}

	// No Content-Length provided; treat as complete once headers and separator are present
	return true, nil
}

// IsResponseComplete checks if we have received a complete response based on Content-Length
func IsResponseComplete(data []byte) (bool, error) {
	dataStr := string(data)
	headerPart, bodyPart, found := strings.Cut(dataStr, "\x03")
	if !found {
		return false, nil
	}

	if len(headerPart) == 0 {
		return false, nil
	}

	// Skip first byte (version + status), then split remaining headers
	stringHeaderPart := headerPart[1:]
	if stringHeaderPart == "" {
		return false, nil
	}

	parts := strings.Split(stringHeaderPart, "\x00")
	// Content-Length is at header position 1 (parts[1])
	// parts[0] = Content-Type, parts[1] = Content-Length, parts[2-10] = other headers
	if len(parts) < 2 {
		return false, nil
	}

	contentLengthStr := parts[1]
	if contentLengthStr == "" {
		return false, errors.New("missing Content-Length header")
	}

	expectedLen, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return false, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
	}

	return len(bodyPart) >= expectedLen, nil
}

func ParseResponse(data []byte) (*Response, error) {
	// Split headers from body using the End of Text character
	dataStr := string(data)
	headerPart, bodyPart, found := strings.Cut(dataStr, "\x03")
	if !found {
		return nil, errors.New("invalid response: missing body separator")
	}

	body := []byte(bodyPart)

	if len(headerPart) == 0 {
		return nil, errors.New("invalid response: empty header part")
	}

	// First byte contains status and version
	firstByte := headerPart[0]
	version := firstByte >> 6               // Upper 2 bits
	compactStatus := firstByte & 0b00111111 // Lower 6 bits

	if version > 3 { // 2 bits can hold values 0-3
		return nil, fmt.Errorf("invalid version: %d", version)
	}

	httpStatusCode := DecodeStatusCode(compactStatus)

	// The rest of the header part is null-separated strings.
	stringHeaderPart := headerPart[1:]
	var parts []string
	// Only split if there's content, otherwise parts will be `[""]` which is not what we want.
	// We want an empty slice if there are no headers.
	if stringHeaderPart != "" {
		parts = strings.Split(stringHeaderPart, "\x00")
	}

	resp := &Response{
		Version:    version,
		StatusCode: httpStatusCode,
		Body:       body,
	}

	if len(parts) > 0 {
		resp.Headers = parts
	}

	// Validate Content-Length if present (header index 1)
	if len(resp.Headers) > 1 && resp.Headers[1] != "" {
		expectedLen, err := strconv.Atoi(resp.Headers[1])
		if err == nil && len(body) < expectedLen {
			return nil, errors.New("incomplete response: not all body data received")
		}
	}

	return resp, nil
}

func ParseRequest(data []byte) (*Request, error) {
	// Split headers from body using the End of Text character
	dataStr := string(data)
	headerPart, bodyPart, found := strings.Cut(dataStr, "\x03")
	if !found {
		return nil, errors.New("invalid request: missing body separator")
	}

	body := []byte(bodyPart)

	if len(headerPart) == 0 {
		return nil, errors.New("invalid request: empty header part")
	}

	// The first byte contains the method (lower 3 bits) and version (upper 5 bits).
	firstByte := headerPart[0]
	version := firstByte >> 6                       // Extract upper 2 bits
	method := Method((firstByte >> 3) & 0b00000111) // Extract middle 3 bits

	if method != GET && method != POST {
		return nil, fmt.Errorf("invalid method value: %d", method)
	}

	// The rest of the header part is null-separated strings.
	stringHeaderPart := headerPart[1:]
	parts := strings.Split(stringHeaderPart, "\x00")
	if len(parts) < 2 { // host, path
		return nil, errors.New("invalid request: not enough parts in header")
	}

	// Validate required fields are not empty
	if parts[0] == "" {
		return nil, errors.New("invalid request: empty host")
	}

	// Default empty path to root
	path := parts[1]
	if path == "" {
		path = "/"
	}

	req := &Request{
		Method:  method,
		Host:    parts[0],
		Path:    path,
		Version: version,
		Body:    body,
	}

	// The rest of the parts are headers
	if len(parts) > 2 {
		req.Headers = parts[2:]
	}

	// Validate Content-Length if present (header index 1)
	if len(req.Headers) > 1 && req.Headers[1] != "" {
		expectedLen, err := strconv.Atoi(req.Headers[1])
		if err == nil && len(body) < expectedLen {
			return nil, errors.New("incomplete request: not all body data received")
		}
	}

	return req, nil
}
