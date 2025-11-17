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
		name    string
		request *Request
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
		},
		{
			name: "GET with Accept header",
			request: &Request{
				Method:  GET,
				Host:    "example.com",
				Path:    "/api",
				Version: 0,
				Headers: map[string]string{
					"accept": JSON.HeaderValue(),
				},
			},
		},
		{
			name: "POST with body",
			request: &Request{
				Method:  POST,
				Host:    "api.example.com",
				Path:    "/submit",
				Version: 0,
				Headers: map[string]string{
					"content-type": JSON.HeaderValue(),
				},
				Body: []byte(`{"key":"val"}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wire := tt.request.Format()
			parsed, err := ParseRequest(wire)
			require.NoError(t, err, "Failed to parse formatted request")

			// Verify round-trip
			require.Equal(t, tt.request.Method, parsed.Method)
			require.Equal(t, tt.request.Host, parsed.Host)
			require.Equal(t, tt.request.Path, parsed.Path)
			require.Equal(t, tt.request.Version, parsed.Version)
			require.Equal(t, tt.request.Headers, parsed.Headers)

			expectedBody := tt.request.Body
			if expectedBody == nil {
				expectedBody = []byte{}
			}
			require.Equal(t, expectedBody, parsed.Body)
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
	}{
		{
			name: "200 OK minimal",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers:    map[string]string{},
				Body:       []byte("OK"),
			},
		},
		{
			name: "200 OK with Content-Type",
			response: &Response{
				Version:    0,
				StatusCode: 200,
				Headers: map[string]string{
					"content-type": TextPlain.HeaderValue(),
				},
				Body: []byte("Hello"),
			},
		},
		{
			name: "404 Not Found",
			response: &Response{
				Version:    0,
				StatusCode: 404,
				Headers: map[string]string{
					"content-type": TextPlain.HeaderValue(),
				},
				Body: []byte("Not Found"),
			},
		},
		{
			name: "204 No Content empty",
			response: &Response{
				Version:    0,
				StatusCode: 204,
				Headers: map[string]string{
					"content-type": Custom.HeaderValue(),
				},
				Body: []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wire := tt.response.Format()
			parsed, err := ParseResponse(wire)
			require.NoError(t, err, "Failed to parse formatted response")
			require.Equal(t, tt.response.Version, parsed.Version)
			require.Equal(t, tt.response.StatusCode, parsed.StatusCode)
			require.Equal(t, tt.response.Headers, parsed.Headers)
			require.Equal(t, tt.response.Body, parsed.Body)
		})
	}
}

func TestParseRequestBasic(t *testing.T) {
	original := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/hello.txt",
		Version: 0,
		Headers: map[string]string{
			"accept":          "1",
			"accept-language": "en-US,en;q=0.5",
		},
		Body: []byte{},
	}

	data := original.Format()
	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, original.Method, req.Method)
	require.Equal(t, original.Host, req.Host)
	require.Equal(t, original.Path, req.Path)
	require.Equal(t, original.Version, req.Version)
	require.Equal(t, original.Headers, req.Headers)
	require.Equal(t, original.Body, req.Body)
}

func TestParseRequestWithBody(t *testing.T) {
	original := &Request{
		Method:  POST,
		Host:    "example.com",
		Path:    "/submit",
		Version: 0,
		Headers: map[string]string{
			"content-type":   JSON.HeaderValue(),
			"content-length": "16",
		},
		Body: []byte(`{"name": "test"}`),
	}

	data := original.Format()
	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, original.Method, req.Method)
	require.Equal(t, original.Host, req.Host)
	require.Equal(t, original.Path, req.Path)
	require.Equal(t, original.Version, req.Version)
	require.Equal(t, original.Headers, req.Headers)
	require.JSONEq(t, `{"name": "test"}`, string(req.Body))
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	original := &Request{
		Method:  POST,
		Host:    "example.com",
		Path:    "/submit",
		Version: 0,
		Headers: map[string]string{},
		Body:    []byte("line1\nline2\nline3"),
	}

	data := original.Format()
	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, original.Method, req.Method)
	require.Equal(t, original.Body, req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	original := &Request{
		Method:  POST,
		Host:    "example.com",
		Path:    "/path",
		Version: 0,
		Headers: map[string]string{},
		Body:    []byte("test body"),
	}

	data := original.Format()
	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, original.Method, req.Method)
	require.Empty(t, req.Headers)
	require.Equal(t, original.Body, req.Body)
}

func TestParseRequestEmptyPathDefaultsToRoot(t *testing.T) {
	original := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "",
		Version: 0,
		Headers: map[string]string{},
		Body:    []byte{},
	}

	data := original.Format()
	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/", req.Path) // empty path should default to "/"
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
	original := &Response{
		Version:    0,
		StatusCode: 200,
		Headers: map[string]string{
			"content-type":   TextPlain.HeaderValue(),
			"content-length": "13",
			"date":           "1758784800",
		},
		Body: []byte("Hello, world!"),
	}

	data := original.Format()
	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, original.Version, resp.Version)
	require.Equal(t, original.StatusCode, resp.StatusCode)
	require.Equal(t, original.Headers, resp.Headers)
	require.Equal(t, original.Body, resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	original := &Response{
		Version:    0,
		StatusCode: 200,
		Headers: map[string]string{
			"content-type": TextPlain.HeaderValue(),
		},
		Body: []byte("Response body"),
	}

	data := original.Format()
	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, original.Version, resp.Version)
	require.Equal(t, original.StatusCode, resp.StatusCode)
	require.Equal(t, original.Headers, resp.Headers)
	require.Equal(t, original.Body, resp.Body)
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
					"accept":          AcceptHeader(JSON, TextPlain),
					"accept-encoding": "gzip, br",
					"user-agent":      "QH-Client/1.0",
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
					"content-type":   JSON.HeaderValue(),
					"content-length": "15",
					"authorization":  "Bearer token123",
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
					"content-type":   JSON.HeaderValue(),
					"content-length": "18",
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
					"content-type":   JSON.HeaderValue(),
					"content-length": "12",
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
					"accept": AcceptHeader(JSON, TextPlain),
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
					"content-type":     TextPlain.HeaderValue(),
					"content-length":   "5",
					"x-custom-header":  "custom-value",
					"x-another-custom": "another-value",
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
					"content-type":   JSON.HeaderValue(),
					"content-length": "15",
					"cache-control":  "max-age=3600",
					"date":           "1758784800",
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
					"content-type":   TextPlain.HeaderValue(),
					"content-length": "9",
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
					"content-type": TextPlain.HeaderValue(),
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
					"content-type":       JSON.HeaderValue(),
					"content-length":     "2",
					"x-custom-response":  "custom-value",
					"x-another-response": "another-value",
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
					"content-type":                 JSON.HeaderValue(),
					"content-length":               "2",
					"access-control-allow-origin":  "*",
					"access-control-allow-methods": "GET, POST, PUT",
					"access-control-allow-headers": "Content-Type, Authorization",
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
