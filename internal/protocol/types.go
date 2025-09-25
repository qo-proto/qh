package protocol

import (
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
	Method  Method
	Host    string
	Path    string
	Version string
	// TODO: Add ContentType to request headers
	Headers []string // Ordered headers by position
	Body    string
}

type Response struct {
	Version    string
	StatusCode int
	Headers    []string // Ordered headers by position
	Body       string
}

// format QH request into wire format
func (r *Request) Format() string {
	var parts []string

	// Request line: <Host>\0<Path>\0<Version>
	requestLine := fmt.Sprintf("%s\x00%s\x00%s", r.Host, r.Path, r.Version)
	parts = append(parts, requestLine)

	parts = append(parts, r.Headers...)

	// Join header parts with null byte, and separate from body with End of Text char.
	headerPart := strings.Join(parts, "\x00")
	return headerPart + "\x03" + r.Body
}

// format QH response into wire format
func (r *Response) Format() string {
	var parts []string

	// Response line: <Version>\0<Status-Code>
	responseLine := fmt.Sprintf("%s\x00%d", r.Version, r.StatusCode)
	parts = append(parts, responseLine)

	parts = append(parts, r.Headers...)

	// Join header parts with null byte, and separate from body with End of Text char.
	headerPart := strings.Join(parts, "\x00")
	return headerPart + "\x03" + r.Body
}

func ParseRequest(data string) (*Request, error) {
	// Split headers from body using the End of Text character
	headerPart, body, found := strings.Cut(data, "\x03")
	if !found {
		return nil, fmt.Errorf("invalid request: missing body separator")
	}

	parts := strings.Split(headerPart, "\x00")
	if len(parts) < 3 { // host, path, version
		return nil, fmt.Errorf("invalid request: not enough parts in header")
	}

	req := &Request{
		Host:    parts[0],
		Path:    parts[1],
		Version: parts[2],
		Body:    body,
	}

	// The rest of the parts are headers
	if len(parts) > 3 {
		req.Headers = parts[3:]
	}

	// Infer method from body presence
	if len(req.Body) > 0 {
		req.Method = POST
	} else {
		req.Method = GET
	}

	return req, nil
}

func ParseResponse(data string) (*Response, error) {
	// Split headers from body using the End of Text character
	headerPart, body, found := strings.Cut(data, "\x03")
	if !found {
		return nil, fmt.Errorf("invalid response: missing body separator")
	}

	parts := strings.Split(headerPart, "\x00")
	if len(parts) < 2 { // version, status
		return nil, fmt.Errorf("invalid response: empty")
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}

	resp := &Response{
		Version:    parts[0],
		StatusCode: statusCode,
		Body:       body,
	}

	// The rest of the parts are headers
	if len(parts) > 2 {
		resp.Headers = parts[2:]
	}

	return resp, nil
}
