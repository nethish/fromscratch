# HTTP2
This is a mini implementation of http2 from scratch and just uses the TCP connection. 
Explores how streams, frames, settings, connections are handled.

h2c - Cleartext protocol - Without TLS

## Run
Run the server, then the client
```bash
go run ./server/main.go
go run ./client/main.go

```

## How
* The client sends a http2 preface indicating that it wants to initiate a http2 connection
```go
// https://datatracker.ietf.org/doc/html/rfc9113#name-http-2-connection-preface
// This is how the connection must start for HTTP2
clientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
```
* The server sends back settings frame which of length 9 bytes
* Then the client keeps sending frames (payload frame, header frame)
* The client tags frames with stream ID, so that it can respond back with that streamID
  * Client initiated streams are odd
  * Server initiated streams are even (for push streams)

## Server
* RFC 9113
* HTTP2 starts with a preface
```bash
// https://datatracker.ietf.org/doc/html/rfc9113#name-http-2-connection-preface
// This is how the connection must start for HTTP2
clientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
```
* The server sends Setting Frame back
* curl --http2-prior-knowledge http://localhost:8080

HEADER Format
+-----------------------------------------------+
|                 Length (24)                   |
+---------------+---------------+---------------+
|   Type (8)    |   Flags (8)   |      R        |
+-+-------------+---------------+---------------+
|                 Stream Identifier (31)        |
+-----------------------------------------------+

Type

Indicates the frame type (e.g., DATA, HEADERS, SETTINGS, etc.).
Common types (defined in RFC 7540):
0x0: DATA
0x1: HEADERS
0x2: PRIORITY
0x3: RST_STREAM
0x4: SETTINGS
0x5: PUSH_PROMISE
0x6: PING
0x7: GOAWAY
0x8: WINDOW_UPDATE
0x9: CONTINUATION

DATA 
Flag - END_STREAM, PADDED
HEADERS - END_STREAM 0x1
PADDED - 0x8


