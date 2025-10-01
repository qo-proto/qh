// Package client implements a QH protocol client over QOTP transport.
package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"

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

	// Resolve hostname to IP if it's not already an IP
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.DefaultResolver.LookupIPAddr(context.Background(), host)
		if err != nil {
			return fmt.Errorf("failed to resolve hostname %s: %w", host, err)
		}
		if len(ips) == 0 {
			return fmt.Errorf("no IP addresses found for hostname: %s", host)
		}
		ip = ips[0].IP // Use the first resolved IP
	}

	// create local listener (auto generates keys)
	listener, err := qotp.Listen()
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	c.listener = listener
	ipAddr := fmt.Sprintf("%s:%d", ip.String(), port)
	// in-band key exchange
	conn, err := listener.DialString(ipAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", ipAddr, err)
	}
	c.conn = conn
	slog.Info("Connected to QH server", "addr", addr, "resolved", ipAddr)
	return nil
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
		if err != nil || len(chunk) == 0 {
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

func (c *Client) GET(host, path string, contentType protocol.ContentType, otherHeaders ...string) (*protocol.Response, error) {
	req := &protocol.Request{
		Method:  protocol.GET,
		Host:    host,
		Path:    path,
		Version: protocol.Version,
		Headers: append([]string{strconv.Itoa(int(contentType))}, otherHeaders...),
	}
	return c.Request(req)
}

func (c *Client) POST(host, path, body string, contentType protocol.ContentType, otherHeaders ...string) (*protocol.Response, error) {
	bodyBytes := []byte(body)
	headers := []string{
		strconv.Itoa(int(contentType)), // [0] Content-Type
		strconv.Itoa(len(bodyBytes)),   // [1] Content-Length
	}
	if len(otherHeaders) > 0 {
		headers = append(headers, otherHeaders...)
	}

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
		c.conn.CloseNow()
	}
	if c.listener != nil {
		return c.listener.CloseNow()
	}
	return nil
}
