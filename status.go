package qh

const (
	StatusMultipleChoices   = 300
	StatusMovedPermanently  = 301
	StatusFound             = 302
	StatusTemporaryRedirect = 307
	StatusPermanentRedirect = 308
)

// map common HTTP status codes to a compact wire format, ordered by frequency
var StatusToCompact = map[int]uint8{
	// 1xx Informational
	100: 27, // Continue
	101: 28, // Switching Protocols
	102: 29, // Processing
	103: 30, // Early Hints

	// 2xx Success
	200: 0,  // OK
	201: 10, // Created
	202: 11, // Accepted
	204: 12, // No Content
	205: 31, // Reset Content
	206: 13, // Partial Content
	207: 32, // Multi-Status
	208: 33, // Already Reported
	226: 34, // IM Used

	// 3xx Redirection
	300: 35, // Multiple Choices
	301: 7,  // Moved Permanently
	302: 3,  // Found (redirect)
	303: 36, // See Other
	304: 8,  // Not Modified
	305: 37, // Use Proxy
	307: 14, // Temporary Redirect
	308: 15, // Permanent Redirect

	// 4xx Client Error
	400: 4,  // Bad Request
	401: 6,  // Unauthorized
	402: 38, // Payment Required
	403: 5,  // Forbidden
	404: 1,  // Not Found
	405: 39, // Method Not Allowed
	406: 40, // Not Acceptable
	407: 41, // Proxy Authentication Required
	408: 42, // Request Timeout
	409: 16, // Conflict
	410: 17, // Gone
	411: 43, // Length Required
	412: 18, // Precondition Failed
	413: 19, // Payload Too Large
	414: 20, // URI Too Long
	415: 21, // Unsupported Media Type
	416: 44, // Range Not Satisfiable
	417: 45, // Expectation Failed
	422: 22, // Unprocessable Entity
	429: 23, // Too Many Requests

	// 5xx Server Error
	500: 2,  // Internal Server Error
	502: 24, // Bad Gateway
	503: 9,  // Service Unavailable
	504: 25, // Gateway Timeout
	505: 26, // QH Version Not Supported
	// room for additional codes up until 63
}

var CompactToStatus map[uint8]int // reverse mapping for decoding

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
	// Fallback: use compact code for 500 Internal Server Error for unmapped codes
	return StatusToCompact[500] // 2
}

// convert compact format to HTTP status code
func DecodeStatusCode(compact uint8) int {
	if httpCode, exists := CompactToStatus[compact]; exists {
		return httpCode
	}
	// Fallback: if the compact code is not in our map, it's an unknown or custom code.
	// A safe default is to return a generic server error.
	return 500 // Internal Server Error
}
