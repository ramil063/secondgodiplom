package config

import "strconv"

// GetAddress получение параметра Address
func (cfg *ServerConfig) GetAddress(defaultValue string) string {
	if cfg.Address != "" {
		return cfg.Address
	}
	return defaultValue
}

// GetDatabaseDSN получение параметра DatabaseURI
func (cfg *ServerConfig) GetDatabaseDSN(defaultValue string) string {
	if cfg.DatabaseURI != "" {
		return cfg.DatabaseURI
	}
	return defaultValue
}

// GetHashKey получение параметра HashKey
func (cfg *ServerConfig) GetHashKey(defaultValue string) string {
	if cfg.HashKey != "" {
		return cfg.HashKey
	}
	return defaultValue
}

// GetCryptoKey получение параметра CryptoKey
func (cfg *ServerConfig) GetCryptoKey(defaultValue string) string {
	if cfg.CryptoKey != "" {
		return cfg.CryptoKey
	}
	return defaultValue
}

// GetStoreInterval получение параметра StoreInterval
func (cfg *ServerConfig) GetStoreInterval(defaultValue int) int {
	if cfg.StoreInterval != "0" {
		if val, err := strconv.Atoi(cfg.StoreInterval); err == nil {
			return val
		}
	}
	return defaultValue
}
