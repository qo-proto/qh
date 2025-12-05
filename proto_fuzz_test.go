package qh

import (
	"bytes"
	"strings"
	"testing"
)

//nolint:gocognit,nestif,cyclop // intentional flat structure
func FuzzParseRequest(f *testing.F) {
	f.Add([]byte("\x00\x0Bexample.com\x06/hello\x00\x00"))                       // Minimal GET
	f.Add([]byte("\x08\x0Bexample.com\x05/echo\x00\x04test"))                    // POST with body
	f.Add([]byte("\x00\x09localhost\x01/\x00\x00"))                              // Minimal path
	f.Add([]byte("\x00\x0Bexample.com\x00\x00\x00"))                             // Empty path (should default to /)
	f.Add([]byte("\x10\x0Bexample.com\x05/data\x03\x46\x01" + "2" + "\x08body")) // PUT with headers
	f.Add([]byte("\x20\x0Bexample.com\x09/resource\x00\x00"))                    // DELETE

	f.Fuzz(func(t *testing.T, data []byte) {
		req, err := ParseRequest(data)

		if err == nil {
			if req == nil {
				t.Fatal("ParseRequest returned nil request with no error")
			}

			if req.Version > maxVersionValue {
				t.Errorf("Invalid version: %d", req.Version)
			}

			if req.Method < GET || req.Method > OPTIONS {
				t.Errorf("Invalid method: %d", req.Method)
			}

			if req.Host == "" {
				t.Error("Empty host in successful parse")
			}

			if len(req.Host) > maxHostLength {
				t.Errorf("Host exceeds max length: %d", len(req.Host))
			}

			if req.Path == "" {
				t.Error("Empty path in successful parse (should default to /)")
			}

			if req.Headers == nil {
				t.Error("Headers map is nil (should be empty map)")
			}

			for k := range req.Headers {
				if k != strings.ToLower(k) {
					t.Errorf("Header key not normalized to lowercase: %s", k)
				}
			}

			// Roundtrip test: parse -> format -> parse
			encoded := req.Format()
			req2, err2 := ParseRequest(encoded)
			if err2 != nil {
				t.Errorf("Roundtrip failed: %v", err2)
			}

			if req2 != nil {
				if req2.Method != req.Method {
					t.Error("Roundtrip: method mismatch")
				}
				if req2.Host != req.Host {
					t.Error("Roundtrip: host mismatch")
				}
				if req2.Path != req.Path {
					t.Error("Roundtrip: path mismatch")
				}
				if req2.Version != req.Version {
					t.Error("Roundtrip: version mismatch")
				}
				if len(req2.Headers) != len(req.Headers) {
					t.Error("Roundtrip: header count mismatch")
				}
				for k, v := range req.Headers {
					if req2.Headers[k] != v {
						t.Errorf("Roundtrip: header %s mismatch", k)
					}
				}
				if !bytes.Equal(req2.Body, req.Body) {
					t.Error("Roundtrip: body mismatch")
				}
			}
		}
	})
}

//nolint:gocognit,nestif,cyclop // intentional flat structure
func FuzzParseResponse(f *testing.F) {
	f.Add([]byte("\x00\x00\x04OK!"))                      // 200 OK minimal
	f.Add([]byte("\x01\x00\x09Not Found"))                // 404 Not Found
	f.Add([]byte("\x00\x03\x90\x01" + "1" + "\x05Hello")) // 200 with content-type header
	f.Add([]byte("\x02\x00\x15Internal Server Error"))    // 500 error
	f.Add([]byte("\x00\x00\x00"))                         // Empty body

	f.Fuzz(func(t *testing.T, data []byte) {
		resp, err := ParseResponse(data)

		if err == nil {
			if resp == nil {
				t.Fatal("ParseResponse returned nil response with no error")
			}

			if resp.Version > maxVersionValue {
				t.Errorf("Invalid version: %d", resp.Version)
			}

			if resp.StatusCode < 100 || resp.StatusCode > 599 {
				t.Errorf("Invalid status code: %d", resp.StatusCode)
			}

			if resp.Headers == nil {
				t.Error("Headers map is nil (should be empty map)")
			}

			for k := range resp.Headers {
				if k != strings.ToLower(k) {
					t.Errorf("Header key not normalized to lowercase: %s", k)
				}
			}

			// Roundtrip test: parse -> format -> parse
			encoded := resp.Format()
			resp2, err2 := ParseResponse(encoded)
			if err2 != nil {
				t.Errorf("Roundtrip failed: %v", err2)
			}

			if resp2 != nil {
				if resp2.StatusCode != resp.StatusCode {
					t.Error("Roundtrip: status code mismatch")
				}
				if resp2.Version != resp.Version {
					t.Error("Roundtrip: version mismatch")
				}
				if len(resp2.Headers) != len(resp.Headers) {
					t.Error("Roundtrip: header count mismatch")
				}
				for k, v := range resp.Headers {
					if resp2.Headers[k] != v {
						t.Errorf("Roundtrip: header %s mismatch", k)
					}
				}
				if !bytes.Equal(resp2.Body, resp.Body) {
					t.Error("Roundtrip: body mismatch")
				}
			}
		}
	})
}

func FuzzIsRequestComplete(f *testing.F) {
	f.Add([]byte("\x00"))                                  // Just first byte
	f.Add([]byte("\x00\x0B"))                              // First byte + host length
	f.Add([]byte("\x00\x0Bexample.com"))                   // Partial request
	f.Add([]byte("\x00\x0Bexample.com\x06/hello\x00\x00")) // Complete request

	f.Fuzz(func(t *testing.T, data []byte) {
		complete, completeErr := IsRequestComplete(data)
		req, parseErr := ParseRequest(data)

		if parseErr == nil && !complete {
			t.Error("ParseRequest succeeded but IsRequestComplete said incomplete")
		}

		if completeErr != nil && parseErr == nil {
			t.Errorf("IsRequestComplete errored (%v) but ParseRequest succeeded", completeErr)
		}

		if parseErr == nil && req == nil {
			t.Error("ParseRequest returned no error but returned nil request")
		}
	})
}

func FuzzIsResponseComplete(f *testing.F) {
	f.Add([]byte("\x00"))             // Just first byte
	f.Add([]byte("\x00\x00"))         // First byte + headers length
	f.Add([]byte("\x00\x00\x04"))     // Partial response
	f.Add([]byte("\x00\x00\x04OK!!")) // Complete response

	f.Fuzz(func(t *testing.T, data []byte) {
		complete, completeErr := IsResponseComplete(data)
		resp, parseErr := ParseResponse(data)

		if parseErr == nil && !complete {
			t.Error("ParseResponse succeeded but IsResponseComplete said incomplete")
		}

		if completeErr != nil && parseErr == nil {
			t.Errorf("IsResponseComplete errored (%v) but ParseResponse succeeded", completeErr)
		}

		if parseErr == nil && resp == nil {
			t.Error("ParseResponse returned no error but returned nil response")
		}
	})
}
