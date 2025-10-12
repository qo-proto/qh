// Package client implements a QH protocol client over QOTP transport.
package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"

	"qh/internal/protocol"

	"github.com/tbocek/qotp"
)

type Client struct {
	listener *qotp.Listener
	conn     *qotp.Conn
	streamID uint32
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
	var version = -1
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

func (c *Client) Request(req *protocol.Request) (*protocol.Response, error) {
	if c.conn == nil {
		return nil, errors.New("client not connected")
	}

	// use next available stream ID
	currentStreamID := c.streamID
	c.streamID++

	stream := c.conn.Stream(currentStreamID)

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

		complete, checkErr := protocol.IsResponseComplete(responseBuffer)
		if checkErr != nil {
			slog.Error("Error checking response completeness", "error", checkErr)
			return false
		}

		if complete {
			return false
		}

		return true
	})

	response, parseErr := protocol.ParseResponse(responseBuffer)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse response: %w", parseErr)
	}
	if response == nil {
		return nil, errors.New("no response received")
	}

	return response, nil
}

func (c *Client) GET(host, path string, accept, acceptEncoding string) (*protocol.Response, error) {
	headers := make([]string, 4) // 4 headers in a Request
	headers[protocol.ReqHeaderAccept] = accept
	headers[protocol.ReqHeaderAcceptEncoding] = acceptEncoding
	// Content-Type and Content-Length are empty for GET requests

	req := &protocol.Request{
		Method:  protocol.GET,
		Host:    host,
		Path:    path,
		Version: protocol.Version,
		Headers: headers,
	}
	return c.Request(req)
}

// TODO: implement accept & acceptEncoding
func (c *Client) POST(host, path, body string, accept, acceptEncoding string, contentType protocol.ContentType) (*protocol.Response, error) {
	bodyBytes := []byte(body)
	headers := make([]string, 4)
	headers[protocol.ReqHeaderAccept] = accept
	headers[protocol.ReqHeaderAcceptEncoding] = acceptEncoding
	headers[protocol.ReqHeaderContentType] = strconv.Itoa(int(contentType))
	headers[protocol.ReqHeaderContentLength] = strconv.Itoa(len(bodyBytes))

	req := &protocol.Request{
		Method:  protocol.POST,
		Host:    host,
		Path:    path,
		Version: protocol.Version,
		Headers: headers,
		Body:    bodyBytes,
	}
	return c.Request(req)
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
