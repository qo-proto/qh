package benchmark

func RunBenchmarks() CategorizedResults {
	edgeCases := GetEdgeCaseTestCases()
	edgeResults := make([]BenchmarkResult, len(edgeCases))
	for i, tc := range edgeCases {
		edgeResults[i] = BenchmarkResult{
			TestCase: tc,
			QH:       EncodeQH(tc),
			HTTP1:    EncodeHTTP1(tc),
			HTTP2:    EncodeHTTP2(tc),
			HTTP3:    EncodeHTTP3(tc),
		}
	}

	trafficCases := GetHTTPTrafficTestCases()
	trafficResults := make([]BenchmarkResult, len(trafficCases))
	for i, tc := range trafficCases {
		trafficResults[i] = BenchmarkResult{
			TestCase: tc,
			QH:       EncodeQH(tc),
			HTTP1:    EncodeHTTP1(tc),
			HTTP2:    EncodeHTTP2(tc),
			HTTP3:    EncodeHTTP3(tc),
		}
	}

	allResults := make([]BenchmarkResult, 0, len(edgeResults)+len(trafficResults))
	allResults = append(allResults, trafficResults...)
	allResults = append(allResults, edgeResults...)

	return CategorizedResults{
		EdgeCases: edgeResults,
		Traffic:   trafficResults,
		All:       allResults,
	}
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
		name       string
		minSize    int
		maxSize    int
		qhTotal    int
		http1Total int
		http2Total int
		http3Total int
		count      int
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

func FindBestWorstCases(results []BenchmarkResult) PerformanceBounds {
	if len(results) == 0 {
		return PerformanceBounds{HasData: false}
	}

	// Initialize with first result
	http1Best := results[0]
	http1Worst := results[0]
	http1BestRatio := float64(results[0].QH.TotalSize) / float64(results[0].HTTP1.TotalSize)
	http1WorstRatio := http1BestRatio

	http2Best := results[0]
	http2Worst := results[0]
	http2BestRatio := float64(results[0].QH.TotalSize) / float64(results[0].HTTP2.TotalSize)
	http2WorstRatio := http2BestRatio

	http3Best := results[0]
	http3Worst := results[0]
	http3BestRatio := float64(results[0].QH.TotalSize) / float64(results[0].HTTP3.TotalSize)
	http3WorstRatio := http3BestRatio

	for i := 1; i < len(results); i++ {
		// vs HTTP/1.1
		ratio1 := float64(results[i].QH.TotalSize) / float64(results[i].HTTP1.TotalSize)
		if ratio1 < http1BestRatio {
			http1BestRatio = ratio1
			http1Best = results[i]
		}
		if ratio1 > http1WorstRatio {
			http1WorstRatio = ratio1
			http1Worst = results[i]
		}

		// vs HTTP/2
		ratio2 := float64(results[i].QH.TotalSize) / float64(results[i].HTTP2.TotalSize)
		if ratio2 < http2BestRatio {
			http2BestRatio = ratio2
			http2Best = results[i]
		}
		if ratio2 > http2WorstRatio {
			http2WorstRatio = ratio2
			http2Worst = results[i]
		}

		// vs HTTP/3
		ratio3 := float64(results[i].QH.TotalSize) / float64(results[i].HTTP3.TotalSize)
		if ratio3 < http3BestRatio {
			http3BestRatio = ratio3
			http3Best = results[i]
		}
		if ratio3 > http3WorstRatio {
			http3WorstRatio = ratio3
			http3Worst = results[i]
		}
	}

	return PerformanceBounds{
		VsHTTP1: ProtocolBounds{
			BestCase:  http1Best,
			WorstCase: http1Worst,
		},
		VsHTTP2: ProtocolBounds{
			BestCase:  http2Best,
			WorstCase: http2Worst,
		},
		VsHTTP3: ProtocolBounds{
			BestCase:  http3Best,
			WorstCase: http3Worst,
		},
		HasData: true,
	}
}

func CalculateHeaderOnlyAnalysis(results []BenchmarkResult) HeaderOnlyAnalysis {
	var qhReqHeaders, http1ReqHeaders, http2ReqHeaders, http3ReqHeaders int
	var qhRespHeaders, http1RespHeaders, http2RespHeaders, http3RespHeaders int

	for _, r := range results {
		qhReqHeaders += r.QH.RequestHeaderSize
		http1ReqHeaders += r.HTTP1.RequestHeaderSize
		http2ReqHeaders += r.HTTP2.RequestHeaderSize
		http3ReqHeaders += r.HTTP3.RequestHeaderSize

		qhRespHeaders += r.QH.ResponseHeaderSize
		http1RespHeaders += r.HTTP1.ResponseHeaderSize
		http2RespHeaders += r.HTTP2.ResponseHeaderSize
		http3RespHeaders += r.HTTP3.ResponseHeaderSize
	}

	count := len(results)

	qhReqAvg := float64(qhReqHeaders) / float64(count)
	http1ReqAvg := float64(http1ReqHeaders) / float64(count)
	http2ReqAvg := float64(http2ReqHeaders) / float64(count)
	http3ReqAvg := float64(http3ReqHeaders) / float64(count)

	qhRespAvg := float64(qhRespHeaders) / float64(count)
	http1RespAvg := float64(http1RespHeaders) / float64(count)
	http2RespAvg := float64(http2RespHeaders) / float64(count)
	http3RespAvg := float64(http3RespHeaders) / float64(count)

	return HeaderOnlyAnalysis{
		QHReqHeaderAvg:     qhReqAvg,
		HTTP1ReqHeaderAvg:  http1ReqAvg,
		HTTP2ReqHeaderAvg:  http2ReqAvg,
		HTTP3ReqHeaderAvg:  http3ReqAvg,
		QHRespHeaderAvg:    qhRespAvg,
		HTTP1RespHeaderAvg: http1RespAvg,
		HTTP2RespHeaderAvg: http2RespAvg,
		HTTP3RespHeaderAvg: http3RespAvg,
		QHReqVsHTTP1Ratio:  (qhReqAvg / http1ReqAvg) * 100,
		QHReqVsHTTP2Ratio:  (qhReqAvg / http2ReqAvg) * 100,
		QHReqVsHTTP3Ratio:  (qhReqAvg / http3ReqAvg) * 100,
		QHRespVsHTTP1Ratio: (qhRespAvg / http1RespAvg) * 100,
		QHRespVsHTTP2Ratio: (qhRespAvg / http2RespAvg) * 100,
		QHRespVsHTTP3Ratio: (qhRespAvg / http3RespAvg) * 100,
	}
}
