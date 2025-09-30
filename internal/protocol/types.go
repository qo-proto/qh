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

	// Per the protocol spec, the body separator and EOT are always present.
	return headerPart + "\x03" + r.Body + "\x04"
}

// format QH response into wire format
func (r *Response) Format() string {
	var parts []string

	compactStatus := EncodeStatusCode(r.StatusCode)
	responseLine := fmt.Sprintf("%s\x00%d", r.Version, compactStatus)
	parts = append(parts, responseLine)

	parts = append(parts, r.Headers...)

	headerPart := strings.Join(parts, "\x00")
	return headerPart + "\x03" + r.Body + "\x04"
}

func ParseResponse(data string) (*Response, error) {
	// The EOT character (\x04) marks the end of the message.
	// It should be trimmed before parsing.
	data, found := strings.CutSuffix(data, "\x04")
	if !found {
		return nil, errors.New("invalid response: missing EOT terminator")
	}
	// Split headers from body using the End of Text character
	headerPart, body, found := strings.Cut(data, "\x03")
	if !found {
		return nil, errors.New("invalid response: missing body separator")
	}

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

	return resp, nil
}

func ParseRequest(data string) (*Request, error) {
	// The EOT character (\x04) marks the end of the message.
	// It should be trimmed before parsing.
	data, found := strings.CutSuffix(data, "\x04")
	if !found {
		return nil, errors.New("invalid request: missing EOT terminator")
	}
	// Split headers from body using the End of Text character.
	// This is now consistent with ParseResponse.
	headerPart, body, found := strings.Cut(data, "\x03")
	if !found {
		return nil, errors.New("invalid request: missing body separator")
	}

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
