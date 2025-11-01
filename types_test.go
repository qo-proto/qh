package qh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Wire format structure Request:
//
//	<firstByte>         - Upper 2 bits: version (0-3), Middle 3 bits: method (0-7), Lower 3 bits: reserved
//	<varint:hostLen>    - Length of host string
//	<host>              - Host string bytes
//	<varint:pathLen>    - Length of path string
//	<path>              - Path string bytes
//	<varint:numHeaders> - Number of headers
//	[headers]           - For each header:
//	                        <headerID>        - Standard header ID (1-15) or 0 for custom
//	                        <varint:valueLen> - Length of value string
//	                        <value>           - Value string bytes
//	                        [if headerID==0:  - Only for custom headers:
//	                          <varint:nameLen>  - Length of custom header name
//	                          <name>]           - Custom header name bytes
//	<varint:bodyLen>    - Length of body
//	<body>              - Body bytes
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
			expected: []byte("\x00\x0Bexample.com\x06/hello\x00\x00"),
		},
		{
			name: "GET with Accept header",
			request: &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/api",
				Version: 0,
				Headers: map[string]string{
					"Accept": JSON.HeaderValue(),
				},
			},
			expected: []byte("\x00\x0Bexample.com\x04/api\x01\x01\x012\x00"),
		},
		{
			name: "POST with body",
			request: &Request{
				Method:  POST,
				Host:    "api.example.com",
				Path:    "/submit",
				Version: 0,
				Headers: map[string]string{
					"Content-Type": JSON.HeaderValue(),
				},
				Body: []byte(`{"key":"val"}`),
			},
			expected: []byte("\x08\x0Fapi.example.com\x07/submit\x01\x05\x012\x0D{\"key\":\"val\"}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.request.Format()
			require.Equal(t, tt.expected, actual, "Wire format mismatch.\nExpected (hex): %x\nActual (hex):   %x", tt.expected, actual)
		})
	}
}

// Wire format structure Response:
//
//	<firstByte>         - Upper 2 bits: version (0-3), Lower 6 bits: compact status code (0-63)
//	<varint:numHeaders> - Number of headers
//	[headers]           - For each header:
//	                        <headerID>        - Standard header ID (1-19) or 0 for custom
//	                        <varint:valueLen> - Length of value string
//	                        <value>           - Value string bytes
//	                        [if headerID==0:  - Only for custom headers:
//	                          <varint:nameLen>  - Length of custom header name
//	                          <name>]           - Custom header name bytes
//	<varint:bodyLen>    - Length of body
//	<body>              - Body bytes
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
			expected: []byte("\x00\x00\x02OK"),
		},
		{
			name: "200 OK with Content-Type",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": TextPlain.HeaderValue(),
				},
				Body: []byte("Hello"),
			},
			expected: []byte("\x00\x01\x01\x011\x05Hello"),
		},
		{
			name: "404 Not Found",
			response: &Response{
				Version:    0,
				StatusCode: 404,
				Headers: map[string]string{
					"Content-Type": TextPlain.HeaderValue(),
				},
				Body: []byte("Not Found"),
			},
			expected: []byte("\x01\x01\x01\x011\x09Not Found"),
		},
		{
			name: "204 No Content empty",
			response: &Response{
				Version:    0,
				StatusCode: 204,
				Headers: map[string]string{
					"Content-Type": Custom.HeaderValue(),
				},
				Body: []byte{},
			},
			expected: []byte("\x0C\x01\x01\x010\x00"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.response.Format()
			require.Equal(t, tt.expected, actual, "Wire format mismatch.\nExpected (hex): %x\nActual (hex):   %x", tt.expected, actual)
		})
	}
}

func TestParseRequestBasic(t *testing.T) {
	data := []byte("\x00\x0Bexample.com\x0A/hello.txt\x02\x01\x011\x00\x0FAccept-Language\x0Een-US,en;q=0.5\x00")

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
	data := []byte("\x08\x0Bexample.com\x07/submit\x02\x05\x012\x06\x0216\x10{\"name\": \"test\"}")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/submit", req.Path)
	require.Equal(t, uint8(0), req.Version)
	require.Equal(t, JSON.HeaderValue(), req.Headers["Content-Type"])
	require.Equal(t, "16", req.Headers["Content-Length"])
	require.JSONEq(t, `{"name": "test"}`, string(req.Body))
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	data := []byte("\x08\x0Bexample.com\x07/submit\x00\x11line1\nline2\nline3")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, []byte("line1\nline2\nline3"), req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	data := []byte("\x08\x0Bexample.com\x05/path\x00\x09test body")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Empty(t, req.Headers)
	require.Equal(t, []byte("test body"), req.Body)
}

