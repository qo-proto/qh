package protocol

// TODO: default print error messages in the body like in http (no need to set body message manually for statusCodes)

// map common HTTP status codes to a compact wire format, ordered by frequency
var StatusToCompact = map[int]uint8{
	200: 0,  // OK
	404: 1,  // Not Found
	500: 2,  // Internal Server Error
	302: 3,  // Found (redirect)
	400: 4,  // Bad Request
	403: 5,  // Forbidden
	401: 6,  // Unauthorized
	301: 7,  // Moved Permanently
	304: 8,  // Not Modified
	503: 9,  // Service Unavailable
	201: 10, // Created
	202: 11, // Accepted
	204: 12, // No Content
	206: 13, // Partial Content
	307: 14, // Temporary Redirect
	308: 15, // Permanent Redirect
	409: 16, // Conflict
	410: 17, // Gone
	412: 18, // Precondition Failed
	413: 19, // Payload Too Large
	414: 20, // URI Too Long
	415: 21, // Unsupported Media Type
	422: 22, // Unprocessable Entity
	429: 23, // Too Many Requests
	502: 24, // Bad Gateway
	504: 25, // Gateway Timeout
	505: 26, // QH Version Not Supported
	100: 27, // Continue
	101: 28, // Switching Protocols
	102: 29, // Processing
	103: 30, // Early Hints
	205: 31, // Reset Content
	207: 32, // Multi-Status
	208: 33, // Already Reported
	226: 34, // IM Used
	300: 35, // Multiple Choices
	303: 36, // See Other
	305: 37, // Use Proxy
	402: 38, // Payment Required
	405: 39, // Method Not Allowed
	406: 40, // Not Acceptable
	407: 41, // Proxy Authentication Required
	408: 42, // Request Timeout
	411: 43, // Length Required
	416: 44, // Range Not Satisfiable
	417: 45, // Expectation Failed
	// room for additional codes up until 63
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
