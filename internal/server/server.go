package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"qh/internal/protocol"
	"strconv"
	"time"

	"github.com/tbocek/qotp"
)

// handles QH requests
type Handler func(*protocol.Request) *protocol.Response

type Server struct { // TODO: add context-based shutdown like http.Server
	// TODO: add context-based shutdown like http.Server
	listener *qotp.Listener
	handlers map[string]map[protocol.Method]Handler // path -> method -> handler
}

// TODO: add context-based shutdown like http.Server
func NewServer() *Server {
	return &Server{
		handlers: make(map[string]map[protocol.Method]Handler),
	}
}

func (s *Server) HandleFunc(path string, method protocol.Method, handler Handler) {
	if s.handlers[path] == nil {
		s.handlers[path] = make(map[protocol.Method]Handler)
	}
	s.handlers[path][method] = handler
	log.Printf("Registered handler for %s %s", method, path)
}

func (s *Server) Listen(addr string) error {
	listener, err := qotp.Listen(qotp.WithListenAddr(addr))
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener
	log.Printf("QH server listening on %s with auto-generated keys", addr)
	return nil
}

func (s *Server) Serve() error {
	if s.listener == nil {
		return fmt.Errorf("server not listening")
	}

	log.Println("Starting QH server loop...")

	s.listener.Loop(func(stream *qotp.Stream) bool {
		if stream == nil {
			return true // continue waiting
		}

		requestData, err := stream.Read()
		if err != nil || len(requestData) == 0 {
			return true // continue waiting for data
		}

		log.Printf("NEW STREAM RECEIVED from client with %d bytes", len(requestData))
		s.handleRequest(stream, requestData)

		return true // continue loop
	})

	return nil
}

// handles a single request/response
func (s *Server) handleRequest(stream *qotp.Stream, requestData []byte) {
	log.Printf("Received request (%d bytes):\n%s", len(requestData), string(requestData))

	// Find the end of the header (double newline)
	headerEnd := bytes.Index(requestData, []byte("\x00\x00"))
	var headerData []byte
	if headerEnd != -1 {
		headerData = requestData[:headerEnd]
	} else {
		headerData = requestData // No body, the whole request is the header part
	}
	log.Printf("Request header (hex):\n%s", hex.EncodeToString(headerData))

	request, err := protocol.ParseRequest(string(requestData))
	if err != nil {
		log.Printf("Failed to parse request: %v", err)
		s.sendErrorResponse(stream, 400, "Bad Request")
		return
	}

	response := s.routeRequest(request) // execute according handler

	// send response
	responseData := response.Format()
	log.Printf("Sending response (%d bytes):\n%s", len(responseData), responseData)

	_, err = stream.Write([]byte(responseData))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
		stream.Close()
		return
	}

	log.Printf("Response sent")
	// Don't close the stream for now, uses qotp's automatic timeout
	// TODO: Add proper connection closing in edge cases
}

func (s *Server) routeRequest(request *protocol.Request) *protocol.Response {
	// check if we have a handler for this path and method
	if pathHandlers, exists := s.handlers[request.Path]; exists {
		if handler, exists := pathHandlers[request.Method]; exists {
			return handler(request)
		}
	}

	// no handler found, return 404
	return ErrorResponse(404, "Not Found")
}

func (s *Server) sendErrorResponse(stream *qotp.Stream, statusCode int, message string) {
	response := ErrorResponse(statusCode, message)
	responseData := response.Format()
	if _, err := stream.Write([]byte(responseData)); err != nil {
		log.Printf("Failed to write error response: %v", err)
	}
	// Don't close the stream for now, uses qotp's automatic timeout
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// TODO: add custom header response method

func Response(statusCode int, contentType protocol.ContentType, body string) *protocol.Response {
	// Initialize headers slice with fixed size, filling with empty strings
	// Using a null byte separator for the start-line components.
	return &protocol.Response{
		Version:    protocol.Version,
		StatusCode: statusCode,
		Headers: []string{ // Ordered headers by position.
			strconv.Itoa(int(contentType)),           // Content-Type (as code)
			"",                                       // Access-Control-Allow-Origin
			strconv.Itoa(len(body)),                  // Content-Length
			"",                                       // Content-Encoding (empty unless compression is used)
			strconv.FormatInt(time.Now().Unix(), 10), // Date (Unix timestamp)
			"",                                       // Set-Cookie
		},
		Body: body,
	}
}

// convenience methods, e.g. write: server.TextResponse(200, "Hello")  instead of server.Response(200, "text/plain", "Hello")
// TODO: research if these makes sense to keep/extend

func OKResponse(contentType protocol.ContentType, body string) *protocol.Response {
	return Response(200, contentType, body)
}

func ErrorResponse(statusCode int, message string) *protocol.Response {
	return Response(statusCode, protocol.TextPlain, message)
}

func JSONResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.JSON, body)
}

func TextResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.TextPlain, body)
}

func HTMLResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.HTML, body)
}
