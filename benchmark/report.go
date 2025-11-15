package benchmark

import (
	"fmt"
	"strings"

	"github.com/qh-project/qh"
)

func GenerateReport(results []BenchmarkResult) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  QH Protocol Benchmark Results\n")
	sb.WriteString("  Comparing QH vs HTTP/1.1, HTTP/2, and HTTP/3\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	sb.WriteString("DETAILED RESULTS BY TEST CASE\n\n")
	sb.WriteString(formatDetailedTable(results))
	sb.WriteString("\n\n")

	summary := CalculateSummary(results)
	sb.WriteString("OVERALL SUMMARY\n\n")
	sb.WriteString(formatOverall(summary))
	sb.WriteString("\n\n")

	sizeCategories := CalculateSizeCategories(results)
	sb.WriteString("BREAKDOWN BY RESPONSE SIZE\n\n")
	sb.WriteString(formatSizeCategories(sizeCategories))
	sb.WriteString("\n\n")

	headerAnalysis := CalculateHeaderAnalysis(results)
	sb.WriteString("REQUEST ENCODING ANALYSIS\n\n")
	sb.WriteString(formatHeaderAnalysis(headerAnalysis))
	sb.WriteString("\n")

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

func formatDifference(ratio float64) string {
	diff := 100 - ratio
	if diff > 0 {
		return fmt.Sprintf("%.1f%% smaller", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%.1f%% larger", -diff)
	}
	return "same size"
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

func formatHeaderAnalysis(h HeaderAnalysis) string {
	var sb strings.Builder

	sb.WriteString("Average request sizes (headers + request body):\n")
	sb.WriteString(fmt.Sprintf("  QH requests:       %6.0f B  (baseline)\n", h.QHAvgHeaders))
	sb.WriteString(fmt.Sprintf("  HTTP/1 requests:   %6.0f B  (%.1f%% of HTTP/1)\n", h.HTTP1AvgHeaders, h.QHVsHTTP1Ratio))
	sb.WriteString(fmt.Sprintf("  HTTP/2 requests:   %6.0f B  (%.1f%% of HTTP/2)\n", h.HTTP2AvgHeaders, h.QHVsHTTP2Ratio))
	sb.WriteString(fmt.Sprintf("  HTTP/3 requests:   %6.0f B  (%.1f%% of HTTP/3)\n\n", h.HTTP3AvgHeaders, h.QHVsHTTP3Ratio))

	http1Savings := h.HTTP1TotalHeaders - h.QHTotalHeaders
	http2Savings := h.HTTP2TotalHeaders - h.QHTotalHeaders
	http3Savings := h.HTTP3TotalHeaders - h.QHTotalHeaders

	sb.WriteString("Total bytes saved on requests:\n")
	sb.WriteString(fmt.Sprintf("  vs HTTP/1.1: %7d B  (%s)\n",
		http1Savings, formatDifference(h.QHVsHTTP1Ratio)))
	sb.WriteString(fmt.Sprintf("  vs HTTP/2:   %7d B  (%s)\n",
		http2Savings, formatDifference(h.QHVsHTTP2Ratio)))
	sb.WriteString(fmt.Sprintf("  vs HTTP/3:   %7d B  (%s)",
		http3Savings, formatDifference(h.QHVsHTTP3Ratio)))

	return sb.String()
}

func formatBytes(bytes float64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%.0f B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", bytes/1024)
	}
	return fmt.Sprintf("%.1f MB", bytes/(1024*1024))
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

func GenerateWireFormatExamples(results []BenchmarkResult, count int) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  WIRE FORMAT EXAMPLES\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	if count > len(results) {
		count = len(results)
	}

	for i := range count {
		r := results[i]
		sb.WriteString(fmt.Sprintf("Example %d: %s\n", i+1, r.TestCase.Name))
		sb.WriteString(strings.Repeat("─", 71) + "\n\n")

		sb.WriteString("REQUEST:\n")
		sb.WriteString(fmt.Sprintf("  QH:       %d bytes\n", r.QH.RequestSize))
		sb.WriteString(fmt.Sprintf("  HTTP/1.1: %d bytes\n", r.HTTP1.RequestSize))
		sb.WriteString(fmt.Sprintf("  HTTP/2:   %d bytes\n", r.HTTP2.RequestSize))
		sb.WriteString(fmt.Sprintf("  HTTP/3:   %d bytes\n\n", r.HTTP3.RequestSize))

		sb.WriteString("  QH Wire Format (annotated):\n")
		req := &qh.Request{
			Method:  parseMethod(r.TestCase.Request.Method),
			Host:    r.TestCase.Request.Host,
			Path:    r.TestCase.Request.Path,
			Version: qh.Version,
			Headers: r.TestCase.Request.Headers,
			Body:    r.TestCase.Request.Body,
		}
		sb.WriteString(req.AnnotateWireFormat(r.QH.RequestBytes))
		sb.WriteString("\n\n")

		sb.WriteString("RESPONSE:\n")
		sb.WriteString(fmt.Sprintf("  QH:       %d bytes\n", r.QH.ResponseSize))
		sb.WriteString(fmt.Sprintf("  HTTP/1.1: %d bytes\n", r.HTTP1.ResponseSize))
		sb.WriteString(fmt.Sprintf("  HTTP/2:   %d bytes\n", r.HTTP2.ResponseSize))
		sb.WriteString(fmt.Sprintf("  HTTP/3:   %d bytes\n\n", r.HTTP3.ResponseSize))

		sb.WriteString("  QH Wire Format (annotated):\n")
		resp := &qh.Response{
			Version:    qh.Version,
			StatusCode: r.TestCase.Response.StatusCode,
			Headers:    r.TestCase.Response.Headers,
			Body:       r.TestCase.Response.Body,
		}
		sb.WriteString(resp.AnnotateWireFormat(r.QH.ResponseBytes))
		sb.WriteString("\n\n")
	}

	return sb.String()
}
