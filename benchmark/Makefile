.PHONY: all build run clean test report

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
	@echo "Generating dated report ..."
	@DATE=$$(date +%Y-%m-%d); \
	DATE_HUMAN=$$(date '+%B %d, %Y'); \
	GO_VERSION=$$(go version | awk '{print $$3}'); \
	GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	mkdir -p ../docs/benchmarks; \
	go run cmd/qhbench/main.go -detailed -examples=2 -o /tmp/benchmark_report.txt; \
	echo "# QH Protocol Benchmark Report" > ../docs/benchmarks/report-$$DATE.txt; \
	echo "" >> ../docs/benchmarks/report-$$DATE.txt; \
	echo "**Date:** $$DATE_HUMAN  " >> ../docs/benchmarks/report-$$DATE.txt; \
	echo "**Commit:** \`$$GIT_COMMIT\`  " >> ../docs/benchmarks/report-$$DATE.txt; \
	echo "**Go Version:** $$GO_VERSION  " >> ../docs/benchmarks/report-$$DATE.txt; \
	echo "" >> ../docs/benchmarks/report-$$DATE.txt; \
	tail -n +3 /tmp/benchmark_report.txt >> ../docs/benchmarks/report-$$DATE.txt; \
	rm /tmp/benchmark_report.txt; \
	echo "report saved to docs/benchmarks/report-$$DATE.txt"

clean:
	@rm -f qhbench benchmark_report.txt

help:
	@echo "QH Protocol Benchmark Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make         - Run benchmarks (default)"
	@echo "  make build   - Build qhbench binary"
	@echo "  make run     - Run benchmarks to stdout"
	@echo "  make detailed- Run with detailed output"
	@echo "  make report - Generate dated report file in docs/benchmarks/"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make test    - Run tests"
	@echo "  make help    - Show this help"
