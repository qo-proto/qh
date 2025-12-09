package qh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerMinCompressionSize(t *testing.T) {
	tests := []struct {
		name               string
		minCompressionSize int
		responseBody       string
		shouldCompress     bool
	}{
		{
			name:               "small response below default threshold (1KB)",
			minCompressionSize: 1024,
			responseBody:       strings.Repeat("a", 500),
			shouldCompress:     false,
		},
		{
			name:               "large response above default threshold",
			minCompressionSize: 1024,
			responseBody:       strings.Repeat("a", 2000),
			shouldCompress:     true,
		},
		{
			name:               "custom threshold - small",
			minCompressionSize: 100,
			responseBody:       strings.Repeat("a", 200),
			shouldCompress:     true,
		},
		{
			name:               "custom threshold - below",
			minCompressionSize: 500,
			responseBody:       strings.Repeat("a", 400),
			shouldCompress:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(WithMinCompressionSize(tt.minCompressionSize))

			req := &Request{
				Method:  GET,
				Host:    "localhost",
				Path:    "/",
				Version: Version,
				Headers: map[string]string{
					"accept-encoding": "zstd, br, gzip",
				},
			}

			resp := TextResponse(200, tt.responseBody)
			server.applyCompression(req, resp)

			if tt.shouldCompress {
				_, hasEncoding := resp.Headers["content-encoding"]
				assert.True(t, hasEncoding, "response should be compressed")
				assert.Less(t, len(resp.Body), len(tt.responseBody), "compressed body should be smaller")
			} else {
				_, hasEncoding := resp.Headers["content-encoding"]
				assert.False(t, hasEncoding, "response should not be compressed")
				assert.Equal(t, tt.responseBody, string(resp.Body), "body should be unchanged")
			}
		})
	}
}

func TestServerSupportedEncodings(t *testing.T) {
	tests := []struct {
		name               string
		supportedEncodings []Encoding
		clientAccepts      string
		expectedEncoding   Encoding
	}{
		{
			name:               "default encodings - client prefers zstd",
			supportedEncodings: []Encoding{Zstd, Brotli, Gzip},
			clientAccepts:      "zstd, br, gzip",
			expectedEncoding:   Zstd,
		},
		{
			name:               "server only supports gzip",
			supportedEncodings: []Encoding{Gzip},
			clientAccepts:      "zstd, br, gzip",
			expectedEncoding:   Gzip,
		},
		{
			name:               "server only supports zstd",
			supportedEncodings: []Encoding{Zstd},
			clientAccepts:      "br, gzip",
			expectedEncoding:   "",
		},
		{
			name:               "custom priority order",
			supportedEncodings: []Encoding{Brotli, Gzip},
			clientAccepts:      "gzip, br, zstd",
			expectedEncoding:   Gzip, // client prefers gzip first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(
				WithSupportedEncodings(tt.supportedEncodings),
				WithMinCompressionSize(100),
			)

			req := &Request{
				Method:  GET,
				Host:    "localhost",
				Path:    "/",
				Version: Version,
				Headers: map[string]string{
					"accept-encoding": tt.clientAccepts,
				},
			}

			// Large enough to compress
			resp := TextResponse(200, strings.Repeat("test data ", 100))
			server.applyCompression(req, resp)

			if tt.expectedEncoding == "" {
				_, hasEncoding := resp.Headers["content-encoding"]
				assert.False(t, hasEncoding, "no compression should occur when no common encoding")
			} else {
				encoding, hasEncoding := resp.Headers["content-encoding"]
				assert.True(t, hasEncoding, "response should be compressed")
				assert.Equal(t, string(tt.expectedEncoding), encoding)
			}
		})
	}
}

func TestServerNoCompressionForBinary(t *testing.T) {
	server := NewServer(WithMinCompressionSize(100))

	req := &Request{
		Method:  GET,
		Host:    "localhost",
		Path:    "/",
		Version: Version,
		Headers: map[string]string{
			"accept-encoding": "zstd, br, gzip",
		},
	}

	resp := NewResponse(200, []byte(strings.Repeat("binary", 200)), map[string]string{
		"content-type": "application/octet-stream",
	})

	server.applyCompression(req, resp)

	_, hasEncoding := resp.Headers["content-encoding"]
	assert.False(t, hasEncoding, "binary content should not be compressed")
}

