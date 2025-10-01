# QH Protocol

**Request for Comments: QH/1.0**
Category: Experimental
Status: Draft

## Table of Contents

- [QH Protocol](#qh-protocol)
  - [Table of Contents](#table-of-contents)
  - [1. Introduction](#1-introduction)
    - [1.1 Purpose](#11-purpose)
    - [1.2 Terminology](#12-terminology)
  - [2. Protocol Parameters](#2-protocol-parameters)
    - [2.1 QH Version](#21-qh-version)
    - [2.2 Media Types](#22-media-types)
    - [2.3 Language Tags](#23-language-tags)
    - [2.4 qh URI Scheme](#24-qh-uri-scheme)
  - [3. Message Format](#3-message-format)
    - [3.1 Message Types](#31-message-types)
    - [3.2 Message Headers](#32-message-headers)
    - [3.3 Message Body](#33-message-body)
    - [3.4 Message Length](#34-message-length)
    - [3.5 General Header Fields](#35-general-header-fields)
  - [4. Request](#4-request)
    - [4.1 Methods](#41-methods)
    - [4.2 Request Format](#42-request-format)
    - [4.3 Request Examples](#43-request-examples)
  - [5. Response](#5-response)
    - [5.1 Status Codes](#51-status-codes)
      - [Status Code Encoding](#status-code-encoding)
      - [5.1.1 Supported Status Codes](#511-supported-status-codes)
    - [5.2 Response Format](#52-response-format)
    - [5.3 Response Examples](#53-response-examples)
  - [6. Headers](#6-headers)
    - [6.1 Request Headers](#61-request-headers)
    - [6.2 Response Headers](#62-response-headers)
  - [7. Transport](#7-transport)
    - [7.1 Connection Establishment](#71-connection-establishment)
      - [7.1.1 Certificate Exchange](#711-certificate-exchange)
    - [7.2 Connection Management](#72-connection-management)
      - [7.2.1 Connection Reuse](#721-connection-reuse)
      - [7.2.2 Reusing Connections for Multiple Origins](#722-reusing-connections-for-multiple-origins)
  - [8. Security Considerations](#8-security-considerations)
  - [9. Versioning](#9-versioning)

## 1. Introduction

The QH Protocol (Quite Ok HTTP Protocol) is a client-server, text-based communication protocol inspired by HTTP. It defines a simple, extensible mechanism for exchanging structured requests and responses over qotp.

QH is designed to be machine readable.

QH uses a **request/response model**.

- The client sends a request message to the server.
- The server replies with a response message.
- Each message is encoded in UTF-8 text unless otherwise specified.
- The header section is separated from the body by a special character.

### 1.1 Purpose

The QH Protocol is an application-level protocol for distributed information systems, inspired by HTTP/1.1. Its primary goal is to provide a much simpler and more compact alternative to HTTP for client-server communication, while retaining the core request/response paradigm.

QH was designed to reduce the verbosity and complexity found in HTTP. It achieves this through a simplified message format, such as using ordered, value-only headers instead of key-value pairs. This makes it suitable for environments where bandwidth is limited or parsing overhead needs to be minimized.

While HTTP is a feature-rich protocol for hypermedia systems, QH focuses on providing a fundamental, extensible mechanism for resource exchange over a secure transport.

### 1.2 Terminology

- **Client**: The initiating party that sends a request.
- **Server**: The receiving party that processes a request and sends back a response.
- **Message**: Either a request or a response, consisting of a start line, headers, and an optional body.
- **Header**: A value-only line providing metadata about a message. The meaning is determined by position.

## 2. Protocol Parameters

### 2.1 QH Version

QH uses a `<number>` numbering scheme to indicate the protocol version. This policy allows communicating parties to know the message format and capabilities of each other.

The protocol version is included in the start-line of every request and response. This document specifies version `0`.

A server that receives a request with a major version higher than what it supports SHOULD respond with a `505 (Version Not Supported)` error.

### 2.2 Media Types

QH uses Internet Media Types (MIME types) to specify the format of a message body. This is communicated via headers like `Content-Type` and `Accept`.

The format is `type/subtype`, optionally followed by parameters.

```text
media-type = type "/" subtype *( ";" parameter )
```

For example: `text/html; charset=utf-8`. The `type`, `subtype`, and parameter names are case-insensitive. Media types are registered with the Internet Assigned Number Authority (IANA).

### 2.3 Language Tags

A language tag, as defined in [RFC5646], identifies a natural language spoken, written, or otherwise conveyed by human beings for communication of information to other human beings. Computer languages are explicitly excluded.

QH uses language tags within the `Accept-Language` header.

```text
language-tag = <Language-Tag, see [RFC5646], Section 2.1>
```

Tags are case-insensitive. Examples include `en` (English), `en-US` (American English), and `fr` (French).

### 2.4 qh URI Scheme

The "qh" URI scheme is defined for identifying resources that are accessible via the QH protocol. Communication is performed over `qotp`, a secure, UDP-based transport.

```text
qh-URI = "qh" "://" authority path-abempty [ "?" query ]
```

The origin server for a "qh" URI is identified by the `authority` component, which includes a host identifier and an optional port number. If the port is not specified, the default port for QH is `8090`.

A sender MUST NOT generate a "qh" URI with an empty host identifier. A recipient that processes such a URI MUST reject it as invalid.

The hierarchical `path` component and optional `query` component identify the target resource within the origin server's namespace.

All communication over QH is inherently secured by the underlying `qotp` transport, which provides mandatory end-to-end encryption. Clients and servers do not need to perform additional steps to secure the channel, as this is a built-in feature of the transport layer.

Resources made available via the "qh" scheme have no shared identity with resources from "http" or "https" schemes. They are distinct origins with separate namespaces.

## 3. Message Format

### 3.1 Message Types

### 3.2 Message Headers

### 3.3 Message Body

### 3.4 Message Length

### 3.5 General Header Fields

## 4. Request

### 4.1 Methods

QH Versoin 0 defines the following methods. The version and method are encoded in the first byte of the request message as follows:

- **Bits 6-7 (upper 2 bits):** Version
- **Bits 3-5 (middle 3 bits):** Method
- **Bits 0-2 (lower 3 bits):** Reserved

| Method | Code | Description                |
| ------ | ---- | -------------------------- |
| GET    | 000  | Retrieve a resource.       |
| POST   | 001  | Submit data to the server. |

For QH/0, the version number is `0`. So, a GET request uses `\x00` (00000000) and a POST request uses `\x08` (00001000).

### 4.2 Request Format

A request message has the following structure:

```text
<1-byte-field><Host>\0<Path>\0<Header-Value-1>\0...<Header-Value-N>\x03<Body>
```

Where:

- `1-byte-field`: A single byte encoding the version (upper 2 bits) and method (middle 3 bits).
- `Host`: Target hostname.
- `Path`: Resource path
- `Header-Value-N`: Header values in a predefined order.
- The separator for all string-based fields is a null byte (`\0`).
- The separator between headers and body is the End of Text character (`\x03`)
- Content-Length: When a request includes a body, the Content-Length header (header index 3) MUST specify the exact number of bytes in the body.
- Requests without a body (e.g., GET) SHOULD have empty Content-Type and Content-Length headers, and the body is considered empty.
- Completeness: A request is considered complete once headers and the ETX (`\x03`) separator are received; if Content-Length is present and non-empty, the receiver waits until the body size matches Content-Length.

### 4.3 Request Examples

TODO: Add mermaid diagrams for simple get request, with/without headers, with body, ...

## 5. Response

### 5.1 Status Codes

QH/1.0 uses HTTP-compatible status codes but encodes them in a compact wire format for efficiency. Status codes are mapped to single-byte compact codes ordered by frequency of use.

The protocol supports standard HTTP status code categories:

- `1xx` Informational — Request received, continuing process
- `2xx` Success — Request successfully processed (e.g., `200 OK`)
- `3xx` Redirection — Further action is needed (e.g., `301 Moved`)
- `4xx` Client Error — The client sent a bad request (e.g., `404 Not Found`)
- `5xx` Server Error — The server failed to process a valid request (e.g., `500 Internal Error`)

#### Status Code Encoding

For wire efficiency, the response version and status code are encoded into a single byte:

- **Bits 6-7 (upper 2 bits):** Version
- **Bits 0-5 (lower 6 bits):** Compact Status Code

#### 5.1.1 Supported Status Codes

The following status codes are supported with their compact wire format encoding:

| HTTP Code | Compact Code | Reason Phrase                 |
| --------- | ------------ | ----------------------------- |
| 200       | 0            | OK                            |
| 404       | 1            | Not Found                     |
| 500       | 2            | Internal Server Error         |
| 302       | 3            | Found                         |
| 400       | 4            | Bad Request                   |
| 403       | 5            | Forbidden                     |
| 401       | 6            | Unauthorized                  |
| 301       | 7            | Moved Permanently             |
| 304       | 8            | Not Modified                  |
| 503       | 9            | Service Unavailable           |
| 201       | 10           | Created                       |
| 202       | 11           | Accepted                      |
| 204       | 12           | No Content                    |
| 206       | 13           | Partial Content               |
| 307       | 14           | Temporary Redirect            |
| 308       | 15           | Permanent Redirect            |
| 409       | 16           | Conflict                      |
| 410       | 17           | Gone                          |
| 412       | 18           | Precondition Failed           |
| 413       | 19           | Payload Too Large             |
| 414       | 20           | URI Too Long                  |
| 415       | 21           | Unsupported Media Type        |
| 422       | 22           | Unprocessable Entity          |
| 429       | 23           | Too Many Requests             |
| 502       | 24           | Bad Gateway                   |
| 504       | 25           | Gateway Timeout               |
| 505       | 26           | QH Version Not Supported      |
| 100       | 27           | Continue                      |
| 101       | 28           | Switching Protocols           |
| 102       | 29           | Processing                    |
| 103       | 30           | Early Hints                   |
| 205       | 31           | Reset Content                 |
| 207       | 32           | Multi-Status                  |
| 208       | 33           | Already Reported              |
| 226       | 34           | IM Used                       |
| 300       | 35           | Multiple Choices              |
| 303       | 36           | See Other                     |
| 305       | 37           | Use Proxy                     |
| 402       | 38           | Payment Required              |
| 405       | 39           | Method Not Allowed            |
| 406       | 40           | Not Acceptable                |
| 407       | 41           | Proxy Authentication Required |
| 408       | 42           | Request Timeout               |
| 411       | 43           | Length Required               |
| 416       | 44           | Range Not Satisfiable         |
| 417       | 45           | Expectation Failed            |

**Encoding Rules:**

- Status codes are ordered by frequency to optimize common cases
- Unmapped status codes default to 500 (Internal Server Error) with compact code 2.
- The compact code and version are packed into the first byte of the response.

### 5.2 Response Format

A response message has the following structure:

```text
<1-byte-field><Header-Value-1>\0<Header-Value-2>\0...<Header-Value-N>\0\x03<Body>
```

Where:

- `Header-Value-N`: Header values in predefined order
- The separator for all fields is a null byte (`\0`)
- The separator between headers and body is the End of Text character (`\x03`)
- Content-Length: When a response includes a body, the Content-Length header (header index 1) MUST specify the exact number of bytes in the body. Receivers use Content-Length to determine completeness.

### 5.3 Response Examples

TODO: Add mermaid diagrams for simple success response, response with headers, empty response, etc.

## 6. Headers

In QH, headers are transmitted as a sequence of values, with their meaning determined by their order in the message. This eliminates the need to send header names, reducing message size.

An empty line still marks the end of the header section. If a header is omitted but a subsequent one is present, an empty line MUST be used as a placeholder.

### 6.1 Request Headers

The following table defines the order and meaning of request headers.

Note: `Host` is not included as it appears in the start-line of the request, not as a header.

| Index | Header            | Description                                | Example                           |
| ----- | ----------------- | ------------------------------------------ | --------------------------------- |
| 0     | `Accept`          | Media types the client can process.        | `text/html,application/xhtml+xml` |
| 1     | `Accept-Encoding` | Content-coding the client can process.     | `gzip, deflate, br, zstd`         |
| 2     | `Content-Type`    | Media type of the request body (POST/PUT). | `application/json`                |
| 3     | `Content-Length`  | Size of the request body in bytes.         | `12`                              |

![QH Message Format](./docs/images/header.svg)

### 6.2 Response Headers

The following table defines the order and meaning of response headers.

| Index | Header                      | Description                             | Example            |
| ----- | --------------------------- | --------------------------------------- | ------------------ |
| 0     | Content-Type                | Numeric content type code               | 1                  |
| 1     | Content-Length              | Size of the message body in bytes       | 13                 |
| 2     | Cache-Control               | Caching directives                      | max-age=3600       |
| 3     | Content-Encoding            | Content encoding used                   | gzip               |
| 4     | Authorization               | Authentication information              | Bearer <token>     |
| 5     | Access-Control-Allow-Origin | CORS allowed origins                    | \*                 |
| 6     | ETag                        | Entity tag for cache validation         | abc123             |
| 7     | Date                        | Unix timestamp of when response created | 1758784800         |
| 8     | Content-Security-Policy     | CSP directives                          | default-src 'self' |
| 9     | X-Content-Type-Options      | MIME sniffing protection                | nosniff            |
| 10    | X-Frame-Options             | Clickjacking protection                 | SAMEORIGIN         |

## 7. Transport

QH is designed to be transported over **qotp**, a secure, reliable, stream-multiplexed protocol running on top of UDP.

`qotp` provides an encrypted transport layer, similar in concept to QUIC, handling reliability and congestion control internally.

A single `qotp` connection can carry multiple concurrent streams, allowing for parallel requests and responses without head-of-line blocking.

### 7.1 Connection Establishment

#### 7.1.1 Certificate Exchange

QOTP enables you to use it without knowing the servers public certificate. But when we don't know the server public certificate it takes some packets for the handshake. A better approach would be to connect to the server knowing the server public certificate.

To get the server certificate before connecting to the server we try to get the certificate from the DNS.

We need a DNS Entry we can get from the client before connecting.

### 7.2 Connection Management

#### 7.2.1 Connection Reuse

QH connections, which are built on the underlying `qotp` transport, are persistent. For best performance, clients SHOULD reuse connections for multiple requests rather than establishing a new connection for each one. Connections should remain open until it is determined that no further communication with the server is necessary (e.g., when a user navigates away from an application) or until the server closes the connection.

Clients SHOULD NOT open more than one QH connection to a given IP address and UDP port for the same transport configuration.

#### 7.2.2 Reusing Connections for Multiple Origins

A single connection to a server endpoint MAY be reused for requests to different origins (i.e., different hostnames) if they resolve to the same IP address. To do this securely, the client MUST validate that the server is authoritative for the new origin. This involves verifying the server's public key against the expected key for the new origin (e.g., by fetching it from a DNS TXT record).

If the server's identity cannot be verified for the new origin, the client MUST NOT reuse the connection for that origin and SHOULD establish a new connection instead.

A server that receives a request for an origin it is not authoritative for can indicate this by sending a `421 (Misdirected Request)` status code.

## 8. Security Considerations

QH inherits its security properties from the underlying **qotp** transport protocol.

`qotp` provides mandatory, built-in, end-to-end encryption for all connections, ensuring both confidentiality and integrity of data in transit. This is analogous to running HTTP over TLS (HTTPS).

While the transport is secure, implementations MUST still validate input to avoid application-level vulnerabilities such as buffer overflows, header injection, and other common attacks.

Future specifications MAY define authentication headers or security extensions.

## 9. Versioning

This document specifies QH/0.

Future versions MAY introduce new methods, headers, or binary framing.

Backward compatibility SHOULD be maintained where possible.

Clients and servers MUST include the protocol version in the request and response start lines.
