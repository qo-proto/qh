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

	summary := CalculateSummary(results)
	sb.WriteString("OVERALL SUMMARY\n\n")
	sb.WriteString(formatOverall(summary))
	sb.WriteString("\n\n")

	sb.WriteString("DETAILED RESULTS BY TEST CASE\n\n")
	sb.WriteString(formatDetailedTable(results))
	sb.WriteString("\n")

	return sb.String()
}

func formatOverall(s Summary) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("  Tests run:        %d\n", s.TotalTests))
	sb.WriteString(fmt.Sprintf("  Total QH bytes:   %d B\n", s.QHTotalBytes))
	sb.WriteString(fmt.Sprintf("  vs HTTP/1.1:      %.1f%% (%.1f%% reduction)\n",
		s.QHVsHTTP1Ratio, 100-s.QHVsHTTP1Ratio))
	sb.WriteString(fmt.Sprintf("  vs HTTP/2:        %.1f%% (%.1f%% reduction)\n",
		s.QHVsHTTP2Ratio, 100-s.QHVsHTTP2Ratio))
	sb.WriteString(fmt.Sprintf("  vs HTTP/3:        %.1f%% (%.1f%% reduction)\n\n",
		s.QHVsHTTP3Ratio, 100-s.QHVsHTTP3Ratio))

	http1Savings := s.HTTP1TotalBytes - s.QHTotalBytes
	http2Savings := s.HTTP2TotalBytes - s.QHTotalBytes
	http3Savings := s.HTTP3TotalBytes - s.QHTotalBytes

	sb.WriteString(fmt.Sprintf("  Bandwidth saved vs HTTP/1.1: %d B\n", http1Savings))
	sb.WriteString(fmt.Sprintf("  Bandwidth saved vs HTTP/2:   %d B\n", http2Savings))
	sb.WriteString(fmt.Sprintf("  Bandwidth saved vs HTTP/3:   %d B", http3Savings))

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
