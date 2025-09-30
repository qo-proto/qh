package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestFormat(t *testing.T) {
	req := &Request{
		Host:    "example.com",
		Path:    "/hello.txt",
		Version: "1.0",
		Headers: []string{"1", "en-US,en;q=0.5"},
		Body:    []byte(""),
	}

	expected := []byte("example.com\x00/hello.txt\x001.0\x001\x00en-US,en;q=0.5\x03\x04")
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestRequestFormatWithBody(t *testing.T) {
	req := &Request{
		Host:    "example.com",
		Path:    "/submit",
		Version: "1.0",
		Headers: []string{"2"},
		Body:    []byte(`{"name": "test"}`),
	}

	expected := []byte("example.com\x00/submit\x001.0\x002\x03{\"name\": \"test\"}\x04")
	actual := req.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormat(t *testing.T) {
	resp := &Response{
		Version:    "1.0",
		StatusCode: 200,
		Headers:    []string{"1", "*", "", "1758784800"},
		Body:       []byte("Hello, world!"),
	}

	expected := []byte("1.0\x001\x001\x00*\x00\x001758784800\x03Hello, world!\x04")
	actual := resp.Format()

	require.Equal(t, expected, actual)
}

func TestResponseFormatEmpty(t *testing.T) {
	resp := &Response{
		Version:    "1.0",
		StatusCode: 204,
		Headers:    []string{"0"},
		Body:       []byte(""),
	}

	expected := []byte("1.0\x0013\x000\x03\x04")
	require.Equal(t, expected, resp.Format())
}

func TestParseRequestBasic(t *testing.T) {
	data := []byte("example.com\x00/hello.txt\x001.0\x001\x00en-US,en;q=0.5\x03\x04")

	req, err := ParseRequest(data)
	require.NoError(t, err)

	// Method is inferred, not stored. We can check the body to confirm.
	require.Empty(t, req.Body, "A request with an empty body should be treated as GET")

	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/hello.txt", req.Path)
	require.Equal(t, "1.0", req.Version)
	require.Equal(t, []string{"1", "en-US,en;q=0.5"}, req.Headers)
	require.Empty(t, req.Body)
}

func TestParseRequestWithBody(t *testing.T) {
	data := []byte("example.com\x00/submit\x001.0\x002\x03{\"name\": \"test\"}\x04")

	req, err := ParseRequest(data)
	require.NoError(t, err)

	// Method is inferred, not stored. We can check the body to confirm.
	require.NotEmpty(t, req.Body, "A request with a non-empty body should be treated as POST")

	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/submit", req.Path)
	require.Equal(t, "1.0", req.Version)
	require.Equal(t, []string{"2"}, req.Headers)
	require.JSONEq(t, `{"name": "test"}`, string(req.Body))
}

func TestParseRequestWithMultilineBody(t *testing.T) {
	data := []byte("example.com\x00/submit\x001.0\x002\x03line1\nline2\nline3\x04")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	// Method is inferred, not stored. We can check the body to confirm.
	require.NotEmpty(t, req.Body, "A request with a non-empty body should be treated as POST")
	require.Equal(t, []byte("line1\nline2\nline3"), req.Body)
}

func TestParseRequestNoHeaders(t *testing.T) {
	data := []byte("example.com\x00/path\x001.0\x03test body\x04")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	// Method is inferred, not stored. We can check the body to confirm.
	require.NotEmpty(t, req.Body, "A request with a non-empty body should be treated as POST")
	require.Empty(t, req.Headers)
	require.Equal(t, []byte("test body"), req.Body)
}

func TestParseRequestEmptyPathDefaultsToRoot(t *testing.T) {
	data := []byte("example.com\x00\x001.0\x03\x04")

	req, err := ParseRequest(data)
	require.NoError(t, err)
	// Method is inferred, not stored. We can check the body to confirm.
	require.Empty(t, req.Body, "A request with an empty body should be treated as GET")
	require.Equal(t, "example.com", req.Host)
	require.Equal(t, "/", req.Path) // Empty path should default to "/"
	require.Equal(t, "1.0", req.Version)
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
		{"invalid request line, too few parts", []byte("example.com")},
		{"invalid request line, too few parts with separator", []byte("example.com\x00/path")},
		{"host missing", []byte("\x00/path\x001.0\x03")},
		{"version missing", []byte("example.com\x00/path\x00\x03")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequest(tt.data)
			require.Error(t, err)
		})
	}
}

func TestParseResponseBasic(t *testing.T) {
	data := []byte("1.0\x001\x001\x00*\x00\x001758784800\x03Hello, world!\x04")

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, "1.0", resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"1", "*", "", "1758784800"}, resp.Headers)
	require.Equal(t, []byte("Hello, world!"), resp.Body)
}

func TestParseResponseSingleHeader(t *testing.T) {
	data := []byte("1.0\x001\x001\x03Response body\x04")

	resp, err := ParseResponse(data)
	require.NoError(t, err)
	require.Equal(t, "1.0", resp.Version)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"1"}, resp.Headers)
	require.Equal(t, []byte("Response body"), resp.Body)
}

func TestParseResponseErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"no body separator", []byte("1.0\x00200")},
		{"empty", []byte("")},
		{"invalid response line, too few parts", []byte("1.0")},
		{"invalid response line, only version", []byte("1.0\x03")},
		// {"invalid status code", []byte("1.0\x00invalid\x03")},
		// {"negative status code", []byte("1.0\x00-200\x03")},
		{"version missing", []byte("\x00200\x03")},
		{"status code missing", []byte("1.0\x00\x03")},
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
