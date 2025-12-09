# API Documentation

For detailed API reference, see the [Go package documentation](https://pkg.go.dev/github.com/qo-proto/qh).

This document covers usage examples and configuration guides.

## Server

### Creating Responses

```go
// Response with headers
qh.NewResponse(200, []byte(`{"message": "success"}`), map[string]string{
    "Content-Type":  "application/json",
    "Cache-Control": "max-age=3600",
})

// Convenience methods
qh.TextResponse(200, "Hello, World!")
qh.JSONResponse(200, `{"data": "value"}`)
```

## Client

### HTTP Methods

```go
// GET, HEAD, DELETE - no body
response, err := client.GET("example.com", "/api/data", headers)
response, err := client.HEAD("example.com", "/api/data", headers)
response, err := client.DELETE("example.com", "/api/data", headers)

// POST, PUT, PATCH - with body
body := []byte(`{"name": "test"}`)
headers := map[string]string{
    "Accept":       "application/json,text/plain",
    "Content-Type": "application/json",
}
response, err := client.POST("example.com", "/submit", body, headers)
response, err := client.PUT("example.com", "/api/user", body, headers)
response, err := client.PATCH("example.com", "/api/user", body, headers)
```

### Compression

QH supports response compression with zstd, brotli, and gzip.

#### Default Behavior

The client automatically adds `Accept-Encoding: zstd, br, gzip` to all requests. The server uses the first client-preferred encoding that it supports.

#### Override or Disable

```go
// Change preference order
headers := map[string]string{
    "Accept-Encoding": "gzip, br",
}

// Disable compression
headers := map[string]string{
    "Accept-Encoding": "",
}
```

#### Server Compression Rules

The server compresses responses when:

1. Client sends a non-empty `Accept-Encoding` header
2. Response body is ≥1KB (configurable via `WithMinCompressionSize`)
3. Content is not binary (`application/octet-stream`)
4. Compressed size is smaller than original

**Limitations:**

- Quality values not supported (e.g., `gzip;q=0.8`) - use ordering instead
- Wildcard `*` and `identity` encodings not supported

## Debugging

### Keylog Support (Wireshark Decryption)

QH provides optional keylog support for decrypting QOTP traffic. Requires the `keylog` build tag.

```go
// Create a keylog file
keylogFile, err := os.Create("qh_server_keylog.txt")
if err != nil {
    log.Fatal(err)
}
defer keylogFile.Close()

// Pass keylog writer to server
srv := qh.NewServer(qh.WithServerKeyLogWriter(keylogFile))
```

**Build and run:**

```bash
go run -tags keylog ./examples/server/main.go
```

The keylog file format is compatible with the QOTP Wireshark dissector:

```
QOTP_SHARED_SECRET <connId_hex> <secret_hex>
```

## DNS-Based Key Exchange

### Server Setup

When a QH server starts, it logs the public key in DNS TXT record format:

```
Server public key for DNS: v=0;k=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnop=
```

Add this as a TXT record at `_qotp.<hostname>`.

### Client Key Discovery

When connecting, the client performs concurrent DNS lookups:

1. **A/AAAA record** - resolve hostname to IP
2. **TXT record at `_qotp.<hostname>`** - retrieve server's public key

If a valid key is found (`v=0;k=<base64-key>`), the client establishes a **0-RTT connection**. Otherwise, it falls back to **1-RTT in-band key exchange**.

### Example DNS Query

```bash
dig TXT _qotp.example.com
```

Expected response:

```
_qotp.example.com.  300  IN  TXT  "v=0;k=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnop="
```

#### Fallback Behavior

DNS lookup failures are non-fatal and the client falls back to in-band key exchange.

**Fallback Scenarios:**

- **TXT record not found**: Client proceeds with 1-RTT handshake using `DialString()`
- **Base64 decode failure**: Client logs warning and falls back to in-band exchange
- **Protocol version mismatch**: Client ignores the key and falls back
- **Any DNS error**: Lookup error is logged but connection continues

#### Implementation Details

- DNS lookup runs in a separate goroutine (`client.go`)
