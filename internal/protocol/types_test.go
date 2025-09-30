package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestFormat(t *testing.T) {
	req := &Request{
		Method:  GET,
		Host:    "example.com",
		Path:    "/hello.txt",
		Version: 0,
		Headers: []string{"1", "en-US,en;q=0.5"},
		Body:    "",
	}

	expected := "\x00example.com\x00/hello.txt\x001\x00en-US,en;q=0.5\x03\x04" // V(00)M(000)R(000) -> 00000000 -> \x00
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestRequestFormatWithBody(t *testing.T) {
	req := &Request{
		Method:  POST,
		Host:    "example.com",
		Path:    "/submit",
		Version: 0,
		Headers: []string{"2"},
		Body:    `{"name": "test"}`,
	}

	expected := "\x08example.com\x00/submit\x002\x03{\"name\": \"test\"}\x04" // V(00)M(001)R(000) -> 00001000 -> \x08
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormat(t *testing.T) {
	resp := &Response{
		Version:    0,
		StatusCode: 200,
		Headers:    []string{"1", "*", "", "1758784800"},
		Body:       "Hello, world!",
	}

	expected := "\x00\x001\x00*\x00\x001758784800\x03Hello, world!\x04" // V(00)S(000000) -> 00000000 -> \x00
	actual := resp.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormatEmpty(t *testing.T) {
	resp := &Response{
		Version:    0,
		StatusCode: 204,
		Headers:    []string{"0"},
		Body:       "",
	}

	expected := "\x0c\x000\x03\x04" // V(00)S(001100) -> 00001100 -> \x0c
	require.Equal(t, expected, resp.Format())
}

func TestParseRequestBasic(t *testing.T) {
	data := "\x00example.com\x00/hello.txt\x001\x00en-US,en;q=0.5\x03\x04" // V(00)M(000)R(000) -> \x00

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)

	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/hello.txt", req.Path)
	require.Equal(t, uint8(0), req.Version)
	require.Equal(t, []string{"1", "en-US,en;q=0.5"}, req.Headers)
	require.Empty(t, req.Body)
}

func TestParseRequestWithBody(t *testing.T) {
	data := "\x08example.com\x00/submit\x002\x03{\"name\": \"test\"}\x04" // V(00)M(001)R(000) -> \x08

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)

	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/submit", req.Path)
	require.Equal(t, uint8(0), req.Version)
	require.Equal(t, []string{"2"}, req.Headers)
	require.JSONEq(t, `{"name": "test"}`, req.Body)
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	data := "\x08example.com\x00/submit\x03line1\nline2\nline3\x04"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "line1\nline2\nline3", req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	data := "\x08example.com\x00/path\x03test body\x04"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Empty(t, req.Headers)
	require.Equal(t, "test body", req.Body)
}

func TestParseRequestEmptyPathDefaultsToRoot(t *testing.T) {
	data := "\x00example.com\x00\x03\x04"

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
		data string
	}{
		{"no body separator", "example.com\x00/path\x001.0"},
		{"empty", ""},
		{"invalid request line, too few parts", "\x00example.com"},
		{"invalid request line, too few parts with separator", "\x00example.com\x03"},
		{"host missing", "\x00\x00/path\x03"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequest(tt.data)
			require.Error(t, err)
		})
	}
}

func TestParseResponseBasic(t *testing.T) {
	data := "\x00\x001\x00*\x00\x001758784800\x03Hello, world!\x04" // V(00)S(000000) -> \x00

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, uint8(0), resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"1", "*", "", "1758784800"}, resp.Headers)
	require.Equal(t, "Hello, world!", resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	data := "\x00\x001\x03Response body\x04" // V(00)S(000000) -> \x00

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, uint8(0), resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"1"}, resp.Headers)
	require.Equal(t, "Response body", resp.Body)
}

func TestParseResponseErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"no body separator", "\x00"},
		{"empty", ""},
		{"invalid response line, only status/version byte", "\x00\x03"},
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
