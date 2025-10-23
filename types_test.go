package qh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestFormat(t *testing.T) {
	tests := []struct {
		name     string
		request  *Request
		expected []byte
	}{
		{
			name: "GET minimal",
			request: &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/hello",
				Version: 0,
				Headers: map[string]string{},
			},
			expected: []byte("\x00example.com\x00/hello\x00\x03"),
		},
		{
			name: "GET with Accept header",
			request: &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/api",
				Version: 0,
				Headers: map[string]string{
					"Accept": "2",
				},
			},
			expected: []byte("\x00example.com\x00/api\x00\x01\x002\x00\x03"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.request.Format()
			require.Equal(t, tt.expected, actual, "Wire format mismatch.\nExpected (hex): %x\nActual (hex):   %x", tt.expected, actual)
		})
	}
}

func TestRequestFormatWithBody(t *testing.T) {
	req := &Request{
		Method:  POST,
		Host:    "api.example.com",
		Path:    "/submit",
		Version: 0,
		Headers: map[string]string{
			"Content-Type": "2",
		},
		Body: []byte(`{"key":"val"}`),
	}

	actual := req.Format()
	// First byte: 0x08 = version 0 (00), method POST (001), reserved (000)
	expected := []byte("\x08api.example.com\x00/submit\x00\x05\x002\x00\x03{\"key\":\"val\"}")
	require.Equal(t, expected, actual, "Wire format mismatch.\nExpected (hex): %x\nActual (hex):   %x", expected, actual)
}

func TestResponseFormat(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		expected []byte
	}{
		{
			name: "200 OK minimal",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers:    map[string]string{},
				Body:       []byte("OK"),
			},
			expected: []byte("\x00\x03OK"),
		},
		{
			name: "200 OK with Content-Type",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "1",
				},
				Body: []byte("Hello"),
			},
			expected: []byte("\x00\x01\x001\x00\x03Hello"),
		},
		{
			name: "404 Not Found",
			response: &Response{
				Version:    0,
				StatusCode: 404,
				Headers: map[string]string{
					"Content-Type": "1",
				},
				Body: []byte("Not Found"),
			},
			// First byte: 0x01 = version 0 (00), compact status 1 (000001) → HTTP 404
			expected: []byte("\x01\x01\x001\x00\x03Not Found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.response.Format()
			require.Equal(t, tt.expected, actual, "Wire format mismatch.\nExpected (hex): %x\nActual (hex):   %x", tt.expected, actual)
		})
	}
}

func TestResponseFormatEmpty(t *testing.T) {
	resp := &Response{
		Version:    0,
		StatusCode: 204,
		Headers: map[string]string{
			"Content-Type": "0",
		},
		Body: []byte{},
	}

	actual := resp.Format()
	expected := []byte("\x0C\x01\x000\x00\x03")
	require.Equal(t, expected, actual, "Wire format mismatch.\nExpected (hex): %x\nActual (hex):   %x", expected, actual)
}

func TestParseRequestBasic(t *testing.T) {
	data := []byte("\x00example.com\x00/hello.txt\x00\x01\x001\x00\x00\x00Accept-Language\x00en-US,en;q=0.5\x00\x03")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/hello.txt", req.Path)
	require.Equal(t, uint8(0), req.Version)
	require.Equal(t, "1", req.Headers["Accept"])
	require.Equal(t, "en-US,en;q=0.5", req.Headers["Accept-Language"])
	require.Empty(t, req.Body)
}

func TestParseRequestWithBody(t *testing.T) {
	data := []byte("\x08example.com\x00/submit\x00\x05\x002\x00\x06\x0016\x00\x03{\"name\": \"test\"}")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/submit", req.Path)
	require.Equal(t, uint8(0), req.Version)
	require.Equal(t, "2", req.Headers["Content-Type"])
	require.Equal(t, "16", req.Headers["Content-Length"])
	require.JSONEq(t, `{"name": "test"}`, string(req.Body))
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	data := []byte("\x08example.com\x00/submit\x03line1\nline2\nline3")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, []byte("line1\nline2\nline3"), req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	data := []byte("\x08example.com\x00/path\x03test body")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Empty(t, req.Headers)
	require.Equal(t, []byte("test body"), req.Body)
}

func TestParseRequestEmptyPathDefaultsToRoot(t *testing.T) {
	data := []byte("\x00example.com\x00\x03")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/", req.Path) // Empty path should default to "/"
	require.Equal(t, uint8(0), req.Version)
	require.Empty(t, req.Headers)
	require.Empty(t, req.Body)
}

func TestParseRequestErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"no body separator", []byte("example.com\x00/path\x001.0")},
		{"empty", []byte("")},
		{"invalid request line, too few parts", []byte("\x00example.com")},
		{"host missing", []byte("\x00\x00/path\x03")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequest(tt.data)
			require.Error(t, err)
		})
	}
}

