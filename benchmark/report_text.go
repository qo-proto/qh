package benchmark

import (
	"fmt"
	"strings"
)

func formatPerformanceBounds(bounds PerformanceBounds) string {
	var sb strings.Builder

	// vs HTTP/1.1
	http1BestRatio := float64(bounds.VsHTTP1.BestCase.QH.TotalSize) / float64(bounds.VsHTTP1.BestCase.HTTP1.TotalSize) * 100
	http1WorstRatio := float64(bounds.VsHTTP1.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP1.WorstCase.HTTP1.TotalSize) * 100

	sb.WriteString("  vs HTTP/1.1:\n")
	sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% smaller (%s)\n", 100-http1BestRatio, bounds.VsHTTP1.BestCase.TestCase.Name))
	sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% smaller (%s)\n", 100-http1WorstRatio, bounds.VsHTTP1.WorstCase.TestCase.Name))
	sb.WriteString("\n")

	// vs HTTP/2
	http2BestRatio := float64(bounds.VsHTTP2.BestCase.QH.TotalSize) / float64(bounds.VsHTTP2.BestCase.HTTP2.TotalSize) * 100
	http2WorstRatio := float64(bounds.VsHTTP2.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP2.WorstCase.HTTP2.TotalSize) * 100

	sb.WriteString("  vs HTTP/2:\n")
	if http2BestRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% smaller (%s)\n", 100-http2BestRatio, bounds.VsHTTP2.BestCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% larger (%s)\n", http2BestRatio-100, bounds.VsHTTP2.BestCase.TestCase.Name))
	}
	if http2WorstRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% smaller (%s)\n", 100-http2WorstRatio, bounds.VsHTTP2.WorstCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% larger (%s)\n", http2WorstRatio-100, bounds.VsHTTP2.WorstCase.TestCase.Name))
	}
	sb.WriteString("\n")

	// vs HTTP/3
	http3BestRatio := float64(bounds.VsHTTP3.BestCase.QH.TotalSize) / float64(bounds.VsHTTP3.BestCase.HTTP3.TotalSize) * 100
	http3WorstRatio := float64(bounds.VsHTTP3.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP3.WorstCase.HTTP3.TotalSize) * 100

	sb.WriteString("  vs HTTP/3:\n")
	if http3BestRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% smaller (%s)\n", 100-http3BestRatio, bounds.VsHTTP3.BestCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Best case:  %.1f%% larger (%s)\n", http3BestRatio-100, bounds.VsHTTP3.BestCase.TestCase.Name))
	}
	if http3WorstRatio < 100 {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% smaller (%s)\n", 100-http3WorstRatio, bounds.VsHTTP3.WorstCase.TestCase.Name))
	} else {
		sb.WriteString(fmt.Sprintf("    Worst case: %.1f%% larger (%s)\n", http3WorstRatio-100, bounds.VsHTTP3.WorstCase.TestCase.Name))
	}

	return sb.String()
}

