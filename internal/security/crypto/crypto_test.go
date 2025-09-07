package crypto

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetGRPCDecryptor(t *testing.T) {
	var decryptor Decryptor
	t.Run("GetGRPCDecryptor", func(t *testing.T) {

		cm := &Manager{
			grpcDecryptor: decryptor,
		}
		assert.Equalf(t, decryptor, cm.GetGRPCDecryptor(), "GetGRPCDecryptor()")
	})
}

func TestManager_GetGRPCEncryptor(t *testing.T) {
	var encryptor Encryptor
	t.Run("GetGRPCEncryptor", func(t *testing.T) {
		cm := &Manager{
			grpcEncryptor: encryptor,
		}
		assert.Equalf(t, encryptor, cm.GetGRPCEncryptor(), "GetGRPCEncryptor()")
	})
}

func TestManager_SetGRPCDecryptor(t *testing.T) {
	var decryptor Decryptor
	t.Run("SetGRPCDecryptor", func(t *testing.T) {
		cm := &Manager{}
		cm.SetGRPCDecryptor(decryptor)
		assert.Equal(t, decryptor, cm.GetGRPCDecryptor(), "GetGRPCDecryptor()")
	})
}

func TestManager_SetGRPCEncryptor(t *testing.T) {
	var encryptor Encryptor

	t.Run("SetGRPCEncryptor", func(t *testing.T) {
		cm := &Manager{}
		cm.SetGRPCEncryptor(encryptor)
		assert.Equalf(t, encryptor, cm.GetGRPCEncryptor(), "SetGRPCEncryptor()")
	})
}

func TestNewCryptoManager(t *testing.T) {
	tests := []struct {
		name string
		want *Manager
	}{
		{
			name: "NewCryptoManager",
			want: &Manager{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewCryptoManager(), "NewCryptoManager()")
		})
	}
}

func TestNewAes256gcmEncryptor(t *testing.T) {
	type args struct {
		encryptionKey []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "NewAes256gcmEncryptor",
			args: args{
				encryptionKey: []byte("test"),
			},
			want: "*aes256gcm.Encryptor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAes256gcmEncryptor(tt.args.encryptionKey)
			assert.NoError(t, err, "NewAes256gcmEncryptor")
			assert.Equalf(t, tt.want, reflect.ValueOf(got).Type().String(), "NewAes256gcmEncryptor(%v)", tt.args.encryptionKey)
		})
	}
}

func TestNewAes256gcmDecryptor(t *testing.T) {
	type args struct {
		encryptionKey []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "NewAes256gcmDecryptor",
			args: args{
				encryptionKey: []byte("test"),
			},
			want: "*aes256gcm.Decryptor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAes256gcmDecryptor(tt.args.encryptionKey)
			assert.NoError(t, err, "NewAes256gcmDecryptor")
			assert.Equalf(t, tt.want, reflect.ValueOf(got).Type().String(), "NewAes256gcmDecryptor(%v)", tt.args.encryptionKey)
		})
	}
}
