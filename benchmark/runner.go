package benchmark

func RunBenchmarks() []BenchmarkResult {
	testCases := GetTestCases()
	results := make([]BenchmarkResult, len(testCases))

	for i, tc := range testCases {
		results[i] = BenchmarkResult{
			TestCase: tc,
			QH:       EncodeQH(tc),
			HTTP1:    EncodeHTTP1(tc),
			HTTP2:    EncodeHTTP2(tc),
			HTTP3:    EncodeHTTP3(tc),
		}
	}

	return results
}

func CalculateSummary(results []BenchmarkResult) Summary {
	var qhTotal, http1Total, http2Total, http3Total int

	for _, r := range results {
		qhTotal += r.QH.TotalSize
		http1Total += r.HTTP1.TotalSize
		http2Total += r.HTTP2.TotalSize
		http3Total += r.HTTP3.TotalSize
	}

	count := len(results)
	qhAvg := float64(qhTotal) / float64(count)
	http1Avg := float64(http1Total) / float64(count)
	http2Avg := float64(http2Total) / float64(count)
	http3Avg := float64(http3Total) / float64(count)

	return Summary{
		TotalTests:      count,
		QHTotalBytes:    qhTotal,
		HTTP1TotalBytes: http1Total,
		HTTP2TotalBytes: http2Total,
		HTTP3TotalBytes: http3Total,
		QHAvgBytes:      qhAvg,
		HTTP1AvgBytes:   http1Avg,
		HTTP2AvgBytes:   http2Avg,
		HTTP3AvgBytes:   http3Avg,
		QHVsHTTP1Ratio:  float64(qhTotal) / float64(http1Total) * 100,
		QHVsHTTP2Ratio:  float64(qhTotal) / float64(http2Total) * 100,
		QHVsHTTP3Ratio:  float64(qhTotal) / float64(http3Total) * 100,
	}
}

func GetDetailedResults(results []BenchmarkResult) []DetailedResult {
	detailed := make([]DetailedResult, len(results))

	for i, r := range results {
		detailed[i] = DetailedResult{
			Name:            r.TestCase.Name,
			QHReqBytes:      r.QH.RequestSize,
			QHRespBytes:     r.QH.ResponseSize,
			QHTotalBytes:    r.QH.TotalSize,
			HTTP1ReqBytes:   r.HTTP1.RequestSize,
			HTTP1RespBytes:  r.HTTP1.ResponseSize,
			HTTP1TotalBytes: r.HTTP1.TotalSize,
			HTTP2ReqBytes:   r.HTTP2.RequestSize,
			HTTP2RespBytes:  r.HTTP2.ResponseSize,
			HTTP2TotalBytes: r.HTTP2.TotalSize,
			HTTP3ReqBytes:   r.HTTP3.RequestSize,
			HTTP3RespBytes:  r.HTTP3.ResponseSize,
			HTTP3TotalBytes: r.HTTP3.TotalSize,
			QHVsHTTP1Ratio:  float64(r.QH.TotalSize) / float64(r.HTTP1.TotalSize) * 100,
			QHVsHTTP2Ratio:  float64(r.QH.TotalSize) / float64(r.HTTP2.TotalSize) * 100,
			QHVsHTTP3Ratio:  float64(r.QH.TotalSize) / float64(r.HTTP3.TotalSize) * 100,
		}
	}

	return detailed
}
