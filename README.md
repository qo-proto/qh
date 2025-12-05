# qh:// - The Quite Ok HTTP Protocol

**qh://** is a simplified HTTP-like protocol. Built on top of [QOTP (Quite Ok Transport Protocol)](https://github.com/qo-proto/qotp), it provides 0-RTT connections, built-in encryption, stream multiplexing, and uses DNS TXT records for key distribution. The protocol uses a compact binary format which eliminates the use of header compression schemes like HPACK or QPACK.

**STATUS:** Experimental

## Comparison

| Feature          | HTTP/1.1       | HTTP/2         | HTTP/3         | QH                                    |
| ---------------- | -------------- | -------------- | -------------- | ------------------------------------- |
| Transport        | TCP            | TCP            | UDP (QUIC)     | UDP (QOTP)                            |
| Encryption       | Optional (TLS) | Optional (TLS) | Built-in       | Built-in                              |
| 0-RTT            | No             | With TLS 1.3   | Yes            | Yes                                   |
| Multiplexing     | No             | Yes            | Yes            | Yes                                   |
| Header Format    | Text           | Binary (HPACK) | Binary (QPACK) | Binary (static table, no compression) |
| Key Distribution | CAs            | CAs            | CAs            | DNS TXT                               |

## Documentation

- **[Protocol Specification](./docs/protocol-definition.md)** - QH protocol definition
- **[Headers Reference](./docs/headers.md)** - Header format
  - [Static header table](./docs/static-tables.md)
- **[API Documentation](./docs/api.md)** - API reference of the Go implementation

## Installation

```bash
go get github.com/qo-proto/qh
```

## Run example

- Prerequisites: Go 1.25 or higher

```bash
# Terminal 1: Start the server
go run ./examples/server/main.go

# Terminal 2: Run the client
go run ./examples/client/main.go

# Or directly in tmux with a shell script (basic client)
./run-demo-tmux.sh

# Or run concurrent examples (multiplexing)
go run ./examples/server-concurrent/main.go
go run ./examples/client-concurrent/main.go
```

### Keylog Support (for Wireshark Decryption)

QH supports QOTP keylog output for decrypting network traffic in Wireshark. This is useful for debugging and protocol analysis.

**Server-side keylog** (recommended):

```bash
# Build with keylog support
go run -tags keylog ./examples/server/main.go

# The server will create qh_server_keylog.txt automatically
```

The keylog file format follows the SSLKEYLOGFILE convention with `QOTP_SHARED_SECRET` entries that can be used with the QOTP Wireshark dissector.

## Benchmarks

QH protocol wire format efficiency compared against HTTP/1.1, HTTP/2, and HTTP/3.

- [Benchmark Documentation](./docs/benchmarks/README.md) - Methodology and documentation
- [Benchmark Report](./docs/benchmarks/report.md) - Generated report
