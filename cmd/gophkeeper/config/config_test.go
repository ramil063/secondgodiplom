package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentConfig_loadConfig(t *testing.T) {
	file, _ := os.CreateTemp("", "config_test.json")
	_, _ = file.Write([]byte(`{
  "address": "localhost:8080",
  "restore": true,
  "store_interval": "1s",
  "store_file": "/path/to/file.db",
  "database_dsn": "",
  "crypto_key": "/path/to/key.pem"
}`))
	defer os.Remove(file.Name())

	type conf struct {
		Address       string
		DatabaseURI   string
		HashKey       string
		CryptoKey     string
		StoreInterval string
	}
	tests := []struct {
		name string
		path string
		conf conf
	}{
		{
			name: "Load Config",
			path: file.Name(),
			conf: conf{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ServerConfig{
				Address:       tt.conf.Address,
				DatabaseURI:   tt.conf.DatabaseURI,
				HashKey:       tt.conf.HashKey,
				CryptoKey:     tt.conf.CryptoKey,
				StoreInterval: tt.conf.StoreInterval,
			}
			err := cfg.loadConfig(tt.path)
			assert.NoError(t, err)
		})
	}
	_ = os.Remove("config_test.json")
}

func TestAgentConfig_prepareConfig(t *testing.T) {
	type conf struct {
		Address       string
		DatabaseURI   string
		HashKey       string
		CryptoKey     string
		StoreInterval string
	}
	tests := []struct {
		conf conf
		name string
	}{
		{
			name: "Prepare Config",
			conf: conf{
				Address:       "testhost:8080",
				DatabaseURI:   "database",
				HashKey:       "test",
				CryptoKey:     "/test/test/test.pem",
				StoreInterval: "1s",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ServerConfig{
				Address:       tt.conf.Address,
				DatabaseURI:   tt.conf.DatabaseURI,
				HashKey:       tt.conf.HashKey,
				CryptoKey:     tt.conf.CryptoKey,
				StoreInterval: tt.conf.StoreInterval,
			}
			err := cfg.prepareConfig()
			assert.NoError(t, err)
			assert.Equal(t, tt.conf.Address, cfg.Address)
			assert.Equal(t, tt.conf.DatabaseURI, cfg.DatabaseURI)
			assert.Equal(t, tt.conf.HashKey, cfg.HashKey)
			assert.Equal(t, tt.conf.CryptoKey, cfg.CryptoKey)
			assert.Equal(t, "1", cfg.StoreInterval)
		})
	}
}

func TestGetConfig(t *testing.T) {
	file, _ := os.CreateTemp("", "config_test.json")
	_, _ = file.Write([]byte(`{
"address": "localhost:8080",
"restore": true,
"store_interval": "1s",
"store_file": "/path/to/file.db",
"database_dsn": "database",
"crypto_key": "/path/to/key.pem",
"hash_key": "test"
}`))
	defer os.Remove(file.Name())

	_ = os.Setenv("CONFIG", file.Name())
	defer os.Unsetenv("CONFIG")

	tests := []struct {
		want *ServerConfig
		name string
	}{
		{
			name: "Get Config",
			want: &ServerConfig{
				Address:       "localhost:8080",
				DatabaseURI:   "database",
				HashKey:       "test",
				CryptoKey:     "/path/to/key.pem",
				StoreInterval: "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig()
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
