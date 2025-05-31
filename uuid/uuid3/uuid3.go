package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strings"
)

// Parse UUID string to 16-byte array
func parseUUID(uuidStr string) [16]byte {
	var uuid [16]byte
	clean := strings.ReplaceAll(uuidStr, "-", "")
	for i := 0; i < 16; i++ {
		fmt.Sscanf(clean[i*2:i*2+2], "%02x", &uuid[i])
	}
	return uuid
}

func generateUUIDv3(namespaceUUID string, name string) string {
	ns := parseUUID(namespaceUUID)
	data := append(ns[:], []byte(name)...)

	hash := md5.Sum(data)

	// Set version to 3 (name-based MD5)
	hash[6] = (hash[6] & 0x0F) | 0x30
	// Set variant to 10xx
	hash[8] = (hash[8] & 0x3F) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(hash[0:4]),
		binary.BigEndian.Uint16(hash[4:6]),
		binary.BigEndian.Uint16(hash[6:8]),
		binary.BigEndian.Uint16(hash[8:10]),
		hash[10:16],
	)
}

func main() {
	// Example: DNS namespace (standard UUID)
	dnsNamespace := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	name := "example.com"

	uuid := generateUUIDv3(dnsNamespace, name)
	fmt.Println("UUIDv3:", uuid)
}
