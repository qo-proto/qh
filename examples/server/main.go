package main

import (
	"fmt"
	"log/slog"
	"os"

	"qh/internal/protocol"
	"qh/internal/server"
)

func main() {
	// TODO: fix slog output format
	// suppress qotp debug logs
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("QH Protocol Server starting")

	srv := server.NewServer()

	srv.HandleFunc("/hello", protocol.GET, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/hello")
		return server.TextResponse(200, "Hello from QH Protocol!")
	})

	srv.HandleFunc("/echo", protocol.POST, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "POST", "path", "/echo", "body", req.Body)
		response := fmt.Sprintf("Echo: %s", req.Body)
		return server.TextResponse(200, response)
	})

	srv.HandleFunc("/status", protocol.GET, func(req *protocol.Request) *protocol.Response {
		slog.Info("Handling request", "method", "GET", "path", "/status")
		return server.TextResponse(200, "QH Server is running!")
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

	// TODO: make a proper example with shutdown, etc.
}
