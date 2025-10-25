// Package qh implements the QH protocol client and server over QOTP transport.
package qh

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/tbocek/qotp"
)

const (
	MaxRedirects = 10
)

type Client struct {
	listener   *qotp.Listener
	conn       *qotp.Conn
	streamID   uint32
	remoteAddr *net.UDPAddr
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// Concurrently resolve hostname to IP and look for a server public key in DNS.
	var ip net.IP
	var serverPubKey string
	var ipLookupErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ip, ipLookupErr = resolveAddr(host)
	}()

	go func() {
		defer wg.Done()
		// This function handles errors internally and just logs them,
		// as failing to find a key is not a critical connection error.
		serverPubKey = lookupPubKey(host)
	}()

	wg.Wait()

	// Check for errors from the IP lookup.
	if ipLookupErr != nil {
		return ipLookupErr
	}

	// create local listener (auto generates keys)
	listener, err := qotp.Listen()
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	c.listener = listener

	ipAddr := fmt.Sprintf("%s:%d", ip.String(), port)

	// Construct the UDP address to store it for potential reconnects.
	udpAddr := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	var conn *qotp.Conn
	if serverPubKey != "" {
		// Out-of-band key exchange (0-RTT)
		slog.Info("Attempting connection with out-of-band key (0-RTT)")
		pubKeyBytes, decodeErr := base64.StdEncoding.DecodeString(serverPubKey)
		if decodeErr != nil {
			slog.Warn("Failed to decode base64 public key from DNS, falling back to in-band handshake", "error", decodeErr)
		} else {
			pubKeyHex := hex.EncodeToString(pubKeyBytes)
			conn, err = listener.DialWithCryptoString(ipAddr, pubKeyHex)
		}
	} else {
		// In-band key exchange
		slog.Info("No DNS key found, attempting connection with in-band key exchange")
		conn, err = listener.DialString(ipAddr)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", ipAddr, err)
	}
	c.conn = conn
	c.remoteAddr = udpAddr
	slog.Info("Connected to QH server", "addr", addr, "resolved", ipAddr)
	return nil
}

// resolveAddr resolves a host to an IP address. It first tries to parse the host
// as a literal IP address to avoid a DNS lookup if possible.
func resolveAddr(host string) (net.IP, error) {
	// First, try parsing as an IP to avoid a DNS lookup if not needed.
	if parsedIP := net.ParseIP(host); parsedIP != nil {
		return parsedIP, nil
	}
	// If not an IP, resolve the hostname.
	ips, err := net.DefaultResolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for hostname: %s", host)
	}
	return ips[0].IP, nil // Use the first resolved IP
}

// lookupPubKey looks for a server's public key in a DNS TXT record.
// It returns the key as a string if a valid record is found, or an empty string otherwise.
func lookupPubKey(host string) string {
	txtRecords, err := net.DefaultResolver.LookupTXT(context.Background(), "_qotp."+host)
	if err != nil || len(txtRecords) == 0 {
		// No record found or an error occurred, just continue without 0-RTT.
		return ""
	}

	// Parse the first TXT record, expecting "v=0;k=..."
	record := txtRecords[0]
	parts := strings.Split(record, ";")
	version := -1
	var key string

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "v":
			v, err := strconv.Atoi(kv[1])
			if err == nil {
				version = v
			}
		case "k":
			key = kv[1]
		}
	}

	// Validate the found values
	if version == qotp.ProtoVersion && key != "" {
		slog.Info("Found valid QOTP public key in DNS TXT record", "host", host, "key", key)
		return key
	}

	if key != "" || version != -1 {
		slog.Warn("DNS TXT record found but is invalid or has mismatched version", "record", record, "expected_version", qotp.ProtoVersion)
	}
	return ""
}

