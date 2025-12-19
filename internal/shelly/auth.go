// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"crypto/rand"
	"math/big"
)

const (
	// DefaultPasswordLength is the default length for generated passwords.
	DefaultPasswordLength = 16
	// PasswordCharset contains characters used for password generation.
	PasswordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
)

// GeneratePassword creates a cryptographically secure random password of the specified length.
// If length is less than 8, it defaults to 8 for security.
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(PasswordCharset)))

	for i := range length {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = PasswordCharset[n.Int64()]
	}

	return string(result), nil
}
