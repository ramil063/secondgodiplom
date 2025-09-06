package aes256gcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encryptor шифровальщик
type Encryptor struct {
	encryptionKey []byte
}

func (e *Encryptor) SetEncryptionKey(encryptionKey []byte) {
	e.encryptionKey = encryptionKey
}

func (e *Encryptor) GetEncryptionKey() []byte {
	return e.encryptionKey
}

// Encrypt функция шифрования
func (e *Encryptor) Encrypt(data []byte) ([]byte, string, []byte, error) {

	// Генерируем случайный IV (12 байт для GCM - рекомендуется)
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Создаем cipher block
	block, err := aes.NewCipher(e.GetEncryptionKey())
	if err != nil {
		return nil, "", nil, err
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", nil, err
	}

	// Шифруем данные
	encryptedData := gcm.Seal(nil, iv, data, nil)

	return encryptedData, algorithm, iv, nil
}
