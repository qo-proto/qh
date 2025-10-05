# API Documentation

## Server

### Response Function

QH uses a single `Response()` function that handles both minimal and header-rich responses:

```go
// Minimal response with required headers only (Content-Type and Content-Length)
server.Response(200, protocol.JSON, []byte(`{"message": "success"}`), nil)
server.Response(404, protocol.TextPlain, []byte("Not Found"), nil)

// Response with optional headers (Date auto-added)
server.Response(200, protocol.JSON, body, map[int]string{
    protocol.RespHeaderCacheControl: "max-age=3600",
    protocol.RespHeaderCORS: "*",
    protocol.RespHeaderETag: "\"abc123\"",
})

// Convenience methods (all use Response() with nil headers internally)
server.TextResponse(200, "Hello, World!")        // text/plain
server.TextResponse(404, "Not Found")            // text/plain errors
server.JSONResponse(200, `{"data": "value"}`)    // application/json
```

**Notes**:

- When `headers` contains any optional headers, `RespHeaderDate` is automatically added with current Unix timestamp
- `RespHeaderDate` cannot be overridden by applications to ensure timestamp accuracy
- Gaps between specified header indices are automatically filled with empty strings
- Content-Type (index 0) and Content-Length (index 1) are always set automatically

#### Available Response Header Indices

```go
const (
    RespHeaderContentType     = 0  // Content type (automatically set)
    RespHeaderContentLength   = 1  // Size of response body (automatically set)
    RespHeaderCacheControl    = 2  // Cache-Control directives
    RespHeaderContentEncoding = 3  // Content encoding used (e.g., "gzip")
    RespHeaderAuthorization   = 4  // Authorization info
    RespHeaderCORS            = 5  // Access-Control-Allow-Origin
    RespHeaderETag            = 6  // Entity tag for cache validation
    RespHeaderDate            = 7  // Unix timestamp (automatically set)
    RespHeaderCSP             = 8  // Content-Security-Policy
    RespHeaderContentTypeOpts = 9  // X-Content-Type-Options
    RespHeaderFrameOptions    = 10 // X-Frame-Options
)
```

## Client
