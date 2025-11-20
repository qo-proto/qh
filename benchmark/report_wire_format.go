package benchmark

import (
	"fmt"
	"strings"

	"github.com/qh-project/qh"
)

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
			Body:    nil,
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
			Body:       nil,
		}
		sb.WriteString(resp.AnnotateWireFormat(r.QH.ResponseBytes))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func GenerateWireFormatExamplesMarkdown(results []BenchmarkResult, count int) string {
	var sb strings.Builder

	if count > len(results) {
		count = len(results)
	}

	for i := range count {
		r := results[i]
		sb.WriteString(fmt.Sprintf("### Example %d: %s\n\n", i+1, r.TestCase.Name))

		sb.WriteString("**Request Sizes:**\n\n")
		sb.WriteString(fmt.Sprintf("- QH: %d bytes\n", r.QH.RequestSize))
		sb.WriteString(fmt.Sprintf("- HTTP/1.1: %d bytes\n", r.HTTP1.RequestSize))
		sb.WriteString(fmt.Sprintf("- HTTP/2: %d bytes\n", r.HTTP2.RequestSize))
		sb.WriteString(fmt.Sprintf("- HTTP/3: %d bytes\n\n", r.HTTP3.RequestSize))

		sb.WriteString("**QH Request Wire Format (annotated):**\n\n")
		sb.WriteString("```\n")
		req := &qh.Request{
			Method:  parseMethod(r.TestCase.Request.Method),
			Host:    r.TestCase.Request.Host,
			Path:    r.TestCase.Request.Path,
			Version: qh.Version,
			Headers: r.TestCase.Request.Headers,
			Body:    nil,
		}
		sb.WriteString(req.AnnotateWireFormat(r.QH.RequestBytes))
		sb.WriteString("```\n\n")

		sb.WriteString("**Response Sizes:**\n\n")
		sb.WriteString(fmt.Sprintf("- QH: %d bytes\n", r.QH.ResponseSize))
		sb.WriteString(fmt.Sprintf("- HTTP/1.1: %d bytes\n", r.HTTP1.ResponseSize))
		sb.WriteString(fmt.Sprintf("- HTTP/2: %d bytes\n", r.HTTP2.ResponseSize))
		sb.WriteString(fmt.Sprintf("- HTTP/3: %d bytes\n\n", r.HTTP3.ResponseSize))

		sb.WriteString("**QH Response Wire Format (annotated):**\n\n")
		sb.WriteString("```\n")
		resp := &qh.Response{
			Version:    qh.Version,
			StatusCode: r.TestCase.Response.StatusCode,
			Headers:    r.TestCase.Response.Headers,
			Body:       nil,
		}
		sb.WriteString(resp.AnnotateWireFormat(r.QH.ResponseBytes))
		sb.WriteString("```\n\n")
	}

	return sb.String()
}
