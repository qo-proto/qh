# Benchmark Methodology

This document describes how to reproduce the QH protocol wire format benchmarks that compare QH against HTTP/1.1, HTTP/2, and HTTP/3.

**NOTE:** These benchmarks measure wire format size using fresh encoders for each request/response pair. HTTP/2 and HTTP/3 can achieve additional compression through dynamic header table reuse across multiple requests within a connection, which is not measured here.

## What We Measure

The benchmarks measure wire format efficiency. It compares the number of bytes needed to encode HTTP-like requests and responses on the wire.

**What is measured:**

- **Request size**: Complete wire-format encoding of method, host, path, and headers
- **Response size**: Complete wire-format encoding of status code and headers
- **Total size**: Request + Response bytes

**What is NOT measured:**

- Message bodies (only headers are compared)
- Connection establishment overhead (TLS handshakes, QUIC setup, etc.)
- Runtime performance (CPU, memory, latency)
- Dynamic table state across multiple requests

The comparison focuses purely on the protocol's header encoding efficiency.

## Test Cases

The benchmark dataset consists of two types of test cases:

### 1. Real Web Traffic

Captured from actual browsing sessions with anonymized/replaced values. The samples are distributed across various content types (JSON API responses, JavaScript files, images, HTML pages, CSS stylesheets, etc.) to reflect typical modern web traffic patterns.

**Current dataset:** 100 test-cases collected from typical web browsing including video streaming (YouTube), search engines (Google), social media, developer documentation sites (GitHub), and blogs. Traffic was captured in Chrome using the HAR (HTTP Archive) format.

**How to generate:**

1. **Capture traffic in Chrome**

- Open DevTools (F12) → Network tab
- Enable "Preserve log" checkbox
- Browse various websites normally
- Right-click → "Save HAR (sanitized)"

2. **Process HAR to extract test cases**

```bash
cd benchmark/cmd/generate-testcases
python3 main.py data.har -o ../../testdata/http_traffic.json -n 100
```

This tool:

- Filters out non-essential resources (.woff, .ttf fonts, tracking pixels, etc.)
- Selects 100 representative requests/responses
- Extracts only headers (bodies are not included)
- Normalizes header names to lowercase
- Outputs in JSON format

3. **Anonymize the data**

Review the generated JSON file and replace any sensitive infos (domains, paths, header values, etc.).

**Output Location:** `benchmark/testdata/http_traffic.json`

### 2. Edge Cases

These consist of manually selected examples to cover edge-cases. Such as worst- and best-case scenarios for each protocol.

**Location:** `benchmark/testdata/edge_cases.json`

## Encoding Implementations

Each protocol is encoded using the according implementations in go:

- HTTP/1.1: `net/http` standard library
- HTTP/2: `golang.org/x/net/http2` with hpack
- HTTP/3: `github.com/lucas-clemente/quic-go/http3` with [qpack](github.com/quic-go/qpack)
- QH: `github.com/qo-proto/qh`

## Running the Benchmarks

Run the benchmarks in the `benchmark/` directory with the following `make` targets:

```
Targets:
  make              - Run benchmarks to stdout (default)
  make build        - Build qhbench binary
  make run          - Run benchmarks to stdout
  make report       - Generate dated markdown report in docs/benchmarks/ (3 wire examples)
  make clean        - Clean build artifacts
  make test         - Run tests
  make help         - Show this help

Options:
  N=<number>        - Number of wire format examples (default: 0 for run, 3 for report)
```

## Interpretation

TODO

## Version History

Benchmark methodology may evolve over time. Each report includes a git commit hash for reproducibility.
