package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"qh/internal/protocol"
	"qh/internal/server"
)

func main() {
	slog.Info("QH Protocol Server starting")

	srv := server.NewServer()

	srv.HandleFunc("/hello", protocol.GET, func(_ *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/hello")
		return server.TextResponse(200, "Hello from QH Protocol!")
	})

	srv.HandleFunc("/echo", protocol.POST, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "POST", "path", "/echo", "body", string(req.Body))
		response := "Echo: " + string(req.Body)
		return server.TextResponse(200, response)
	})

	srv.HandleFunc("/status", protocol.GET, func(_ *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/status")
		return server.TextResponse(200, "QH Server is running!")
	})

	srv.HandleFunc("/api/user", protocol.GET, func(_ *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/api/user")
		return server.JSONResponse(200, `{"name": "John Doe", "id": 123, "active": true}`)
	})

	srv.HandleFunc("/data", protocol.POST, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "POST", "path", "/data", "body", string(req.Body))
		response := fmt.Sprintf("Updated data: %s ; %s; Hallo Welt;", string(req.Body), strings.Repeat("a", 1900))
		return server.TextResponse(200, response)
	})

	srv.HandleFunc("/large-post", protocol.POST, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling large POST request", "method", "POST", "path", "/large-post", "body_size", len(req.Body))
		response := fmt.Sprintf("Received %d bytes successfully", len(req.Body))
		return server.TextResponse(200, response)
	})

	srv.HandleFunc("/file", protocol.GET, func(_ *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/file")
		content, err := os.ReadFile("examples/server/files/text.txt")
		if err != nil {
			slog.Error("Failed to read file", "error", err)
			return server.ErrorResponse(500, "Internal Server Error")
		}
		return server.Response(200, protocol.TextPlain, content)
	})

	srv.HandleFunc("/image", protocol.GET, func(_ *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/image")
		content, err := os.ReadFile("examples/server/files/cloud.jpeg")
		if err != nil {
			slog.Error("Failed to read image", "error", err)
			return server.ErrorResponse(500, "Internal Server Error")
		}
		slog.Info("Serving image", "bytes", len(content))
		return server.Response(200, protocol.OctetStream, content)
	})

	// listening with auto-generated keys
	addr := "127.0.0.1:8090"
	// You can provide a seed for deterministic keys
	//if err := srv.Listen(addr); err != nil {
	if err := srv.Listen(addr, "my-secret-server-seed"); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	slog.Info("QH Server started", "address", addr)

	if err := srv.Serve(); err != nil {
		slog.Error("Server error", "error", err)
	}
}
