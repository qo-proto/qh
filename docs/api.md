# API Documentation

## Server

### Response Function

QH uses a single `Response()` function that accepts headers as a string map:

```go
// Minimal response with automatic Content-Length
server.Response(200, []byte(`{"message": "success"}`), map[string]string{
    "Content-Type": strconv.Itoa(int(protocol.JSON)),
})

// Response with multiple headers
server.Response(200, []byte(body), map[string]string{
    "Content-Type":  strconv.Itoa(int(protocol.JSON)),
    "Cache-Control": "max-age=3600",
    "Date":          strconv.FormatInt(time.Now().Unix(), 10),
    "ETag":          "\"abc123\"",
})

// Convenience methods (automatically set Content-Type)
server.TextResponse(200, "Hello, World!")        // Content-Type: 1 (text/plain)
server.TextResponse(404, "Not Found")            // Content-Type: 1 (text/plain)
server.JSONResponse(200, `{"data": "value"}`)    // Content-Type: 2 (application/json)
```

**Notes**:

- `Content-Length` is automatically calculated and set
- `Date` header is optional, set it manually if needed for caching
- All header values are strings (use `strconv.Itoa()` for numeric values)

**Example:**

```go
headers := map[string]string{
    "Content-Type": strconv.Itoa(int(protocol.JSON)),  // "2"
}
```

## Client

### GET Request

```go
headers := map[string]string{
    "Accept": "3,2,1",  // HTML, JSON, text/plain (in order of preference)
}
response, err := client.GET("example.com", "/api/data", headers)
```

### POST Request

```go
body := []byte(`{"name": "test"}`)
headers := map[string]string{
    "Accept":       "2,1",  // JSON, text/plain
    "Content-Type": strconv.Itoa(int(protocol.JSON)),  // "2"
}
response, err := client.POST("example.com", "/submit", body, headers)
```

**Notes**:

- `Content-Length` is automatically set for POST requests
- Body is `[]byte` (convert strings with `[]byte()`)

### Compression

**Note:** QH currently only supports compression for responses, not requests. This matches how most HTTP traffic operates, as request bodies are typically small.

The client **automatically** adds the `Accept-Encoding: zstd, br, gzip` header to all requests. The server supports the same encodings by default and uses the first client-preferred encoding.

**Example - Override client preferences:**

```go
headers := map[string]string{
    "Accept-Encoding": "gzip, br",  // Only accept gzip or brotli
}
// Server will use gzip (client's first choice), not zstd (server default)
response, err := client.GET("example.com", "/data", headers)
```

**To disable compression:**

```go
headers := map[string]string{
    "Accept-Encoding": "",  // Explicitly disable compression
}
response, err := client.GET("example.com", "/data", headers)
```

**Notes:**

- Server only compresses responses when beneficial (≥1KB, non-binary content, actual size savings)
- The `deflate` encoding is available in the compression library but not enabled by default

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
