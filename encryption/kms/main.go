package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

type SimpleKMS struct {
	keyStorePath string
	keys         map[string]string
}

func NewSimpleKMS(keyStorePath string) *SimpleKMS {
	kms := &SimpleKMS{keyStorePath: keyStorePath}
	kms.loadKeys()
	return kms
}

func (kms *SimpleKMS) loadKeys() {
	// Load keys from the key store file
	if _, err := os.Stat(kms.keyStorePath); err == nil {
		data, err := os.ReadFile(kms.keyStorePath)
		if err != nil {
			log.Fatalf("Error reading key store: %v", err)
		}
		if err := json.Unmarshal(data, &kms.keys); err != nil {
			log.Fatalf("Error unmarshaling key store: %v", err)
		}
	} else {
		kms.keys = make(map[string]string)
	}
}

func (kms *SimpleKMS) saveKeys() {
	// Save keys to the key store file
	data, err := json.Marshal(kms.keys)
	if err != nil {
		log.Fatalf("Error marshaling keys: %v", err)
	}
	err = os.WriteFile(kms.keyStorePath, data, 0644)
	if err != nil {
		log.Fatalf("Error writing key store: %v", err)
	}
}

func (kms *SimpleKMS) generateKey(keyID string) string {
	// Generate a random AES key
	key := make([]byte, 32) // AES-256 key
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Error generating key: %v", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(key)
	kms.keys[keyID] = encodedKey
	kms.saveKeys()
	log.Printf("Generated new key: %s", keyID)
	return keyID
}

func (kms *SimpleKMS) getKey(keyID string) ([]byte, error) {
	encodedKey, exists := kms.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key %s not found", keyID)
	}
	return base64.StdEncoding.DecodeString(encodedKey)
}

func (kms *SimpleKMS) encrypt(keyID, plaintext string) (string, error) {
	key, err := kms.getKey(keyID)
	if err != nil {
		return "", err
	}

	// AES Encryption
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Generate random IV (Initialization Vector)
	iv := make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}

	// Pad plaintext to be multiple of AES block size
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padText := append([]byte(plaintext), bytes.Repeat([]byte{byte(padding)}, padding)...)

	// Encrypt the plaintext
	ciphertext := make([]byte, len(padText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padText)

	// Combine IV and ciphertext
	ciphertext = append(iv, ciphertext...)

	// Encode the result in base64 for storage or transmission
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (kms *SimpleKMS) decrypt(keyID, encryptedText string) (string, error) {
	key, err := kms.getKey(keyID)
	if err != nil {
		return "", err
	}

	// Decode the base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Split IV and ciphertext
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// AES Decryption
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Decrypt the ciphertext
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove padding
	padding := int(ciphertext[len(ciphertext)-1])
	ciphertext = ciphertext[:len(ciphertext)-padding]

	return string(ciphertext), nil
}

func (kms *SimpleKMS) rotateKey(keyID string) (string, error) {
	_, exists := kms.keys[keyID]
	if !exists {
		return "", fmt.Errorf("key %s not found", keyID)
	}

	// Generate new key
	newKeyID := keyID + "_new"
	kms.generateKey(newKeyID)

	// Optionally, handle re-encrypting data with the new key here

	// Remove the old key (if needed)
	delete(kms.keys, keyID)
	kms.saveKeys()

	log.Printf("Rotated key: %s -> %s", keyID, newKeyID)
	return newKeyID, nil
}

func main() {
	// Create a new KMS instance
	kms := NewSimpleKMS("keys.json")

	// Generate a new key
	keyID := "myKey"
	kms.generateKey(keyID)

	// Encrypt some data
	plaintext := "Hello, KMS!"
	encryptedData, err := kms.encrypt(keyID, plaintext)
	if err != nil {
		log.Fatalf("Error encrypting data: %v", err)
	}
	fmt.Println("Encrypted Data:", encryptedData)

	// Decrypt the data
	decryptedData, err := kms.decrypt(keyID, encryptedData)
	if err != nil {
		log.Fatalf("Error decrypting data: %v", err)
	}
	fmt.Println("Decrypted Data:", decryptedData)

	// Rotate the key
	newKeyID, err := kms.rotateKey(keyID)
	if err != nil {
		log.Fatalf("Error rotating key: %v", err)
	}
	fmt.Println("New Key ID after rotation:", newKeyID)

	// Encrypt with the new key
	encryptedDataNew, err := kms.encrypt(newKeyID, plaintext)
	if err != nil {
		log.Fatalf("Error encrypting data with new key: %v", err)
	}
	fmt.Println("Encrypted Data with New Key:", encryptedDataNew)
}
