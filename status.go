package qh

const (
	// 1xx Informational
	StatusContinue           = 100
	StatusSwitchingProtocols = 101
	StatusProcessing         = 102
	StatusEarlyHints         = 103

	// 2xx Success
	StatusOK              = 200
	StatusCreated         = 201
	StatusAccepted        = 202
	StatusNoContent       = 204
	StatusResetContent    = 205
	StatusPartialContent  = 206
	StatusMultiStatus     = 207
	StatusAlreadyReported = 208
	StatusIMUsed          = 226

	// 3xx Redirection
	StatusMultipleChoices   = 300
	StatusMovedPermanently  = 301
	StatusFound             = 302
	StatusSeeOther          = 303
	StatusNotModified       = 304
	StatusUseProxy          = 305
	StatusTemporaryRedirect = 307
	StatusPermanentRedirect = 308

	// 4xx Client Errors
	StatusBadRequest           = 400
	StatusUnauthorized         = 401
	StatusPaymentRequired      = 402
	StatusForbidden            = 403
	StatusNotFound             = 404
	StatusMethodNotAllowed     = 405
	StatusNotAcceptable        = 406
	StatusProxyAuthRequired    = 407
	StatusRequestTimeout       = 408
	StatusConflict             = 409
	StatusGone                 = 410
	StatusLengthRequired       = 411
	StatusPreconditionFailed   = 412
	StatusPayloadTooLarge      = 413
	StatusURITooLong           = 414
	StatusUnsupportedMediaType = 415
	StatusRangeNotSatisfiable  = 416
	StatusExpectationFailed    = 417
	StatusUnprocessableEntity  = 422
	StatusTooManyRequests      = 429

	// 5xx Server Error
	StatusInternalServerError   = 500
	StatusBadGateway            = 502
	StatusServiceUnavailable    = 503
	StatusGatewayTimeout        = 504
	StatusQHVersionNotSupported = 505
)

// map common HTTP status codes to a compact wire format, ordered by frequency
var StatusToCompact = map[int]uint8{
	// 1xx Informational
	100: 10, // Continue
	101: 11, // Switching Protocols
	102: 12, // Processing
	103: 13, // Early Hints

	// 2xx Success
	200: 20, // OK
	201: 21, // Created
	202: 22, // Accepted
	204: 24, // No Content
	205: 25, // Reset Content
	206: 26, // Partial Content
	207: 27, // Multi-Status
	208: 28, // Already Reported
	226: 29, // IM Used

	// 3xx Redirection
	300: 30, // Multiple Choices
	301: 31, // Moved Permanently
	302: 32, // Found (redirect)
	303: 33, // See Other
	304: 34, // Not Modified
	305: 35, // Use Proxy
	307: 37, // Temporary Redirect
	308: 38, // Permanent Redirect

	// 4xx Client Error
	400: 40, // Bad Request
	401: 41, // Unauthorized
	402: 42, // Payment Required
	403: 43, // Forbidden
	404: 44, // Not Found
	405: 45, // Method Not Allowed
	406: 46, // Not Acceptable
	407: 47, // Proxy Authentication Required
	408: 48, // Request Timeout
	409: 49, // Conflict
	410: 80, // Gone
	411: 81, // Length Required
	412: 81, // Precondition Failed
	413: 83, // Payload Too Large
	414: 84, // URI Too Long
	415: 85, // Unsupported Media Type
	416: 86, // Range Not Satisfiable
	417: 87, // Expectation Failed
	422: 88, // Unprocessable Entity
	429: 89, // Too Many Requests

	// 5xx Server Error
	500: 50, // Internal Server Error
	502: 52, // Bad Gateway
	503: 53, // Service Unavailable
	504: 54, // Gateway Timeout
	505: 55, // QH Version Not Supported
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
	return StatusToCompact[StatusInternalServerError]
}

// convert compact format to HTTP status code
func DecodeStatusCode(compact uint8) int {
	if httpCode, exists := CompactToStatus[compact]; exists {
		return httpCode
	}
	// Fallback: if the compact code is not in our map, it's an unknown or custom code.
	// A safe default is to return a generic server error.
	return StatusInternalServerError
}
