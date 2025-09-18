package client

import (
	"fmt"
	"log"

	"qh/internal/protocol"

	"github.com/tbocek/qotp"
)

type Client struct {
	listener *qotp.Listener
	conn     *qotp.Connection
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string) error {
	// create local listener (auto generates keys)
	listener, err := qotp.Listen()
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	c.listener = listener

	// in-band key exchange
	conn, err := listener.DialString(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	c.conn = conn
	log.Printf("Connected to QH server at %s", addr)
	return nil
}

func (c *Client) Request(req *protocol.Request) (*protocol.Response, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	// always use stream 0 for now
	stream := c.conn.Stream(0)

	// send request
	requestData := req.Format()
	log.Printf("Sending request on stream 0 (%d bytes):\n%s", len(requestData), requestData)

	_, err := stream.Write([]byte(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// wait for response
	var response *protocol.Response
	var parseErr error

	c.listener.Loop(func(s *qotp.Stream) bool {
		if s == nil {
			return true // continue waiting
		}

		responseData, err := s.Read()
		if err != nil || len(responseData) == 0 {
			return true // continue waiting
		}

		log.Printf("Received response on stream 0 (%d bytes):\n%s", len(responseData), string(responseData))
		response, parseErr = protocol.ParseResponse(string(responseData))
		return false // got response, exit loop
	})

	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse response: %w", parseErr)
	}
	if response == nil {
		return nil, fmt.Errorf("no response received")
	}

	return response, nil
}

func (c *Client) GET(host, path string, headers ...string) (*protocol.Response, error) {
	req := &protocol.Request{
		Method:  protocol.GET,
		Host:    host,
		Path:    path,
		Version: protocol.Version,
		Headers: headers,
	}
	return c.Request(req)
}

func (c *Client) POST(host, path, body string, headers ...string) (*protocol.Response, error) {
	req := &protocol.Request{
		Method:  protocol.POST,
		Host:    host,
		Path:    path,
		Version: protocol.Version,
		Headers: headers,
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
