package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/ramil063/secondgodiplom/internal/logger"
)

type envConfig struct {
	GRPCConfigPath string `env:"GRPC_CONFIG_PATH"`
}

// ServerConfig структура для парсинга файла конфигурации
type ServerConfig struct {
	Address          string `json:"address"`
	DatabaseURI      string `json:"database_uri"`
	HashKey          string `json:"hash_key"`
	CryptoKey        string `json:"crypto_key"`
	StoreInterval    string `json:"store_interval"`
	Secret           string `json:"secret"`
	WorkersCount     int    `json:"workers_count"`
	DbMaxConnections int32  `json:"db_max_connections"`
	DbMinConnections int32  `json:"db_min_connections"`
}

// loadConfig загружает конфигурацию из файла
func (cfg *ServerConfig) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read the config file %s: %w", path, err)
	}

	if err = json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmurshal the config file %s: %w", path, err)
	}

	return nil
}

// prepareConfig подготавливает параметры конфигурации для дальнейшей работы
func (cfg *ServerConfig) prepareConfig() error {
	storeInterval, err := time.ParseDuration(cfg.StoreInterval)
	if err != nil {
		return fmt.Errorf("failed to parse ReportInterval: %w", err)
	}
	cfg.StoreInterval = strconv.FormatFloat(storeInterval.Seconds(), 'f', 0, 64)

	return nil
}

// getConfigName получение названия файла конфигурации
func getConfigName() string {
	var configPath = ""
	var ev envConfig

	err := env.Parse(&ev)
	if err != nil {
		logger.WriteErrorLog("failed to parse config vars")
	}

	configPath = ev.GRPCConfigPath

	return configPath
}

// GetConfig установка значений конфигурации
func GetConfig() (*ServerConfig, error) {
	configName := getConfigName()

	var config ServerConfig
	var err error

	if configName == "" {
		return &config, nil
	}

	if err = config.loadConfig(configName); err != nil {
		return nil, err
	}

	if err = config.prepareConfig(); err != nil {
		return nil, err
	}

	return &config, err
}
