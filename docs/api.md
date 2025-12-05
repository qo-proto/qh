# API Documentation

## Server

### Response Function

QH uses a single `Response()` function that accepts headers as a string map:

```go
// Minimal response with automatic Content-Length
server.Response(200, []byte(`{"message": "success"}`), map[string]string{
    "Content-Type": "application/json",
})

// Response with multiple headers
server.Response(200, []byte(body), map[string]string{
    "Content-Type":  "application/json",
    "Cache-Control": "max-age=3600",
    "Date":          strconv.FormatInt(time.Now().Unix(), 10),
    "ETag":          "\"abc123\"",
})

// Convenience methods (automatically set Content-Type)
server.TextResponse(200, "Hello, World!")        // Content-Type: text/plain
server.TextResponse(404, "Not Found")            // Content-Type: text/plain
server.JSONResponse(200, `{"data": "value"}`)    // Content-Type: application/json
```

**Notes**:

- `Content-Length` is automatically calculated and set
- `Date` header is optional, set it manually if needed for caching
- Use standard MIME type strings for Content-Type (e.g., `"application/json"`, `"text/plain"`)

## Client

QH supports the following HTTP methods for RESTful APIs:

### GET Request

```go
headers := map[string]string{
    "Accept": "text/html,application/json,text/plain",
}
response, err := client.GET("example.com", "/api/data", headers)
```

### HEAD Request

```go
headers := map[string]string{
    "Accept": "application/json,text/plain",
}
response, err := client.HEAD("example.com", "/api/user", headers)
// Response contains headers but no body
```

### POST Request

```go
body := []byte(`{"name": "test"}`)
headers := map[string]string{
    "Accept":       "application/json,text/plain",
    "Content-Type": "application/json",
}
response, err := client.POST("example.com", "/submit", body, headers)
```

### PUT Request

```go
body := []byte(`{"name": "test", "id": 123}`)
headers := map[string]string{
    "Accept":       "application/json,text/plain",
    "Content-Type": "application/json",
}
response, err := client.PUT("example.com", "/api/user", body, headers)
```

### PATCH Request

```go
body := []byte(`{"name": "updated"}`)
headers := map[string]string{
    "Accept":       "application/json,text/plain",
    "Content-Type": "application/json",
}
response, err := client.PATCH("example.com", "/api/user", body, headers)
```

### DELETE Request

```go
headers := map[string]string{
    "Accept": "application/json,text/plain",
}
response, err := client.DELETE("example.com", "/api/user", headers)
```

**Notes**:

- `Content-Length` is automatically set for POST, PUT, and PATCH requests
- Body is `[]byte` (convert strings with `[]byte()`)
- GET, HEAD, and DELETE methods don't have a body parameter

### Compression

**Note:** QH currently only supports compression for responses, not requests. This matches how most HTTP traffic operates, as request bodies are typically small.

#### Default Behavior

The client **automatically** adds the `Accept-Encoding: zstd, br, gzip` header to all requests. The server supports the same encodings by default and uses the first client-preferred encoding.

```go
// Compression happens automatically
response, err := client.GET("example.com", "/api/data", nil)
// Client sends: Accept-Encoding: zstd, br, gzip
// Server compresses with zstd (if beneficial)
```

#### Override Client Preferences

Specify different encodings or change the preference order:

```go
headers := map[string]string{
    "Accept-Encoding": "gzip, br",  // Only accept gzip or brotli
}
response, err := client.GET("example.com", "/data", headers)
// Server will use gzip (client's first choice), not zstd
```

#### Disable Compression

Set the `Accept-Encoding` header to an empty string:

```go
headers := map[string]string{
    "Accept-Encoding": "",  // No compression
}
response, err := client.GET("example.com", "/data", headers)
// Server will send uncompressed response
```

#### Server Compression Rules

The server only compresses responses when all conditions are met:

1. **Client supports it**: `Accept-Encoding` header is present and non-empty
2. **Size threshold**: Response body is ≥1KB (smaller responses aren't worth compressing)
3. **Content is compressible**: Skips binary content like `application/octet-stream`
4. **Actual savings**: Compressed size must be smaller than original (otherwise uncompressed is sent)

**Notes:**

- QH does not support HTTP quality values (e.g., `gzip;q=0.8`) - use ordering instead
- QH does not support wildcard encodings (`*`)
- QH does not support `identity` encoding - use empty string `""` to disable compression

## Debugging and Analysis

### Keylog Support (Wireshark Decryption)

QH provides optional keylog support for decrypting QOTP traffic in Wireshark. This requires building with the `keylog` build tag.

#### Server-side Keylog

Server-side keylog is recommended as it has access to all encryption keys during connection acceptance:

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

The keylog file will contain entries in the format:

```
QOTP_SHARED_SECRET <connId_hex> <secret_hex>
```

**Notes:**

- Keylog is only available when building with `-tags keylog`
- Without the build tag, keylog functions are no-ops
- The keylog format is compatible with the QOTP Wireshark dissector

## DNS

### Server Configuration

#### Generating the DNS Record

When a QH server starts, it automatically generates an X25519 key pair and logs the public key in DNS TXT record format (`internal/server/server.go`):

```
Server public key for DNS: v=0;k=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnop=
```

#### Deterministic Keys with Seeds

### Client Behavior

#### Key Discovery Flow

When a QH client connects to a server:

1. **Address Resolution**: Client splits the target address into hostname and port
2. **Concurrent DNS Lookups**: Client performs two DNS lookups in parallel:
   - DNS A/AAAA record lookup to resolve hostname to IP address
   - DNS TXT record lookup for `_qotp.<hostname>` to find the server's public key
   - Both lookups complete before proceeding
3. **Key Validation** (if TXT record found):
   - Parse the TXT record format `v=0;k=<base64-key>`
   - Verify protocol version matches (currently 0)
   - Decode the base64 public key
4. **0-RTT Connection** (if valid key found):
   - Client establishes connection with the server's public key
   - First request includes data immediately (no handshake round-trip)
5. **1-RTT Fallback** (if no key or validation fails):
   - Client negotiates keys with the server during connection
   - Adds one round-trip time for key exchange
   - DNS lookup failures are non-fatal and automatically trigger this fallback

#### Example DNS Query

To query for a QH server's public key:

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
