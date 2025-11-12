# Benchmark Methodology

This document describes how to reproduce the QH protocol wire format benchmarks that compare QH against HTTP/1.1, HTTP/2, and HTTP/3.

## What We Measure

The benchmarks measure wire format efficiency. It compares the number of bytes needed to encode HTTP-like requests and responses on the wire (headers and bodies).

## Test Cases
