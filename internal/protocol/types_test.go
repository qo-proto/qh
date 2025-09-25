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
		Headers: []string{"text/plain", "en-US,en;q=0.5"},
		Body:    "",
	}

	expected := "1 example.com /hello.txt 1.0\ntext/plain\nen-US,en;q=0.5\n\n"
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestRequestFormatWithBody(t *testing.T) {
	req := &Request{
		Method:  POST,
		Host:    "example.com",
		Path:    "/submit",
		Version: "1.0",
		Headers: []string{"application/json"},
		Body:    `{"name": "test"}`,
	}

	expected := "2 example.com /submit 1.0\napplication/json\n\n{\"name\": \"test\"}"
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormat(t *testing.T) {
	resp := &Response{
		Version:    "1.0",
		StatusCode: 200,
		Headers:    []string{"*", "", "text/plain", "Mon, 17 Sep 2025 10:00:00 CET"},
		Body:       "Hello, world!",
	}

	expected := "1.0 200\n*\n\ntext/plain\nMon, 17 Sep 2025 10:00:00 CET\n\nHello, world!"
	actual := resp.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormatEmpty(t *testing.T) {
	resp := &Response{
		Version:    "1.0",
		StatusCode: 204,
		Headers:    []string{},
		Body:       "",
	}

	expected := "1.0 204\n\n"
	require.Equal(t, expected, resp.Format())
}

func TestParseRequestBasic(t *testing.T) {
	data := "1 example.com /hello.txt 1.0\ntext/plain\nen-US,en;q=0.5\n\n"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/hello.txt", req.Path)
	require.Equal(t, "1.0", req.Version)
	require.Equal(t, []string{"text/plain", "en-US,en;q=0.5"}, req.Headers)
	require.Empty(t, req.Body)
}

func TestParseRequestWithBody(t *testing.T) {
	data := "2 example.com /submit 1.0\napplication/json\n\n{\"name\": \"test\"}"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/submit", req.Path)
	require.Equal(t, "1.0", req.Version)
	require.Equal(t, []string{"application/json"}, req.Headers)
	require.JSONEq(t, `{"name": "test"}`, req.Body)
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	data := "2 example.com /submit 1.0\napplication/json\n\nline1\nline2\nline3"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, POST, req.Method)
	require.Equal(t, "line1\nline2\nline3", req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	data := "1 example.com /path 1.0\n\ntest body"

	req, err := ParseRequest(data)
	require.NoError(t, err)
	require.Equal(t, GET, req.Method)
	require.Empty(t, req.Headers)
	require.Equal(t, "test body", req.Body)
}

func TestParseRequestErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"invalid request line, too few parts", "1 example.com"},
		{"invalid request line, too many parts", "1 example.com /path 1.0 extra"},
		{"invalid method", "GET example.com /path 1.0"}, // method should be an integer
		{"invalid method zero", "0 example.com /path 1.0\n\n"},
		{"invalid method negative", "-1 example.com /path 1.0\n\n"},
		{"invalid method too large", "999 example.com /path 1.0\n\n"},
		{"empty host", "1  /path 1.0\n\n"},
		{"empty path", "1 example.com  1.0\n\n"},
		{"empty version", "1 example.com /path \n\n"},
		{"whitespace in host", "1 exam ple.com /path 1.0\n\n"},
		{"missing empty line separator", "1 example.com /path 1.0\nheader1\nheader2"},
		{"missing empty line with body", "1 example.com /path 1.0\nheader1\nbody content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequest(tt.data)
			require.Error(t, err)
		})
	}
}

func TestParseResponseBasic(t *testing.T) {
	data := "1.0 200\n*\n\nHello, world!"

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, "1.0", resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"*"}, resp.Headers)
	require.Equal(t, "Hello, world!", resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	data := "1.0 200\ntext/plain\n\nResponse body"

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, "1.0", resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"text/plain"}, resp.Headers)
	require.Equal(t, "Response body", resp.Body)
}

func TestParseResponseErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"invalid response line, too few parts", "1.0"},
		{"invalid response line, too many parts", "1.0 200 extra"},
		{"invalid status code", "1.0 invalid"},
		{"empty version", " 200\n\n"},
		{"whitespace in version", "1. 0 200\n\n"},
		{"missing empty line separator", "1.0 200\nheader1\nheader2"},
		{"missing empty line with body", "1.0 200\nheader1\nbody content"},
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
