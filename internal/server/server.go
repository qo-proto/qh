package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"qh/internal/protocol"

	"github.com/tbocek/qotp"
)

// Handler handles QH requests.
type Handler func(*protocol.Request) *protocol.Response

// connectionState holds the cached Host and Path for a qotp.Conn.
type connectionState struct {
	lastHost string
	lastPath string
}

type Server struct {
	listener *qotp.Listener
	handlers map[string]map[protocol.Method]Handler // path -> method -> handler (method parsed from request first byte)
	// connStates stores connection-specific state like cached headers.
	connStates map[*qotp.Conn]*connectionState
	// streamToConn maps a stream to its parent connection.
	streamToConn map[*qotp.Stream]*qotp.Conn
	// knownConns tracks connections that have been processed to avoid re-mapping.
	knownConns map[*qotp.Conn]bool
}

func NewServer() *Server {
	return &Server{
		handlers:     make(map[string]map[protocol.Method]Handler),
		connStates:   make(map[*qotp.Conn]*connectionState),
		streamToConn: make(map[*qotp.Stream]*qotp.Conn),
		knownConns:   make(map[*qotp.Conn]bool),
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

	streamBuffers := make(map[*qotp.Stream][]byte)

	s.listener.Loop(func(stream *qotp.Stream) bool {
		// On stream close, we could clean up the connection state if no other streams are active.
		// For simplicity, we'll let them be garbage collected when the connection eventually closes.
		// A more robust implementation might use a finalizer on the qotp.Conn.
		// stream.Conn().OnClose(func() { delete(s.connStates, stream.Conn()) })

		if stream == nil {
			return true
		}

		// Since we cannot call stream.Conn(), we must find the parent connection
		// by iterating through the listener's connections and their streams.
		// We do this only once per new stream.
		// This uses reflection as a workaround for unexported fields.
		if _, exists := s.streamToConn[stream]; !exists {
			// Use reflection to access the unexported 'conn' field of the stream.
			streamVal := reflect.ValueOf(stream).Elem()
			connVal := streamVal.FieldByName("conn")
			if connVal.IsValid() && !connVal.IsNil() {
				// Use unsafe to get the pointer value of the unexported field.
				conn := (*qotp.Conn)(unsafe.Pointer(connVal.Pointer()))
				if conn != nil {
					s.streamToConn[stream] = conn
				}
			}
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
	// Create a binary string representation for logging
	var binStr strings.Builder
	for _, b := range requestData {
		binStr.WriteString(fmt.Sprintf("%08b ", b))
	}

	slog.Info("Received request", "bytes", len(requestData),
		"data_hex", hex.EncodeToString(requestData), "data_binary", binStr.String())

	request, err := protocol.ParseRequest(requestData)
	if err != nil {
		slog.Error("Failed to parse request", "error", err)
		s.sendErrorResponse(stream, 400, "Bad Request")
		return
	}

	// Get or create the state for this connection.
	conn, ok := s.streamToConn[stream]
	if !ok {
		slog.Error("BUG: Could not find parent connection for stream")
		s.sendErrorResponse(stream, 500, "Internal Server Error")
		return
	}
	state, ok := s.connStates[conn]
	if !ok {
		state = &connectionState{}
		s.connStates[conn] = state

		// Use reflection to get the unexported connId for logging
		connVal := reflect.ValueOf(conn).Elem()
		connIDVal := connVal.FieldByName("connId")
		if connIDVal.IsValid() {
			slog.Debug("Created new connection state", "conn", connIDVal.Uint())
		}

	}

	// Apply caching logic from the protocol definition.
	if request.Host == "" {
		if state.lastHost == "" {
			slog.Error("Received request with empty host on a new connection")
			s.sendErrorResponse(stream, 400, "Bad Request: Host is required on first request")
			return
		}
		request.Host = state.lastHost
		slog.Debug("Using cached host", "host", request.Host)
	} else {
		state.lastHost = request.Host // Update cache
	}

	if request.Path == "/" && state.lastPath != "" { // Check for empty path sentinel
		request.Path = state.lastPath
		slog.Debug("Using cached path", "path", request.Path)
	} else {
		state.lastPath = request.Path // Update cache
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
	return TextResponse(404, "Not Found")
}

func (s *Server) sendErrorResponse(stream *qotp.Stream, statusCode int, message string) {
	response := TextResponse(statusCode, message)
	responseData := response.Format()
	if _, err := stream.Write(responseData); err != nil {
		slog.Error("Failed to write error response", "error", err)
	}
}

// TODO: determine content type and convert body to bytes if needed
// so no need to manually write []byte: server.Response(404, protocol.TextPlain, []byte("Not Found"))
// TODO: make content-type optional like in http?
func Response(statusCode int, contentType protocol.ContentType, body []byte, headers map[int]string) *protocol.Response {
	// For minimal responses (no optional headers), just return Content-Type and Content-Length
	if len(headers) == 0 {
		return &protocol.Response{
			Version:    protocol.Version,
			StatusCode: statusCode,
			Headers: []string{
				strconv.Itoa(int(contentType)), // [0] Content-Type
				strconv.Itoa(len(body)),        // [1] Content-Length
			},
			Body: body,
		}
	}

	// Calculate max header index needed
	maxIdx := 1 // At minimum we need Content-Type (0) and Content-Length (1)
	for idx := range headers {
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	// Ensure room for date header
	if protocol.RespHeaderDate > maxIdx {
		maxIdx = protocol.RespHeaderDate
	}

	// Build header array (auto-fills with empty strings)
	headerArray := make([]string, maxIdx+1)
	headerArray[0] = strconv.Itoa(int(contentType)) // [0] Content-Type
	headerArray[1] = strconv.Itoa(len(body))        // [1] Content-Length

	// Add user headers (and prevent overriding Date)
	for idx, val := range headers {
		if idx > 1 && idx != protocol.RespHeaderDate {
			headerArray[idx] = val
		}
	}

	// Auto add Date when optional headers are present
	headerArray[protocol.RespHeaderDate] = strconv.FormatInt(time.Now().Unix(), 10)

	return &protocol.Response{
		Version:    protocol.Version,
		StatusCode: statusCode,
		Headers:    headerArray,
		Body:       body,
	}
}

// Convenience methods for common response types
func TextResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.TextPlain, []byte(body), nil)
}

func JSONResponse(statusCode int, body string) *protocol.Response {
	return Response(statusCode, protocol.JSON, []byte(body), nil)
}
