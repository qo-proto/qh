package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/qh-project/qh"
)

func main() {
	slog.Info("QH Protocol Server starting")

	srv := qh.NewServer()

	srv.HandleFunc("/hello", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/hello")
		return qh.TextResponse(200, "Hello from QH Protocol!")
	})

	srv.HandleFunc("/echo", qh.POST, func(req *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "POST", "path", "/echo", "body", string(req.Body))
		response := "Echo: " + string(req.Body)
		return qh.TextResponse(200, response)
	})

	srv.HandleFunc("/status", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/status")
		return qh.TextResponse(200, "QH Server is running!")
	})

	srv.HandleFunc("/api/user", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/api/user")
		headers := map[string]string{
			"Content-Type":  qh.JSON.HeaderValue(),
			"Cache-Control": "max-age=3600",
			"Date":          strconv.FormatInt(time.Now().Unix(), 10),
		}
		body := `{"name": "John Doe", "id": 123, "active": true}`
		return qh.NewResponse(200, []byte(body), headers)
	})

	srv.HandleFunc("/data", qh.POST, func(req *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "POST", "path", "/data", "body", string(req.Body))
		response := fmt.Sprintf("Updated data: %s ; %s; Hallo Welt;", string(req.Body), strings.Repeat("a", 1900))
		return qh.TextResponse(200, response)
	})

	srv.HandleFunc("/large-post", qh.POST, func(req *qh.Request) *qh.Response {
		slog.Info("Handling large POST request", "method", "POST", "path", "/large-post", "body_size", len(req.Body))
		response := fmt.Sprintf("Received %d bytes successfully", len(req.Body))
		return qh.TextResponse(200, response)
	})

	srv.HandleFunc("/file", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/file")
		content, err := os.ReadFile("examples/server/files/text.txt")
		if err != nil {
			slog.Error("Failed to read file", "error", err)
			return qh.TextResponse(500, "Internal Server Error")
		}
		headers := map[string]string{
			"Content-Type": qh.TextPlain.HeaderValue(),
		}
		return qh.NewResponse(200, content, headers)
	})

	srv.HandleFunc("/image", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/image")
		content, err := os.ReadFile("examples/server/files/cloud.jpeg")
		if err != nil {
			slog.Error("Failed to read image", "error", err)
			return qh.TextResponse(500, "Internal Server Error")
		}
		slog.Info("Serving image", "bytes", len(content))
		headers := map[string]string{
			"Content-Type": qh.OctetStream.HeaderValue(),
		}
		return qh.NewResponse(200, content, headers)
	})

	srv.HandleFunc("/redirect", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/redirect")
		headers := map[string]string{
			"host": "qh2.gianhunold.ch",
			"path": "/permanent-hello",
		}
		// A redirect response typically has an empty body.
		return qh.NewResponse(301, nil, headers)
	})

	srv.HandleFunc("/permanent-hello", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/permanent-hello")
		return qh.TextResponse(200, "Hello from the new, permanent location!")
	})

	srv.HandleFunc("/api/user", qh.PUT, func(req *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "PUT", "path", "/api/user", "body", string(req.Body))
		return qh.JSONResponse(200, `{"message": "User updated", "id": 123}`)
	})

	// listening with auto-generated keys
	addr := "127.0.0.1:8090"
	keyLogFile, err := os.OpenFile("qotp_keylog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open key log file", "error", err)
		os.Exit(1)
	}
	defer keyLogFile.Close()

	// You can provide a seed for deterministic keys
	//if err := srv.Listen(addr, keyLogFile); err != nil {
	seed := "Start123"
	if err := srv.Listen(addr, keyLogFile, seed); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	slog.Info("QH Server started", "address", addr)

	if err := srv.Serve(); err != nil {
		slog.Error("Server error", "error", err)
	}
}
