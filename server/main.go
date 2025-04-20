package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/net/http2/hpack"
)

const (

	// https://datatracker.ietf.org/doc/html/rfc9113#name-http-2-connection-preface
	// This is how the connection must start for HTTP2
	clientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Listening for h2c (HTTP/2 over TCP) on http://localhost:8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// Step 1: Read client preface
	preface := make([]byte, len(clientPreface))
	if _, err := io.ReadFull(conn, preface); err != nil {
		log.Println("Failed to read client preface:", err)
		return
	}
	if string(preface) != clientPreface {
		log.Printf("Invalid client preface: %q\n", preface)
		return
	}
	log.Println("Received valid HTTP/2 client preface")

	// Step 2: Send SETTINGS frame (empty for now)
	if err := sendSettingsFrame(conn); err != nil {
		log.Println("Failed to send SETTINGS frame:", err)
		return
	}
	log.Println("Sent SETTINGS frame")

	for {
		if err := readFrame(conn); err != nil {
			log.Println("Connection closed or error:", err)
			return
		}
	}
}

func readFrame(conn net.Conn) error {
	// Step 1: Read 9-byte frame header
	header := make([]byte, 9)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("error reading frame header: %w", err)
	}

	length := int(header[0])<<16 | int(header[1])<<8 | int(header[2])
	frameType := header[3]
	flags := header[4]
	streamID := int(header[5]&0x7F)<<24 | int(header[6])<<16 | int(header[7])<<8 | int(header[8])

	// Step 2: Read payload
	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(conn, payload); err != nil {
			return fmt.Errorf("error reading frame payload: %w", err)
		}
	}

	// Step 3: Handle known frame types
	switch frameType {
	case 0x4: // SETTINGS
		if flags&0x1 == 0x1 {
			log.Printf("Received SETTINGS ACK (stream=%d)", streamID)
		} else {
			log.Printf("Received SETTINGS frame with %d bytes (stream=%d)", length, streamID)
		}
	case 0x6: // PING
		log.Printf("Received PING frame: %x (ack=%t)", payload, flags&0x1 == 0x1)
	case 0x0: // DATA
		log.Printf("Stream %d: Received DATA (len=%d)", streamID, len(payload))
		stream, ok := streams[streamID]
		if !ok {
			log.Printf("Stream %d not found for DATA frame", streamID)
			break
		}
		stream.data = append(stream.data, payload...)

		if flags&0x1 == 0x1 { // END_STREAM
			stream.closed = true
			log.Printf("Stream %d: END_STREAM received. Full data: %q", streamID, stream.data)

			// Respond with echo
			respondWithEcho(conn, stream)
		}
	case 0x1: // HEADERS
		log.Printf("Received HEADERS frame (len=%d)", len(payload))
		headers, err := decodeHeaders(payload)
		if err != nil {
			log.Println("Failed to decode HPACK headers:", err)
		} else {
			log.Printf("Stream %d: Received HEADERS:\n", streamID)
			for _, hf := range headers {
				log.Printf("  %s: %s", hf.Name, hf.Value)
			}

			// Create stream
			stream := &streamState{
				id:      streamID,
				headers: headers,
			}
			streams[streamID] = stream
			// respondHello(conn, streamID)
		}

	default:
		log.Printf("Received unknown frame type: 0x%x (len=%d)", frameType, len(payload))
	}

	return nil
}

func sendSettingsFrame(conn net.Conn) error {
	// SETTINGS frame: type = 0x4, flags = 0x0, stream ID = 0
	header := make([]byte, 9)

	// Length: 0 (no payload)
	header[0] = 0x00
	header[1] = 0x00
	header[2] = 0x00

	// Type: SETTINGS (0x4)
	header[3] = 0x4

	// Flags: 0
	header[4] = 0x0

	// Stream Identifier: 0 (connection-level frame)
	header[5] = 0x0
	header[6] = 0x0
	header[7] = 0x0
	header[8] = 0x0

	_, err := conn.Write(header)
	return err
}

func decodeHeaders(payload []byte) ([]hpack.HeaderField, error) {
	// Skip the first part of the payload: PADDED and PRIORITY flags may affect layout.
	// For now we assume: no padding, no priority (works with curl).

	// HEADERS frame starts with:
	// [0] ... optional flags (we assume no PRIORITY info)

	// Create HPACK decoder
	decoder := hpack.NewDecoder(4096, nil)
	headers, err := decoder.DecodeFull(payload)
	if err != nil {
		return nil, err
	}
	return headers, nil
}

func respondHello(conn net.Conn, streamID int) {
	// Step 1: Encode HEADERS with HPACK
	var buf bytes.Buffer
	encoder := hpack.NewEncoder(&buf)

	headers := []hpack.HeaderField{
		{Name: ":status", Value: "200"},
		{Name: "content-type", Value: "text/plain"},
	}

	for _, hf := range headers {
		if err := encoder.WriteField(hf); err != nil {
			log.Println("Failed to encode header:", err)
			return
		}
	}
	headerBlock := buf.Bytes()

	// Step 2: Send HEADERS frame
	sendFrame(conn, 0x1, 0x4, streamID, headerBlock) // 0x4 = END_HEADERS

	// Step 3: Send DATA frame with "Hello, world!"
	data := []byte("Hello, world!\n")
	sendFrame(conn, 0x0, 0x1, streamID, data) // 0x1 = END_STREAM
}

func sendFrame(conn net.Conn, frameType byte, flags byte, streamID int, payload []byte) {
	length := len(payload)
	header := []byte{
		byte(length >> 16), byte(length >> 8), byte(length),
		frameType,
		flags,
		byte(streamID >> 24 & 0x7F), byte(streamID >> 16), byte(streamID >> 8), byte(streamID),
	}
	conn.Write(header)
	conn.Write(payload)
}

type streamState struct {
	id      int
	headers []hpack.HeaderField
	data    []byte
	closed  bool
}

var streams = make(map[int]*streamState)

func respondWithEcho(conn net.Conn, stream *streamState) {
	// Step 1: Headers
	var buf bytes.Buffer
	encoder := hpack.NewEncoder(&buf)

	headers := []hpack.HeaderField{
		{Name: ":status", Value: "200"},
		{Name: "content-type", Value: "text/plain"},
	}

	for _, hf := range headers {
		_ = encoder.WriteField(hf)
	}
	headerBlock := buf.Bytes()

	// 0x4 - END_HEADERS
	sendFrame(conn, 0x1, 0x4, stream.id, headerBlock)

	// Step 2: Send DATA
	// 0x1 END_STREAM
	sendFrame(conn, 0x0, 0x1, stream.id, stream.data)

	// If you want to send more streams then you'll do
	// sendFrame(conn, 0x0, 0x0, stream.id, stream.data)
	// sendFrame(conn, 0x0, 0x1, stream.id, stream.data) // END_STREAM
}
