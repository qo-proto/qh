package main

import (
	"fmt"
	"log/slog"
	"net"
	"strings"

	"qh/internal/client"
	"qh/internal/protocol"
)

func main() {
	slog.Info("QH Protocol Client starting")

	hostname := "qh.gianhunold.ch"
	port := 8090

	ips, err := net.LookupIP(hostname)
	if err != nil {
		slog.Error("Failed to resolve hostname", "hostname", hostname, "error", err)
		return
	}
	if len(ips) == 0 {
		slog.Error("No IP addresses found for hostname", "hostname", hostname)
		return
	}
	// Use the first resolved IP address
	ip := ips[0]
	slog.Info("Resolved hostname", "hostname", hostname, "ip", ip.String())
	addr := fmt.Sprintf("%s:%d", ip.String(), port)

	// ptr is a helper to create a pointer to a string literal.
	ptr := func(s string) *string { return &s }

	requests := []struct {
		method string
		path   string
		body   *string
	}{
		{method: "GET", path: "/hello"},
		{method: "GET", path: "/status"},
		{method: "POST", path: "/echo", body: ptr("Hello QH World!")},
		{method: "POST", path: "/data", body: ptr(strings.Repeat("a", 2000))},
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
			response, err = c.GET(hostname, req.path, "text/plain")
		case "POST":
			body := ""
			if req.body != nil {
				body = *req.body
			}
			response, err = c.POST(hostname, req.path, body, "text/plain")
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
		} else {
			slog.Info("Request success", "method", req.method, "path", req.path, "status", response.StatusCode, "body", response.Body)
		}
	}

	slog.Info("All tests completed")
}
