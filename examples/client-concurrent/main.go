package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/qo-proto/qh"
)

func main() {
	slog.Info("QH Protocol Concurrent Client Test - Testing Stream Multiplexing")

	hostname := "127.0.0.1"
	port := 8090
	addr := fmt.Sprintf("%s:%d", hostname, port)

	requests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"GET-hello", "GET", "/hello", ""},
		{"GET-status", "GET", "/status", ""},
		{"POST-echo", "POST", "/echo", "Hello from concurrent POST!"},
		{"GET-api-user", "GET", "/api/user", ""},
		{"POST-data", "POST", "/data", "Concurrent data update"},
		{"GET-hello-2", "GET", "/hello", ""},
		{"GET-status-2", "GET", "/status", ""},
		{"POST-echo-2", "POST", "/echo", "Another concurrent POST!"},
	}

	c := qh.NewClient()
	defer c.Close()

	if err := c.Connect(addr, nil); err != nil {
		slog.Error("Failed to connect", "error", err)
		return
	}

	slog.Info("Connected successfully, starting concurrent requests test")
	// single connection for all requests
	// all requests will use different stream IDs on the same connection

	var wg sync.WaitGroup
	start := time.Now()

	for _, req := range requests {
		wg.Go(func() {
			slog.Info("request starting", "name", req.name, "method", req.method, "path", req.path)

			var resp *qh.Response
			var err error

			switch req.method {
			case "GET":
				resp, err = c.GET(hostname, req.path, nil)
			case "POST":
				headers := map[string]string{
					"Content-Type": "text/plain",
				}
				resp, err = c.POST(hostname, req.path, []byte(req.body), headers)
			}

			if err != nil {
				slog.Error("request failed", "name", req.name, "error", err)
			} else {
				slog.Info("response received", "name", req.name, "status", resp.StatusCode, "body_len", len(resp.Body))
			}
		})
	}

	wg.Wait()
	duration := time.Since(start)

	slog.Info("Concurrent requests completed",
		"total_requests", len(requests),
		"duration", duration)

	// let QOTP finish sending ACKs before closing
	time.Sleep(100 * time.Millisecond)
}