func TestParseRequestEmptyPathDefaultsToRoot(t *testing.T) {
	data := []byte("\x00\x0Bexample.com\x00\x00\x00")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/", req.Path)
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
	data := []byte("\x00\x03\x01\x011\x02\x0213\x06\x0A1758784800\x0DHello, world!")

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, uint8(0), resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, TextPlain.HeaderValue(), resp.Headers["Content-Type"])
	require.Equal(t, "13", resp.Headers["Content-Length"])
	require.Equal(t, "1758784800", resp.Headers["Date"])
	require.Equal(t, []byte("Hello, world!"), resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	data := []byte("\x00\x01\x01\x011\x0DResponse body")

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, uint8(0), resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, TextPlain.HeaderValue(), resp.Headers["Content-Type"])
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
		{PUT, "PUT"},
		{PATCH, "PATCH"},
		{DELETE, "DELETE"},
		{HEAD, "HEAD"},
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

func TestContentTypeHeaderValue(t *testing.T) {
	tests := []struct {
		name        string
		contentType ContentType
		expected    string
	}{
		{"Custom", Custom, "0"},
		{"TextPlain", TextPlain, "1"},
		{"JSON", JSON, "2"},
		{"HTML", HTML, "3"},
		{"OctetStream", OctetStream, "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.contentType.HeaderValue()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func assertRequestEqual(t *testing.T, expected, actual *Request) {
	t.Helper()
	require.Equal(t, expected.Method, actual.Method)
	require.Equal(t, expected.Host, actual.Host)
	require.Equal(t, expected.Path, actual.Path)
	require.Equal(t, expected.Version, actual.Version)
	require.Equal(t, expected.Headers, actual.Headers)
	require.Equal(t, expected.Body, actual.Body)
}

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
					"Accept":          AcceptHeader(JSON, TextPlain),
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
					"Content-Type":   JSON.HeaderValue(),
					"Content-Length": "15",
					"Authorization":  "Bearer token123",
				},
				Body: []byte(`{"name":"test"}`),
			},
		},
		{
			name: "PUT with body",
			request: &Request{
				Method:  PUT,
				Host:    "api.example.com",
				Path:    "/user/123",
				Version: 0,
				Headers: map[string]string{
					"Content-Type":   JSON.HeaderValue(),
					"Content-Length": "18",
				},
				Body: []byte(`{"name":"updated"}`),
			},
		},
		{
			name: "PATCH with body",
			request: &Request{
				Method:  PATCH,
				Host:    "api.example.com",
				Path:    "/user/123",
				Version: 0,
				Headers: map[string]string{
					"Content-Type":   JSON.HeaderValue(),
					"Content-Length": "12",
				},
				Body: []byte(`{"age":"30"}`),
			},
		},
		{
			name: "DELETE without body",
			request: &Request{
				Method:  DELETE,
				Host:    "api.example.com",
				Path:    "/user/123",
				Version: 0,
				Headers: map[string]string{},
				Body:    []byte{},
			},
		},
		{
			name: "HEAD without body",
			request: &Request{
				Method:  HEAD,
				Host:    "example.com",
				Path:    "/api/data",
				Version: 0,
				Headers: map[string]string{
					"Accept": AcceptHeader(JSON, TextPlain),
				},
				Body: []byte{},
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
					"Content-Type":     TextPlain.HeaderValue(),
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
			assertRequestEqual(t, tt.request, parsed)
		})
	}
}

func assertResponseEqual(t *testing.T, expected, actual *Response) {
	t.Helper()
	require.Equal(t, expected.Version, actual.Version)
	require.Equal(t, expected.StatusCode, actual.StatusCode)
	require.Equal(t, expected.Headers, actual.Headers)
	require.Equal(t, expected.Body, actual.Body)
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
					"Content-Type":   JSON.HeaderValue(),
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
					"Content-Type":   TextPlain.HeaderValue(),
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
					"Content-Type": TextPlain.HeaderValue(),
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
					"Content-Type":       JSON.HeaderValue(),
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
					"Content-Type":                 JSON.HeaderValue(),
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
			assertResponseEqual(t, tt.response, parsed)
		})
	}
}
