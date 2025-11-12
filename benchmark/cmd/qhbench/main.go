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
		outputFile     = flag.String("o", "", "Output file for benchmark report (default: stdout)")
		wireExamples   = flag.Int("examples", 3, "Number of wire format examples to include")
		detailedOutput = flag.Bool("detailed", false, "Include detailed per-test results")
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

	fmt.Printf("Completed %d test cases\n", len(results))
	fmt.Println()

	var report string
	report = fmt.Sprintf(
		"QH Protocol Benchmark Report\nVersion: %s\nTimestamp: %s\n\n",
		version,
		timestamp,
	)
	report += benchmark.GenerateReport(results)
	if *detailedOutput {
		report += "\n" + benchmark.GenerateWireFormatExamples(results, *wireExamples)
	}

	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(report), 0o600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to: %s\n", *outputFile)
	} else {
		fmt.Println(report)
	}
}
