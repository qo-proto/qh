## Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  (Web Apps, APIs, Single Page Applications)                 │
├─────────────────────────────────────────────────────────────┤
│                     qh:// Protocol                          │
│  • HTTP-like request/response semantics                     │
│  • GET, POST, PUT, DELETE methods                           │
│  • Header and body parsing                                  │
│  • Status codes and content types                           │
├─────────────────────────────────────────────────────────────┤
│                    QOTP Transport                           │
│  • 0-RTT connection establishment                           │
│  • Built-in encryption and authentication                   │
│  • UDP-based communication                                  │
│  • Stream multiplexing                                      │
├─────────────────────────────────────────────────────────────┤
│                      UDP/IP                                 │
│  • Standard network layer                                   │
└─────────────────────────────────────────────────────────────┘
```
