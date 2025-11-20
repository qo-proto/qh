package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/qh-project/qh/benchmark"
)

const version = "0.0.1"

func main() {
	var (
		outputFile   = flag.String("o", "", "Output file for benchmark report (default: stdout)")
		wireExamples = flag.Int("examples", 2, "Number of wire format examples to include (0 to disable)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "qhbench - QH Protocol Benchmark Tool\n\n")
		fmt.Fprintf(
			os.Stderr,
			"Compares QH protocol wire format efficiency against HTTP/1.1, HTTP/2, and HTTP/3\n",
		)
		fmt.Fprintf(os.Stderr, "Usage: qhbench [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	timestamp := time.Now().Format(time.DateTime)
	fmt.Printf("Running QH Protocol Benchmarks (v%s)\n", version)
	fmt.Printf("Timestamp: %s\n", timestamp)
	fmt.Println()

	results := benchmark.RunBenchmarks()

	fmt.Printf("Completed %d test cases\n", len(results.All))
	fmt.Println()

	// stdout: use multi-section report
	if *outputFile == "" {
		report := benchmark.GenerateMultiSectionReport(results.EdgeCases, results.Traffic, results.All)
		if *wireExamples > 0 {
			report += "\n" + benchmark.GenerateWireFormatExamples(results.All, *wireExamples)
		}
		fmt.Println(report)
		return
	}

	// file output: use multi-section report with wire examples
	report := benchmark.GenerateMultiSectionReportMarkdown(results.EdgeCases, results.Traffic, results.All)

	if *wireExamples > 0 {
		report += "\n\n## Wire Format Examples\n\n"
		report += benchmark.GenerateWireFormatExamplesMarkdown(results.All, *wireExamples)
	}

	err := os.WriteFile(*outputFile, []byte(report), 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Report written to: %s\n", *outputFile)
}
