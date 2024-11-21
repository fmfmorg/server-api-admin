package middlewares

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"server-api-admin/config"
)

var secretKey = []byte(config.SessionIDSecretKey)

// EncryptSessionID encrypts the session ID using AES-256 in GCM mode.
func EncryptSessionID(sessionID string) (string, error) {
	if len(secretKey) != 16 && len(secretKey) != 24 && len(secretKey) != 32 {
		return "", errors.New("invalid key size for AES encryption")
	}

	block, err := aes.NewCipher(secretKey)
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(sessionID), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSessionID decrypts the session ID using AES-256 in GCM mode.
func DecryptSessionID(encryptedSessionID string) (string, error) {
	if len(secretKey) != 16 && len(secretKey) != 24 && len(secretKey) != 32 {
		return "", errors.New("invalid key size for AES decryption")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedSessionID)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
