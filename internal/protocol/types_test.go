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

// TODO: add tests for ParseRequest and ParseResponse
