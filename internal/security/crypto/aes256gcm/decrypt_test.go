package aes256gcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"reflect"
	"testing"
)

func TestDecryptor_GetDecryptionKey(t *testing.T) {
	type fields struct {
		decryptionKey []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "test 1",
			fields: fields{
				decryptionKey: []byte("test"),
			},
			want: []byte("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decryptor{
				decryptionKey: tt.fields.decryptionKey,
			}
			if got := d.GetDecryptionKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDecryptionKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecryptor_SetDecryptionKey(t *testing.T) {
	type fields struct {
		decryptionKey []byte
	}
	type args struct {
		key []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test 1",
			fields: fields{
				decryptionKey: []byte("test"),
			},
			args: args{
				key: []byte("test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decryptor{
				decryptionKey: tt.fields.decryptionKey,
			}
			d.SetDecryptionKey(tt.args.key)
		})
	}
}

func TestDecryptor_Decrypt_VariousData(t *testing.T) {
	decryptor := &Decryptor{}
	key := make([]byte, 24)
	rand.Read(key)
	decryptor.SetDecryptionKey(key)

	testCases := []struct {
		name string
		data []byte
	}{
		{"short text", []byte("hello")},
		{"long text", []byte("this is a very long secret message that needs to be encrypted and decrypted properly")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}},
		{"empty data", []byte{}},
		{"special chars", []byte("secret with !@#$%^&*()_+")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			iv := make([]byte, 12)
			rand.Read(iv)

			// Шифруем
			block, _ := aes.NewCipher(key)
			gcm, _ := cipher.NewGCM(block)
			encrypted := gcm.Seal(nil, iv, tc.data, nil)

			// Дешифруем
			decrypted, err := decryptor.Decrypt(encrypted, iv)
			if err != nil {
				t.Fatalf("Decryption failed for %s: %v", tc.name, err)
			}

			// Проверяем
			if len(decrypted) != len(tc.data) {
				t.Errorf("Length mismatch for %s: expected %d, got %d",
					tc.name, len(tc.data), len(decrypted))
			}
			if string(decrypted) != string(tc.data) {
				t.Errorf("Data mismatch for %s", tc.name)
			}
		})
	}
}

func TestDecryptor_Integration(t *testing.T) {
	// Создаем encryptor для теста (простая реализация)
	encryptor := struct {
		encrypt func(key, iv, plaintext []byte) ([]byte, error)
	}{
		encrypt: func(key, iv, plaintext []byte) ([]byte, error) {
			if len(key) != 24 {
				return nil, fmt.Errorf("key must be 24 bytes")
			}
			block, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}
			gcm, err := cipher.NewGCM(block)
			if err != nil {
				return nil, err
			}
			return gcm.Seal(nil, iv, plaintext, nil), nil
		},
	}

	decryptor := &Decryptor{}

	// Тестовые данные
	testMessages := [][]byte{
		[]byte("test message"),
		[]byte(""),
		[]byte("very long message with special chars !@#$%^&*()"),
	}

	for i, message := range testMessages {
		t.Run(fmt.Sprintf("Message_%d", i), func(t *testing.T) {
			// Генерируем ключ и IV
			key := make([]byte, 24)
			rand.Read(key)
			iv := make([]byte, 12)
			rand.Read(iv)

			// Шифруем
			encrypted, err := encryptor.encrypt(key, iv, message)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Настраиваем дешифратор
			decryptor.SetDecryptionKey(key)

			// Дешифруем
			decrypted, err := decryptor.Decrypt(encrypted, iv)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Проверяем
			if string(decrypted) != string(message) {
				t.Errorf("Expected %s, got %s", string(message), string(decrypted))
			}
		})
	}
}
