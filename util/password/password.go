package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"server-api-admin/config"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength   = 16
	timeParam    = 1
	memoryParam  = 64 * 1024
	threadsParam = 4
	keyLength    = 32
)

func GeneratePasswordHash(password string) (string, string, error) {
	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	// Combine password and pepper
	passwordWithPepper := []byte(password + config.Pepper)

	// Generate the hash
	hash := argon2.IDKey(passwordWithPepper, salt, timeParam, memoryParam, threadsParam, keyLength)

	// Encode salt and hash as base64
	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedHash := base64.StdEncoding.EncodeToString(hash)

	// Create the complete hash string
	completeHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memoryParam, timeParam, threadsParam, encodedSalt, encodedHash)

	return completeHash, encodedSalt, nil
}

// ComparePasswordWithHash compares the input password with the stored hash
func ComparePasswordWithHash(password, storedHash, storedSalt string) (bool, error) {
	// Decode the stored hash parts
	parts := strings.Split(storedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Combine password and pepper
	passwordWithPepper := []byte(password + config.Pepper)

	// Compute the hash of the input password
	computedHash := argon2.IDKey(passwordWithPepper, salt, time, memory, threads, uint32(len(decodedHash)))

	// Compare the computed hash with the stored hash
	return subtle.ConstantTimeCompare(decodedHash, computedHash) == 1, nil
}
