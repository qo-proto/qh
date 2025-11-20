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
	sb.WriteString("  1. Executive Summary\n")
	sb.WriteString(fmt.Sprintf("  2. Edge Case Analysis (%d test cases)\n", len(edgeResults)))
	sb.WriteString(fmt.Sprintf("  3. Real HTTP Traffic Analysis (%d test cases)\n", len(trafficResults)))
	sb.WriteString(fmt.Sprintf("  4. Combined Results (%d test cases)\n\n", len(allResults)))

	// Section 1: Executive Summary
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 1: EXECUTIVE SUMMARY\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(formatExecutiveSummary(edgeResults, trafficResults, allResults))
	sb.WriteString("\n\n")

	// Section 2: Edge Case Analysis
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 2: EDGE CASE ANALYSIS\n")
	sb.WriteString("  Protocol boundary conditions and stress tests\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(formatEdgeCaseSection(edgeResults))
	sb.WriteString("\n\n")

	// Section 3: Real Traffic Analysis
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 3: REAL HTTP TRAFFIC ANALYSIS\n")
	sb.WriteString("  101 requests captured from 2025 internet traffic\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(formatTrafficSection(trafficResults))
	sb.WriteString("\n\n")

	// Section 4: Combined Results
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SECTION 4: COMBINED RESULTS\n")
	sb.WriteString(fmt.Sprintf("  All %d test cases\n", len(allResults)))
	sb.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	sb.WriteString(formatCombinedSection(allResults))

	return sb.String()
}

func GenerateMultiSectionReportMarkdown(edgeResults, trafficResults, allResults []BenchmarkResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# QH Protocol Benchmark Report\n\n")
	sb.WriteString(fmt.Sprintf("**%d Test Cases** (%d Edge Cases + %d Real Traffic)\n\n",
		len(allResults), len(edgeResults), len(trafficResults)))

	// Table of Contents
	sb.WriteString("## Table of Contents\n\n")
	sb.WriteString("1. [Executive Summary](#executive-summary)\n")
	sb.WriteString(fmt.Sprintf("2. [Edge Case Analysis](#edge-case-analysis) (%d test cases)\n", len(edgeResults)))
	sb.WriteString(fmt.Sprintf("3. [Real HTTP Traffic Analysis](#real-http-traffic-analysis) (%d test cases)\n", len(trafficResults)))
	sb.WriteString(fmt.Sprintf("4. [Combined Results](#combined-results) (%d test cases)\n\n", len(allResults)))

	// Section 1: Executive Summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(formatExecutiveSummaryMarkdown(edgeResults, trafficResults, allResults))
	sb.WriteString("\n\n")

	// Section 2: Edge Case Analysis
	sb.WriteString("## Edge Case Analysis\n\n")
	sb.WriteString("*Protocol boundary conditions and stress tests*\n\n")
	sb.WriteString(formatEdgeCaseSectionMarkdown(edgeResults))
	sb.WriteString("\n\n")

	// Section 3: Real Traffic Analysis
	sb.WriteString("## Real HTTP Traffic Analysis\n\n")
	sb.WriteString(fmt.Sprintf("*%d requests captured from 2025 internet traffic*\n\n", len(trafficResults)))
	sb.WriteString(formatTrafficSectionMarkdown(trafficResults))
	sb.WriteString("\n\n")

	// Section 4: Combined Results
	sb.WriteString("## Combined Results\n\n")
	sb.WriteString(fmt.Sprintf("*All %d test cases*\n\n", len(allResults)))
	sb.WriteString(formatCombinedSectionMarkdown(allResults))

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

func formatDifference(ratio float64) string {
	diff := 100 - ratio
	if diff > 0 {
		return fmt.Sprintf("%.1f%% smaller", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%.1f%% larger", -diff)
	}
	return "same size"
}
