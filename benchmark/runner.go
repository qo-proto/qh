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

func CalculateSizeCategories(results []BenchmarkResult) []SizeCategory {
	categories := []struct {
		name      string
		minSize   int
		maxSize   int
		qhTotal   int
		http1Total int
		http2Total int
		http3Total int
		count     int
	}{
		{name: "Tiny (<1KB)", minSize: 0, maxSize: 1024},
		{name: "Small (1-10KB)", minSize: 1024, maxSize: 10240},
		{name: "Medium (10-100KB)", minSize: 10240, maxSize: 102400},
		{name: "Large (>100KB)", minSize: 102400, maxSize: 1<<31 - 1},
	}

	for _, r := range results {
		totalSize := r.QH.TotalSize
		for i := range categories {
			if totalSize >= categories[i].minSize && totalSize < categories[i].maxSize {
				categories[i].qhTotal += r.QH.TotalSize
				categories[i].http1Total += r.HTTP1.TotalSize
				categories[i].http2Total += r.HTTP2.TotalSize
				categories[i].http3Total += r.HTTP3.TotalSize
				categories[i].count++
				break
			}
		}
	}

	sizeCategories := make([]SizeCategory, 0)
	for _, c := range categories {
		if c.count == 0 {
			continue
		}
		qhAvg := float64(c.qhTotal) / float64(c.count)
		http1Avg := float64(c.http1Total) / float64(c.count)
		http2Avg := float64(c.http2Total) / float64(c.count)
		http3Avg := float64(c.http3Total) / float64(c.count)

		sizeCategories = append(sizeCategories, SizeCategory{
			Name:           c.name,
			Count:          c.count,
			QHAvg:          qhAvg,
			HTTP1Avg:       http1Avg,
			HTTP2Avg:       http2Avg,
			HTTP3Avg:       http3Avg,
			QHVsHTTP1Ratio: (qhAvg / http1Avg) * 100,
			QHVsHTTP2Ratio: (qhAvg / http2Avg) * 100,
			QHVsHTTP3Ratio: (qhAvg / http3Avg) * 100,
		})
	}

	return sizeCategories
}

func CalculateHeaderAnalysis(results []BenchmarkResult) HeaderAnalysis {
	var qhTotal, http1Total, http2Total, http3Total int

	for _, r := range results {
		qhTotal += r.QH.RequestSize
		http1Total += r.HTTP1.RequestSize
		http2Total += r.HTTP2.RequestSize
		http3Total += r.HTTP3.RequestSize
	}

	count := len(results)
	qhAvg := float64(qhTotal) / float64(count)
	http1Avg := float64(http1Total) / float64(count)
	http2Avg := float64(http2Total) / float64(count)
	http3Avg := float64(http3Total) / float64(count)

	return HeaderAnalysis{
		QHAvgHeaders:      qhAvg,
		HTTP1AvgHeaders:   http1Avg,
		HTTP2AvgHeaders:   http2Avg,
		HTTP3AvgHeaders:   http3Avg,
		QHTotalHeaders:    qhTotal,
		HTTP1TotalHeaders: http1Total,
		HTTP2TotalHeaders: http2Total,
		HTTP3TotalHeaders: http3Total,
		QHVsHTTP1Ratio:    (qhAvg / http1Avg) * 100,
		QHVsHTTP2Ratio:    (qhAvg / http2Avg) * 100,
		QHVsHTTP3Ratio:    (qhAvg / http3Avg) * 100,
	}
}
