# qh:// - The Quite Ok HTTP Protocol

**qh://** is a simplified HTTP-like protocol. Built on top of [QOTP (Quite Ok Transport Protocol)](https://github.com/tbocek/qotp), it provides 0-RTT connections, built-in encryption and uses DNS TXT records for key distribution. The protocol uses a compact binary format which eliminates the use of header compression schemes like HPACK or QPACK.

**STATUS:** Experimental - Under active development

## Documentation

- **[Protocol Specification](./docs/protocol-definition.md)** - QH protocol definition
- **[Headers Reference](./docs/headers.md)** - Header format
  - [Static header table](./docs/static-tables.md)
- **[API Documentation](./docs/api.md)** - API reference of the Go implementation

## Run example

- Prerequisites: Go 1.25 or higher

```bash
# Start the server
go run ./examples/server/main.go

# In another terminal, run the client
go run ./examples/client/main.go

# Or directly in tmux with a shell script
./run-demo-tmux.sh
```

## Benchmarks

- `./benchmark/` - Benchmarking tool for QH vs HTTP/1.1, HTTP/2 and HTTP/3

NOTE: These benchmarks are not 100% representative of real-world bytes. Both HTTP/2 and HTTP/3 reuse dynamic header tables (HPACK/QPACK) and include additional control stream data, while these tests use fresh encoders and omit those overheads for consistency.

```
make         - Run benchmarks (default)
make build   - Build qhbench binary
make run     - Run benchmarks to stdout
make detailed- Run with detailed output
make report  - Generate report file
make clean   - Clean build artifacts
make test    - Run tests
make help    - Show this help
```
