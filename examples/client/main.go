package main

import (
	"fmt"
	"log/slog"
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
		}

		if err != nil {
			slog.Error("Request failed", "method", req.method, "path", req.path, "error", err)
		} else {
			logResponse(req.method, req.path, response)
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
	sb.WriteString(fmt.Sprintf("Version:    %s\n", response.Version))
	sb.WriteString(fmt.Sprintf("StatusCode: %d\n", response.StatusCode))
	if formattedDate != "" {
		sb.WriteString(fmt.Sprintf("Timestamp:  %s\n", formattedDate))
	}
	sb.WriteString(fmt.Sprintf("Body:       %s\n", response.Body))
	slog.Info(sb.String())
}
