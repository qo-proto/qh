package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const Version = "1.0"

type Method int

const (
	GET  Method = 1
	POST Method = 2
)

// String returns the string representation of the QH protocols method. Implements fmt.Stringer interface, used for logging and debugging.
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

type Request struct {
	Host    string
	Path    string
	Version string
	// TODO: Add ContentType to request headers
	Headers []string // Ordered headers by position
	Body    []byte
}

type Response struct {
	Version    string
	StatusCode int
	Headers    []string // Ordered headers by position
	Body       []byte
}

// format QH request into wire format
func (r *Request) Format() []byte {
	var parts []string

	// Request line: <Host>\0<Path>\0<Version>
	requestLine := fmt.Sprintf("%s\x00%s\x00%s", r.Host, r.Path, r.Version)
	parts = append(parts, requestLine)

	parts = append(parts, r.Headers...)

	// Join header parts with null byte, and separate from body with End of Text char.
	headerPart := strings.Join(parts, "\x00")

	// Build message: headers + ETX + body
	result := []byte(headerPart)
	result = append(result, '\x03')
	result = append(result, r.Body...)
	return result
}

// format QH response into wire format
func (r *Response) Format() []byte {
	var parts []string

	compactStatus := EncodeStatusCode(r.StatusCode)
	responseLine := fmt.Sprintf("%s\x00%d", r.Version, compactStatus)
	parts = append(parts, responseLine)

	parts = append(parts, r.Headers...)

	headerPart := strings.Join(parts, "\x00")

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

	parts := strings.Split(headerPart, "\x00")
	if len(parts) < 3 {
		return false, nil
	}

	if len(bodyPart) == 0 {
		return true, nil
	}

	return true, nil
}

// IsResponseComplete checks if we have received a complete response based on Content-Length
func IsResponseComplete(data []byte) (bool, error) {
	dataStr := string(data)
	headerPart, bodyPart, found := strings.Cut(dataStr, "\x03")
	if !found {
		return false, nil
	}

	parts := strings.Split(headerPart, "\x00")
	if len(parts) < 5 {
		return false, nil
	}

	contentLengthStr := parts[4]
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

	parts := strings.Split(headerPart, "\x00")
	if len(parts) < 2 { // version, status
		return nil, errors.New("invalid response: empty")
	}

	// Validate required fields are not empty
	if parts[0] == "" {
		return nil, errors.New("invalid response: empty version")
	}
	if parts[1] == "" {
		return nil, errors.New("invalid response: empty status code")
	}

	compactStatus, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}

	if compactStatus < 0 || compactStatus > 255 {
		return nil, fmt.Errorf("compact status code out of range: %d", compactStatus)
	}

	httpStatusCode := DecodeStatusCode(uint8(compactStatus))

	resp := &Response{
		Version:    parts[0],
		StatusCode: httpStatusCode,
		Body:       body,
	}

	// The rest of the parts are headers
	if len(parts) > 2 {
		resp.Headers = parts[2:]
	}

	// Validate Content-Length if present (header index 2)
	if len(resp.Headers) > 2 && resp.Headers[2] != "" {
		expectedLen, err := strconv.Atoi(resp.Headers[2])
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

	parts := strings.Split(headerPart, "\x00")
	if len(parts) < 3 { // host, path, version
		return nil, errors.New("invalid request: not enough parts in header")
	}

	// Validate required fields are not empty
	if parts[0] == "" {
		return nil, errors.New("invalid request: empty host")
	}
	if parts[2] == "" {
		return nil, errors.New("invalid request: empty version")
	}

	// Default empty path to root
	path := parts[1]
	if path == "" {
		path = "/"
	}

	req := &Request{
		Host:    parts[0],
		Path:    path,
		Version: parts[2],
		Body:    body,
	}

	// The rest of the parts are headers
	if len(parts) > 3 {
		req.Headers = parts[3:]
	}

	return req, nil
}