func formatExecutiveSummary(edgeResults, trafficResults, allResults []BenchmarkResult) string {
	var sb strings.Builder

	allSummary := CalculateSummary(allResults)
	trafficSummary := CalculateSummary(trafficResults)

	sb.WriteString("Overall Performance (all test cases):\n")
	sb.WriteString(fmt.Sprintf("  • %.1f%% bandwidth savings vs HTTP/1.1 (%d B saved)\n",
		100-allSummary.QHVsHTTP1Ratio, allSummary.HTTP1TotalBytes-allSummary.QHTotalBytes))
	sb.WriteString(fmt.Sprintf("  • %.1f%% bandwidth savings vs HTTP/2 (%d B saved)\n",
		100-allSummary.QHVsHTTP2Ratio, allSummary.HTTP2TotalBytes-allSummary.QHTotalBytes))
	sb.WriteString(fmt.Sprintf("  • Achieves %.1f%% of HTTP/3 efficiency\n", allSummary.QHVsHTTP3Ratio))
	sb.WriteString("\n")

	sb.WriteString("Real-World Traffic Performance:\n")
	sb.WriteString(fmt.Sprintf("  • %.1f%% better than HTTP/1.1 on typical requests\n", 100-trafficSummary.QHVsHTTP1Ratio))
	sb.WriteString(fmt.Sprintf("  • %.1f%% better than HTTP/2\n", 100-trafficSummary.QHVsHTTP2Ratio))

	trafficHeaders := CalculateHeaderOnlyAnalysis(trafficResults)
	sb.WriteString(fmt.Sprintf("  • Header overhead: %.1f%% smaller than HTTP/1.1 (requests)\n",
		100-trafficHeaders.QHReqVsHTTP1Ratio))
	sb.WriteString(fmt.Sprintf("  • Header overhead: %.1f%% smaller than HTTP/1.1 (responses)\n",
		100-trafficHeaders.QHRespVsHTTP1Ratio))
	sb.WriteString("\n")

	sb.WriteString("\n")
	bounds := FindBestWorstCases(edgeResults)
	if bounds.HasData {
		sb.WriteString("Edge Case Performance Bounds:\n\n")
		// Format with bullet points for executive summary
		http1BestRatio := float64(bounds.VsHTTP1.BestCase.QH.TotalSize) / float64(bounds.VsHTTP1.BestCase.HTTP1.TotalSize) * 100
		sb.WriteString(fmt.Sprintf("  • vs HTTP/1.1: %.1f%% to %.1f%% smaller\n",
			100-http1BestRatio, 100-float64(bounds.VsHTTP1.WorstCase.QH.TotalSize)/float64(bounds.VsHTTP1.WorstCase.HTTP1.TotalSize)*100))

		http2BestRatio := float64(bounds.VsHTTP2.BestCase.QH.TotalSize) / float64(bounds.VsHTTP2.BestCase.HTTP2.TotalSize) * 100
		http2WorstRatio := float64(bounds.VsHTTP2.WorstCase.QH.TotalSize) / float64(bounds.VsHTTP2.WorstCase.HTTP2.TotalSize) * 100
		if http2WorstRatio > 100 {
			sb.WriteString(fmt.Sprintf("  • vs HTTP/2: %.1f%% smaller to %.1f%% larger\n",
				100-http2BestRatio, http2WorstRatio-100))
		} else {
			sb.WriteString(fmt.Sprintf("  • vs HTTP/2: %.1f%% to %.1f%% smaller\n",
				100-http2BestRatio, 100-http2WorstRatio))
		}
	}

	return sb.String()
}

func formatEdgeCaseSection(results []BenchmarkResult) string {
	var sb strings.Builder

	sb.WriteString("EDGE CASE DETAILED RESULTS\n\n")
	sb.WriteString(formatDetailedTable(results))
	sb.WriteString("\n\n")

	summary := CalculateSummary(results)
	sb.WriteString("EDGE CASE SUMMARY\n\n")
	sb.WriteString(formatOverall(summary))
	sb.WriteString("\n\n")

	// Header analysis for edge cases
	headerOnly := CalculateHeaderOnlyAnalysis(results)
	sb.WriteString("REQUEST/RESPONSE BREAKDOWN\n\n")
	sb.WriteString(formatHeaderOnlyAnalysis(headerOnly))

	return sb.String()
}

func formatTrafficSection(results []BenchmarkResult) string {
	var sb strings.Builder

	summary := CalculateSummary(results)
	sb.WriteString("OVERALL SUMMARY\n\n")
	sb.WriteString(formatOverall(summary))
	sb.WriteString("\n\n")

	bounds := FindBestWorstCases(results)
	if bounds.HasData {
		sb.WriteString("PERFORMANCE BOUNDS\n\n")
		sb.WriteString(formatPerformanceBounds(bounds))
		sb.WriteString("\n")
	}

	sizeCategories := CalculateSizeCategories(results)
	sb.WriteString("BREAKDOWN BY SIZE CATEGORY\n\n")
	sb.WriteString(formatSizeCategories(sizeCategories))
	sb.WriteString("\n\n")

	headerOnly := CalculateHeaderOnlyAnalysis(results)
	sb.WriteString("REQUEST/RESPONSE BREAKDOWN\n\n")
	sb.WriteString(formatHeaderOnlyAnalysis(headerOnly))

	return sb.String()
}

func formatCombinedSection(results []BenchmarkResult) string {
	var sb strings.Builder

	summary := CalculateSummary(results)
	sb.WriteString("OVERALL SUMMARY\n\n")
	sb.WriteString(formatOverall(summary))
	sb.WriteString("\n\n")

	bounds := FindBestWorstCases(results)
	if bounds.HasData {
		sb.WriteString("PERFORMANCE BOUNDS\n\n")
		sb.WriteString(formatPerformanceBounds(bounds))
		sb.WriteString("\n")
	}

	sizeCategories := CalculateSizeCategories(results)
	sb.WriteString("BREAKDOWN BY SIZE CATEGORY\n\n")
	sb.WriteString(formatSizeCategories(sizeCategories))
	sb.WriteString("\n\n")

	headerOnly := CalculateHeaderOnlyAnalysis(results)
	sb.WriteString("REQUEST/RESPONSE BREAKDOWN\n\n")
	sb.WriteString(formatHeaderOnlyAnalysis(headerOnly))

	return sb.String()
}

