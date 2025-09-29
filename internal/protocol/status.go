package protocol

// map common HTTP status codes to a compact wire format, ordered by frequency
var StatusToCompact = map[int]uint8{
	200: 1,  // OK
	404: 2,  // Not Found
	500: 3,  // Internal Server Error
	302: 4,  // Found (redirect)
	400: 5,  // Bad Request
	403: 6,  // Forbidden
	401: 7,  // Unauthorized
	301: 8,  // Moved Permanently
	304: 9,  // Not Modified
	503: 10, // Service Unavailable
	201: 11, // Created
	202: 12, // Accepted
	204: 13, // No Content
	206: 14, // Partial Content
	307: 15, // Temporary Redirect
	308: 16, // Permanent Redirect
	409: 17, // Conflict
	410: 18, // Gone
	412: 19, // Precondition Failed
	413: 20, // Payload Too Large
	414: 21, // URI Too Long
	415: 22, // Unsupported Media Type
	422: 23, // Unprocessable Entity
	429: 24, // Too Many Requests
	502: 25, // Bad Gateway
	504: 26, // Gateway Timeout
	505: 27, // QH Version Not Supported
	100: 31, // Continue
	101: 32, // Switching Protocols
	102: 33, // Processing
	103: 34, // Early Hints
	205: 35, // Reset Content
	207: 36, // Multi-Status
	208: 37, // Already Reported
	226: 38, // IM Used
	300: 39, // Multiple Choices
	303: 40, // See Other
	305: 41, // Use Proxy
	402: 42, // Payment Required
	405: 43, // Method Not Allowed
	406: 44, // Not Acceptable
	407: 45, // Proxy Authentication Required
	408: 46, // Request Timeout
	411: 47, // Length Required
	416: 48, // Range Not Satisfiable
	417: 49, // Expectation Failed
	// room for additional codes up until 99
}

var CompactToStatus map[uint8]int // reverse mapping for decoding

// TODO: maybe use manual reverse map over init method
func init() {
	CompactToStatus = make(map[uint8]int, len(StatusToCompact))
	for httpCode, compactCode := range StatusToCompact {
		CompactToStatus[compactCode] = httpCode
	}
}

// convert HTTP status code to compact format
func EncodeStatusCode(httpCode int) uint8 {
	if compact, exists := StatusToCompact[httpCode]; exists {
		return compact
	}
	// Fallback: use 500 Internal Server Error for unmapped codes
	return StatusToCompact[500]
}

// convert compact format to HTTP status code
func DecodeStatusCode(compact uint8) int {
	if httpCode, exists := CompactToStatus[compact]; exists {
		return httpCode
	}
	// Fallback: treat as direct HTTP code
	return int(compact)
}
