package qh

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"strconv"

	"github.com/qo-proto/qotp"
)

const (
	// Default server configuration values
	defaultMaxRequestSize     = 10 * 1024 * 1024 // 10MB
	defaultMinCompressionSize = 1024             // 1KB
)

// handles QH requests
type Handler func(*Request) *Response

type Server struct {
	listener           *qotp.Listener
	handlers           map[string]map[Method]Handler // path -> method -> handler (method parsed from request first byte)
	supportedEncodings []Encoding                    // compression algorithms this server supports, in order of preference
	maxRequestSize     int
	minCompressionSize int
	keylogWriter       io.Writer
}

type ServerOption func(*Server)

func WithMaxRequestSize(size int) ServerOption {
	return func(s *Server) {
		s.maxRequestSize = size
	}
}

func WithMinCompressionSize(size int) ServerOption {
	return func(s *Server) {
		s.minCompressionSize = size
	}
}

func WithSupportedEncodings(encodings []Encoding) ServerOption {
	return func(s *Server) {
		s.supportedEncodings = encodings
	}
}

func WithServerKeyLogWriter(w io.Writer) ServerOption {
	return func(s *Server) {
		s.keylogWriter = w
	}
}

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		handlers:           make(map[string]map[Method]Handler),
		supportedEncodings: []Encoding{Zstd, Brotli, Gzip},
		maxRequestSize:     defaultMaxRequestSize,
		minCompressionSize: defaultMinCompressionSize,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// HandleFunc registers a handler for a given path and method.
func (s *Server) HandleFunc(path string, method Method, handler Handler) {
	if s.handlers[path] == nil {
		s.handlers[path] = make(map[Method]Handler)
	}
	s.handlers[path][method] = handler
	slog.Info("Registered handler", "method", method.String(), "path", path)
}

// Listen starts listening on the given address.
// It uses qotp with auto-generated keys.
//
// Deprecated: keyLogWriter parameter is unused. Use WithServerKeyLogWriter option instead.
func (s *Server) Listen(addr string, _ io.Writer, seed ...string) error {
	opts := []qotp.ListenFunc{qotp.WithListenAddr(addr)}

	if len(seed) > 0 && seed[0] != "" {
		opts = append(opts, qotp.WithSeedStr(seed[0]))
		slog.Info("QH server listening with provided seed")
	}
	if s.keylogWriter != nil {
		s.addKeyLogWriter(&opts)
	}
	listener, err := qotp.Listen(opts...)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener
	slog.Info("QH server listening", "address", addr)
	slog.Info("Server public key for DNS", "pubKey", s.getPublicKeyDNS())
	return nil
}

// Serve starts the server's main loop, accepting and handling incoming streams.
func (s *Server) Serve() error {
	if s.listener == nil {
		return errors.New("server not listening")
	}

	slog.Info("Starting QH server loop")

	streamBuffers := make(map[*qotp.Stream][]byte)

	s.listener.Loop(func(stream *qotp.Stream) (bool, error) {
		if stream == nil {
			return true, nil
		}

		data, err := stream.Read()
		if err != nil {
			slog.Error("Stream read error", "error", err)
			delete(streamBuffers, stream) // Clean up buffer on error
			return true, nil
		}

		if len(data) > 0 {
			// Get or create buffer for this stream
			buffer := streamBuffers[stream]

			if len(buffer)+len(data) > s.maxRequestSize {
				slog.Error("Request size exceeds limit", "bytes", len(buffer)+len(data), "limit", s.maxRequestSize)
				s.sendErrorResponse(stream, StatusPayloadTooLarge, "Payload Too Large")
				delete(streamBuffers, stream)
				stream.Close()
				return true, nil
			}

			buffer = append(buffer, data...)
			streamBuffers[stream] = buffer

			slog.Debug("Received data fragment", "fragment_bytes", len(data), "total_bytes", len(buffer))

			complete, checkErr := isRequestComplete(buffer)
			if checkErr != nil {
				slog.Error("Request validation error", "error", checkErr)
				s.sendErrorResponse(stream, StatusBadRequest, "Bad Request")
				delete(streamBuffers, stream) // Clear buffer on error
				return true, nil
			}

			if complete {
				slog.Info("Complete request received", "bytes", len(buffer))
				s.handleRequest(stream, buffer)
				delete(streamBuffers, stream) // Clear buffer
			}
		}

		return true, nil
	})

	return nil
}

// Close shuts down the server's listener.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) getPublicKeyDNS() string {
	if s.listener == nil || s.listener.PubKey() == nil {
		return ""
	}
	return fmt.Sprintf("v=%d;k=%s", qotp.ProtoVersion, base64.StdEncoding.EncodeToString(s.listener.PubKey().Bytes()))
}

