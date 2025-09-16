1. Clone `https://github.com/tbocek/qotp` locally
2. Clone `qh` repo and run `go get github.com/tbocek/qotp@v0.0.0`



# QH Protocol  
**Request for Comments: QH/1.0**  
Category: Experimental  
Status: Draft  

---

## 1. Introduction
The QH Protocol (Quite Ok HTTP Protocol) is a client-server, text-based communication protocol inspired by HTTP. It defines a simple, extensible mechanism for exchanging structured requests and responses over qotp.  

QH is designed to be machine readable.  

---

## 2. Terminology
- **Client**: The initiating party that sends a request.  
- **Server**: The receiving party that processes a request and sends back a response.  
- **Message**: Either a request or a response, consisting of a start line, headers, and an optional body.  
- **Header**: A key-value pair providing metadata about a message.  

---

## 3. Protocol Overview
QH uses a **request/response model**.  
- The client sends a request message to the server.  
- The server replies with a response message.  
- Each message is encoded in UTF-8 text unless otherwise specified.  
- Messages are delimited by a blank line between headers and body.  

---

## 4. Message Format

### 4.1 Request
A request consists of:

```text
<Method> <Path> QH/1.0
<Header-Name>: <Header-Value>
<Header-Name>: <Header-Value>
<Optional Body>
```

Example:

```text
GET /hello.txt QH/1.0
Host: example.com
Accept: text/plain
```

### 4.2 Response

A response consists of:

```text
QH/1.0 <Status-Code> <Reason-Phrase>
<Header-Name>: <Header-Value>
<Header-Name>: <Header-Value>

<Optional Body>
```

Example:

```text
QH/1.0 200 OK
Content-Type: text/plain
Content-Length: 13

Hello, world!
```

## 5. Methods

QH/1.0 defines the following methods:

- `GET` → Retrieve a resource.
- `POST` → Submit data to the server.
- `PUT` → Replace a resource.
- `DELETE` → Remove a resource.

Future extensions MAY define additional methods.

## 6. Status Codes

QH/1.0 status codes are three-digit integers grouped by category:

- `1xx` Informational — Request received, continuing process.
- `2xx` Success — Request successfully processed (e.g., `200 OK`).
- `3xx` Redirection — Further action is needed (e.g., `301 Moved`).
- `4xx` Client Error — The client sent a bad request (e.g., `404 Not Found`).
- `5xx` Server Error — The server failed to process a valid request (e.g., `500 Internal Error`).

## 7. Headers

Headers are key-value pairs separated by `:`.

Examples:

```text
Host: example.com
Content-Type: application/json
Content-Length: 42
```

Header names are case-insensitive.

Header values MAY be free-form text unless otherwise specified.

An empty line (CRLF CRLF) marks the end of the header section.

## 8. Transport

QH is designed to be transported over **qotp**, a secure, reliable, stream-multiplexed protocol running on top of UDP.

`qotp` provides an encrypted transport layer, similar in concept to QUIC, handling reliability and congestion control internally.

A single `qotp` connection can carry multiple concurrent streams, allowing for parallel requests and responses without head-of-line blocking.

## 9. Security Considerations

QH inherits its security properties from the underlying **qotp** transport protocol.

`qotp` provides mandatory, built-in, end-to-end encryption for all connections, ensuring both confidentiality and integrity of data in transit. This is analogous to running HTTP over TLS (HTTPS).

While the transport is secure, implementations MUST still validate input to avoid application-level vulnerabilities such as buffer overflows, header injection, and other common attacks.

Future specifications MAY define authentication headers or security extensions.

## 10. Versioning

This document specifies QH/1.0.

Future versions MAY introduce new methods, headers, or binary framing.

Backward compatibility SHOULD be maintained where possible.

Clients and servers MUST include the protocol version in the request and response start lines.



packetdiag {
  colwidth = 32;
  node_height = 64;

  0-31: Start-Line\n(Request-Line or Status-Line);
  32-63: Header Field 1:\nHost: www.example.com;
  64-95: Header Field 2:\nUser-Agent: ...;
  96-127: Header Field 3:\nAccept: ...;
  128-159: Header Field N:\n(field-name: value);
  160-191: CRLF\n(blank line);
  192-223: Message Body\n(optional);
}
