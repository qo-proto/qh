package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"qh/internal/client"
	"qh/internal/protocol"
)

func main() {
	slog.Info("QH Protocol Client starting")

	hostname := "qh.gianhunold.ch"
	port := 8090

	addr := fmt.Sprintf("%s:%d", hostname, port)

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
		{method: "POST", path: "/data", body: ptr("Updated data!")},
		{method: "GET", path: "/file"},
		{method: "GET", path: "/image"},
		{method: "GET", path: "/not-found"}, // This will trigger a 404
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
			response, err = c.GET(hostname, req.path, protocol.TextPlain)
		case "POST":
			body := ""
			if req.body != nil {
				body = *req.body
			}
			response, err = c.POST(hostname, req.path, body, protocol.TextPlain)
		default:
			slog.Error("Unsupported method", "method", req.method, "path", req.path)
			continue
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
		} else {
			logResponse(req.method, req.path, response)

			// save image
			if req.path == "/image" {
				filename := "downloaded_cloud.jpeg"
				err := os.WriteFile(filename, response.Body, 0o644)
				if err != nil {
					slog.Error("Failed to save file", "path", filename, "error", err)
				} else {
					slog.Info("Saved response to file", "path", filename, "bytes", len(response.Body))
				}
			}
		}
	}

	slog.Info("All tests completed")
}

func logResponse(method, path string, response *protocol.Response) {
	// Format the successful response for better readability
	var formattedDate string
	// The date is the 5th header (index 4)
	if len(response.Headers) > 4 && response.Headers[4] != "" {
		unixTime, err := strconv.ParseInt(response.Headers[4], 10, 64)
		if err == nil {
			// Format to DD.MM.YYYY HH:MM
			formattedDate = time.Unix(unixTime, 0).Format("02.01.2006 15:04")
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n--- Response for %s %s ---\n", method, path))
	sb.WriteString(fmt.Sprintf("Version:    %d\n", response.Version))
	sb.WriteString(fmt.Sprintf("StatusCode: %d\n", response.StatusCode))
	if formattedDate != "" {
		sb.WriteString(fmt.Sprintf("Timestamp:  %s\n", formattedDate))
	}
	// Show body size and preview for binary data
	bodyPreview := string(response.Body)
	if len(response.Body) > 100 {
		bodyPreview = string(response.Body[:100]) + "... (truncated)"
	}
	sb.WriteString(fmt.Sprintf("Body (%d bytes): %s\n", len(response.Body), bodyPreview))
	slog.Info(sb.String())
}
