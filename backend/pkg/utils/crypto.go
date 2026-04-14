package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// EncryptAES encrypts plaintext using AES-GCM with the given key
func EncryptAES(plaintext string, key string) (string, error) {
	// Pad key to 32 bytes if needed
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		keyBytes = append(keyBytes, make([]byte, 32-len(keyBytes))...)
	}

	block, err := aes.NewCipher(keyBytes[:32])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts ciphertext using AES-GCM with the given key
func DecryptAES(ciphertext string, key string) (string, error) {
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		keyBytes = append(keyBytes, make([]byte, 32-len(keyBytes))...)
	}

	block, err := aes.NewCipher(keyBytes[:32])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonce, ciphertextData := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
