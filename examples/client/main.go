package main

import (
	"log/slog"

	"qh/internal/client"
	"qh/internal/protocol"
)

func main() {
	slog.Info("QH Protocol Client starting")

	addr := "127.0.0.1:8090"

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/hello", ""},
		{"GET", "/status", ""},
		{"POST", "/echo", "Hello QH World!"},
		{"PUT", "/data", "Updated resource data"},
		{"DELETE", "/data", ""},
		{"HEAD", "/info", ""},
	}

	c := client.NewClient()
	defer c.Close()

	if err := c.Connect(addr); err != nil {
		slog.Error("Failed to connect", "error", err)
		return
	}

	for _, req := range requests {
		slog.Info("Testing request", "method", req.method, "path", req.path)

		var response *protocol.Response
		var err error
		switch req.method {
		case "GET":
			response, err = c.GET("127.0.0.1", req.path, "text/plain")
		case "POST":
			response, err = c.POST("127.0.0.1", req.path, req.body, "text/plain")
		case "PUT":
			response, err = c.PUT("127.0.0.1", req.path, req.body, "text/plain")
		case "DELETE":
			response, err = c.DELETE("127.0.0.1", req.path, "text/plain")
		case "HEAD":
			response, err = c.HEAD("127.0.0.1", req.path, "text/plain")
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
		} else {
			slog.Info("Request success", "method", req.method, "path", req.path, "status", response.StatusCode, "body", response.Body)
		}
	}

	slog.Info("All tests completed")
}
