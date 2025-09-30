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

QH uses a `<major>.<minor>` numbering scheme to indicate the protocol version. This policy allows communicating parties to know the message format and capabilities of each other.

- The **`<major>`** number is incremented when a change breaks the fundamental message parsing, such as a change to the overall message structure.
- The **`<minor>`** number is incremented for backward-compatible changes, like adding new methods or headers.

The protocol version is included in the start-line of every request and response. This document specifies version `1.0`.

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

---

## 4. Request

### 4.1 Methods

QH/1.0 supports the following methods:

| Method | Description                |
| ------ | -------------------------- |
| GET    | Retrieve a resource.       |
| POST   | Submit data to the server. |

The method is inferred based on the presence of a message body: a request with a non-empty body is treated as `POST`, and a request with an empty body is treated as `GET`.

### 4.2 Request Format

A request message has the following structure:

```text
<Host>\0<Path>\0<Version>\0<Header-Value-1>\0<Header-Value-2>\0...<Header-Value-N>\0\x03<Body>\x04
```

Where:

- `Host`: Target hostname
- `Path`: Resource path
- `Version`: Protocol version (currently "1.0")
- `Header-Value-N`: Header values in predefined order
- The separator for all fields is a null byte (`\0`)
- The separator between headers and body is the End of Text character (`\x03`)
- The end of the entire message is marked by the End of Transmission character (`\x04`)

### 4.3 Request Examples

**Simple GET request:**

```text
example.com\0/hello.txt\01.0\0\x03\x04
```

**GET request with headers:**

```text
example.com\0/hello.txt\01.0\0
text/html,application/xhtml+xml\0
en-US,en;q=0.5\0
\x03\x04
```

**POST request with body:**

```text
example.com\0/submit\01.0\0
application/json\0
\x03
{"name": "test"}
\x04
```

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

For wire efficiency, common HTTP status codes are mapped to compact single-byte representations:

#### 5.1.1 Supported Status Codes

The following status codes are supported with their compact wire format encoding:

| HTTP Code | Compact Code | Reason Phrase                 |
| --------- | ------------ | ----------------------------- |
| 200       | 1            | OK                            |
| 404       | 2            | Not Found                     |
| 500       | 3            | Internal Server Error         |
| 302       | 4            | Found                         |
| 400       | 5            | Bad Request                   |
| 403       | 6            | Forbidden                     |
| 401       | 7            | Unauthorized                  |
| 301       | 8            | Moved Permanently             |
| 304       | 9            | Not Modified                  |
| 503       | 10           | Service Unavailable           |
| 201       | 11           | Created                       |
| 202       | 12           | Accepted                      |
| 204       | 13           | No Content                    |
| 206       | 14           | Partial Content               |
| 307       | 15           | Temporary Redirect            |
| 308       | 16           | Permanent Redirect            |
| 409       | 17           | Conflict                      |
| 410       | 18           | Gone                          |
| 412       | 19           | Precondition Failed           |
| 413       | 20           | Payload Too Large             |
| 414       | 21           | URI Too Long                  |
| 415       | 22           | Unsupported Media Type        |
| 422       | 23           | Unprocessable Entity          |
| 429       | 24           | Too Many Requests             |
| 502       | 25           | Bad Gateway                   |
| 504       | 26           | Gateway Timeout               |
| 505       | 27           | QH Version Not Supported      |
| 100       | 31           | Continue                      |
| 101       | 32           | Switching Protocols           |
| 102       | 33           | Processing                    |
| 103       | 34           | Early Hints                   |
| 205       | 35           | Reset Content                 |
| 207       | 36           | Multi-Status                  |
| 208       | 37           | Already Reported              |
| 226       | 38           | IM Used                       |
| 300       | 39           | Multiple Choices              |
| 303       | 40           | See Other                     |
| 305       | 41           | Use Proxy                     |
| 402       | 42           | Payment Required              |
| 405       | 43           | Method Not Allowed            |
| 406       | 44           | Not Acceptable                |
| 407       | 45           | Proxy Authentication Required |
| 408       | 46           | Request Timeout               |
| 411       | 47           | Length Required               |
| 416       | 48           | Range Not Satisfiable         |
| 417       | 49           | Expectation Failed            |

