package qh

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/tbocek/qotp"
)

// handles QH requests
type Handler func(*Request) *Response

type Server struct {
	listener           *qotp.Listener
	handlers           map[string]map[Method]Handler // path -> method -> handler (method parsed from request first byte)
	supportedEncodings []Encoding                    // compression algorithms this server supports, in order of preference
}

func NewServer() *Server {
	return &Server{
		handlers:           make(map[string]map[Method]Handler),
		supportedEncodings: []Encoding{Zstd, Brotli, Gzip},
	}
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
func (s *Server) Listen(addr string, keyLogWriter io.Writer, seed ...string) error {
	opts := []qotp.ListenFunc{qotp.WithListenAddr(addr)}
	if keyLogWriter != nil {
		opts = append(opts, qotp.WithKeyLogWriter(keyLogWriter))
	}
	if len(seed) > 0 && seed[0] != "" {
		opts = append(opts, qotp.WithSeedStr(seed[0]))
		slog.Info("QH server listening with provided seed")
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

			complete, checkErr := IsRequestComplete(buffer)
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

	req, err := ParseRequest(requestData)
	if err != nil {
		slog.Error("Failed to parse request", "error", err)
		s.sendErrorResponse(stream, 400, "Bad Request")
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
	if parseErr != nil || !IsValidContentType(contentType) {
		slog.Error("Invalid content-type", "value", contentTypeStr)
		s.sendErrorResponse(stream, 415, "Unsupported Media Type")
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
	return TextResponse(404, "Not Found")
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
	const minCompressionSize = 1024 // 1KB - typical HTTP server threshold
	if len(resp.Body) < minCompressionSize {
		slog.Debug("Skipping compression for small response", "bytes", len(resp.Body), "threshold", minCompressionSize)
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

	acceptedEncodings := ParseAcceptEncoding(acceptEncodingStr)
	selectedEncoding := SelectEncoding(acceptedEncodings, s.supportedEncodings)

	if selectedEncoding == "" {
		slog.Debug("No common encoding between client and server")
		return // No matching encoding
	}

	originalSize := len(resp.Body)
	compressed, err := Compress(resp.Body, selectedEncoding)
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

	for key, value := range headers {
		headerMap[key] = value
	}

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
