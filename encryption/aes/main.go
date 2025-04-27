package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Generate a block from Key
func encryptAES(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	fmt.Println("Block Size:", block.BlockSize()) // 16 bytes
	if err != nil {
		return nil, nil, err
	}

	paddedPlaintext := pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(paddedPlaintext))
	iv := ciphertext[:aes.BlockSize]

	// Generate random IV
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	// Cipher Block Chain
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedPlaintext)

	dec := cipher.NewCBCDecrypter(block, iv)
	buf := make([]byte, len(ciphertext)-aes.BlockSize)
	dec.CryptBlocks(buf, ciphertext[aes.BlockSize:])
	fmt.Println("Decrypted:", string(buf))

	return ciphertext, iv, nil
}

func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func main() {
	// The length of the key determines whether it's AES-128, AES-192 or AES-256
	key := []byte("examplekey123456") // 16 bytes = 128-bit key
	plaintext := []byte("Hello, World!")

	ciphertext, iv, err := encryptAES(plaintext, key)
	if err != nil {
		panic(err)
	}

	fmt.Printf("IV: %x | %s\n", iv, string(iv))
	fmt.Printf("Ciphertext: %x | %s\n", ciphertext, string(ciphertext))
}
