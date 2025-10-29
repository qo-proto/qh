# QH - Header Static Table

QH uses a **static header table** that maps single-byte IDs to either complete header key-value pairs or header names.

For detailed header format specifications, see [Section 6.1 - Header Format](./protocol-definition.md#61-header-format) in the protocol definition.

## Header ID Space Allocation

TODO: update X and tables below once data is ready and examples

```
0x00           = Custom header (key and value both transmitted)
0x01 - X    = Complete key-value pairs (X most common combinations)
X - 0xFF    = Header names only (X header names, value transmitted separately)
```

## Header Formats Summary

QH uses three header formats:

1. **Complete pairs (0x01-X)**: Single byte → `\x01` = `Content-Type: application/json`
2. **Name + value (X-0xFF)**: ID + varint length + value → `\x72 \x0A 1758784800` = `Date: 1758784800`
3. **Custom (0x00)**: Full key and value → `\x00 \x0C X-Request-ID \x06 abc123` = `X-Request-ID: abc123`

The tables below define the static ID-to-header mappings.

## Complete Key-Value Pairs

The most common header combinations mapped to single bytes:

| ID   | Complete Header                  | Use Case           |
| ---- | -------------------------------- | ------------------ |
| 0x01 | `Content-Type: application/json` | JSON API responses |
| 0x02 | `Content-Type: text/html`        | HTML pages         |

**Example usage:**

```
Response with JSON content type:
Wire: \x01
Means: Content-Type: application/json
Size: 1 byte
```

## Header Names Only

Headers where the name is common but values vary:

### Request Headers

| ID   | Header Name     | Description                 | Example Value   |
| ---- | --------------- | --------------------------- | --------------- |
| 0x20 | Accept          | Custom accept types         | `3,2,1`         |
| 0x21 | Accept-Encoding | Custom encoding preferences | `gzip, deflate` |

### Response Headers

| ID        | Header Name        | Description              | Example Value     |
| --------- | ------------------ | ------------------------ | ----------------- |
| 0x70      | Content-Type       | Custom content types     | `5` (custom code) |
| 0x71      | Cache-Control      | Custom cache directives  | `max-age=7200`    |
| 0x7E      | X-Payment-Response | x402 settlement response | `<base64-json>`   |
| 0x7F-0xFF | (reserved)         | Future response headers  |                   |

**Example usage:**

```
Response with custom cache control:
Wire: \x71 \x0C max-age=7200
Means: Cache-Control: max-age=7200
Size: 14 bytes (ID + varint length + value)
```

## Content Type Codes

When using numeric content type codes (for custom Content-Type values):

| Code | MIME Type                | Description          | Common Use       |
| ---- | ------------------------ | -------------------- | ---------------- |
| 0    | custom                   | Custom type          | Rare             |
| 1    | text/plain               | Plain text           | Text responses   |
| 2    | application/json         | JSON data            | API responses    |
| 3    | text/html                | HTML markup          | Web pages        |
| 4    | application/octet-stream | Binary data          | Files, downloads |
| 5-15 | (reserved)               | Future content types |                  |

## Custom Headers (0x00)

For headers not in the static table:

**Wire format:** `\x00 <varint:keyLen> <key> <varint:valueLen> <value>`

**Example:**

```
Header: X-Request-ID: abc123

Wire: \x00 \x0C X-Request-ID \x06 abc123

Breakdown:
- \x00: Custom header indicator
- \x0C: Key length (12 bytes)
- X-Request-ID: Key name
- \x06: Value length (6 bytes)
- abc123: Value

Total: 21 bytes
```

## Wire Format Examples

### Example 1: JSON Response with Standard Headers

```
HTTP equivalent:
Status: 200 OK
Content-Type: application/json
Cache-Control: public, max-age=3600
Date: 1758784800

{"data":"hello"}

QH wire format:
TODO

Breakdown:
TODO
```

### Example 2: Request with Authorization

```
HTTP equivalent:
GET /api/users
Host: example.com
Accept: application/json
Authorization: Bearer token123

QH wire format:
TODO

Breakdown:
TODO
```

### Example 3: Response with Custom Header

```
HTTP equivalent:
Status: 200 OK
X-Trace-ID: trace-12345
Content-Type: application/json

(no body)

QH wire format:
\x00 \x02 \x00 \x0A X-Trace-ID \x0C trace-12345 \x01 \x00

Breakdown:
- \x00: Status byte (200 OK)
- \x02: Num headers (2)
- \x00: Custom header indicator
- \x0A: Key length (10)
- X-Trace-ID: Key
- \x0C: Value length (12)
- trace-12345: Value
- \x01: Content-Type: application/json (complete pair, 1 byte)
- \x00: Body length (0 bytes, no body)
```

## Efficiency Comparison

| Format        | Example                          | Wire Size |
| ------------- | -------------------------------- | --------- |
| Complete pair | `Content-Type: application/json` | 1 byte    |
| Name + value  | `Date: 1758784800`               | 12 bytes  |
| Custom        | `X-Request-ID: abc123`           | 21 bytes  |

## Payment Protocol Support (x402)

QH supports the [x402 payment protocol](https://github.com/coinbase/x402).

**Required headers:**

- `X-Payment` (Request: TODO ADD CODE)
- `X-Payment-Response` (Response: TODO ADD CODE)
- `402 Payment Required` status code

**Flow:**

1. Client requests → Server responds 402 with payment requirements
2. Client sends payment via `X-Payment` header
3. Server validates and returns resource with `X-Payment-Response`

For details: [x402 repository](https://github.com/coinbase/x402) | [ecosystem](https://x402.org/ecosystem)
