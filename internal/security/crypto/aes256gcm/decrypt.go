package aes256gcm

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Decryptor шифровальщик
type Decryptor struct {
	decryptionKey []byte
}

func (d *Decryptor) SetDecryptionKey(key []byte) {
	d.decryptionKey = key
}

func (d *Decryptor) GetDecryptionKey() []byte {
	return d.decryptionKey
}

// Decrypt функция дешифровки
func (d *Decryptor) Decrypt(encryptedData []byte, iv []byte) ([]byte, error) {
	// Проверяем длину ключа
	if len(d.GetDecryptionKey()) != 24 {
		return nil, fmt.Errorf("encryption key must be 24 bytes")
	}

	// Создаем cipher block
	block, err := aes.NewCipher(d.GetDecryptionKey())
	if err != nil {
		return nil, err
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Дешифруем данные
	decryptedData, err := gcm.Open(nil, iv, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return decryptedData, nil
}
