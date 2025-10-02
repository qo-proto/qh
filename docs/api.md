# API Documentation

## Server

### Basic Response

```go
// Minimal response with required headers only (Content-Type and Content-Length)
server.Response(200, protocol.JSON, []byte(`{"message": "success"}`))
server.Response(404, protocol.TextPlain, []byte("Not Found"))

// Convenience methods (all use Response() internally)
server.TextResponse(200, "Hello, World!")
server.JSONResponse(200, `{"data": "value"}`)
server.HTMLResponse(200, "<h1>Welcome</h1>")
server.ErrorResponse(404, "Not Found")
server.OKResponse(protocol.TextPlain, []byte("Success"))
```

### ResponseWithHeaders

```go
server.ResponseWithHeaders(200, protocol.JSON, body, map[int]string{
    protocol.RespHeaderCacheControl: "max-age=3600",
    protocol.RespHeaderCORS: "*",
    protocol.RespHeaderETag: "\"abc123\"",
})
```

**Notes**:

- `RespHeaderDate` is always set automatically and cannot be overridden
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
