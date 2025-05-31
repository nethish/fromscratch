package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

var counter uint32 = randomCounterSeed()

func main() {
	oid := NewObjectID()
	fmt.Printf("Mongo ObjectId: %s\n", oid.Hex())
}

// ObjectID is a 12-byte MongoDB ObjectId
type ObjectID [12]byte

// NewObjectID generates a new ObjectID
func NewObjectID() ObjectID {
	var b ObjectID

	// 1. Timestamp (first 4 bytes)
	timestamp := uint32(time.Now().Unix())
	b[0] = byte(timestamp >> 24)
	b[1] = byte(timestamp >> 16)
	b[2] = byte(timestamp >> 8)
	b[3] = byte(timestamp)

	// 2. Machine ID (3 bytes - random for portability)
	copy(b[4:7], machineID())

	// 3. Process ID (2 bytes)
	pid := uint16(os.Getpid())
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)

	// 4. Counter (3 bytes)
	i := atomic.AddUint32(&counter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)

	return b
}

// Hex returns the ObjectID as a hex string
func (id ObjectID) Hex() string {
	return hex.EncodeToString(id[:])
}

// machineID generates a random 3-byte machine ID
func machineID() []byte {
	var sum [3]byte
	_, err := rand.Read(sum[:])
	if err != nil {
		panic("cannot generate machine ID")
	}
	return sum[:]
}

// randomCounterSeed gives a random starting point for the counter
func randomCounterSeed() uint32 {
	var b [3]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic("cannot generate random counter seed")
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}
