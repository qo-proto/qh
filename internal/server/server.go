package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"qh/internal/compression"
	"qh/internal/protocol"

	"github.com/tbocek/qotp"
)

// handles QH requests
type Handler func(*protocol.Request) *protocol.Response

type Server struct {
	listener           *qotp.Listener
	handlers           map[string]map[protocol.Method]Handler // path -> method -> handler (method parsed from request first byte)
	supportedEncodings []compression.Encoding                 // compression algorithms this server supports, in order of preference
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]map[protocol.Method]Handler),
		supportedEncodings: []compression.Encoding{
			compression.Zstd,
			compression.Brotli,
			compression.Gzip,
		},
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

	s.applyCompression(request, response)

	// send response
	responseData := response.Format()
	slog.Debug("Sending response", "bytes", len(responseData))

	_, err = stream.Write(responseData)
	if err != nil {
		slog.Error("Failed to write response", "error", err)
		stream.Close()
		return
	}

	slog.Debug("Response sent, stream kept open for reuse")
}

func (s *Server) validateContentType(request *protocol.Request, stream *qotp.Stream) error {
	contentTypeStr, hasContentType := request.Headers["Content-Type"]

	if !hasContentType || contentTypeStr == "" {
		slog.Debug("Content-Type missing for POST, defaulting to octet-stream")
		request.Headers["Content-Type"] = strconv.Itoa(int(protocol.OctetStream))
		return nil
	}

	contentType, parseErr := strconv.Atoi(contentTypeStr)
	if parseErr != nil || !protocol.IsValidContentType(contentType) {
		slog.Error("Invalid Content-Type", "value", contentTypeStr)
		s.sendErrorResponse(stream, 415, "Unsupported Media Type")
		return fmt.Errorf("invalid content-type: %s", contentTypeStr)
	}

	return nil
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

func (s *Server) applyCompression(request *protocol.Request, response *protocol.Response) {
	if len(response.Body) == 0 {
		return
	}

	// Don't compress very small responses (overhead not worth it)
	const minCompressionSize = 1024 // 1KB - typical HTTP server threshold
	if len(response.Body) < minCompressionSize {
		slog.Debug("Skipping compression for small response", "bytes", len(response.Body), "threshold", minCompressionSize)
		return
	}

	contentTypeStr, ok := response.Headers["Content-Type"]
	contentType, err := strconv.Atoi(contentTypeStr)
	if ok && err == nil && contentType == int(protocol.OctetStream) {
		slog.Debug("Skipping compression for binary media", "content_type", "octet-stream")
		return
	}

	acceptEncodingStr, ok := request.Headers["Accept-Encoding"]
	if !ok || acceptEncodingStr == "" {
		return
	}

	acceptedEncodings := compression.ParseAcceptEncoding(acceptEncodingStr)
	selectedEncoding := compression.SelectEncoding(acceptedEncodings, s.supportedEncodings)

	if selectedEncoding == "" {
		slog.Debug("No common encoding between client and server")
		return // No matching encoding
	}

	originalSize := len(response.Body)
	compressed, err := compression.Compress(response.Body, selectedEncoding)
	if err != nil {
		slog.Error("Compression failed", "encoding", selectedEncoding, "error", err)
		return
	}

	if len(compressed) >= originalSize {
		slog.Debug("Compression not beneficial", "encoding", selectedEncoding,
			"original", originalSize, "compressed", len(compressed))
		return
	}

	response.Body = compressed
	response.Headers["Content-Encoding"] = string(selectedEncoding)
	response.Headers["Content-Length"] = strconv.Itoa(len(compressed))

	savings := float64(originalSize-len(compressed)) / float64(originalSize) * 100
	slog.Info("Compressed", "encoding", selectedEncoding,
		"original_bytes", originalSize, "compressed_bytes", len(compressed),
		"saved", fmt.Sprintf("%.1f%%", savings))
}

func Response(statusCode int, body []byte, headers map[string]string) *protocol.Response {
	headerMap := make(map[string]string)
	headerMap["Content-Length"] = strconv.Itoa(len(body))

	for key, value := range headers {
		headerMap[key] = value
	}

	return &protocol.Response{
		Version:    protocol.Version,
		StatusCode: statusCode,
		Headers:    headerMap,
		Body:       body,
	}
}

// Convenience methods for common response types
func TextResponse(statusCode int, body string) *protocol.Response {
	headers := map[string]string{
		"Content-Type": strconv.Itoa(int(protocol.TextPlain)),
	}
	return Response(statusCode, []byte(body), headers)
}

func JSONResponse(statusCode int, body string) *protocol.Response {
	headers := map[string]string{
		"Content-Type": strconv.Itoa(int(protocol.JSON)),
	}
	return Response(statusCode, []byte(body), headers)
}
