package benchmark

import (
	"fmt"
	"strings"
)

func fmtPerformanceBounds(bounds PerformanceBounds) string {
	var sb strings.Builder

	// vs HTTP/1.1
	http1BestRatio := float64(bounds.VsHTTP1.BestCase.QH.TotalSize) / float64(bounds.VsHTTP1.BestCase.HTTP1.TotalSize) * 100
	http1WorstRatio := float64(bounds.VsHTTP1.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP1.WorstCase.HTTP1.TotalSize) * 100

	sb.WriteString("  vs HTTP/1.1:\n")
	sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% smaller - QH: %d B vs HTTP/1.1: %d B (%s)\n",
		100-http1BestRatio, bounds.VsHTTP1.BestCase.QH.TotalSize, bounds.VsHTTP1.BestCase.HTTP1.TotalSize, bounds.VsHTTP1.BestCase.TestCase.Name))
	sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% smaller - QH: %d B vs HTTP/1.1: %d B (%s)\n",
		100-http1WorstRatio, bounds.VsHTTP1.WorstCase.QH.TotalSize, bounds.VsHTTP1.WorstCase.HTTP1.TotalSize, bounds.VsHTTP1.WorstCase.TestCase.Name))
	sb.WriteString("\n")

	// vs HTTP/2
	http2BestRatio := float64(bounds.VsHTTP2.BestCase.QH.TotalSize) / float64(bounds.VsHTTP2.BestCase.HTTP2.TotalSize) * 100
	http2WorstRatio := float64(bounds.VsHTTP2.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP2.WorstCase.HTTP2.TotalSize) * 100

	sb.WriteString("  vs HTTP/2:\n")
	if http2BestRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% smaller - QH: %d B vs HTTP/2: %d B (%s)\n",
			100-http2BestRatio, bounds.VsHTTP2.BestCase.QH.TotalSize, bounds.VsHTTP2.BestCase.HTTP2.TotalSize, bounds.VsHTTP2.BestCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% larger - QH: %d B vs HTTP/2: %d B (%s)\n",
			http2BestRatio-100, bounds.VsHTTP2.BestCase.QH.TotalSize, bounds.VsHTTP2.BestCase.HTTP2.TotalSize, bounds.VsHTTP2.BestCase.TestCase.Name))
	}
	if http2WorstRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% smaller - QH: %d B vs HTTP/2: %d B (%s)\n",
			100-http2WorstRatio, bounds.VsHTTP2.WorstCase.QH.TotalSize, bounds.VsHTTP2.WorstCase.HTTP2.TotalSize, bounds.VsHTTP2.WorstCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% larger - QH: %d B vs HTTP/2: %d B (%s)\n",
			http2WorstRatio-100, bounds.VsHTTP2.WorstCase.QH.TotalSize, bounds.VsHTTP2.WorstCase.HTTP2.TotalSize, bounds.VsHTTP2.WorstCase.TestCase.Name))
	}
	sb.WriteString("\n")

	// vs HTTP/3
	http3BestRatio := float64(bounds.VsHTTP3.BestCase.QH.TotalSize) / float64(bounds.VsHTTP3.BestCase.HTTP3.TotalSize) * 100
	http3WorstRatio := float64(bounds.VsHTTP3.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP3.WorstCase.HTTP3.TotalSize) * 100

	sb.WriteString("  vs HTTP/3:\n")
	if http3BestRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% smaller - QH: %d B vs HTTP/3: %d B (%s)\n",
			100-http3BestRatio, bounds.VsHTTP3.BestCase.QH.TotalSize, bounds.VsHTTP3.BestCase.HTTP3.TotalSize, bounds.VsHTTP3.BestCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% larger - QH: %d B vs HTTP/3: %d B (%s)\n",
			http3BestRatio-100, bounds.VsHTTP3.BestCase.QH.TotalSize, bounds.VsHTTP3.BestCase.HTTP3.TotalSize, bounds.VsHTTP3.BestCase.TestCase.Name))
	}
	if http3WorstRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% smaller - QH: %d B vs HTTP/3: %d B (%s)\n",
			100-http3WorstRatio, bounds.VsHTTP3.WorstCase.QH.TotalSize, bounds.VsHTTP3.WorstCase.HTTP3.TotalSize, bounds.VsHTTP3.WorstCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% larger - QH: %d B vs HTTP/3: %d B (%s)\n",
			http3WorstRatio-100, bounds.VsHTTP3.WorstCase.QH.TotalSize, bounds.VsHTTP3.WorstCase.HTTP3.TotalSize, bounds.VsHTTP3.WorstCase.TestCase.Name))
	}

	return sb.String()
}

