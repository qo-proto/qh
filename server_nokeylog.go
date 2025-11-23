//go:build !keylog

package qh

import (
	"github.com/qo-proto/qotp"
)

// addKeyLogWriter is a no-op when building without keylog support
func (s *Server) addKeyLogWriter(_ *[]qotp.ListenFunc) {
	// Keylog not available in this build
	// To enable keylog:
	// 1. Add to go.mod: replace github.com/qo-proto/qotp => ../qotp
	// 2. Build with: go build -tags keylog
}
