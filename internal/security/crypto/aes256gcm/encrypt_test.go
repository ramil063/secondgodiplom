package aes256gcm

import (
	"crypto/rand"
	"reflect"
	"strings"
	"testing"
)

func TestEncryptor_Integration_EncryptDecryptCycle(t *testing.T) {
	encryptor := &Encryptor{}
	decryptor := &Decryptor{} // Предполагаем, что у вас есть Decryptor

	testCases := []struct {
		name string
		data []byte
	}{
		{"short text", []byte("hello")},
		{"long text", []byte("this is a very long secret message that needs to be encrypted")},
		{"empty data", []byte{}},
		{"binary data", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}},
		{"special chars", []byte("secret with !@#$%^&*()_+")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Генерируем 24-байтный ключ
			key := make([]byte, 24)
			_, err := rand.Read(key)
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}

			// Настраиваем encryptor
			encryptor.SetEncryptionKey(key)
			decryptor.SetDecryptionKey(key)

			// Шифруем
			encryptedData, algorithm, iv, err := encryptor.Encrypt(tc.data)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Проверяем возвращаемые значения
			if algorithm != "AES-256-GCM" {
				t.Errorf("Expected algorithm AES-256-GCM, got %s", algorithm)
			}

			if len(iv) != 12 {
				t.Errorf("Expected IV length 12, got %d", len(iv))
			}

			// Дешифруем
			decryptedData, err := decryptor.Decrypt(encryptedData, iv)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Проверяем что данные совпадают
			if string(decryptedData) != string(tc.data) {
				t.Errorf("Expected %s, got %s", string(tc.data), string(decryptedData))
			}

			// Проверяем длину
			if len(decryptedData) != len(tc.data) {
				t.Errorf("Length mismatch: expected %d, got %d", len(tc.data), len(decryptedData))
			}
		})
	}
}

func TestEncryptor_TableDriven(t *testing.T) {
	encryptor := &Encryptor{}

	testCases := []struct {
		name          string
		key           []byte
		data          []byte
		expectError   bool
		errorContains string
		setup         func(*Encryptor)
	}{
		{
			name:        "valid encryption",
			key:         make([]byte, 24),
			data:        []byte("test message"),
			expectError: false,
			setup:       func(e *Encryptor) { rand.Read(make([]byte, 24)) },
		},
		{
			name:        "empty data",
			key:         make([]byte, 24),
			data:        []byte{},
			expectError: false,
			setup:       func(e *Encryptor) { rand.Read(make([]byte, 24)) },
		},
		{
			name:        "nil data",
			key:         make([]byte, 24),
			data:        nil,
			expectError: false,
			setup:       func(e *Encryptor) { rand.Read(make([]byte, 24)) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка ключа
			if tc.key != nil && len(tc.key) == 24 {
				rand.Read(tc.key)
			}
			encryptor.SetEncryptionKey(tc.key)

			// Выполняем дополнительную настройку если нужно
			tc.setup(encryptor)

			// Шифруем
			encryptedData, algorithm, iv, err := encryptor.Encrypt(tc.data)

			// Проверяем ошибки
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

			// Проверяем возвращаемые значения
			if algorithm == "" {
				t.Error("Algorithm should not be empty")
			}

			if len(iv) != 12 {
				t.Errorf("Expected IV length 12, got %d", len(iv))
			}

			if tc.data != nil && len(encryptedData) <= len(tc.data) {
				t.Errorf("Encrypted data should be longer than original")
			}

			// Для валидных случаев проверяем что можно дешифровать
			if !tc.expectError && tc.data != nil {
				decryptor := &Decryptor{}
				decryptor.SetDecryptionKey(tc.key)

				decrypted, err := decryptor.Decrypt(encryptedData, iv)
				if err != nil {
					t.Fatalf("Failed to decrypt: %v", err)
				}

				if string(decrypted) != string(tc.data) {
					t.Errorf("Decryption mismatch: expected %s, got %s", string(tc.data), string(decrypted))
				}
			}
		})
	}
}

func TestEncryptor_GetEncryptionKey(t *testing.T) {
	type fields struct {
		encryptionKey []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "test 1",
			fields: fields{
				encryptionKey: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			},
			want: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encryptor{
				encryptionKey: tt.fields.encryptionKey,
			}
			if got := e.GetEncryptionKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEncryptionKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptor_SetEncryptionKey(t *testing.T) {
	type fields struct {
		encryptionKey []byte
	}
	type args struct {
		encryptionKey []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test 1",
			fields: fields{
				encryptionKey: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			},
			args: args{
				encryptionKey: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encryptor{
				encryptionKey: tt.fields.encryptionKey,
			}
			e.SetEncryptionKey(tt.args.encryptionKey)
		})
	}
}