// handleRequest parses a request from a stream, routes it, and sends a response.
func (s *Server) handleRequest(stream *qotp.Stream, requestData []byte) {
	slog.Debug("Received request", "bytes", len(requestData), "data", string(requestData))

	req, err := parseRequest(requestData)
	if err != nil {
		slog.Error("Failed to parse request", "error", err)
		s.sendErrorResponse(stream, StatusBadRequest, "Bad Request")
		return
	}

	// Validate and normalize Content-Type for requests with body
	if (req.Method == POST || req.Method == PUT || req.Method == PATCH) && s.validateContentType(req, stream) != nil {
		return // error response already sent
	}

	resp := s.routeRequest(req) // execute according handler

	s.applyCompression(req, resp)

	// send response
	respData := resp.Format()
	slog.Debug("Sending response", "bytes", len(respData))

	_, err = stream.Write(respData)
	if err != nil {
		slog.Error("Failed to write response", "error", err)
		stream.Close()
		return
	}

	slog.Debug("Response sent, stream kept open for reuse")
}

func (s *Server) validateContentType(req *Request, stream *qotp.Stream) error {
	contentTypeStr, hasContentType := req.Headers["content-type"]

	if !hasContentType || contentTypeStr == "" {
		slog.Debug("content-type missing for POST, defaulting to octet-stream")
		req.Headers["content-type"] = strconv.Itoa(int(OctetStream))
		return nil
	}

	contentType, parseErr := strconv.Atoi(contentTypeStr)
	if parseErr != nil || !isValidContentType(contentType) {
		slog.Error("Invalid content-type", "value", contentTypeStr)
		s.sendErrorResponse(stream, StatusUnsupportedMediaType, "Unsupported Media Type")
		return fmt.Errorf("invalid content-type: %s", contentTypeStr)
	}

	return nil
}

func (s *Server) routeRequest(req *Request) *Response {
	slog.Debug("Routing request", "path", req.Path, "method", req.Method.String())

	// check if we have a handler for this path and method
	if pathHandlers, exists := s.handlers[req.Path]; exists {
		if handler, methodExists := pathHandlers[req.Method]; methodExists {
			return handler(req) // Execute the handler for the method
		}
	}

	// no handler found, return 404
	return TextResponse(StatusNotFound, "Not Found")
}

func (s *Server) sendErrorResponse(stream *qotp.Stream, statusCode int, message string) {
	response := TextResponse(statusCode, message)
	responseData := response.Format()
	if _, err := stream.Write(responseData); err != nil {
		slog.Error("Failed to write error response", "error", err)
	}
}

func (s *Server) applyCompression(req *Request, resp *Response) {
	if len(resp.Body) == 0 {
		return
	}

	// Don't compress very small responses (overhead not worth it)
	if len(resp.Body) < s.minCompressionSize {
		slog.Debug("Skipping compression for small response", "bytes", len(resp.Body), "threshold", s.minCompressionSize)
		return
	}

	contentTypeStr, ok := resp.Headers["content-type"]
	contentType, err := strconv.Atoi(contentTypeStr)
	if ok && err == nil && contentType == int(OctetStream) {
		slog.Debug("Skipping compression for binary media", "content_type", "octet-stream")
		return
	}

	acceptEncodingStr, ok := req.Headers["accept-encoding"]
	if !ok || acceptEncodingStr == "" {
		return
	}

	acceptedEncodings := parseAcceptEncoding(acceptEncodingStr)
	selectedEncoding := selectEncoding(acceptedEncodings, s.supportedEncodings)

	if selectedEncoding == "" {
		slog.Debug("No common encoding between client and server")
		return // No matching encoding
	}

	originalSize := len(resp.Body)
	compressed, err := compress(resp.Body, selectedEncoding)
	if err != nil {
		slog.Error("Compression failed", "encoding", selectedEncoding, "error", err)
		return
	}

	if len(compressed) >= originalSize {
		slog.Debug("Compression not beneficial", "encoding", selectedEncoding,
			"original", originalSize, "compressed", len(compressed))
		return
	}

	resp.Body = compressed
	resp.Headers["content-encoding"] = string(selectedEncoding)
	resp.Headers["content-length"] = strconv.Itoa(len(compressed))

	savings := float64(originalSize-len(compressed)) / float64(originalSize) * 100
	slog.Info("Compressed", "encoding", selectedEncoding,
		"original_bytes", originalSize, "compressed_bytes", len(compressed),
		"saved", fmt.Sprintf("%.1f%%", savings))
}

func NewResponse(statusCode int, body []byte, headers map[string]string) *Response {
	headerMap := make(map[string]string)
	headerMap["content-length"] = strconv.Itoa(len(body))

	maps.Copy(headerMap, headers)

	return &Response{
		Version:    Version,
		StatusCode: statusCode,
		Headers:    headerMap,
		Body:       body,
	}
}

// Convenience methods for common response types
func TextResponse(statusCode int, body string) *Response {
	headers := map[string]string{
		"content-type": strconv.Itoa(int(TextPlain)),
	}
	return NewResponse(statusCode, []byte(body), headers)
}

func JSONResponse(statusCode int, body string) *Response {
	headers := map[string]string{
		"content-type": strconv.Itoa(int(JSON)),
	}
	return NewResponse(statusCode, []byte(body), headers)
}