func TestParseResponseBasic(t *testing.T) {
	data := []byte("\x00\x01\x001\x00\x02\x0013\x00\x06\x001758784800\x00\x03Hello, world!")

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, uint8(0), resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "1", resp.Headers["Content-Type"])
	require.Equal(t, "13", resp.Headers["Content-Length"])
	require.Equal(t, "1758784800", resp.Headers["Date"])
	require.Equal(t, []byte("Hello, world!"), resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	data := []byte("\x00\x01\x001\x00\x03Response body")

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, uint8(0), resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "1", resp.Headers["Content-Type"])
	require.Equal(t, []byte("Response body"), resp.Body)
}

func TestParseResponseErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"no body separator", []byte("\x00")},
		{"empty", []byte("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseResponse(tt.data)
			require.Error(t, err)
		})
	}
}

func TestMethodString(t *testing.T) {
	tests := []struct {
		method   Method
		expected string
	}{
		{GET, "GET"},
		{POST, "POST"},
		/*{PUT, "PUT"},
		{DELETE, "DELETE"},
		{HEAD, "HEAD"},*/
		{Method(123), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.method.String())
		})
	}
}

func TestIsValidContentType(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		valid bool
	}{
		{"Custom", 0, true},
		{"TextPlain", 1, true},
		{"JSON", 2, true},
		{"HTML", 3, true},
		{"OctetStream", 4, true},
		{"MaxValid", 15, true},
		{"TooHigh", 16, false},
		{"Invalid99", 99, false},
		{"Negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.valid, IsValidContentType(tt.code))
		})
	}
}

// TODO: add tests for IsRequestComplete and IsResponseComplete

// Round-trip tests verify that Format() and Parse() are consistent

func TestRequestRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		request *Request
	}{
		{
			name: "GET with headers",
			request: &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/api/data",
				Version: 0,
				Headers: map[string]string{
					"Accept":          "2,1",
					"Accept-Encoding": "gzip, br",
					"User-Agent":      "QH-Client/1.0",
				},
				Body: []byte{},
			},
		},
		{
			name: "POST with body and headers",
			request: &Request{
				Method:  POST,
				Host:    "api.example.com",
				Path:    "/submit",
				Version: 0,
				Headers: map[string]string{
					"Content-Type":   "2",
					"Content-Length": "15",
					"Authorization":  "Bearer token123",
				},
				Body: []byte(`{"name":"test"}`),
			},
		},
		{
			name: "GET with minimal headers",
			request: &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/",
				Version: 0,
				Headers: map[string]string{},
				Body:    []byte{},
			},
		},
		{
			name: "POST with custom header",
			request: &Request{
				Method:  POST,
				Host:    "example.com",
				Path:    "/custom",
				Version: 0,
				Headers: map[string]string{
					"Content-Type":     "1",
					"Content-Length":   "5",
					"X-Custom-Header":  "custom-value",
					"X-Another-Custom": "another-value",
				},
				Body: []byte("hello"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.request.Format()
			parsed, err := ParseRequest(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.request.Method, parsed.Method)
			require.Equal(t, tt.request.Host, parsed.Host)
			require.Equal(t, tt.request.Path, parsed.Path)
			require.Equal(t, tt.request.Version, parsed.Version)
			require.Equal(t, tt.request.Headers, parsed.Headers)
			require.Equal(t, tt.request.Body, parsed.Body)
		})
	}
}

func TestResponseRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
	}{
		{
			name: "200 OK with headers",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type":   "2",
					"Content-Length": "15",
					"Cache-Control":  "max-age=3600",
					"Date":           "1758784800",
				},
				Body: []byte(`{"status":"ok"}`),
			},
		},
		{
			name: "404 Not Found",
			response: &Response{
				Version:    0,
				StatusCode: 404,
				Headers: map[string]string{
					"Content-Type":   "1",
					"Content-Length": "9",
				},
				Body: []byte("Not Found"),
			},
		},
		{
			name: "204 No Content",
			response: &Response{
				Version:    0,
				StatusCode: 204,
				Headers: map[string]string{
					"Content-Type": "1",
				},
				Body: []byte{},
			},
		},
		{
			name: "Response with custom headers",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type":       "2",
					"Content-Length":     "2",
					"X-Custom-Response":  "custom-value",
					"X-Another-Response": "another-value",
				},
				Body: []byte("{}"),
			},
		},
		{
			name: "Response with CORS headers",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type":                 "2",
					"Content-Length":               "2",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "GET, POST, PUT",
					"Access-Control-Allow-Headers": "Content-Type, Authorization",
				},
				Body: []byte("{}"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.response.Format()
			parsed, err := ParseResponse(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.response.Version, parsed.Version)
			require.Equal(t, tt.response.StatusCode, parsed.StatusCode)
			require.Equal(t, tt.response.Headers, parsed.Headers)
			require.Equal(t, tt.response.Body, parsed.Body)
		})
	}
}
