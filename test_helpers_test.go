package qh

import (
	"fmt"
	"net"
	"testing"
)

// newTestServer creates a test server on a random available port.
// Returns the server and its address (e.g., "127.0.0.1:12345").
//
//nolint:unparam // opts kept for future extensibility
func newTestServer(t *testing.T, opts ...ServerOption) (*Server, string) {
	t.Helper()

	// Find an available port by briefly binding to port 0
	//nolint:noctx // Context not needed
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	srv := NewServer(opts...)

	err = srv.Listen(addr, nil, "test")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	go func() {
		_ = srv.Serve()
	}()

	return srv, addr
}
