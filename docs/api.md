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
