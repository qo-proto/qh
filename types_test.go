package qh

import (
	"encoding/base64"
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
					"Accept": JSON.HeaderValue(),
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
			"Content-Type": JSON.HeaderValue(),
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
					"Content-Type": TextPlain.HeaderValue(),
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
					"Content-Type": TextPlain.HeaderValue(),
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
			"Content-Type": Custom.HeaderValue(),
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
	require.Equal(t, JSON.HeaderValue(), req.Headers["Content-Type"])
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
	require.Equal(t, TextPlain.HeaderValue(), resp.Headers["Content-Type"])
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
			require.Equal(t, tt.response.Version, parsed.Version)
			require.Equal(t, tt.response.StatusCode, parsed.StatusCode)
			require.Equal(t, tt.response.Headers, parsed.Headers)
			require.Equal(t, tt.response.Body, parsed.Body)
		})
	}
}

// bytes.Cut splits at the first \x03, so any \x03 bytes in the body are preserved
func TestETXWithBinaryData(t *testing.T) {
	t.Run("Response with ETX byte in body", func(t *testing.T) {
		binaryBody := []byte{0xAB, 0xCD, 0x03, 0xEF, 0x12, 0x34}
		resp := &Response{
			Version:    0,
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":     OctetStream.HeaderValue(),
				"Content-Encoding": "zstd",
				"Content-Length":   "6",
			},
			Body: binaryBody,
		}
		wireFormat := resp.Format()
		parsed, err := ParseResponse(wireFormat)
		require.NoError(t, err)
		require.Equal(t, binaryBody, parsed.Body, "Body should be preserved completely, including \\x03 bytes")
	})

	t.Run("Request with ETX byte in compressed body", func(t *testing.T) {
		compressedBody := []byte{0x28, 0xB5, 0x2F, 0xFD, 0x03, 0x00, 0x59, 0x00}
		req := &Request{
			Method:  POST,
			Host:    "api.example.com",
			Path:    "/submit",
			Version: 0,
			Headers: map[string]string{
				"Content-Type":     JSON.HeaderValue(),
				"Content-Encoding": "zstd",
				"Content-Length":   "8",
			},
			Body: compressedBody,
		}

		wireFormat := req.Format()
		parsed, err := ParseRequest(wireFormat)
		require.NoError(t, err)
		require.Equal(t, compressedBody, parsed.Body, "Compressed body should be preserved with \\x03 bytes intact")
	})

	t.Run("Real gzip data with ETX byte", func(t *testing.T) {
		// \x03 at position 9
		gzipData := []byte{
			0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x03, 0xf3, 0x48, 0xcd, 0xc9, 0xc9, 0x57,
			0x08, 0xcf, 0x2f, 0xca, 0x49, 0x01, 0x00, 0x85,
			0x11, 0x4a, 0x0d, 0x0c, 0x00, 0x00, 0x00,
		}
		resp := &Response{
			Version:    0,
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":     TextPlain.HeaderValue(),
				"Content-Encoding": "gzip",
			},
			Body: gzipData,
		}
		wireFormat := resp.Format()
		parsed, err := ParseResponse(wireFormat)
		require.NoError(t, err)
		require.Equal(t, gzipData, parsed.Body, "Real gzip data should be preserved completely, including \\x03 byte at position 9")
	})
}