func TestServerRouting(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/hello", GET, func(_ *Request) *Response {
		return TextResponse(200, "Hello World")
	})

	srv.HandleFunc("/api/data", GET, func(_ *Request) *Response {
		return JSONResponse(200, `{"status":"ok"}`)
	})

	srv.HandleFunc("/api/user", POST, func(req *Request) *Response {
		return TextResponse(201, "User created: "+string(req.Body))
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	t.Run("Route to /hello", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/hello", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "Hello World", string(resp.Body))
	})

	t.Run("Route to /api/data", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/api/data", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.JSONEq(t, `{"status":"ok"}`, string(resp.Body))
	})

	t.Run("Route to /api/user with POST", func(t *testing.T) {
		resp, err := client.POST("127.0.0.1", "/api/user", []byte("John"), nil)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, "User created: John", string(resp.Body))
	})
}

func TestServerMethodMatching(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	// Register GET and POST handlers for same path
	srv.HandleFunc("/api/resource", GET, func(_ *Request) *Response {
		return TextResponse(200, "GET response")
	})

	srv.HandleFunc("/api/resource", POST, func(_ *Request) *Response {
		return TextResponse(201, "POST response")
	})

	srv.HandleFunc("/api/resource", PUT, func(_ *Request) *Response {
		return TextResponse(200, "PUT response")
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	t.Run("GET matches GET handler", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/api/resource", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "GET response", string(resp.Body))
	})

	t.Run("POST matches POST handler", func(t *testing.T) {
		resp, err := client.POST("127.0.0.1", "/api/resource", []byte("data"), nil)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, "POST response", string(resp.Body))
	})

	t.Run("PUT matches PUT handler", func(t *testing.T) {
		resp, err := client.PUT("127.0.0.1", "/api/resource", []byte("data"), nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "PUT response", string(resp.Body))
	})

	t.Run("DELETE to path with no DELETE handler returns 404", func(t *testing.T) {
		resp, err := client.DELETE("127.0.0.1", "/api/resource", nil)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestServer404Handling(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/exists", GET, func(_ *Request) *Response {
		return TextResponse(200, "Found")
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	t.Run("Unregistered path returns 404", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/does-not-exist", nil)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "Not Found")
	})

	t.Run("Similar but different path returns 404", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/exist", nil)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Path with trailing slash returns 404 if not registered", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/exists/", nil)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestServerResponseHelpers(t *testing.T) {
	t.Run("TextResponse sets correct content-type", func(t *testing.T) {
		resp := TextResponse(200, "Hello")
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "Hello", string(resp.Body))
		assert.Equal(t, "text/plain", resp.Headers["content-type"])
	})

	t.Run("JSONResponse sets correct content-type", func(t *testing.T) {
		resp := JSONResponse(201, `{"key":"value"}`)
		assert.Equal(t, 201, resp.StatusCode)
		assert.JSONEq(t, `{"key":"value"}`, string(resp.Body))
		assert.Equal(t, "application/json", resp.Headers["content-type"])
	})

	t.Run("NewResponse constructs properly", func(t *testing.T) {
		headers := map[string]string{
			"content-type":  "text/html",
			"cache-control": "max-age=3600",
		}
		resp := NewResponse(404, []byte("Not Found"), headers)
		assert.Equal(t, 404, resp.StatusCode)
		assert.Equal(t, "Not Found", string(resp.Body))
		assert.Equal(t, "text/html", resp.Headers["content-type"])
		assert.Equal(t, "max-age=3600", resp.Headers["cache-control"])
		assert.Equal(t, uint8(Version), resp.Version)
	})

	t.Run("NewResponse with nil headers", func(t *testing.T) {
		resp := NewResponse(204, nil, nil)
		assert.Equal(t, 204, resp.StatusCode)
		assert.Empty(t, resp.Body)
		assert.NotNil(t, resp.Headers)
		assert.Empty(t, resp.Headers)
	})
}
