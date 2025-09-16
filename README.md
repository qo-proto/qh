1. Clone `https://github.com/tbocek/qotp` locally
2. Clone `qh` repo and run `go get github.com/tbocek/qotp@v0.0.0`

- see [architecture](./docs/architecture.md) for more details.

# QH Protocol

**Request for Comments: QH/1.0**
Category: Experimental
Status: Draft

---

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
  - [3. Message Format](#3-message-format)
    - [3.1 Message Types](#31-message-types)
    - [3.2 Message Headers](#32-message-headers)
    - [3.3 Messsage Body](#33-messsage-body)
    - [3.4 Messsage Length](#34-messsage-length)
    - [3.5 General Header Fields](#35-general-header-fields)
  - [4. Request](#4-request)
    - [4.1 Methods](#41-methods)
    - [4.2 Request Header Fields](#42-request-header-fields)
    - [4.3 Reqeust Example](#43-reqeust-example)
  - [5 Response](#5-response)
    - [5.1 Status Codes](#51-status-codes)
      - [5.1.1 Status Codes List](#511-status-codes-list)
    - [5.2 Response Header Fields](#52-response-header-fields)
    - [5.3 Response Example](#53-response-example)
  - [6. Headers](#6-headers)
    - [6.1 Request Headers](#61-request-headers)
    - [6.2 Response Headers](#62-response-headers)
  - [7. Transport](#7-transport)
  - [8. Security Considerations](#8-security-considerations)
  - [9. Versioning](#9-versioning)

---

## 1. Introduction

The QH Protocol (Quite Ok HTTP Protocol) is a client-server, text-based communication protocol inspired by HTTP. It defines a simple, extensible mechanism for exchanging structured requests and responses over qotp.

QH is designed to be machine readable.

QH uses a **request/response model**.

- The client sends a request message to the server.
- The server replies with a response message.
- Each message is encoded in UTF-8 text unless otherwise specified.
- Messages are delimited by a blank line between headers and body.

### 1.1 Purpose

The QH Protocol is an application-level protocol for distributed information systems, inspired by HTTP/1.1. Its primary goal is to provide a much simpler and more compact alternative to HTTP for client-server communication, while retaining the core request/response paradigm.

QH was designed to reduce the verbosity and complexity found in HTTP. It achieves this through a simplified message format, such as using ordered, value-only headers instead of key-value pairs. This makes it suitable for environments where bandwidth is limited or parsing overhead needs to be minimized.

While HTTP is a feature-rich protocol for hypermedia systems, QH focuses on providing a fundamental, extensible mechanism for resource exchange over a secure transport.

### 1.2 Terminology

- **Client**: The initiating party that sends a request.
- **Server**: The receiving party that processes a request and sends back a response.
- **Message**: Either a request or a response, consisting of a start line, headers, and an optional body.
- **Header**: A key-value pair providing metadata about a message.

---

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

QH uses language tags to indicate the natural language of the content, primarily within the `Accept-Language` header. A language tag identifies a natural language spoken or written by humans.

The syntax is composed of a primary tag and optional subtags, separated by a hyphen.

```text
language-tag = primary-tag *("-" subtag)
```

Tags are case-insensitive. Examples include `en` (English), `en-US` (American English), and `fr` (French).

---

## 3. Message Format

### 3.1 Message Types

### 3.2 Message Headers

### 3.3 Messsage Body

### 3.4 Messsage Length

### 3.5 General Header Fields

---

## 4. Request

### 4.1 Methods

QH/1.0 defines the following methods:

| QH  | HTTP   | Description                |
| --- | ------ | -------------------------- |
| 1   | `GET`  | Retrieve a resource.       |
| 2   | `POST` | Submit data to the server. |

Future extensions MAY define additional methods.

### 4.2 Request Header Fields

### 4.3 Reqeust Example

A request consists of:
To reduce verbosity, the `Host` is included directly in the start-line, and subsequent header lines contain only the value, omitting the name. The meaning of each header is determined by its order.

```text
<Method> <Host> <Path> <Version>
<Header-1-Value>
...

<Optional Body>
```

Example:

```text
1 example.com /hello.txt 1.0
```

HTTP Example:

```text
GET /hello.txt QH/1.0
Host: example.com
Accept: text/plain
```

---

## 5 Response

### 5.1 Status Codes

QH/1.0 status codes are the same as HTTP, three-digit integers grouped by category:

- `1xx` Informational — Request received, continuing process.
- `2xx` Success — Request successfully processed (e.g., `200 OK`).
- `3xx` Redirection — Further action is needed (e.g., `301 Moved`).
- `4xx` Client Error — The client sent a bad request (e.g., `404 Not Found`).
- `5xx` Server Error — The server failed to process a valid request (e.g., `500 Internal Error`).

#### 5.1.1 Status Codes List

| Status Code | Reason Phrase                   |
| ----------- | ------------------------------- |
| 100         | Continue                        |
| 200         | OK                              |
| 201         | Created                         |
| 202         | Accepted                        |
| 203         | Non-Authoritative Information   |
| 204         | No Content                      |
| 205         | Reset Content                   |
| 206         | Partial Content                 |
| 300         | Multiple Choices                |
| 301         | Moved Permanently               |
| 302         | Found                           |
| 303         | See Other                       |
| 304         | Not Modified                    |
| 305         | Use Proxy                       |
| 307         | Temporary Redirect              |
| 400         | Bad Request                     |
| 401         | Unauthorized                    |
| 402         | Payment Required                |
| 403         | Forbidden                       |
| 404         | Not Found                       |
| 405         | Method Not Allowed              |
| 406         | Not Acceptable                  |
| 407         | Proxy Authentication Required   |
| 408         | Request Time-out                |
| 409         | Conflict                        |
| 410         | Gone                            |
| 411         | Length Required                 |
| 412         | Precondition Failed             |
| 413         | Request Entity Too Large        |
| 414         | Request-URI Too Large           |
| 415         | Unsupported Media Type          |
| 416         | Requested range not satisfiable |
| 417         | Expectation Failed              |
| 500         | Internal Server Error           |
| 501         | Not Implemented                 |
| 502         | Bad Gateway                     |
| 503         | Service Unavailable             |
| 504         | Gateway Time-out                |
| 505         | QH Version not supported        |

### 5.2 Response Header Fields

### 5.3 Response Example

Similar to requests, the response format is optimized for size. The reason phrase is omitted, and headers consist only of their values, identified by order.

```text
<Version> <Status-Code>
<Header-1-Value>
...

<Optional Body>
```

Example:

```text
1.0 200
text/plain
13

Hello, world!
```

---

## 6. Headers

In QH, headers are transmitted as a sequence of values, with their meaning determined by their order in the message. This eliminates the need to send header names, reducing message size.

An empty line still marks the end of the header section. If a header is omitted but a subsequent one is present, an empty line MUST be used as a placeholder.

### 6.1 Request Headers

The following table defines the order and meaning of request headers.

| Order | HTTP              | Description                              | Example                           |
| ----- | ----------------- | ---------------------------------------- | --------------------------------- |
| 1     | `Host`            | The domain name of the server.           | `developer.mozilla.org`           |
| 2     | `Accept`          | Media types the client can process.      | `text/html,application/xhtml+xml` |
| 3     | `Accept-Language` | The preferred language for the response. | `en-US,en;q=0.5`                  |
| 4     | `Accept-Encoding` | Content-coding the client can process.   | `gzip, deflate, br`               |

![QH Message Format](./docs/images/header.svg)

### 6.2 Response Headers

The following table defines the order and meaning of response headers.

| Order | Header                        | Description                                            | Example                         |
| ----- | ----------------------------- | ------------------------------------------------------ | ------------------------------- |
| 1     | `Access-Control-Allow-Origin` | Specifies which origins can access the resource.       | `*`                             |
| 2     | `Content-Encoding`            | The encoding format of the content.                    | `gzip`                          |
| 3     | `Content-Type`                | The MIME type of the resource.                         | `text/html; charset=utf-8`      |
| 4     | `Date`                        | The date and time at which the message was originated. | `Mon, 18 Jul 2016 16:06:00 GMT` |
| 5     | `Set-Cookie`                  | Sends a cookie from the server to the user agent.      | `my-key=my value; ...`          |

## 7. Transport

QH is designed to be transported over **qotp**, a secure, reliable, stream-multiplexed protocol running on top of UDP.

`qotp` provides an encrypted transport layer, similar in concept to QUIC, handling reliability and congestion control internally.

A single `qotp` connection can carry multiple concurrent streams, allowing for parallel requests and responses without head-of-line blocking.

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
