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
	"sync/atomic"

	"github.com/qo-proto/qotp"
)

const (
	// Default client configuration values
	defaultMaxResponseSize = 50 * 1024 * 1024 // 50MB
	defaultMaxRedirects    = 10

	x25519KeySize         = 32 // X25519 public key size in bytes
	maxDNSTXTRecordLength = 80
)

// Client is a QH protocol client that manages connections to QH servers.
// It supports connection establishment with optional 0-RTT via DNS-based key exchange,
// automatic response decompression, and redirect handling.
type Client struct {
	listener        *qotp.Listener
	conn            *qotp.Conn
	streamID        atomic.Uint32
	remoteAddr      *net.UDPAddr
	maxResponseSize int
	maxRedirects    int
}

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*Client)

// WithMaxResponseSize sets the maximum allowed response size in bytes.
// Responses exceeding this limit will return an error.
// Default is 50MB.
func WithMaxResponseSize(size int) ClientOption {
	return func(c *Client) {
		c.maxResponseSize = size
	}
}

// WithMaxRedirects sets the maximum number of redirects to follow.
// If this limit is exceeded, the request will return an error.
// Default is 10.
func WithMaxRedirects(limit int) ClientOption {
	return func(c *Client) {
		c.maxRedirects = limit
	}
}

// NewClient creates a new QH client with the specified options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		maxResponseSize: defaultMaxResponseSize,
		maxRedirects:    defaultMaxRedirects,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect establishes a connection to a QH server at the specified address.
// The address should be in "host:port" format.
//
// Connect performs concurrent DNS lookups to resolve the hostname and
// optionally retrieve the server's public key from a DNS TXT record
// (at _qotp.<host>) for 0-RTT connection establishment. If no valid DNS key
// is found, Connect falls back to a standard in-band key exchange handshake.
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
		switch {
		case decodeErr != nil:
			slog.Warn("Failed to decode base64 public key from DNS, falling back to in-band handshake", "error", decodeErr)
		case len(pubKeyBytes) != x25519KeySize:
			slog.Warn("Invalid public key length from DNS, expected X25519 key size, falling back to in-band handshake", "expected", x25519KeySize, "got", len(pubKeyBytes))
		default:
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

	if len(record) > maxDNSTXTRecordLength {
		slog.Warn("DNS TXT record too long, ignoring", "max_length", maxDNSTXTRecordLength, "actual_length", len(record), "host", host)
		return ""
	}
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

// Request sends a QH request and returns the response.
// The redirectCount parameter tracks the number of redirects followed
// and should typically be 0 for initial requests.
//
// This method handles automatic decompression of responses if the server
// uses compression and the client advertised support via Accept-Encoding.
func (c *Client) Request(req *Request, redirectCount int) (*Response, error) {
	if c.conn == nil {
		return nil, errors.New("client not connected")
	}

	if _, ok := req.Headers["accept-encoding"]; !ok {
		req.Headers["accept-encoding"] = "zstd, br, gzip"
	}

	// Get next available stream ID
	currentStreamID := c.streamID.Add(1) - 1

	stream := c.conn.Stream(currentStreamID)

	requestData := req.Format()
	slog.Debug("Sending request", "stream_id", currentStreamID, "bytes", len(requestData))

	_, err := stream.Write(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var responseBuffer []byte

	c.listener.Loop(func(s *qotp.Stream) (bool, error) {
		if s == nil {
			return true, nil
		}

		chunk, err := s.Read()
		if err != nil {
			slog.Debug("Read error in response loop", "error", err)
			return true, nil
		}
		if len(chunk) == 0 {
			return true, nil
		}

		slog.Debug("Received chunk from server", "bytes", len(chunk))

		if len(responseBuffer)+len(chunk) > c.maxResponseSize {
			return false, fmt.Errorf("response size exceeds limit of %d bytes", c.maxResponseSize)
		}

		responseBuffer = append(responseBuffer, chunk...)

		complete, checkErr := isResponseComplete(responseBuffer)
		if checkErr != nil {
			slog.Error("Error checking response completeness", "error", checkErr)
			return false, nil
		}

		if complete {
			return false, nil
		}

		return true, nil
	})

	resp, parseErr := parseResponse(responseBuffer)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse response: %w", parseErr)
	}
	if resp == nil {
		return nil, errors.New("no response received")
	}

	// Handle redirects
	switch resp.StatusCode {
	case StatusMultipleChoices, StatusMovedPermanently, StatusFound, StatusTemporaryRedirect, StatusPermanentRedirect:
		return c.handleRedirect(req, resp, redirectCount)
	}

	if err := c.decompressResponse(resp); err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	return resp, nil
}

// GET performs a GET request to the specified host and path.
// Returns the server's response or an error if the request fails.
func (c *Client) GET(host, path string, headers map[string]string) (*Response, error) {
	return c.do(GET, host, path, headers, nil)
}

// POST performs a POST request with the given body to the specified host and path.
// The Content-Length header is automatically set based on the body size.
func (c *Client) POST(host, path string, body []byte, headers map[string]string) (*Response, error) {
	return c.do(POST, host, path, headers, body)
}

// PUT performs a PUT request with the given body to the specified host and path.
// The Content-Length header is automatically set based on the body size.
func (c *Client) PUT(host, path string, body []byte, headers map[string]string) (*Response, error) {
	return c.do(PUT, host, path, headers, body)
}

// PATCH performs a PATCH request with the given body to the specified host and path.
// The Content-Length header is automatically set based on the body size.
func (c *Client) PATCH(host, path string, body []byte, headers map[string]string) (*Response, error) {
	return c.do(PATCH, host, path, headers, body)
}

// DELETE performs a DELETE request to the specified host and path.
// Returns the server's response or an error if the request fails.
func (c *Client) DELETE(host, path string, headers map[string]string) (*Response, error) {
	return c.do(DELETE, host, path, headers, nil)
}

// HEAD performs a HEAD request to the specified host and path.
// Returns only headers without a response body.
func (c *Client) HEAD(host, path string, headers map[string]string) (*Response, error) {
	return c.do(HEAD, host, path, headers, nil)
}

// Close closes the client connection and releases associated resources.
// After calling Close, the client should not be used for further requests.
func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.listener != nil {
		return c.listener.Close()
	}
	return nil
}

