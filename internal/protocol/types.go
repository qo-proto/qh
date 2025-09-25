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

type Request struct {
	Method  Method
	Host    string
	Path    string
	Version string
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

	// Request line: <Method> <Host> <Path> <Version>
	requestLine := fmt.Sprintf("%d %s %s %s", int(r.Method), r.Host, r.Path, r.Version)
	parts = append(parts, requestLine)

	parts = append(parts, r.Headers...)

	parts = append(parts, "")     // separate headers from body
	parts = append(parts, r.Body) // always add body

	return strings.Join(parts, "\n")
}

// format QH response into wire format
func (r *Response) Format() string {
	var parts []string

	// Response line: <Version> <Status-Code>
	responseLine := fmt.Sprintf("%s %d", r.Version, r.StatusCode)
	parts = append(parts, responseLine)

	parts = append(parts, r.Headers...)

	parts = append(parts, "")     // separate headers from body
	parts = append(parts, r.Body) // always add body

	return strings.Join(parts, "\n")
}

func ParseRequest(data string) (*Request, error) {
	lines := strings.Split(data, "\n")
	if len(lines) < 1 {
		return nil, errors.New("invalid request: empty")
	}

	requestParts := strings.Fields(lines[0])
	if len(requestParts) != 4 {
		return nil, fmt.Errorf("invalid request line: %s", lines[0])
	}

	method, err := strconv.Atoi(requestParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid method: %s", requestParts[0])
	}

	req := &Request{
		Method:  Method(method),
		Host:    requestParts[1],
		Path:    requestParts[2],
		Version: requestParts[3],
	}

	// find empty line separating headers from body
	headerEnd := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			headerEnd = i
			break
		}
		req.Headers = append(req.Headers, lines[i])
	}

	// validate that the empty line spearator was found
	if headerEnd == -1 {
		return nil, errors.New("invalid request: missing empty line separator between headers and body")
	}

	// parse body if exists
	if headerEnd+1 < len(lines) {
		bodyLines := lines[headerEnd+1:]
		req.Body = strings.Join(bodyLines, "\n")
	}

	return req, nil
}

func ParseResponse(data string) (*Response, error) {
	lines := strings.Split(data, "\n")
	if len(lines) < 1 {
		return nil, errors.New("invalid response: empty")
	}

	responseParts := strings.Fields(lines[0])
	if len(responseParts) != 2 {
		return nil, fmt.Errorf("invalid response line: %s", lines[0])
	}

	statusCode, err := strconv.Atoi(responseParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", responseParts[1])
	}

	resp := &Response{
		Version:    responseParts[0],
		StatusCode: statusCode,
	}

	// find empty line separating headers from body
	headerEnd := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			headerEnd = i
			break
		}
		resp.Headers = append(resp.Headers, lines[i])
	}

	// validate that the empty line spearator was found
	if headerEnd == -1 {
		return nil, errors.New("invalid response: missing empty line separator between headers and body")
	}

	// parse body if exists
	if headerEnd+1 < len(lines) {
		bodyLines := lines[headerEnd+1:]
		resp.Body = strings.Join(bodyLines, "\n")
	}

	return resp, nil
}
