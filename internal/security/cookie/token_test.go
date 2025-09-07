package cookie

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSaveTokens_TableDriven(t *testing.T) {
	testCases := []struct {
		name         string
		accessToken  string
		refreshToken string
		expiresIn    int64
		expectError  bool
	}{
		{
			name:         "valid tokens",
			accessToken:  "valid_access",
			refreshToken: "valid_refresh",
			expiresIn:    3600,
			expectError:  false,
		},
		{
			name:         "empty tokens",
			accessToken:  "",
			refreshToken: "",
			expiresIn:    0,
			expectError:  false,
		},
		{
			name:         "special characters",
			accessToken:  "token_with_~!@#$%^&*()",
			refreshToken: "refresh_with_特殊字符",
			expiresIn:    999999,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "table_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			err = SaveTokens(tc.accessToken, tc.refreshToken, tmpFile.Name(), tc.expiresIn)

			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Если ошибки не ожидалось, проверяем содержимое
			if !tc.expectError && err == nil {
				verifyFileContent(t, tmpFile.Name(), tc.accessToken, tc.refreshToken, tc.expiresIn)
			}
		})
	}
}

func verifyFileContent(t *testing.T, filename, expectedAccess, expectedRefresh string, expectedExpires int64) {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	var tokens TokenStorage
	if err := json.Unmarshal(data, &tokens); err != nil {
		t.Fatal(err)
	}

	if tokens.AccessToken != expectedAccess {
		t.Errorf("Access token mismatch: expected %s, got %s", expectedAccess, tokens.AccessToken)
	}
	if tokens.RefreshToken != expectedRefresh {
		t.Errorf("Refresh token mismatch: expected %s, got %s", expectedRefresh, tokens.RefreshToken)
	}
	if tokens.ExpiresIn != expectedExpires {
		t.Errorf("Expires in mismatch: expected %d, got %d", expectedExpires, tokens.ExpiresIn)
	}
}

func TestLoadTokens_IntegrationWithSave(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "integration_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Сохраняем токены
	expectedAccess := "integration_access"
	expectedRefresh := "integration_refresh"
	expectedExpires := time.Now().Add(time.Hour).Unix()

	err = SaveTokens(expectedAccess, expectedRefresh, tmpFile.Name(), expectedExpires)
	if err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	// Загружаем токены
	access, refresh, expires, err := LoadTokens(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}

	// Проверяем что данные совпадают
	if access != expectedAccess {
		t.Errorf("Access token mismatch: expected %s, got %s", expectedAccess, access)
	}
	if refresh != expectedRefresh {
		t.Errorf("Refresh token mismatch: expected %s, got %s", expectedRefresh, refresh)
	}
	if expires != expectedExpires {
		t.Errorf("Expires in mismatch: expected %d, got %d", expectedExpires, expires)
	}
}

func TestLoadTokens_TableDriven(t *testing.T) {
	testCases := []struct {
		name          string
		prepareFile   func(string) error
		expectError   bool
		errorContains string
	}{
		{
			name: "valid file",
			prepareFile: func(filename string) error {
				tokens := TokenStorage{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    100,
				}
				data, _ := json.Marshal(tokens)
				return os.WriteFile(filename, data, 0644)
			},
			expectError: false,
		},
		{
			name: "wrong JSON structure",
			prepareFile: func(filename string) error {
				return os.WriteFile(filename, []byte(`{"wrong_field": "value"}`), 0644)
			},
			expectError: false, // Должно работать, неизвестные поля игнорируются
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "table_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			if err := tc.prepareFile(tmpFile.Name()); err != nil {
				t.Fatalf("Failed to prepare file: %v", err)
			}

			_, _, _, err = LoadTokens(tmpFile.Name())

			if tc.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error should contain '%s', got: %v", tc.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
