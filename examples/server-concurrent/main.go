package main

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/qo-proto/qh"
)

var errServerStart = errors.New("server startup failed")

// Intentional delay to verify concurrent execution
// With multiplexing it takes ~200-300ms to complete all 8 requests
// Without it takes ~1.6s (8 × 200ms).
const requestDelay = 200 * time.Millisecond

func main() {
	slog.Info("QH Protocol Concurrent Test Server starting")
	slog.Info("Each request will have an artificial delay", "delay", requestDelay)

	srv := qh.NewServer()

	srv.HandleFunc("/hello", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Request processing", "method", "GET", "path", "/hello")
		time.Sleep(requestDelay)
		return qh.TextResponse(200, "Hello from QH Protocol!")
	})

	srv.HandleFunc("/echo", qh.POST, func(req *qh.Request) *qh.Response {
		slog.Info("Request processing", "method", "POST", "path", "/echo", "body", string(req.Body))
		time.Sleep(requestDelay)
		response := "Echo: " + string(req.Body)
		return qh.TextResponse(200, response)
	})

	srv.HandleFunc("/status", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Request processing", "method", "GET", "path", "/status")
		time.Sleep(requestDelay)
		return qh.TextResponse(200, "QH Server is running!")
	})

	srv.HandleFunc("/api/user", qh.GET, func(_ *qh.Request) *qh.Response {
		slog.Info("Request processing", "method", "GET", "path", "/api/user")
		time.Sleep(requestDelay)
		headers := map[string]string{
			"Content-Type":  qh.JSON.HeaderValue(),
			"Cache-Control": "max-age=3600",
			"Date":          strconv.FormatInt(time.Now().Unix(), 10),
		}
		body := `{"name": "John Doe", "id": 123, "active": true}`
		return qh.NewResponse(200, []byte(body), headers)
	})

	srv.HandleFunc("/data", qh.POST, func(req *qh.Request) *qh.Response {
		slog.Info("Request processing", "method", "POST", "path", "/data", "body", string(req.Body))
		time.Sleep(requestDelay)
		response := "Updated data: " + string(req.Body)
		return qh.TextResponse(200, response)
	})

	addr := "127.0.0.1:8090"
	seed := "Start123"

	//nolint:staticcheck
	if err := srv.Listen(addr, nil, seed); err != nil {
		slog.Error("Failed to start server", "error", err)
		return
	}

	slog.Info("QH Concurrent Server started", "address", addr)

	if err := srv.Serve(); err != nil && !errors.Is(err, errServerStart) {
		slog.Error("Server error", "error", err)
	}
}
