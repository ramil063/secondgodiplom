package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// GetTokenHash рандомный набор символов
func GetTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
