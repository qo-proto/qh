package server

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"qh/internal/protocol"

	"github.com/tbocek/qotp"
)

// handles QH requests
type Handler func(*protocol.Request) *protocol.Response

type Server struct {
	listener *qotp.Listener
	handlers map[string]map[protocol.Method]Handler // path -> method -> handler, where method is inferred
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]map[protocol.Method]Handler), // path -> method -> handler
	}
}

// HandleFunc registers a handler for a given path and method.
func (s *Server) HandleFunc(path string, method protocol.Method, handler Handler) {
	if s.handlers[path] == nil {
		s.handlers[path] = make(map[protocol.Method]Handler)
	}
	s.handlers[path][method] = handler
	slog.Info("Registered handler", "method", method.String(), "path", path)
}

// Listen starts listening on the given address.
// It uses qotp with auto-generated keys.
func (s *Server) Listen(addr string) error {
	listener, err := qotp.Listen(qotp.WithListenAddr(addr))
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener
	slog.Info("QH server listening with auto-generated keys", "address", addr)
	return nil
}

// Serve starts the server's main loop, accepting and handling incoming streams.
func (s *Server) Serve() error {
	if s.listener == nil {
		return errors.New("server not listening")
	}

	slog.Info("Starting QH server loop")

	requestBuffers := make(map[*qotp.Stream][]byte)

	s.listener.Loop(func(stream *qotp.Stream) bool {
		if stream == nil {
			return true
		}

		chunk, err := stream.Read()
		if err != nil || len(chunk) == 0 {
			return true
		}

		requestBuffers[stream] = append(requestBuffers[stream], chunk...)

		complete, checkErr := protocol.IsRequestComplete(requestBuffers[stream])
		if checkErr != nil {
			slog.Error("Error checking request completeness", "error", checkErr)
			delete(requestBuffers, stream)
			return true
		}

		if complete {
			slog.Info("Complete request received", "bytes", len(requestBuffers[stream]))
			s.handleRequest(stream, requestBuffers[stream])
			delete(requestBuffers, stream)
		}

		return true
	})

	return nil
}

// Close shuts down the server's listener.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.CloseNow()
	}
	return nil
}

// handleRequest parses a request from a stream, routes it, and sends a response.
func (s *Server) handleRequest(stream *qotp.Stream, requestData []byte) {
	slog.Debug("Received request", "bytes", len(requestData), "data", string(requestData))

	request, err := protocol.ParseRequest(requestData)
	if err != nil {
		slog.Error("Failed to parse request", "error", err)
		s.sendErrorResponse(stream, 400, "Bad Request")
		return
	}

	response := s.routeRequest(request) // execute according handler

	// send response
	responseData := response.Format()
	slog.Debug("Sending response", "bytes", len(responseData))

	_, err = stream.Write(responseData)
	if err != nil {
		slog.Error("Failed to write response", "error", err)
		stream.CloseNow()
		return
	}

	slog.Debug("Response sent, stream kept open for reuse")
}

func (s *Server) routeRequest(request *protocol.Request) *protocol.Response {
	slog.Debug("Routing request", "path", request.Path, "method", request.Method.String())

	// check if we have a handler for this path and method
	if pathHandlers, exists := s.handlers[request.Path]; exists {
		if handler, methodExists := pathHandlers[request.Method]; methodExists {
			return handler(request) // Execute the handler for the method
		}
	}

	// no handler found, return 404
	return ErrorResponse(404, "Not Found")
}

func (s *Server) sendErrorResponse(stream *qotp.Stream, statusCode int, message string) {
	response := ErrorResponse(statusCode, message)
	responseData := response.Format()
	if _, err := stream.Write(responseData); err != nil {
		slog.Error("Failed to write error response", "error", err)
	}
}

func Response(statusCode int, contentType protocol.ContentType, body []byte) *protocol.Response {
	// Initialize headers slice with fixed size, filling with empty strings
	// Using a null byte separator for the start-line components.
	return &protocol.Response{
		Version:    0,
		StatusCode: statusCode,
		Headers: []string{ // Ordered headers by position.
			strconv.Itoa(int(contentType)),           // Content-Type (as code)
			"",                                       // Access-Control-Allow-Origin
			strconv.Itoa(len(body)),                  // Content-Length
			"",                                       // Content-Encoding (empty unless compression is used)
			strconv.FormatInt(time.Now().Unix(), 10), // Date (Unix timestamp)
			"",                                       // Set-Cookie
			"",                                       // Connection
		},
		Body: body,
	}
}

// convenience methods, e.g. write: server.TextResponse(200, "Hello")  instead of server.Response(200, "text/plain", "Hello")

func OKResponse(contentType protocol.ContentType, body []byte) *protocol.Response {
	return Response(200, contentType, body)
}

func ErrorResponse(statusCode int, message string) *protocol.Response {
	return Response(statusCode, protocol.TextPlain, []byte(message))
}

func JSONResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.JSON, []byte(body))
}

func TextResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.TextPlain, []byte(body))
}

func HTMLResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.HTML, []byte(body))
}
