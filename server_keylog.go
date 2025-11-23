//go:build keylog
// +build keylog

package qh

import (
	"log/slog"

	"github.com/qo-proto/qotp"
)

// addKeyLogWriter adds keylog support when building with -tags keylog
func (s *Server) addKeyLogWriter(opts *[]qotp.ListenFunc) {
	if s.keylogWriter != nil {
		*opts = append(*opts, qotp.WithKeyLogWriter(s.keylogWriter))
		slog.Info("Keylog writer enabled for server")
	}
}