func (c *Client) do(method Method, host, path string, headers map[string]string, body []byte) (*Response, error) {
	if headers == nil {
		headers = map[string]string{}
	}

	// Normalize body based on method - body is only allowed for POST, PUT, PATCH
	// NOTE: content-length header is not needed in QH - body length is determined by varint prefix
	if method != POST && method != PUT && method != PATCH {
		body = nil // ensure no body for non-body methods
	}

	req := &Request{
		Method:  method,
		Host:    host,
		Path:    path,
		Version: Version,
		Headers: headers,
		Body:    body,
	}
	return c.Request(req, 0)
}

func (c *Client) decompressResponse(resp *Response) error {
	contentEncoding, ok := resp.Headers["content-encoding"]
	if !ok || contentEncoding == "" {
		return nil // No compression
	}

	originalSize := len(resp.Body)
	slog.Debug("Decompressing response", "encoding", contentEncoding, "compressed_bytes", originalSize)

	decompressed, err := decompress(resp.Body, Encoding(contentEncoding), c.maxResponseSize)
	if err != nil {
		return fmt.Errorf("failed to decompress with %s: %w", contentEncoding, err)
	}

	resp.Body = decompressed
	delete(resp.Headers, "content-encoding") // Remove encoding header after decompression

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
	if redirectCount >= c.maxRedirects {
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
	} else if location, ok := resp.Headers["location"]; ok {
		// Fallback to standard location header.
		slog.Info("Redirecting (location header)", "status", resp.StatusCode, "location", location)
		newURL, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("invalid location header: %w", err)
		}
		newHostname = newURL.Hostname()
		newPath = newURL.Path
	} else {
		return nil, errors.New("redirect response missing location or host/path headers")
	}

	if newPath == "" {
		newPath = "/"
	}

	// Preserve method and body for 307/308; switch to GET for 300/301/302.
	preserve := resp.StatusCode == StatusTemporaryRedirect || resp.StatusCode == StatusPermanentRedirect
	newMethod := GET
	var newBody []byte
	headers := req.Headers
	if preserve {
		newMethod = req.Method
		newBody = req.Body
	}
	newReq := &Request{
		Method:  newMethod,
		Host:    newHostname,
		Path:    newPath,
		Version: Version,
		Headers: headers,
		Body:    newBody,
	}

	// Reconnect if the host has changed.
	if newHostname != "" && newHostname != req.Host {
		if err := c.reconnect(newHostname, c.remoteAddr.Port); err != nil {
			return nil, err
		}
	}
	return c.Request(newReq, redirectCount+1)
}
