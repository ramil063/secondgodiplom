package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
)

func GenerateAccessToken(userID int, secret string) (string, error) {
	expired := time.Duration(auth.TokenExpiredSeconds)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"sub":     userID,
		"exp":     time.Now().Add(expired * time.Second).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	})
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
