package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	FrameHeaderLen = 9
	// https://datatracker.ietf.org/doc/html/rfc9113#name-http-2-connection-preface
	// This is how the connection must start for HTTP2
	ConnectionPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

type FrameHeader struct {
	Length   uint32
	Type     uint8
	Flags    uint8
	StreamID uint32
}

func readFrameHeader(r io.Reader) (*FrameHeader, error) {
	head := make([]byte, FrameHeaderLen)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, err
	}

	length := uint32(head[0])<<16 | uint32(head[1])<<8 | uint32(head[2])
	frameType := head[3]
	flags := head[4]
	streamID := binary.BigEndian.Uint32(head[5:]) & 0x7FFFFFFF

	return &FrameHeader{
		Length:   length,
		Type:     frameType,
		Flags:    flags,
		StreamID: streamID,
	}, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the connection preface
	preface := make([]byte, len(ConnectionPreface))
	if _, err := io.ReadFull(conn, preface); err != nil {
		log.Println("Failed to read preface:", err)
		return
	}
	if string(preface) != ConnectionPreface {
		log.Println("Invalid connection preface")
		return
	}
	fmt.Println("Received valid HTTP/2 connection preface")

	// Loop and read frames
	for {
		header, err := readFrameHeader(conn)
		if err != nil {
			log.Println("Error reading frame header:", err)
			return
		}
		fmt.Printf("\n--- Frame ---\nLength: %d\nType: %d\nFlags: %d\nStreamID: %d\n",
			header.Length, header.Type, header.Flags, header.StreamID)

		payload := make([]byte, header.Length)
		if _, err := io.ReadFull(conn, payload); err != nil {
			log.Println("Error reading frame payload:", err)
			return
		}
		fmt.Printf("Payload: %x\n", payload)
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Listening on :8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
