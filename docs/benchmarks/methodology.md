# Benchmark Methodology

This document describes how to reproduce the QH protocol wire format benchmarks that compare QH against HTTP/1.1, HTTP/2, and HTTP/3.

NOTE: These benchmarks measure wire format size using fresh encoders. HTTP/2 and HTTP/3 can achieve additional compression through dynamic header table reuse across multiple requests, which is not measured here.

## What We Measure

The benchmarks measure wire format efficiency. It compares the number of bytes needed to encode HTTP-like requests and responses on the wire (headers and bodies).

## Test Cases

The benchmark dataset consists of two types of test cases:

### 1. Real Web Traffic (100 cases)

Captured from actual browsing sessions with anonymized/replaced values. The samples were distributed (JSON API responses, JS files, Images, HTML pages, CSS stylesheets and other content types) to reflect typical web traffic.

The current dataset was collected from typical web browsing including: Video streaming (YouTube), search engines, blogs, etc. It was captured in Chrome.

**How to generate:**

1. Capture traffic in Chrome

- Open DevTools (F12) → Network tab
- Enable "Preserve log" checkbox
- Browse various websites normally
- Right-click → "Save HAR (sanitized)"

2. Generate test cases

```bash
python3 main.py data.har -o ../../testdata/http_traffic.json -n 100
```

**Output Location:** `benchmark/testdata/edge_cases.json`

### 2. Edge Cases (manual)

These consist of manually selected examples to cover edge-cases. Such as worst- and best-case scenarios for each protocol.

**Location:** `benchmark/testdata/edge_cases.json`
