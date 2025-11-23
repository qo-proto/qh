package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/qo-proto/qh"
)

var errServerStart = errors.New("server startup failed")

func main() {
	slog.Info("QH Protocol Server starting")

	// Optionally enable keylog for Wireshark decryption
	// Run with: go run -tags keylog .\examples\server\
	var serverOpts []qh.ServerOption
	keylogFile, err := os.Create("qh_server_keylog.txt")
	if err == nil {
		defer keylogFile.Close()
		serverOpts = append(serverOpts, qh.WithServerKeyLogWriter(keylogFile))
		slog.Info("Keylog file created", "path", "qh_server_keylog.txt")
	}

	srv := qh.NewServer(serverOpts...)

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
		response := fmt.Sprintf(
			"Updated data: %s ; %s; Hallo Welt;",
			string(req.Body),
			strings.Repeat("a", 1900),
		)
		return qh.TextResponse(200, response)
	})

	srv.HandleFunc("/large-post", qh.POST, func(req *qh.Request) *qh.Response {
		slog.Info(
			"Handling large POST request",
			"method",
			"POST",
			"path",
			"/large-post",
			"body_size",
			len(req.Body),
		)
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

	// A HEAD request should return the same headers as a GET request, but with no body.
	// This is useful for checking resource metadata without downloading the content.
	srv.HandleFunc("/file", qh.HEAD, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "HEAD", "path", "/file")
		// We don't need to read the file, just set the headers.
		headers := map[string]string{
			"Content-Type": qh.TextPlain.HeaderValue(),
		}
		// For HEAD, the body is nil, and Content-Length should be set to 0.
		return qh.NewResponse(200, nil, headers)
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

	// A HEAD request should return the same headers as a GET request, but with no body.
	// This is useful for checking resource metadata without downloading the content.
	srv.HandleFunc("/file", qh.HEAD, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "HEAD", "path", "/file")
		// We don't need to read the file, just set the headers.
		headers := map[string]string{
			"Content-Type": qh.TextPlain.HeaderValue(),
		}
		// For HEAD, the body is nil, and Content-Length should be set to 0.
		return qh.NewResponse(200, nil, headers)
	})

	srv.HandleFunc("/permanent-hello", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "GET", "path", "/permanent-hello")
		return qh.TextResponse(200, "Hello from the new, permanent location!")
	})

	srv.HandleFunc("/api/user", qh.PUT, func(req *qh.Request) *qh.Response {
		slog.Info(
			"Handling request",
			"method",
			"PUT",
			"path",
			"/api/user",
			"body",
			string(req.Body),
		)
		return qh.JSONResponse(200, `{"message": "User updated", "id": 123}`)
	})

	srv.HandleFunc("/api/user", qh.PATCH, func(req *qh.Request) *qh.Response {
		slog.Info(
			"Handling request",
			"method",
			"PATCH",
			"path",
			"/api/user",
			"body",
			string(req.Body),
		)
		return qh.JSONResponse(200, `{"message": "User partially updated", "id": 123}`)
	})

	srv.HandleFunc("/api/user", qh.DELETE, func(_ *qh.Request) *qh.Response {
		slog.Info("Handling request", "method", "DELETE", "path", "/api/user")
		// A 204 No Content response is appropriate for a successful DELETE with no body.
		return qh.NewResponse(204, nil, nil)
	})

	// listening with auto-generated keys
	addr := "127.0.0.1:8090"

	// You can provide a seed for deterministic keys
	seed := "Start123"
	//nolint:staticcheck // keyLogWriter parameter deprecated, use WithServerKeyLogWriter instead
	if err := srv.Listen(addr, nil, seed); err != nil {
		slog.Error("Failed to start server", "error", err)
		return
	}

	slog.Info("QH Server started", "address", addr)

	if err := srv.Serve(); err != nil && !errors.Is(err, errServerStart) {
		slog.Error("Server error", "error", err)
	}
}