func (c *Client) Request(req *Request, redirectCount int) (*Response, error) {
	if c.conn == nil {
		return nil, errors.New("client not connected")
	}

	if _, ok := req.Headers["Accept-Encoding"]; !ok {
		req.Headers["Accept-Encoding"] = "zstd, br, gzip"
	}

	// use next available stream ID
	currentStreamID := c.streamID
	c.streamID++

	stream := c.conn.Stream(currentStreamID)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	requestData := req.Format()
	slog.Debug("Sending request", "stream_id", currentStreamID, "bytes", len(requestData))

	_, err := stream.Write(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var responseBuffer []byte

	c.listener.Loop(func(s *qotp.Stream) bool {
		if s == nil {
			return true
		}

		chunk, err := s.Read()
		if err != nil {
			slog.Debug("Read error in response loop", "error", err)
			return true
		}
		if len(chunk) == 0 {
			return true
		}

		slog.Debug("Received chunk from server", "bytes", len(chunk))
		responseBuffer = append(responseBuffer, chunk...)

		complete, checkErr := IsResponseComplete(responseBuffer)
		if checkErr != nil {
			slog.Error("Error checking response completeness", "error", checkErr)
			return false
		}

		if complete {
			return false
		}

		return true
	})

	response, parseErr := ParseResponse(responseBuffer)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse response: %w", parseErr)
	}
	if response == nil {
		return nil, errors.New("no response received")
	}

	// Handle redirects
	switch response.StatusCode {
	case StatusMultipleChoices, StatusMovedPermanently, StatusFound, StatusTemporaryRedirect, StatusPermanentRedirect:
		return c.handleRedirect(req, response, redirectCount)
	}

	if err := c.decompressResponse(response); err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	return response, nil
}

func (c *Client) GET(host, path string, headers map[string]string) (*Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	req := &Request{
		Method:  GET,
		Host:    host,
		Path:    path,
		Version: Version,
		Headers: headers,
	}
	return c.Request(req, 0)
}

func (c *Client) POST(host, path string, body []byte, headers map[string]string) (*Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	// Auto-set Content-Length if not provided
	if _, exists := headers["Content-Length"]; !exists {
		headers["Content-Length"] = strconv.Itoa(len(body))
	}

	req := &Request{
		Method:  POST,
		Host:    host,
		Path:    path,
		Version: Version,
		Headers: headers,
		Body:    body,
	}
	return c.Request(req, 0)
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.listener != nil {
		return c.listener.Close()
	}
	return nil
}

func (c *Client) decompressResponse(response *Response) error {
	contentEncoding, ok := response.Headers["Content-Encoding"]
	if !ok || contentEncoding == "" {
		return nil // No compression
	}

	originalSize := len(response.Body)
	slog.Debug("Decompressing response", "encoding", contentEncoding, "compressed_bytes", originalSize)

	decompressed, err := Decompress(response.Body, Encoding(contentEncoding))
	if err != nil {
		return fmt.Errorf("failed to decompress with %s: %w", contentEncoding, err)
	}

	response.Body = decompressed
	delete(response.Headers, "Content-Encoding") // Remove encoding header after decompression
	response.Headers["Content-Length"] = strconv.Itoa(len(decompressed))

	slog.Info("Response decompressed", "encoding", contentEncoding,
		"compressed_bytes", originalSize, "decompressed_bytes", len(decompressed))

	return nil
}

func (c *Client) reconnect(host string, port int) error {
	slog.Info("Reconnecting to new host", "host", host, "port", port)
	c.Close()

	// Create a new listener for the new connection
	listener, err := qotp.Listen()
	if err != nil {
		return fmt.Errorf("failed to create new listener for reconnect: %w", err)
	}
	c.listener = listener
	return c.Connect(fmt.Sprintf("%s:%d", host, port))
}

func (c *Client) handleRedirect(req *Request, resp *Response, redirectCount int) (*Response, error) {
	if redirectCount >= MaxRedirects {
		return nil, errors.New("too many redirects")
	}

	var newHostname, newPath string

	// Prioritize custom host/path headers as requested.
	if host, ok := resp.Headers["host"]; ok {
		if path, ok := resp.Headers["path"]; ok {
			slog.Info("Redirecting (custom headers)", "status", resp.StatusCode, "host", host, "path", path)
			newHostname = host
			newPath = path
		}
	} else if location, ok := resp.Headers["Location"]; ok {
		// Fallback to standard Location header.
		slog.Info("Redirecting (Location header)", "status", resp.StatusCode, "location", location)
		newURL, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("invalid Location header: %w", err)
		}
		newHostname = newURL.Hostname()
		newPath = newURL.Path
	} else {
		return nil, errors.New("redirect response missing Location or host/path headers")
	}

	if newPath == "" {
		newPath = "/"
	}

	// For 307 and 308, the method should be preserved. For simplicity, we'll
	// always use GET for redirects, which is common practice for 301/302.
	newReq := &Request{
		Method:  GET,
		Host:    newHostname,
		Path:    newPath,
		Version: Version,
		Headers: req.Headers, // Re-use headers from original request
	}

	// Reconnect if the host has changed.
	if newHostname != "" && newHostname != req.Host {
		if err := c.reconnect(newHostname, c.remoteAddr.Port); err != nil {
			return nil, err
		}
	}
	return c.Request(newReq, redirectCount+1)
}
