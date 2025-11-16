package benchmark

type TestCase struct {
	Name        string
	Description string
	Request     RequestData
	Response    ResponseData
}

type RequestData struct {
	Method  string
	Host    string
	Path    string
	Headers map[string]string
	Body    string `json:"body,omitempty"`
}

type ResponseData struct {
	StatusCode int
	Headers    map[string]string
	Body       string `json:"body,omitempty"`
}

type BenchmarkResult struct {
	TestCase TestCase
	QH       EncodedResult
	HTTP1    EncodedResult
	HTTP2    EncodedResult
	HTTP3    EncodedResult
}

type EncodedResult struct {
	RequestBytes  []byte
	ResponseBytes []byte
	RequestSize   int
	ResponseSize  int
	TotalSize     int
}

type Summary struct {
	TotalTests      int
	QHTotalBytes    int
	HTTP1TotalBytes int
	HTTP2TotalBytes int
	HTTP3TotalBytes int
	QHAvgBytes      float64
	HTTP1AvgBytes   float64
	HTTP2AvgBytes   float64
	HTTP3AvgBytes   float64
	QHVsHTTP1Ratio  float64
	QHVsHTTP2Ratio  float64
	QHVsHTTP3Ratio  float64
}

type DetailedResult struct {
	Name            string
	QHReqBytes      int
	QHRespBytes     int
	QHTotalBytes    int
	HTTP1ReqBytes   int
	HTTP1RespBytes  int
	HTTP1TotalBytes int
	HTTP2ReqBytes   int
	HTTP2RespBytes  int
	HTTP2TotalBytes int
	HTTP3ReqBytes   int
	HTTP3RespBytes  int
	HTTP3TotalBytes int
	QHVsHTTP1Ratio  float64
	QHVsHTTP2Ratio  float64
	QHVsHTTP3Ratio  float64
}

type SizeCategory struct {
	Name            string
	Count           int
	QHAvg           float64
	HTTP1Avg        float64
	HTTP2Avg        float64
	HTTP3Avg        float64
	QHVsHTTP1Ratio  float64
	QHVsHTTP2Ratio  float64
	QHVsHTTP3Ratio  float64
}

type HeaderAnalysis struct {
	QHAvgHeaders      float64
	HTTP1AvgHeaders   float64
	HTTP2AvgHeaders   float64
	HTTP3AvgHeaders   float64
	QHTotalHeaders    int
	HTTP1TotalHeaders int
	HTTP2TotalHeaders int
	HTTP3TotalHeaders int
	QHVsHTTP1Ratio    float64
	QHVsHTTP2Ratio    float64
	QHVsHTTP3Ratio    float64
}
