package server

import (
	"fmt"
	"log"
	"net/http"
	"qh/internal/protocol"
	"time"

	"github.com/tbocek/qotp"
)

// handles QH requests
type Handler func(*protocol.Request) *protocol.Response

type Server struct {
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
		if stream != nil {
			log.Printf("NEW STREAM RECEIVED from client")
			s.handleStream(stream)
		}

		return true // continue loop
	})

	return nil
}

// handles a single stream (request/response)
func (s *Server) handleStream(stream *qotp.Stream) {
	requestData, err := stream.Read()
	if err != nil {
		log.Printf("Failed to read from stream: %v", err)
		return
	}

	if len(requestData) == 0 {
		log.Printf("Empty request received")
		return
	}

	log.Printf("Received request (%d bytes):\n%s", len(requestData), string(requestData))

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
	stream.Close()
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
	stream.Close() // Close stream after error response
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// TODO: add custom header response method

func Response(statusCode int, contentType, body string) *protocol.Response {
	return &protocol.Response{
		Version:    protocol.Version,
		StatusCode: statusCode,
		Headers: []string{
			"", // Access-Control-Allow-Origin
			"", // Content-Encoding (empty unless compression is used)
			contentType,
			time.Now().UTC().Format(http.TimeFormat), // Date (RFC 7231 HTTP date format)
			"",                                       // Set-Cookie
		},
		Body: body,
	}
}

// convenience methods, e.g. write: server.TextResponse(200, "Hello")  instead of server.Response(200, "text/plain", "Hello")
// TODO: research if these makes sense to keep/extend

func OKResponse(contentType, body string) *protocol.Response {
	return Response(200, contentType, body)
}

func ErrorResponse(statusCode int, message string) *protocol.Response {
	return Response(statusCode, "text/plain", message)
}

func JSONResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, "application/json", body)
}

func TextResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, "text/plain", body)
}

func HTMLResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, "text/html", body)
}
