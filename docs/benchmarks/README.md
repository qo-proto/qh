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
  make report       - Generate benchmark report in docs/benchmarks/report.md
  make clean        - Clean build artifacts
  make test         - Run tests
  make help         - Show this help

Options:
  N=<number>        - Number of wire format examples (default: 0 for run, 3 for report)
```

## Interpretation

These interpretations are based on the benchmark report found in `docs/benchmarks/`.

### Considerations

While the byte size comparison is important, the encoding/decoding speed is not measured here which would be another advantage of QH.

### Key Findings

The benchmarks reveal QH's wire format efficiency characteristics across different scenarios:

#### Comparison with HTTP/1.1

QH consistently outperforms HTTP/1.1 in all scenarios (edge-cases and actual traffic). The core advantage over the text-based HTTP/1.1 comres from various improvements/design decisions in the protocols design.

#### Comparison with HTTP/2 & HTTP/3

Against modern binary protocols, QH's performance varies by scenario: In the edge-cases QH outperforms both HTTP/2 and HTTP/3 as the test cases were designed to highlight QH's strengths (static table matches, simple headers). In real traffic however QH is slightly bigger than both HTTP/2 and HTTP/3.

This difference can be explained by the following reasons:

- **Static table coverage**: QH's static table is optimized for common patterns. When requests match the static table (Edge Case 1) or have a header name only entry (Format 2), QH greatly benefits from it. Completely custom headers (Format 3) are less efficient compared to HTTP/2/3's compression.
- **HPACK/QPACK dynamic tables**: HTTP/2/3 benefit from dynamic header compression across multiple requests on the same connection. These benchmarks use fresh encoders per request, so dynamic table benefits are not measured. In production, HTTP/2/3 would likely achieve better compression ratios over long-lived connections.

### Conclusions

QH dramatically improves over HTTP/1 in terms of wire format size and is competitive compared to HTTP/2 and HTTP/3 for simpler requests/responses.

For maximum compression HTTP/3 is the best choice.

The main benefits of QH are:

- Simpler implementation (no HPACK/QPACK complexity)
- Faster encoder/decoder (no dynamic table lookups or updates)

## Version History

Benchmark methodology may evolve over time. Each report includes a git commit hash for reproducibility.