**Encoding Rules:**

- Status codes are ordered by frequency to optimize common cases
- Unmapped status codes default to 500 (Internal Server Error) with compact code 3
- The compact code is transmitted in the wire format, then decoded to the HTTP status code

### 5.2 Response Format

A response message has the following structure:

```text
<Version>\0<Compact-Status-Code>\0<Header-Value-1>\0<Header-Value-2>\0...<Header-Value-N>\0\x03<Body>\x04
```

Where:

- `Version`: Protocol version (currently "1.0")
- `Compact-Status-Code`: Single-byte encoded status code (see encoding table above)
- `Header-Value-N`: Header values in predefined order
- The separator for all fields is a null byte (`\0`)
- The separator between headers and body is the End of Text character (`\x03`)
- The end of the entire message is marked by the End of Transmission character (`\x04`)

### 5.3 Response Examples

**Simple successful response:**

```text
1.0\01\0
\x03
Hello, world!
\x04
```

**Response with headers:**

```text
1.0\01\0
*\0
text/plain\0
\x03
Hello, world!
\x04
```

**Empty response (No Content):**

```text
1.0\013\0
\x03
\x04
```

**Error response:**

```text
1.0\02\0
\x03
Page not found
\x04
```

## 6. Headers

In QH, headers are transmitted as a sequence of values, with their meaning determined by their order in the message. This eliminates the need to send header names, reducing message size.

An empty line still marks the end of the header section. If a header is omitted but a subsequent one is present, an empty line MUST be used as a placeholder.

### 6.1 Request Headers

The following table defines the order and meaning of request headers.

Note: `Host` is not included as it appears in the start-line of the request, not as a header.

| Order | Header            | Description                                 | Example                           |
| ----- | ----------------- | ------------------------------------------- | --------------------------------- |
| 1     | `Accept`          | Media types the client can process.         | `text/html,application/xhtml+xml` |
| 2     | `Accept-Language` | The preferred language for the response.    | `en-US,en;q=0.5`                  |
| 3     | `Accept-Encoding` | Content-coding the client can process.      | `gzip, deflate, br`               |
| 4     | `Fragment-Offset` | The byte offset for a fragmented body.      | `1200`                            |
| 5     | `Connection`      | Control options for the current connection. | `close`                           |

![QH Message Format](./docs/images/header.svg)

### 6.2 Response Headers

The following table defines the order and meaning of response headers.

| Order | Header                        | Description                                            | Example                       |
| ----- | ----------------------------- | ------------------------------------------------------ | ----------------------------- |
| 1     | `Access-Control-Allow-Origin` | Specifies which origins can access the resource.       | `*`                           |
| 2     | `Content-Length`              | The size of the message body in octets (8-bit bytes).  | `1234`                        |
| 3     | `Content-Encoding`            | The encoding format of the content.                    | `gzip`                        |
| 4     | `Content-Type`                | The MIME type of the resource.                         | `text/html; charset=utf-8`    |
| 5     | `Date`                        | The date and time at which the message was originated. | `1468857960` (Unix timestamp) |
| 6     | `Content-Language`            | The preferred language for the response.               | `en-US,en;q=0.5`              |
| 7     | `Fragment-Offset`             | The byte offset for a fragmented body.                 | `1200`                        |
| 8     | `Fragment-Request-ID`         | A unique ID to correlate fragments.                    | `42`                          |
| 9     | `Date`                        | The date and time at which the message was originated. | `1468857960` (Unix timestamp) |
| 10    | `Connection`                  | Control options for the current connection.            | `close`                       |

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

This document specifies QH/1.0.

Future versions MAY introduce new methods, headers, or binary framing.

Backward compatibility SHOULD be maintained where possible.

Clients and servers MUST include the protocol version in the request and response start lines.
