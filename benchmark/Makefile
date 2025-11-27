.PHONY: all build run clean test report

N ?=

all: run # default with no args

build:
	@echo "Building qhbench..."
	@cd cmd/qhbench && go build -o ../../qhbench

run:
	@go run cmd/qhbench/main.go -examples=$(if $(N),$(N),0)

report:
	@echo "Generating benchmark report..."
	@DATE_HUMAN=$$(date '+%B %d, %Y at %H:%M:%S'); \
	GO_VERSION=$$(go version | awk '{print $$3}'); \
	GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	OS_NAME=$$(uname -s); \
	OS_ARCH=$$(uname -m); \
	mkdir -p ../docs/benchmarks; \
	go run cmd/qhbench/main.go -examples=$(if $(N),$(N),3) -o .benchmark_report_tmp.md; \
	{ \
		echo "# QH Protocol Benchmark Report"; \
		echo ""; \
		echo "**Generated:** $$DATE_HUMAN  "; \
		echo "**Commit:** \`$$GIT_COMMIT\`  "; \
		echo "**Go Version:** $$GO_VERSION  "; \
		echo "**Platform:** $$OS_NAME/$$OS_ARCH  "; \
		echo ""; \
		cat .benchmark_report_tmp.md; \
	} > ../docs/benchmarks/report.md; \
	rm .benchmark_report_tmp.md; \
	echo "Report saved to docs/benchmarks/report.md"

clean:
	@rm -f qhbench .benchmark_report_tmp.md

help:
	@echo "QH Protocol Benchmark Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make              - Run benchmarks to stdout (default)"
	@echo "  make build        - Build qhbench binary"
	@echo "  make run          - Run benchmarks to stdout"
	@echo "  make report       - Generate benchmark report in docs/benchmarks/report.md"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Options:"
	@echo "  N=<number>        - Number of wire format examples (default: 0 for run, 3 for report)"
