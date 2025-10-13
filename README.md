# qh:// - The Quite Ok HTTP Protocol

**qh://** is a simplified HTTP-like protocol. Built on top of [QOTP](https://github.com/tbocek/qotp) (Quite Ok Transport Protocol), it provides 0-RTT connections, built-in encryption and uses DNS TXT records for key distribution. The protocol uses a compact binary format which eliminates the use of header compression schemes like HPACK or QPACK.

**STATUS:** Experimental - Under active development

## Documentation

- **[Protocol Specification](./docs/protocol-definition.md)** - Complete QH protocol definition
- **[API Documentation](./docs/api.md)** - Go API reference
- **[Headers Reference](./docs/headers.md)** - Supported headers and encoding

## Run example

- Prerequisites: Go 1.25 or higher

```bash
# Start the server
go run ./examples/server/main.go

# In another terminal, run the client
go run ./examples/client/main.go
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│            Application Layer: qh:// Protocol                │
│  • HTTP-inspired request/response semantics                 │
│  • Compact binary encoding                                  │
│  • DNS TXT record key distribution                          │
├─────────────────────────────────────────────────────────────┤
│            Transport Layer: QOTP                            │
│  • 0-RTT connection establishment                           │
│  • Built-in encryption (curve25519/chacha20-poly1305)       │
│  • UDP-based communication                                  │
│  • Stream multiplexing                                      │
├─────────────────────────────────────────────────────────────┤
│               Network Layer: UDP/IP                         │
│  • Standard network layer                                   │
└─────────────────────────────────────────────────────────────┘
```
