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

	hostname := "qh2.gianhunold.ch" // 127.0.0.1 with public key from seed: my-secret-server-seed
	//hostname := "qh2.gianhunold.ch" // 127.0.0.1 but no public key
	port := 8090

	addr := fmt.Sprintf("%s:%d", hostname, port)

	// ptr is a helper to create a pointer to a string literal.
	//ptr := func(s string) *string { return &s }

	//largePayload := strings.Repeat("LARGE_DATA_", 20000) // ~220KB

	requests := []struct {
		method string
		path   string
		body   *string
	}{
		{method: "GET", path: "/hello"},
		//{method: "GET", path: "/status"},
		//{method: "GET", path: "/api/user"}, // JSON response
		//{method: "POST", path: "/echo", body: ptr("Hello QH World!")},
		//{method: "POST", path: "/data", body: ptr("Updated data!")},
		//{method: "POST", path: "/large-post", body: &largePayload},
		//{method: "GET", path: "/file"},
		//{method: "GET", path: "/image"},
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
			response, err = c.GET(hostname, req.path, "3,2,1", "") // Accept: HTML, JSON, text/plain
		case "POST":
			body := ""
			if req.body != nil {
				body = *req.body
			}
			response, err = c.POST(hostname, req.path, body, "2,1", "", protocol.TextPlain) // Accept: JSON, text/plain
		default:
			slog.Error("Unsupported method", "method", req.method, "path", req.path)
			continue
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
			continue
		}

		logResponse(req.method, req.path, response)

		// save files
		var filename string
		switch req.path {
		case "/file":
			filename = "examples/client/downloaded_files/file_response.txt"
		case "/image":
			filename = "examples/client/downloaded_files/downloaded_cloud.jpeg"
		}

		if filename != "" {
			if err := os.WriteFile(filename, response.Body, 0o600); err != nil {
				slog.Error("Failed to save file", "path", filename, "error", err)
			} else {
				slog.Info("Saved response to file", "path", filename, "bytes", len(response.Body))
			}
		}
	}

	slog.Info("All tests completed")
}

func logResponse(method, path string, response *protocol.Response) {
	// Format the successful response for better readability
	var formattedDate string
	if len(response.Headers) > protocol.RespHeaderDate && response.Headers[protocol.RespHeaderDate] != "" {
		unixTime, err := strconv.ParseInt(response.Headers[protocol.RespHeaderDate], 10, 64)
		if err == nil {
			// Format to DD.MM.YYYY HH:MM
			formattedDate = time.Unix(unixTime, 0).Format("02.01.2006 15:04")
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n--- Response for %s %s ---\n", method, path))
	sb.WriteString(fmt.Sprintf("Version:    %d\n", response.Version))
	sb.WriteString(fmt.Sprintf("StatusCode: %d\n", response.StatusCode))

	// Show Content-Type (decode from header position 0)
	if len(response.Headers) > 0 && response.Headers[0] != "" {
		contentTypeCode, err := strconv.Atoi(response.Headers[0])
		if err == nil {
			sb.WriteString(fmt.Sprintf("Content:    %s\n", protocol.ContentType(contentTypeCode).String()))
		}
	}

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
