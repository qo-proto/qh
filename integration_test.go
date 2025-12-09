package qh

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAllHTTPMethods(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/resource", GET, func(_ *Request) *Response {
		return TextResponse(200, "GET OK")
	})
	srv.HandleFunc("/resource", POST, func(req *Request) *Response {
		return TextResponse(201, "POST: "+string(req.Body))
	})
	srv.HandleFunc("/resource", PUT, func(req *Request) *Response {
		return TextResponse(200, "PUT: "+string(req.Body))
	})
	srv.HandleFunc("/resource", PATCH, func(req *Request) *Response {
		return TextResponse(200, "PATCH: "+string(req.Body))
	})
	srv.HandleFunc("/resource", DELETE, func(_ *Request) *Response {
		return NewResponse(204, nil, nil)
	})
	srv.HandleFunc("/resource", HEAD, func(_ *Request) *Response {
		return NewResponse(200, nil, map[string]string{
			"content-type":   "text/plain",
			"content-length": "6",
		})
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	tests := []struct {
		name           string
		method         string
		body           []byte
		expectedStatus int
		expectedBody   string
	}{
		{"GET", "GET", nil, 200, "GET OK"},
		{"POST", "POST", []byte("test data"), 201, "POST: test data"},
		{"PUT", "PUT", []byte("updated"), 200, "PUT: updated"},
		{"PATCH", "PATCH", []byte("partial"), 200, "PATCH: partial"},
		{"DELETE", "DELETE", nil, 204, ""},
		{"HEAD", "HEAD", nil, 200, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *Response
			var err error

			switch tt.method {
			case "GET":
				resp, err = client.GET("127.0.0.1", "/resource", nil)
			case "POST":
				resp, err = client.POST("127.0.0.1", "/resource", tt.body, nil)
			case "PUT":
				resp, err = client.PUT("127.0.0.1", "/resource", tt.body, nil)
			case "PATCH":
				resp, err = client.PATCH("127.0.0.1", "/resource", tt.body, nil)
			case "DELETE":
				resp, err = client.DELETE("127.0.0.1", "/resource", nil)
			case "HEAD":
				resp, err = client.HEAD("127.0.0.1", "/resource", nil)
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, string(resp.Body))
		})
	}
}

func TestIntegrationLargeRequestBody(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/upload", POST, func(req *Request) *Response {
		return TextResponse(200, strings.Repeat("X", len(req.Body)))
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	sizes := []int{1024, 10 * 1024, 100 * 1024, 500 * 1024} // 1KB, 10KB, 100KB, 500KB
	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dB", size), func(t *testing.T) {
			largeBody := []byte(strings.Repeat("A", size))
			resp, err := client.POST("127.0.0.1", "/upload", largeBody, nil)
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			assert.Len(t, resp.Body, size, "Response should echo same size")
		})
	}
}

func TestIntegrationLargeResponseBody(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	responseSize := 500 * 1024 // 500KB
	largeResponse := strings.Repeat("R", responseSize)
	srv.HandleFunc("/large", GET, func(_ *Request) *Response {
		return TextResponse(200, largeResponse)
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	resp, err := client.GET("127.0.0.1", "/large", nil)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, resp.Body, responseSize)
}

func TestIntegrationCompressionNegotiation(t *testing.T) {
	srv, addr := newTestServer(t,
		WithMinCompressionSize(100),
		WithSupportedEncodings([]Encoding{Brotli}),
	)
	defer srv.Close()

	responseBody := strings.Repeat("compressible data ", 100)
	srv.HandleFunc("/compress", GET, func(_ *Request) *Response {
		return TextResponse(200, responseBody)
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	resp, err := client.GET("127.0.0.1", "/compress", nil)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, responseBody, string(resp.Body))
}

func TestIntegrationHeaderHandling(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/headers", GET, func(req *Request) *Response {
		respHeaders := map[string]string{
			"content-type":    "application/json",
			"x-custom-header": req.Headers["x-custom-header"],
			"user-agent":      req.Headers["user-agent"],
			"cache-control":   "max-age=3600",
			"x-request-id":    "test-123",
			"x-server":        "qh-test",
		}
		return NewResponse(200, []byte(`{"status":"ok"}`), respHeaders)
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	t.Run("Request headers echoed back", func(t *testing.T) {
		headers := map[string]string{
			"x-custom-header": "custom-value-123",
			"user-agent":      "QH-Test-Client/1.0",
		}
		resp, err := client.GET("127.0.0.1", "/headers", headers)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "custom-value-123", resp.Headers["x-custom-header"])
		assert.Equal(t, "QH-Test-Client/1.0", resp.Headers["user-agent"])
	})

	t.Run("Multiple response headers", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/headers", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Headers["content-type"])
		assert.Equal(t, "max-age=3600", resp.Headers["cache-control"])
		assert.Equal(t, "test-123", resp.Headers["x-request-id"])
		assert.Equal(t, "qh-test", resp.Headers["x-server"])
	})
}