func TestRequestValidation(t *testing.T) {
	t.Run("Valid request passes", func(t *testing.T) {
		req := &Request{
			Method:  GET,
			Host:    "example.com",
			Path:    "/path",
			Version: 0,
			Headers: map[string]string{
				"Accept":       AcceptHeader(JSON, TextPlain),
				"User-Agent":   "QH-Client/1.0",
				"X-Custom-Key": "custom-value",
			},
		}
		err := req.Validate()
		require.NoError(t, err)
	})

	t.Run("Host with control characters rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			host      string
			errSubstr string
		}{
			{"null byte", "example\x00.com", "null byte"},
			{"ETX byte", "example\x03.com", "ETX"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := &Request{
					Method:  GET,
					Host:    tt.host,
					Path:    "/path",
					Version: 0,
				}
				err := req.Validate()
				require.Contains(t, err.Error(), "invalid host")
				require.Contains(t, err.Error(), tt.errSubstr)
			})
		}
	})

	t.Run("Path with control characters rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			path      string
			errSubstr string
		}{
			{"null byte", "/path\x00/sub", "null byte"},
			{"ETX byte", "/path\x03/sub", "ETX"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := &Request{
					Method:  GET,
					Host:    "example.com",
					Path:    tt.path,
					Version: 0,
				}
				err := req.Validate()
				require.Contains(t, err.Error(), "invalid path")
				require.Contains(t, err.Error(), tt.errSubstr)
			})
		}
	})

	t.Run("Predefined header value with control character rejected", func(t *testing.T) {
		req := &Request{
			Method:  GET,
			Host:    "example.com",
			Path:    "/path",
			Version: 0,
			Headers: map[string]string{
				"User-Agent": "QH\x03Client",
			},
		}
		err := req.Validate()
		require.Contains(t, err.Error(), "User-Agent")
		require.Contains(t, err.Error(), "ETX")
	})

	t.Run("Custom header key with control character rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			key       string
			errSubstr string
		}{
			{"null byte", "X-Custom\x00Key", "null byte"},
			{"ETX byte", "X-Custom\x03Key", "ETX"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := &Request{
					Method:  GET,
					Host:    "example.com",
					Path:    "/path",
					Version: 0,
					Headers: map[string]string{
						tt.key: "value",
					},
				}
				err := req.Validate()
				require.Contains(t, err.Error(), "invalid header key")
				require.Contains(t, err.Error(), tt.errSubstr)
			})
		}
	})

	t.Run("Custom header value with control character rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			value     string
			errSubstr string
		}{
			{"null byte", "value\x00data", "null byte"},
			{"ETX byte", "value\x03data", "ETX"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := &Request{
					Method:  GET,
					Host:    "example.com",
					Path:    "/path",
					Version: 0,
					Headers: map[string]string{
						"X-Custom": tt.value,
					},
				}
				err := req.Validate()
				require.Contains(t, err.Error(), "X-Custom")
				require.Contains(t, err.Error(), tt.errSubstr)
			})
		}
	})
}

func TestResponseValidation(t *testing.T) {
	t.Run("Valid response passes", func(t *testing.T) {
		resp := &Response{
			Version:    0,
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":      JSON.HeaderValue(),
				"Cache-Control":     "max-age=3600",
				"X-Custom-Response": "custom-value",
			},
			Body: []byte("test body"),
		}

		err := resp.Validate()
		require.NoError(t, err)
	})

	t.Run("Header value with control character rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			header    map[string]string
			errHeader string
			errSubstr string
		}{
			{"null byte", map[string]string{"Cache-Control": "max-age\x00=3600"}, "Cache-Control", "null byte"},
			{"ETX byte", map[string]string{"Content-Type": JSON.HeaderValue() + "\x03"}, "Content-Type", "ETX"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp := &Response{
					Version:    0,
					StatusCode: 200,
					Headers:    tt.header,
				}
				err := resp.Validate()
				require.Contains(t, err.Error(), tt.errHeader)
				require.Contains(t, err.Error(), tt.errSubstr)
			})
		}
	})

	t.Run("Custom header key with control character rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			key       string
			errSubstr string
		}{
			{"null byte", "X-Bad\x00Key", "null byte"},
			{"ETX byte", "X-Bad\x03Key", "ETX"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp := &Response{
					Version:    0,
					StatusCode: 200,
					Headers: map[string]string{
						tt.key: "value",
					},
				}
				err := resp.Validate()
				require.Contains(t, err.Error(), "invalid header key")
				require.Contains(t, err.Error(), tt.errSubstr)
			})
		}
	})

	t.Run("Body can contain control characters", func(t *testing.T) {
		resp := &Response{
			Version:    0,
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": OctetStream.HeaderValue(),
			},
			Body: []byte{0x00, 0x01, 0x03, 0xFF},
		}

		err := resp.Validate()
		require.NoError(t, err, "Body should be allowed to contain any bytes")
	})
}

func TestWireFormatCorruption(t *testing.T) {
	// Without validation: ETX in headers corrupts wire format
	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/test",
		Headers: map[string]string{"X-Data": "before\x03after"},
	}

	wireFormat := req.Format()
	parsed, _ := ParseRequest(wireFormat)
	require.NotEqual(t, "before\x03after", parsed.Headers["X-Data"])
	require.Equal(t, "before", parsed.Headers["X-Data"], "Truncated at ETX byte")

	// With validation: Error before corruption
	err := req.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ETX")

	// Fix with base64 encoding
	req.Headers = make(map[string]string)
	req.Headers["Content-Type"] = TextPlain.HeaderValue()
	encoded := base64.StdEncoding.EncodeToString([]byte("before\x03after"))
	req.Headers["X-Data"] = encoded
	require.NoError(t, req.Validate())

	// Round-trip preserves data
	wireFormat = req.Format()
	parsed, _ = ParseRequest(wireFormat)
	data, _ := base64.StdEncoding.DecodeString(parsed.Headers["X-Data"])
	require.Equal(t, []byte("before\x03after"), data)
}
