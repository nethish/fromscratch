package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"golang.org/x/net/http2/hpack"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Step 1: Send HTTP/2 client connection preface
	preface := "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
	_, err = conn.Write([]byte(preface))
	checkErr(err)
	fmt.Println("✔ Sent HTTP/2 client preface")

	// Step 2: Send SETTINGS frame (empty)
	settings := buildSettingsFrame()
	_, err = conn.Write(settings)
	checkErr(err)
	fmt.Println("✔ Sent SETTINGS frame")

	// Step 3: Send HEADERS frame on Stream 1
	headers := buildHeadersFrame(1)
	_, err = conn.Write(headers)
	checkErr(err)
	fmt.Println("✔ Sent HEADERS frame")

	// Step 4: Send DATA frame
	data := buildDataFrame(1, []byte("Hello Serverrrr!"))
	_, err = conn.Write(data)
	checkErr(err)
	fmt.Println("✔ Sent DATA frame with END_STREAM")

	// Optional: Read response (not decoding HPACK yet)
	readResponse(conn)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}

// SETTINGS frame (no payload)
func buildSettingsFrame() []byte {
	payload := []byte{}
	frame := make([]byte, 9)
	frame[3] = 0x4 // SETTINGS
	// Length already 0, Flags = 0, StreamID = 0
	return append(frame, payload...)
}

func buildHeadersFrame(streamID uint32) []byte {
	var buf bytes.Buffer
	encoder := hpack.NewEncoder(&buf)

	// HPACK encode the pseudo-headers
	encoder.WriteField(hpack.HeaderField{Name: ":method", Value: "POST"})
	encoder.WriteField(hpack.HeaderField{Name: ":path", Value: "/"})
	encoder.WriteField(hpack.HeaderField{Name: ":scheme", Value: "http"})
	encoder.WriteField(hpack.HeaderField{Name: ":authority", Value: "localhost"})

	headerBlock := buf.Bytes()

	// Frame header
	frame := make([]byte, 9)
	frame[3] = 0x1 // HEADERS
	frame[4] = 0x4 // END_HEADERS
	binary.BigEndian.PutUint32(frame[5:], streamID&0x7FFFFFFF)

	// Length
	length := len(headerBlock)
	frame[0] = byte(length >> 16)
	frame[1] = byte(length >> 8)
	frame[2] = byte(length)

	return append(frame, headerBlock...)
}

func buildDataFrame(streamID uint32, data []byte) []byte {
	frame := make([]byte, 9)
	frame[3] = 0x0 // DATA
	frame[4] = 0x1 // END_STREAM
	binary.BigEndian.PutUint32(frame[5:], streamID&0x7FFFFFFF)

	length := len(data)
	frame[0] = byte(length >> 16)
	frame[1] = byte(length >> 8)
	frame[2] = byte(length)

	return append(frame, data...)
}

func readResponse(conn net.Conn) {
	decoder := hpack.NewDecoder(4096, nil)
	for {
		header := make([]byte, 9)
		_, err := conn.Read(header)
		if err != nil {
			fmt.Println("✅ Connection closed")
			return
		}

		length := int(header[0])<<16 | int(header[1])<<8 | int(header[2])
		ftype := header[3]
		flags := header[4]
		streamID := binary.BigEndian.Uint32(header[5:]) & 0x7FFFFFFF

		payload := make([]byte, length)
		_, err = conn.Read(payload)
		if err != nil {
			fmt.Println("❌ Error reading payload:", err)
			return
		}

		switch ftype {
		case 0x1: // HEADERS
			fmt.Printf("\n📦 HEADERS frame on stream %d:\n", streamID)
			fields, err := decoder.DecodeFull(payload)
			if err != nil {
				fmt.Println("❌ HPACK decode error:", err)
			}
			for _, f := range fields {
				fmt.Printf("  %s: %s\n", f.Name, f.Value)
			}
		case 0x0: // DATA
			fmt.Printf("\n📦 DATA frame on stream %d:\n", streamID)
			fmt.Println("  ", string(payload))

		case 0x3: // RST_STREAM
			if len(payload) < 4 {
				fmt.Printf("❌ RST_STREAM frame too short\n")
				break
			}
			errorCode := binary.BigEndian.Uint32(payload)
			fmt.Printf("\n⛔️ RST_STREAM on stream %d: error code 0x%x\n", streamID, errorCode)
		case 0x4: // SETTINGS
			fmt.Println("\n⚙️ SETTINGS frame (usually ACK)")

		case 0x5: // PUSH_PROMISE
			fmt.Printf("\n📦 PUSH_PROMISE frame on stream %d:\n", streamID)

			// First 4 bytes = Promised Stream ID
			promisedID := binary.BigEndian.Uint32(payload[:4]) & 0x7FFFFFFF
			fmt.Printf("  📬 Promised stream ID: %d\n", promisedID)

			hpackPayload := payload[4:]
			fields, err := decoder.DecodeFull(hpackPayload)
			if err != nil {
				fmt.Println("❌ HPACK decode error:", err)
			}
			for _, f := range fields {
				fmt.Printf("    %s: %s\n", f.Name, f.Value)
			}
		case 0x6: // PING
			fmt.Println("\n🏓 PING frame")
		default:
			fmt.Printf("\n❓ Unknown frame type 0x%x on stream %d\n", ftype, streamID)
		}

		if flags&0x1 != 0 {
			fmt.Println("🚪 END_STREAM received – done reading")
			return
		}
	}
}
