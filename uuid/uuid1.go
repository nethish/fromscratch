package main

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const uuidEpochStart = 122192928000000000 // Offset between UUID and Unix epoch in 100-ns

func getTimestamp100ns() uint64 {
	now := time.Now().UTC()
	unix100ns := uint64(now.UnixNano() / 100)
	return unix100ns + uuidEpochStart
}

func getMacAddress() [6]byte {
	var mac [6]byte
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) >= 6 {
			copy(mac[:], iface.HardwareAddr[:6])
			return mac
		}
	}
	// fallback to random MAC
	_, _ = cryptoRand.Read(mac[:])
	mac[0] |= 0x01 // set multicast bit to indicate it's not globally unique
	return mac
}

func generateClockSequence() uint16 {
	var b [2]byte
	_, _ = cryptoRand.Read(b[:])
	return binary.BigEndian.Uint16(b[:]) & 0x3FFF // 14 bits
}

func generateUUIDv1() string {
	timestamp := getTimestamp100ns()

	timeLow := uint32(timestamp & 0xFFFFFFFF)
	timeMid := uint16((timestamp >> 32) & 0xFFFF)
	timeHi := uint16((timestamp >> 48) & 0x0FFF)
	timeHi |= 0x1000 // version 1

	clockSeq := generateClockSequence()
	clockSeqHi := byte((clockSeq >> 8) & 0x3F)
	clockSeqHi |= 0x80 // variant 10xxxxxx
	clockSeqLow := byte(clockSeq & 0xFF)

	node := getMacAddress()

	// Format to UUID string
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%012x",
		timeLow, timeMid, timeHi,
		clockSeqHi, clockSeqLow,
		binary.BigEndian.Uint64(append([]byte{0, 0}, node[:]...))&0xFFFFFFFFFFFF,
	)
}

func main() {
	fmt.Println("UUIDv1 (manual):", generateUUIDv1())
}
