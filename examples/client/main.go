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

	largePayload := strings.Repeat("LARGE_DATA_", 20000) // ~220KB

	requests := []struct {
		method string
		path   string
		body   *string
	}{
		{method: "GET", path: "/hello"},
		{method: "GET", path: "/status"},
		{method: "GET", path: "/api/user"}, // JSON response
		{method: "POST", path: "/echo", body: ptr("Hello QH World!")},
		{method: "POST", path: "/data", body: ptr("Updated data!")},
		{method: "POST", path: "/large-post", body: &largePayload},
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
			headers := map[string]string{
				"Accept": "3,2,1", // HTML, JSON, text/plain
			}
			response, err = c.GET(hostname, req.path, headers)
		case "POST":
			body := []byte("")
			if req.body != nil {
				body = []byte(*req.body)
			}
			headers := map[string]string{
				"Accept":       "2,1", // JSON, text/plain
				"Content-Type": strconv.Itoa(int(protocol.TextPlain)),
			}
			response, err = c.POST(hostname, req.path, body, headers)
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
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n--- Response for %s %s ---\n", method, path))
	sb.WriteString(fmt.Sprintf("Version:    %d\n", response.Version))
	sb.WriteString(fmt.Sprintf("StatusCode: %d\n", response.StatusCode))

	if contentTypeStr, ok := response.Headers["Content-Type"]; ok && contentTypeStr != "" {
		contentTypeCode, err := strconv.Atoi(contentTypeStr)
		if err == nil {
			sb.WriteString(fmt.Sprintf("Content:    %s\n", protocol.ContentType(contentTypeCode).String()))
		}
	}

	if dateStr, ok := response.Headers["Date"]; ok && dateStr != "" {
		unixTime, err := strconv.ParseInt(dateStr, 10, 64)
		if err == nil {
			formattedDate := time.Unix(unixTime, 0).Format("02.01.2006 15:04")
			sb.WriteString(fmt.Sprintf("Timestamp:  %s\n", formattedDate))
		}
	}

	bodyPreview := string(response.Body)
	if len(response.Body) > 100 {
		bodyPreview = string(response.Body[:100]) + "... (truncated)"
	}
	sb.WriteString(fmt.Sprintf("Body (%d bytes): %s\n", len(response.Body), bodyPreview))
	slog.Info(sb.String())
}
