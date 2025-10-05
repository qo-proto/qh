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
	handlers map[string]map[protocol.Method]Handler // path -> method -> handler (method parsed from request first byte)
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]map[protocol.Method]Handler),
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
func (s *Server) Listen(addr string, seed ...string) error {
	opts := []qotp.ListenFunc{qotp.WithListenAddr(addr)}
	if len(seed) > 0 && seed[0] != "" {
		opts = append(opts, qotp.WithSeedStr(seed[0]))
		slog.Info("QH server listening with provided seed")
	}
	listener, err := qotp.Listen(opts...)
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

	streamBuffers := make(map[*qotp.Stream][]byte)

	s.listener.Loop(func(stream *qotp.Stream) bool {
		if stream == nil {
			return true
		}

		data, err := stream.Read()
		if err != nil {
			slog.Error("Stream read error", "error", err)
			delete(streamBuffers, stream) // Clean up buffer on error
			return true
		}

		if len(data) > 0 {
			// Get or create buffer for this stream
			buffer := streamBuffers[stream]
			buffer = append(buffer, data...)
			streamBuffers[stream] = buffer

			slog.Debug("Received data fragment", "fragment_bytes", len(data), "total_bytes", len(buffer))

			complete, checkErr := protocol.IsRequestComplete(buffer)
			if checkErr != nil {
				slog.Error("Request validation error", "error", checkErr)
				s.sendErrorResponse(stream, 400, "Bad Request")
				delete(streamBuffers, stream) // Clear buffer on error
				return true
			}

			if complete {
				slog.Info("Complete request received", "bytes", len(buffer))
				s.handleRequest(stream, buffer)
				delete(streamBuffers, stream) // Clear buffer
			}
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

	// Validate and normalize Content-Type for POST requests
	if request.Method == protocol.POST && s.validateContentType(request, stream) != nil {
		return // error response already sent
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

func (s *Server) validateContentType(request *protocol.Request, stream *qotp.Stream) error {
	// Check if Content-Type header exists and is not empty
	hasContentType := len(request.Headers) > protocol.ReqHeaderContentType &&
		request.Headers[protocol.ReqHeaderContentType] != ""

	if !hasContentType {
		// Content-Type missing or empty, default to octet-stream
		slog.Debug("Content-Type missing for POST, defaulting to octet-stream")
		s.ensureHeadersSize(request, protocol.ReqHeaderContentType)
		request.Headers[protocol.ReqHeaderContentType] = strconv.Itoa(int(protocol.OctetStream)) // default
		return nil
	}

	// Validate Content-Type is in valid range (0-15)
	contentTypeStr := request.Headers[protocol.ReqHeaderContentType]
	contentType, parseErr := strconv.Atoi(contentTypeStr)
	if parseErr != nil || !protocol.IsValidContentType(contentType) {
		slog.Error("Invalid Content-Type", "value", contentTypeStr)
		s.sendErrorResponse(stream, 415, "Unsupported Media Type")
		return fmt.Errorf("invalid content-type: %s", contentTypeStr)
	}

	return nil
}

func (s *Server) ensureHeadersSize(request *protocol.Request, minIndex int) {
	if len(request.Headers) <= minIndex {
		newHeaders := make([]string, minIndex+1)
		copy(newHeaders, request.Headers)
		request.Headers = newHeaders
	}
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

// TODO: determine content type and convert body to bytes if needed
// so no need to manually write []byte: server.Response(404, protocol.TextPlain, []byte("Not Found"))

// Minimal response
func Response(statusCode int, contentType protocol.ContentType, body []byte) *protocol.Response {
	return &protocol.Response{ // required headers only
		Version:    protocol.Version,
		StatusCode: statusCode,
		Headers: []string{
			strconv.Itoa(int(contentType)), // [0] Content-Type (as code)
			strconv.Itoa(len(body)),        // [1] Content-Length
		},
		Body: body,
	}
}

func ResponseWithHeaders(statusCode int, contentType protocol.ContentType, body []byte, extraHeaders map[int]string) *protocol.Response {
	maxIdx := 1
	for idx := range extraHeaders {
		if idx != protocol.RespHeaderDate && idx > maxIdx {
			maxIdx = idx
		}
	}
	if protocol.RespHeaderDate > maxIdx {
		maxIdx = protocol.RespHeaderDate
	}
	headers := make([]string, maxIdx+1)         // initialize with empty strings
	headers[0] = strconv.Itoa(int(contentType)) // [0] Content-Type
	headers[1] = strconv.Itoa(len(body))        // [1] Content-Length
	// TODO: see if it makes sense to always include the Date
	headers[protocol.RespHeaderDate] = strconv.FormatInt(time.Now().Unix(), 10) // Unix timestamp
	for idx, val := range extraHeaders {
		if idx != protocol.RespHeaderDate { // Prevent user from overriding Date
			headers[idx] = val
		}
	}

	return &protocol.Response{
		Version:    protocol.Version,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
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
