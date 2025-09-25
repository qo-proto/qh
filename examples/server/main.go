package main

import (
	"log/slog"
	"os"

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
		slog.Info("Handling request", "method", "POST", "path", "/echo", "body", req.Body)
		response := "Echo: " + req.Body
		return server.TextResponse(200, response)
	})

	srv.HandleFunc("/status", protocol.GET, func(_ *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/status")
		return server.TextResponse(200, "QH Server is running!")
	})

	srv.HandleFunc("/data", protocol.POST, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "PUT", "path", "/data", "body", req.Body)
		response := "Updated data: " + req.Body
		return server.TextResponse(200, response)
	})

	// listening with auto-generated keys
	addr := "127.0.0.1:8090"
	if err := srv.Listen(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	slog.Info("QH Server started", "address", addr)

	if err := srv.Serve(); err != nil {
		slog.Error("Server error", "error", err)
	}
}