func TestIntegrationErrorResponses(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/bad-request", POST, func(_ *Request) *Response {
		return TextResponse(400, "Invalid request format")
	})
	srv.HandleFunc("/forbidden", GET, func(_ *Request) *Response {
		return TextResponse(403, "Access denied")
	})
	srv.HandleFunc("/error", GET, func(_ *Request) *Response {
		return TextResponse(500, "Internal server error")
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		{"404 Not Found", "GET", "/nonexistent", 404, "Not Found"},
		{"400 Bad Request", "POST", "/bad-request", 400, "Invalid request format"},
		{"403 Forbidden", "GET", "/forbidden", 403, "Access denied"},
		{"500 Internal Error", "GET", "/error", 500, "Internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *Response
			var err error
			switch tt.method {
			case "GET":
				resp, err = client.GET("127.0.0.1", tt.path, nil)
			case "POST":
				resp, err = client.POST("127.0.0.1", tt.path, nil, nil)
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Contains(t, string(resp.Body), tt.expectedBody)
		})
	}
}

func TestIntegrationEmptyBody(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	srv.HandleFunc("/empty", GET, func(_ *Request) *Response {
		return NewResponse(204, nil, nil)
	})
	srv.HandleFunc("/empty", POST, func(req *Request) *Response {
		if len(req.Body) == 0 {
			return TextResponse(200, "empty body received")
		}
		return TextResponse(200, "body: "+string(req.Body))
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	t.Run("GET empty response", func(t *testing.T) {
		resp, err := client.GET("127.0.0.1", "/empty", nil)
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
		assert.Empty(t, resp.Body)
	})

	t.Run("POST empty body", func(t *testing.T) {
		resp, err := client.POST("127.0.0.1", "/empty", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "empty body received", string(resp.Body))
	})
}

func TestIntegrationBinaryData(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	binaryData := make([]byte, 256)
	for i := range binaryData {
		binaryData[i] = byte(i)
	}

	srv.HandleFunc("/binary", POST, func(req *Request) *Response {
		return NewResponse(200, req.Body, map[string]string{
			"content-type": "application/octet-stream",
		})
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	headers := map[string]string{"content-type": "application/octet-stream"}
	resp, err := client.POST("127.0.0.1", "/binary", binaryData, headers)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, binaryData, resp.Body)
}

func TestIntegrationConnectionReuse(t *testing.T) {
	srv, addr := newTestServer(t)
	defer srv.Close()

	callCount := 0
	srv.HandleFunc("/reuse", GET, func(_ *Request) *Response {
		callCount++
		return TextResponse(200, "OK")
	})

	client := NewClient()
	defer client.Close()
	require.NoError(t, client.Connect(addr, nil))

	connBefore := client.conn
	initialStreamID := client.streamID.Load()

	numRequests := 20
	for i := range numRequests {
		resp, err := client.GET("127.0.0.1", "/reuse", nil)
		require.NoError(t, err, "Request %d should succeed", i)
		assert.Equal(t, 200, resp.StatusCode)
	}

	assert.Equal(t, numRequests, callCount, "All requests should reach the server")
	assert.Same(t, connBefore, client.conn, "Should reuse same connection")

	finalStreamID := client.streamID.Load()
	assert.Equal(t, uint32(numRequests), finalStreamID-initialStreamID,
		"Stream IDs should increment sequentially on reused connection")
}
