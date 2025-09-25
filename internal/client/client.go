package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"qh/internal/protocol"
	"strconv"
	"strings"

	"github.com/tbocek/qotp"
)

type Client struct {
	listener *qotp.Listener
	conn     *qotp.Connection
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
	log.Printf("Connected to QH server at %s (resolved to %s)", addr, ipAddr)
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

	// send request
	requestData := req.Format()
	slog.Debug("Sending request", "stream_id", currentStreamID, "bytes", len(requestData), "data", requestData)

	_, err := stream.Write([]byte(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// wait for response by reading directly from stream
	var response *protocol.Response
	var parseErr error
	var responseBuffer strings.Builder

	c.listener.Loop(func(s *qotp.Stream) bool {
		if s != stream {
			return true // continue waiting
		}

		// Loop to read all available data from the stream until we have a complete message.
		for {
			responseData, err := s.Read()
			if err != nil || len(responseData) == 0 {
				return true // No more data right now, continue listening.
			}

			slog.Info("Received frame from server", "stream_id", currentStreamID, "bytes", len(responseData))

			responseBuffer.Write(responseData)

			// Check if we have received the end-of-transmission character.
			if strings.HasSuffix(responseBuffer.String(), "\x04") {
				fullMessage := strings.TrimSuffix(responseBuffer.String(), "\x04")
				slog.Debug("Reassembled full response", "bytes", len(fullMessage))
				response, parseErr = protocol.ParseResponse(fullMessage)
				return false // We have the complete message, exit the listener loop.
			}
		}
	})

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
	req := &protocol.Request{
		Method:  protocol.POST,
		Host:    host,
		Path:    path,
		Version: protocol.Version,
		// The first header for a POST request is the Content-Type of the body.
		Headers: append([]string{strconv.Itoa(int(contentType))}, otherHeaders...),
		Body:    body,
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
