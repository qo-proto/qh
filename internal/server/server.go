package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"qh/internal/protocol"

	"github.com/tbocek/qotp"
)

// handles QH requests
type Handler func(*protocol.Request) *protocol.Response

type Server struct {
	listener *qotp.Listener
	handlers map[string]map[protocol.Method]Handler // path -> method -> handler
}

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
	slog.Info("Registered handler", "method", method, "path", path)
}

func (s *Server) Listen(addr string) error {
	listener, err := qotp.Listen(qotp.WithListenAddr(addr))
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener
	slog.Info("QH server listening with auto-generated keys", "address", addr)
	return nil
}

func (s *Server) Serve() error {
	if s.listener == nil {
		return errors.New("server not listening")
	}

	slog.Info("Starting QH server loop")

	s.listener.Loop(func(stream *qotp.Stream) bool {
		if stream == nil {
			return true // continue waiting
		}

		requestData, err := stream.Read()
		if err != nil || len(requestData) == 0 {
			return true // continue waiting for data
		}

		slog.Info("New stream received from client", "bytes", len(requestData))
		s.handleRequest(stream, requestData)

		return true // continue loop
	})

	return nil
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handles a single request/response
func (s *Server) handleRequest(stream *qotp.Stream, requestData []byte) {
	slog.Debug("Received request", "bytes", len(requestData), "data", string(requestData))

	request, err := protocol.ParseRequest(string(requestData))
	if err != nil {
		slog.Error("Failed to parse request", "error", err)
		s.sendErrorResponse(stream, 400, "Bad Request")
		return
	}

	response := s.routeRequest(request) // execute according handler

	// send response
	responseData := response.Format()
	slog.Debug("Sending response", "bytes", len(responseData), "data", responseData)

	_, err = stream.Write([]byte(responseData))
	if err != nil {
		slog.Error("Failed to write response", "error", err)
		stream.Close()
		return
	}

	slog.Debug("Response sent")
	// Don't close the stream for now, uses qotp's automatic timeout
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
		slog.Error("Failed to write error response", "error", err)
	}
	// Don't close the stream for now, uses qotp's automatic timeout
}

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