func formatOverall(s Summary) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("  Tests run:        %d\n", s.TotalTests))
	sb.WriteString(fmt.Sprintf("  Total QH bytes:   %d B\n", s.QHTotalBytes))

	sb.WriteString(fmt.Sprintf("  vs HTTP/1.1:      %.1f%% (%s)\n",
		s.QHVsHTTP1Ratio, formatDifference(s.QHVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  vs HTTP/2:        %.1f%% (%s)\n",
		s.QHVsHTTP2Ratio, formatDifference(s.QHVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  vs HTTP/3:        %.1f%% (%s)\n\n",
		s.QHVsHTTP3Ratio, formatDifference(s.QHVsHTTP3Ratio)))

	http1Savings := s.HTTP1TotalBytes - s.QHTotalBytes
	http2Savings := s.HTTP2TotalBytes - s.QHTotalBytes
	http3Savings := s.HTTP3TotalBytes - s.QHTotalBytes

	sb.WriteString(fmt.Sprintf("  Bandwidth saved vs HTTP/1.1: %d B\n", http1Savings))
	sb.WriteString(fmt.Sprintf("  Bandwidth saved vs HTTP/2:   %d B\n", http2Savings))
	sb.WriteString(fmt.Sprintf("  Bandwidth saved vs HTTP/3:   %d B", http3Savings))

	return sb.String()
}

func formatSizeCategories(categories []SizeCategory) string {
	var sb strings.Builder

	sb.WriteString("Category          Count   QH Avg      H1 Avg      H2 Avg      H3 Avg      QH/H1   QH/H2   QH/H3\n")
	sb.WriteString("────────────────────────────────────────────────────────────────────────────────────────────────────\n")

	for _, c := range categories {
		sb.WriteString(fmt.Sprintf("%-16s %6d %10s %10s %10s %10s %6.1f%% %6.1f%% %6.1f%%\n",
			c.Name,
			c.Count,
			formatBytes(c.QHAvg),
			formatBytes(c.HTTP1Avg),
			formatBytes(c.HTTP2Avg),
			formatBytes(c.HTTP3Avg),
			c.QHVsHTTP1Ratio,
			c.QHVsHTTP2Ratio,
			c.QHVsHTTP3Ratio,
		))
	}

	return sb.String()
}

func formatHeaderOnlyAnalysis(h HeaderOnlyAnalysis) string {
	var sb strings.Builder

	sb.WriteString("REQUEST HEADERS:\n")
	sb.WriteString(fmt.Sprintf("  QH avg:       %6.0f B  (baseline)\n", h.QHReqHeaderAvg))
	sb.WriteString(fmt.Sprintf("  HTTP/1 avg:   %6.0f B  (%s)\n",
		h.HTTP1ReqHeaderAvg, formatDifference(h.QHReqVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 avg:   %6.0f B  (%s)\n",
		h.HTTP2ReqHeaderAvg, formatDifference(h.QHReqVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 avg:   %6.0f B  (%s)\n\n",
		h.HTTP3ReqHeaderAvg, formatDifference(h.QHReqVsHTTP3Ratio)))

	sb.WriteString("RESPONSE HEADERS:\n")
	sb.WriteString(fmt.Sprintf("  QH avg:       %6.0f B  (baseline)\n", h.QHRespHeaderAvg))
	sb.WriteString(fmt.Sprintf("  HTTP/1 avg:   %6.0f B  (%s)\n",
		h.HTTP1RespHeaderAvg, formatDifference(h.QHRespVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 avg:   %6.0f B  (%s)\n",
		h.HTTP2RespHeaderAvg, formatDifference(h.QHRespVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 avg:   %6.0f B  (%s)\n\n",
		h.HTTP3RespHeaderAvg, formatDifference(h.QHRespVsHTTP3Ratio)))

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
		http1Total, formatDifference(totalVsHTTP1)))
	sb.WriteString(fmt.Sprintf("  HTTP/2 avg:   %6.0f B  (%s)\n",
		http2Total, formatDifference(totalVsHTTP2)))
	sb.WriteString(fmt.Sprintf("  HTTP/3 avg:   %6.0f B  (%s)",
		http3Total, formatDifference(totalVsHTTP3)))

	return sb.String()
}

func formatDetailedTable(results []BenchmarkResult) string {
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
