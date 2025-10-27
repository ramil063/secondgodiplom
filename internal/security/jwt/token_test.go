package jwt

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestGenerateAccessToken_Integration(t *testing.T) {
	// Тестовые данные
	testUserID := 123
	testSecret := "my_super_secret_key_123456"

	// Генерируем токен
	tokenString, err := GenerateAccessToken(testUserID, testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Верифицируем токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(testSecret), nil
	})

	if err != nil {
		t.Fatalf("Token verification failed: %v", err)
	}

	// Проверяем claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверяем user_id
		if int(claims["user_id"].(float64)) != testUserID {
			t.Errorf("Expected user_id %d, got %v", testUserID, claims["user_id"])
		}

		// Проверяем sub
		if int(claims["sub"].(float64)) != testUserID {
			t.Errorf("Expected sub %d, got %v", testUserID, claims["sub"])
		}

		// Проверяем type
		if claims["type"] != "access" {
			t.Errorf("Expected type 'access', got %v", claims["type"])
		}

		// Проверяем exp (должен быть в будущем)
		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		if exp.Before(time.Now()) {
			t.Errorf("Token expiration is in the past: %v", exp)
		}

		// Проверяем iat (должен быть в прошлом/настоящем)
		iat := time.Unix(int64(claims["iat"].(float64)), 0)
		if iat.After(time.Now()) {
			t.Errorf("Token issued at is in the future: %v", iat)
		}

	} else {
		t.Error("Invalid token claims")
	}
}

func TestGenerateAccessToken_TableDriven(t *testing.T) {
	testCases := []struct {
		name          string
		userID        int
		secret        string
		expectError   bool
		errorContains string
		setup         func() // Дополнительная настройка
	}{
		{
			name:        "valid token generation",
			userID:      1,
			secret:      "valid_secret_key_1234567890",
			expectError: false,
		},
		{
			name:        "another valid user",
			userID:      999,
			secret:      "different_secret_123456",
			expectError: false,
		},
		{
			name:        "zero user id",
			userID:      0,
			secret:      "secret_key",
			expectError: false,
		},
		{
			name:        "negative user id",
			userID:      -1,
			secret:      "secret_key",
			expectError: false,
		},
		{
			name:        "very long secret",
			userID:      1,
			secret:      strings.Repeat("a", 1000), // Длинный секрет
			expectError: false,
		},
		{
			name:        "special chars in secret",
			userID:      1,
			secret:      "secret!@#$%^&*()_+-=[]{}|;:,.<>?",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			tokenString, err := GenerateAccessToken(tc.userID, tc.secret)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error should contain '%s', got: %v", tc.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Basic validation of the token
			if tokenString == "" {
				t.Error("Generated token should not be empty")
			}

			// Check if it looks like a JWT token (3 parts separated by dots)
			parts := strings.Split(tokenString, ".")
			if len(parts) != 3 {
				t.Errorf("JWT token should have 3 parts, got %d", len(parts))
			}

			// Verify we can parse it back with the same secret
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(tc.secret), nil
			})

			if err != nil {
				t.Errorf("Generated token should be verifiable: %v", err)
			}

			if !token.Valid {
				t.Error("Generated token should be valid")
			}

			// Verify claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if int(claims["user_id"].(float64)) != tc.userID {
					t.Errorf("Expected user_id %d, got %v", tc.userID, claims["user_id"])
				}
				if claims["type"] != "access" {
					t.Errorf("Expected type 'access', got %v", claims["type"])
				}
			}
		})
	}
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем базовые свойства токена
	if token1 == "" {
		t.Error("Token should not be empty")
	}

	if token2 == "" {
		t.Error("Token should not be empty")
	}

	// Токены должны быть разными (высокая вероятность)
	if token1 == token2 {
		t.Error("Tokens should be different")
	}

	// Проверяем что это hex строка
	if len(token1) != 64 { // 32 байта * 2 (hex)
		t.Errorf("Token should be 64 chars long, got %d", len(token1))
	}

	// Проверяем что это валидный hex
	_, err = hex.DecodeString(token1)
	if err != nil {
		t.Errorf("Token should be valid hex: %v", err)
	}

	_, err = hex.DecodeString(token2)
	if err != nil {
		t.Errorf("Token should be valid hex: %v", err)
	}
}
