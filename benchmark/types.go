package benchmark

type TestCase struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Request     RequestData  `json:"request"`
	Response    ResponseData `json:"response"`
}

type RequestData struct {
	Method  string            `json:"method"`
	Host    string            `json:"host"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}

type BenchmarkResult struct {
	TestCase TestCase
	QH       EncodedResult
	HTTP1    EncodedResult
	HTTP2    EncodedResult
	HTTP3    EncodedResult
}

type EncodedResult struct {
	RequestBytes       []byte
	ResponseBytes      []byte
	RequestSize        int
	ResponseSize       int
	TotalSize          int
	RequestHeaderSize  int
	ResponseHeaderSize int
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
	Name           string
	Count          int
	QHAvg          float64
	HTTP1Avg       float64
	HTTP2Avg       float64
	HTTP3Avg       float64
	QHVsHTTP1Ratio float64
	QHVsHTTP2Ratio float64
	QHVsHTTP3Ratio float64
}

type HeaderOnlyAnalysis struct {
	QHReqHeaderAvg     float64
	HTTP1ReqHeaderAvg  float64
	HTTP2ReqHeaderAvg  float64
	HTTP3ReqHeaderAvg  float64
	QHRespHeaderAvg    float64
	HTTP1RespHeaderAvg float64
	HTTP2RespHeaderAvg float64
	HTTP3RespHeaderAvg float64
	QHReqVsHTTP1Ratio  float64
	QHReqVsHTTP2Ratio  float64
	QHReqVsHTTP3Ratio  float64
	QHRespVsHTTP1Ratio float64
	QHRespVsHTTP2Ratio float64
	QHRespVsHTTP3Ratio float64
}

type ProtocolBounds struct {
	BestCase  BenchmarkResult
	WorstCase BenchmarkResult
}

type PerformanceBounds struct {
	VsHTTP1 ProtocolBounds
	VsHTTP2 ProtocolBounds
	VsHTTP3 ProtocolBounds
	HasData bool
}

type CategorizedResults struct {
	EdgeCases []BenchmarkResult
	Traffic   []BenchmarkResult
	All       []BenchmarkResult
}
