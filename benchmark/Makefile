.PHONY: all build run clean test

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
	@echo "  make clean   - Clean build artifacts"
	@echo "  make test    - Run tests"
	@echo "  make help    - Show this help"
