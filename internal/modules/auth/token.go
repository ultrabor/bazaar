package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func newToken() (string, error) {
	buff := make([]byte, 32)

	_, err := rand.Read(buff)
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buff), nil
}

func hashToken(token string) string {
	hash256 := sha256.Sum256([]byte(token))

	hash := hex.EncodeToString(hash256[:])

	return hash
}
