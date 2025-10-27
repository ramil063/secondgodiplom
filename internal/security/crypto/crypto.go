package crypto

import (
	"github.com/ramil063/secondgodiplom/internal/security/crypto/aes256gcm"
)

// Encryptor общий интерфейс для шифрования
type Encryptor interface {
	Encrypt(data []byte) ([]byte, string, []byte, error)
}

// Decryptor общий интерфейс для шифрования
type Decryptor interface {
	Decrypt(encryptedData []byte, iv []byte) ([]byte, error)
}

// Manager содержит все шифровальщики и дешифровщики
type Manager struct {
	grpcEncryptor Encryptor
	grpcDecryptor Decryptor
}

func NewCryptoManager() *Manager {
	return &Manager{}
}

func (cm *Manager) SetGRPCEncryptor(enc Encryptor) {
	cm.grpcEncryptor = enc
}

func (cm *Manager) SetGRPCDecryptor(decr Decryptor) {
	cm.grpcDecryptor = decr
}

func (cm *Manager) GetGRPCEncryptor() Encryptor {
	return cm.grpcEncryptor
}

func (cm *Manager) GetGRPCDecryptor() Decryptor {
	return cm.grpcDecryptor
}

// NewAes256gcmEncryptor фабрика для алгоритма Aes256gcm
func NewAes256gcmEncryptor(encryptionKey []byte) (Encryptor, error) {
	encryptor := &aes256gcm.Encryptor{}
	var err error

	encryptor.SetEncryptionKey(encryptionKey)

	return encryptor, err
}

// NewAes256gcmDecryptor фабрика для алгоритма Aes256gcm
func NewAes256gcmDecryptor(encryptionKey []byte) (Decryptor, error) {
	decryptor := &aes256gcm.Decryptor{}
	var err error

	decryptor.SetDecryptionKey(encryptionKey)

	return decryptor, err
}
