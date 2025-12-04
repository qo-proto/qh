package benchmark

import (
	"fmt"
	"strings"
)

func GenerateMultiSectionReport(edgeResults, trafficResults, allResults []BenchmarkResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  QH Protocol Benchmark Report\n")
	sb.WriteString(fmt.Sprintf("  %d Test Cases (%d Edge Cases + %d Real Traffic)\n",
		len(allResults), len(edgeResults), len(trafficResults)))
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Table of Contents
	sb.WriteString("TABLE OF CONTENTS\n")
	sb.WriteString(fmt.Sprintf("  1. Edge Case Analysis (%d test cases)\n", len(edgeResults)))
	sb.WriteString(fmt.Sprintf("  2. Real HTTP Traffic Analysis (%d test cases)\n", len(trafficResults)))
	sb.WriteString(fmt.Sprintf("  3. Combined Results (%d test cases)\n\n", len(allResults)))

	// Section 1: Edge Case Analysis
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 1: EDGE CASE ANALYSIS\n")
	sb.WriteString("  Protocol boundary conditions and stress tests\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(fmtEdgeCaseSection(edgeResults))
	sb.WriteString("\n\n")

	// Section 2: Real Traffic Analysis
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 2: REAL HTTP TRAFFIC ANALYSIS\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(fmtTrafficSection(trafficResults))
	sb.WriteString("\n\n")

	// Section 3: Combined Results
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 3: COMBINED RESULTS\n")
	sb.WriteString(fmt.Sprintf("  All %d test cases\n", len(allResults)))
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(fmtCombinedSection(allResults))

	return sb.String()
}

func GenerateMultiSectionReportMarkdown(edgeResults, trafficResults, allResults []BenchmarkResult) string {
	var sb strings.Builder

	// Test case summary
	sb.WriteString(fmt.Sprintf("**%d Test Cases** (%d Edge Cases + %d Real Traffic)\n\n",
		len(allResults), len(edgeResults), len(trafficResults)))

	// Table of Contents
	sb.WriteString("## Table of Contents\n\n")
	sb.WriteString(fmt.Sprintf("1. [Edge Case Analysis](#edge-case-analysis) (%d test cases)\n", len(edgeResults)))
	sb.WriteString(fmt.Sprintf("2. [Real HTTP Traffic Analysis](#real-http-traffic-analysis) (%d test cases)\n", len(trafficResults)))
	sb.WriteString(fmt.Sprintf("3. [Combined Results](#combined-results) (%d test cases)\n\n", len(allResults)))

	// Section 1: Edge Case Analysis
	sb.WriteString("## Edge Case Analysis\n\n")
	sb.WriteString(fmtEdgeCaseSectionMarkdown(edgeResults))
	sb.WriteString("\n\n")

	// Section 2: Real Traffic Analysis
	sb.WriteString("## Real HTTP Traffic Analysis\n\n")
	sb.WriteString(fmtTrafficSectionMarkdown(trafficResults))
	sb.WriteString("\n\n")

	// Section 3: Combined Results
	sb.WriteString("## Combined Results\n\n")
	sb.WriteString(fmtCombinedSectionMarkdown(allResults))

	return sb.String()
}

func fmtBytes(bytes float64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%.0f B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", bytes/1024)
	}
	return fmt.Sprintf("%.1f MB", bytes/(1024*1024))
}

func fmtDifference(ratio float64) string {
	if ratio < 100 {
		percentLarger := (100.0/ratio - 1.0) * 100.0
		return fmt.Sprintf("%.1f%% larger", percentLarger)
	} else if ratio > 100 {
		percentSmaller := (1.0 - 100.0/ratio) * 100.0
		return fmt.Sprintf("%.1f%% smaller", percentSmaller)
	}
	return "same size"
}