func fmtEdgeCaseSection(results []BenchmarkResult) string {
	var sb strings.Builder

	summary := CalculateSummary(results)
	sb.WriteString("SUMMARY\n\n")
	sb.WriteString(fmt.Sprintf("  Tests run:         %d (manually selected)\n", summary.TotalTests))
	sb.WriteString(fmt.Sprintf("  QH total:          %d B\n", summary.QHTotalBytes))
	sb.WriteString(fmt.Sprintf("  HTTP/1.1 total:    %d B (%s)\n", summary.HTTP1TotalBytes, fmtDifference(summary.QHVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 total:      %d B (%s)\n", summary.HTTP2TotalBytes, fmtDifference(summary.QHVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 total:      %d B (%s)\n", summary.HTTP3TotalBytes, fmtDifference(summary.QHVsHTTP3Ratio)))
	sb.WriteString("\n\n")

	bounds := FindBestWorstCases(results)
	if bounds.HasData {
		sb.WriteString("PERFORMANCE BOUNDS\n\n")
		sb.WriteString(fmtPerformanceBounds(bounds))
		sb.WriteString("\n")
	}

	sb.WriteString("EDGE CASE DETAILED RESULTS\n\n")
	sb.WriteString(fmtDetailedTable(results))
	sb.WriteString("\n\n")

	// Header analysis for edge cases
	headerOnly := CalculateHeaderOnlyAnalysis(results)
	sb.WriteString("REQUEST/RESPONSE BREAKDOWN\n\n")
	sb.WriteString(fmtHeaderOnlyAnalysis(headerOnly))

	return sb.String()
}

func fmtTrafficSection(results []BenchmarkResult) string {
	var sb strings.Builder

	summary := CalculateSummary(results)
	sb.WriteString("SUMMARY\n\n")
	sb.WriteString(fmt.Sprintf("  Tests run:         %d (collected from actual internet traffic)\n", summary.TotalTests))
	sb.WriteString(fmt.Sprintf("  QH total:          %d B\n", summary.QHTotalBytes))
	sb.WriteString(fmt.Sprintf("  HTTP/1.1 total:    %d B (%s)\n", summary.HTTP1TotalBytes, fmtDifference(summary.QHVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 total:      %d B (%s)\n", summary.HTTP2TotalBytes, fmtDifference(summary.QHVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 total:      %d B (%s)\n", summary.HTTP3TotalBytes, fmtDifference(summary.QHVsHTTP3Ratio)))
	sb.WriteString("\n\n")

	bounds := FindBestWorstCases(results)
	if bounds.HasData {
		sb.WriteString("PERFORMANCE BOUNDS\n\n")
		sb.WriteString(fmtPerformanceBounds(bounds))
		sb.WriteString("\n")
	}

	sizeCategories := CalculateSizeCategories(results)
	sb.WriteString("BREAKDOWN BY SIZE CATEGORY\n\n")
	sb.WriteString(fmtSizeCategories(sizeCategories))
	sb.WriteString("\n\n")

	headerOnly := CalculateHeaderOnlyAnalysis(results)
	sb.WriteString("REQUEST/RESPONSE BREAKDOWN\n\n")
	sb.WriteString(fmtHeaderOnlyAnalysis(headerOnly))

	return sb.String()
}

func fmtCombinedSection(results []BenchmarkResult) string {
	var sb strings.Builder

	summary := CalculateSummary(results)
	sb.WriteString("SUMMARY\n\n")
	sb.WriteString(fmt.Sprintf("  Tests run:         %d\n", summary.TotalTests))
	sb.WriteString(fmt.Sprintf("  QH total:          %d B\n", summary.QHTotalBytes))
	sb.WriteString(fmt.Sprintf("  HTTP/1.1 total:    %d B (%s)\n", summary.HTTP1TotalBytes, fmtDifference(summary.QHVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 total:      %d B (%s)\n", summary.HTTP2TotalBytes, fmtDifference(summary.QHVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 total:      %d B (%s)\n", summary.HTTP3TotalBytes, fmtDifference(summary.QHVsHTTP3Ratio)))
	sb.WriteString("\n\n")

	bounds := FindBestWorstCases(results)
	if bounds.HasData {
		sb.WriteString("PERFORMANCE BOUNDS\n\n")
		sb.WriteString(fmtPerformanceBounds(bounds))
		sb.WriteString("\n")
	}

	sizeCategories := CalculateSizeCategories(results)
	sb.WriteString("BREAKDOWN BY SIZE CATEGORY\n\n")
	sb.WriteString(fmtSizeCategories(sizeCategories))
	sb.WriteString("\n\n")

	headerOnly := CalculateHeaderOnlyAnalysis(results)
	sb.WriteString("REQUEST/RESPONSE BREAKDOWN\n\n")
	sb.WriteString(fmtHeaderOnlyAnalysis(headerOnly))

	return sb.String()
}

func fmtSizeCategories(categories []SizeCategory) string {
	var sb strings.Builder

	sb.WriteString("Category          Count   QH Avg      H1 Avg      H2 Avg      H3 Avg      QH/H1   QH/H2   QH/H3\n")
	sb.WriteString("────────────────────────────────────────────────────────────────────────────────────────────────────\n")

	for _, c := range categories {
		sb.WriteString(fmt.Sprintf("%-16s %6d %10s %10s %10s %10s %6.1f%% %6.1f%% %6.1f%%\n",
			c.Name,
			c.Count,
			fmtBytes(c.QHAvg),
			fmtBytes(c.HTTP1Avg),
			fmtBytes(c.HTTP2Avg),
			fmtBytes(c.HTTP3Avg),
			c.QHVsHTTP1Ratio,
			c.QHVsHTTP2Ratio,
			c.QHVsHTTP3Ratio,
		))
	}

	return sb.String()
}

func fmtHeaderOnlyAnalysis(h HeaderOnlyAnalysis) string {
	var sb strings.Builder

	sb.WriteString("REQUEST HEADERS:\n")
	sb.WriteString(fmt.Sprintf("  QH avg:       %6.0f B  (baseline)\n", h.QHReqHeaderAvg))
	sb.WriteString(fmt.Sprintf("  HTTP/1 avg:   %6.0f B  (%s)\n",
		h.HTTP1ReqHeaderAvg, fmtDifference(h.QHReqVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 avg:   %6.0f B  (%s)\n",
		h.HTTP2ReqHeaderAvg, fmtDifference(h.QHReqVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 avg:   %6.0f B  (%s)\n\n",
		h.HTTP3ReqHeaderAvg, fmtDifference(h.QHReqVsHTTP3Ratio)))

	sb.WriteString("RESPONSE HEADERS:\n")
	sb.WriteString(fmt.Sprintf("  QH avg:       %6.0f B  (baseline)\n", h.QHRespHeaderAvg))
	sb.WriteString(fmt.Sprintf("  HTTP/1 avg:   %6.0f B  (%s)\n",
		h.HTTP1RespHeaderAvg, fmtDifference(h.QHRespVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 avg:   %6.0f B  (%s)\n",
		h.HTTP2RespHeaderAvg, fmtDifference(h.QHRespVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 avg:   %6.0f B  (%s)\n\n",
		h.HTTP3RespHeaderAvg, fmtDifference(h.QHRespVsHTTP3Ratio)))

	qhTotal := h.QHReqHeaderAvg + h.QHRespHeaderAvg
	http1Total := h.HTTP1ReqHeaderAvg + h.HTTP1RespHeaderAvg
	http2Total := h.HTTP2ReqHeaderAvg + h.HTTP2RespHeaderAvg
	http3Total := h.HTTP3ReqHeaderAvg + h.HTTP3RespHeaderAvg

	totalVsHTTP1 := (qhTotal / http1Total) * 100
	totalVsHTTP2 := (qhTotal / http2Total) * 100
	totalVsHTTP3 := (qhTotal / http3Total) * 100

	sb.WriteString("TOTAL HEADERS (Request + Response):\n")
	sb.WriteString(fmt.Sprintf("  QH avg:       %6.0f B  (baseline)\n", qhTotal))
	sb.WriteString(fmt.Sprintf("  HTTP/1 avg:   %6.0f B  (%s)\n",
		http1Total, fmtDifference(totalVsHTTP1)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 avg:   %6.0f B  (%s)\n",
		http2Total, fmtDifference(totalVsHTTP2)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 avg:   %6.0f B  (%s)",
		http3Total, fmtDifference(totalVsHTTP3)))

	return sb.String()
}

func fmtDetailedTable(results []BenchmarkResult) string {
	var sb strings.Builder

	sb.WriteString(
		"Test Case                              QH (bytes)  HTTP/1  HTTP/2  HTTP/3  QH/H1   QH/H2   QH/H3\n",
	)
	sb.WriteString(
		"───────────────────────────────────────────────────────────────────────────────────────────────\n",
	)

	detailed := GetDetailedResults(results)
	for _, d := range detailed {
		name := d.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}

		sb.WriteString(fmt.Sprintf("%-38s %7d %7d %7d %7d %6.1f%% %6.1f%% %6.1f%%\n",
			name,
			d.QHTotalBytes,
			d.HTTP1TotalBytes,
			d.HTTP2TotalBytes,
			d.HTTP3TotalBytes,
			d.QHVsHTTP1Ratio,
			d.QHVsHTTP2Ratio,
			d.QHVsHTTP3Ratio,
		))
	}

	return sb.String()
}
