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
	}

	// TODO: currently using new connections for each request, until multiplexing correctly implemented
	for _, req := range requests {
		slog.Info("Testing request", "method", req.method, "path", req.path)

		c := client.NewClient()
		// TODO: check return value
		defer c.Close()

		if err := c.Connect(addr); err != nil {
			slog.Error("Failed to connect", "error", err)
			continue
		}

		var response *protocol.Response
		var err error
		if req.method == "GET" {
			response, err = c.GET("127.0.0.1", req.path, "text/plain")
		} else if req.method == "POST" {
			response, err = c.POST("127.0.0.1", req.path, req.body, "text/plain")
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
		} else {
			slog.Info("Request success", "method", req.method, "path", req.path, "status", response.StatusCode, "body", response.Body)
		}
	}

	slog.Info("All tests completed")
}
