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

var errUnsupportedMethod = errors.New("unsupported method")

//nolint:funlen
func main() {
	slog.Info("QH Protocol Client starting")

	hostname := "qh.gianhunold.ch" // 127.0.0.1 with public key from seed: my-secret-server-seed
	// hostname := "qh2.gianhunold.ch" // 127.0.0.1 but no public key
	port := 8090

	addr := fmt.Sprintf("%s:%d", hostname, port)

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{method: "GET", path: "/hello"},
		{method: "GET", path: "/status"},
		{method: "GET", path: "/api/user"}, // JSON response
		{method: "POST", path: "/echo", body: "Hello QH World!"},
		{method: "POST", path: "/data", body: "Updated data!"},
		{method: "PUT", path: "/api/user", body: `{"name": "Jane Doe", "id": 123}`},
		{method: "PATCH", path: "/api/user", body: `{"status": "inactive"}`},
		{method: "POST", path: "/large-post", body: strings.Repeat("LARGE_DATA_", 20000)}, // ~220KB
		{method: "HEAD", path: "/file"},
		{method: "GET", path: "/file"},
		{method: "GET", path: "/image"},
		{method: "GET", path: "/not-found"}, // This will trigger a 404
		{
			method: "GET",
			path:   "/redirect",
		}, // This should return a 301 and hostname from the new site
	}

	c := qh.NewClient()
	defer c.Close()

	if err := c.Connect(addr); err != nil {
		slog.Error("Failed to connect", "error", err)
		return
	}

	for _, req := range requests {
		slog.Info("Testing request", "method", req.method, "path", req.path)

		var resp *qh.Response
		var err error
		switch req.method {
		case "GET":
			headers := map[string]string{
				"Accept": "text/html,application/json,text/plain",
			}
			resp, err = c.GET(hostname, req.path, headers)
		case "POST":
			headers := map[string]string{
				"Accept":       "application/json,text/plain",
				"Content-Type": "text/plain",
			}
			resp, err = c.POST(hostname, req.path, []byte(req.body), headers)
		case "PUT":
			headers := map[string]string{
				"Accept":       "application/json,text/plain",
				"Content-Type": "application/json",
			}
			resp, err = c.PUT(hostname, req.path, []byte(req.body), headers)
		case "PATCH":
			headers := map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			}
			resp, err = c.PATCH(hostname, req.path, []byte(req.body), headers)
		case "DELETE":
			// DELETE requests often have no body and expect a 204 No Content response.
			resp, err = c.DELETE(hostname, req.path, nil)
		case "HEAD":
			// HEAD requests have no body and expect headers only.
			resp, err = c.HEAD(hostname, req.path, nil)
		default:
			slog.Error(
				"Unsupported method",
				"error",
				errUnsupportedMethod,
				"method",
				req.method,
				"path",
				req.path,
			)
			continue
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
			continue
		}

		logResponse(req.method, req.path, resp)

		// save files
		var filename string
		switch req.path {
		case "/file":
			filename = "examples/client/downloaded_files/file_response.txt"
		case "/image":
			filename = "examples/client/downloaded_files/downloaded_cloud.jpeg"
		}

		if filename != "" && len(resp.Body) > 0 {
			// Create directory if it doesn't exist
			if err := os.MkdirAll("examples/client/downloaded_files", 0o755); err != nil {
				slog.Error(
					"Failed to create directory",
					"path",
					"examples/client/downloaded_files",
					"error",
					err,
				)
			} else if err := os.WriteFile(filename, resp.Body, 0o600); err != nil {
				slog.Error("Failed to save file", "path", filename, "error", err)
			} else {
				slog.Info("Saved response to file", "path", filename, "bytes", len(resp.Body))
			}
		}
	}

	slog.Info("All tests completed")
}

func logResponse(method, path string, response *qh.Response) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n--- Response for %s %s ---\n", method, path))
	sb.WriteString(fmt.Sprintf("Version:    %d\n", response.Version))
	sb.WriteString(fmt.Sprintf("StatusCode: %d\n", response.StatusCode))

	if contentTypeStr, ok := response.Headers["content-type"]; ok && contentTypeStr != "" {
		sb.WriteString(fmt.Sprintf("Content:    %s\n", contentTypeStr))
	}

	if dateStr, ok := response.Headers["Date"]; ok && dateStr != "" {
		unixTime, err := strconv.ParseInt(dateStr, 10, 64)
		if err == nil {
			formattedDate := time.Unix(unixTime, 0).Format("02.01.2006 15:04")
			sb.WriteString(fmt.Sprintf("Timestamp:  %s\n", formattedDate))
		}
	}

	// Special handling for redirect response logging
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		if location, ok := response.Headers["Location"]; ok {
			sb.WriteString(fmt.Sprintf("Location:   %s\n", location))
		}
	}

	bodyPreview := string(response.Body)
	if len(response.Body) > 100 {
		bodyPreview = string(response.Body[:100]) + "... (truncated)"
	}
	sb.WriteString(fmt.Sprintf("Body (%d bytes): %s\n", len(response.Body), bodyPreview))
	slog.Info(sb.String())
}
