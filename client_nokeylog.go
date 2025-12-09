//go:build !keylog

package qh

import "github.com/qo-proto/qotp"

// addKeyLogWriter is a no-op when building without keylog support
func (c *Client) addKeyLogWriter(_ *[]qotp.ListenFunc) {
	// Keylog not available in this build
	// To enable keylog:
	// 1. Set QH_ENABLE_KEYLOG=1 environment variable
	// 2. Build with: go build -tags keylog
}
