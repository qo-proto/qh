//go:build keylog

package qh

import (
	"log/slog"

	"github.com/qo-proto/qotp"
)

// addKeyLogWriter adds keylog support when building with -tags keylog
func (c *Client) addKeyLogWriter(opts *[]qotp.ListenFunc) {
	if c.keylogWriter != nil {
		*opts = append(*opts, qotp.WithKeyLogWriter(c.keylogWriter))
		slog.Info("Keylog writer enabled for client")
	}
}
