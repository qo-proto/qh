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
		Version: "1.0",
		Headers: []string{"1", "en-US,en;q=0.5"},
		Body:    "",
	}

	expected := "example.com\x00/hello.txt\x001.0\x001\x00en-US,en;q=0.5\x03"
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestRequestFormatWithBody(t *testing.T) {
	req := &Request{
		Method:  POST,
		Host:    "example.com",
		Path:    "/submit",
		Version: "1.0",
		Headers: []string{"2"},
		Body:    `{"name": "test"}`,
	}

	expected := "example.com\x00/submit\x001.0\x002\x03{\"name\": \"test\"}"
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormat(t *testing.T) {
	resp := &Response{
		Version:    "1.0",
		StatusCode: 200,
		Headers:    []string{"1", "*", "", "1758784800"},
		Body:       "Hello, world!",
	}

	expected := "1.0\x00200\x001\x00*\x00\x001758784800\x03Hello, world!"
	actual := resp.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormatEmpty(t *testing.T) {
	resp := &Response{
		Version:    "1.0",
		StatusCode: 204,
		Headers:    []string{"0"},
		Body:       "",
	}

	expected := "1.0\x00204\x000\x03"
	require.Equal(t, expected, resp.Format())
}

func TestParseRequestBasic(t *testing.T) {
	data := "example.com\x00/hello.txt\x001.0\x001\x00en-US,en;q=0.5\x03"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/hello.txt", req.Path)
	require.Equal(t, "1.0", req.Version)
	require.Equal(t, []string{"1", "en-US,en;q=0.5"}, req.Headers)
	require.Empty(t, req.Body)
}

func TestParseRequestWithBody(t *testing.T) {
	data := "example.com\x00/submit\x001.0\x002\x03{\"name\": \"test\"}"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/submit", req.Path)
	require.Equal(t, "1.0", req.Version)
	require.Equal(t, []string{"2"}, req.Headers)
	require.JSONEq(t, `{"name": "test"}`, req.Body)
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	data := "example.com\x00/submit\x001.0\x002\x03line1\nline2\nline3"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "line1\nline2\nline3", req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	data := "example.com\x00/path\x001.0\x03test body"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method) // Has a body, so it's POST
	require.Empty(t, req.Headers)
	require.Equal(t, "test body", req.Body)
}

func TestParseRequestErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"no body separator", "example.com\x00/path\x001.0"},
		{"empty", ""},
		{"invalid request line, too few parts", "example.com"},
		{"invalid request line, too many parts", "example.com\x00/path\x001.0\x00extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequest(tt.data)
			require.Error(t, err)
		})
	}
}

func TestParseResponseBasic(t *testing.T) {
	data := "1.0\x00200\x001\x00*\x00\x001758784800\x03Hello, world!"

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, "1.0", resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"1", "*", "", "1758784800"}, resp.Headers)
	require.Equal(t, "Hello, world!", resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	data := "1.0\x00200\x001\x03Response body"

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, "1.0", resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"1"}, resp.Headers)
	require.Equal(t, "Response body", resp.Body)
}

func TestParseResponseErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"no body separator", "1.0\x00200"},
		{"empty", ""},
		{"invalid response line, too few parts", "1.0"},
		{"invalid response line, too many parts", "1.0\x00200\x00extra"},
		{"invalid status code", "1.0\x00invalid"},
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
