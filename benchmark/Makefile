.PHONY: all build run clean test results

all: run # default with no args

build:
	@echo "Building qhbench..."
	@cd cmd/qhbench && go build -o ../../qhbench

run:
	@go run cmd/qhbench/main.go

detailed:
	@echo "Running detailed benchmarks..."
	@go run cmd/qhbench/main.go -detailed -examples 10

report:
	@echo "Generating benchmark report..."
	@go run cmd/qhbench/main.go -o benchmark_results.txt
	@echo "Report saved to benchmark_results.txt"

results:
	@echo "Generating dated results file..."
	@DATE=$$(date +%Y-%m-%d); \
	DATE_HUMAN=$$(date '+%B %d, %Y'); \
	GO_VERSION=$$(go version | awk '{print $$3}'); \
	GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	mkdir -p ../docs/benchmarks; \
	go run cmd/qhbench/main.go -detailed -examples=2 -o /tmp/benchmark_results.txt; \
	echo "# QH Protocol Benchmark Results" > ../docs/benchmarks/results-$$DATE.md; \
	echo "" >> ../docs/benchmarks/results-$$DATE.md; \
	echo "**Date:** $$DATE_HUMAN  " >> ../docs/benchmarks/results-$$DATE.md; \
	echo "**Commit:** \`$$GIT_COMMIT\`  " >> ../docs/benchmarks/results-$$DATE.md; \
	echo "**Go Version:** $$GO_VERSION  " >> ../docs/benchmarks/results-$$DATE.md; \
	echo "" >> ../docs/benchmarks/results-$$DATE.md; \
	tail -n +3 /tmp/benchmark_results.txt >> ../docs/benchmarks/results-$$DATE.md; \
	rm /tmp/benchmark_results.txt; \
	echo "Results saved to docs/benchmarks/results-$$DATE.md"

clean:
	@rm -f qhbench benchmark_results.txt

# TODO: add tests and uncomment
# test:
# 	@echo "Running tests..."
# 	@go test -v ./...

help:
	@echo "QH Protocol Benchmark Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make         - Run benchmarks (default)"
	@echo "  make build   - Build qhbench binary"
	@echo "  make run     - Run benchmarks to stdout"
	@echo "  make detailed- Run with detailed output"
	@echo "  make report  - Generate report file"
	@echo "  make results - Generate dated results file in docs/benchmarks/"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make test    - Run tests"
	@echo "  make help    - Show this help"
