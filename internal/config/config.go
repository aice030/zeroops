package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
}

type ServerConfig struct {
	BindAddr string `json:"bindAddr"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

func Load() (*Config, error) {
	configFile := flag.String("f", "", "Path to configuration file")
	flag.Parse()

	cfg := &Config{
		Server: ServerConfig{
			BindAddr: getEnv("SERVER_BIND_ADDR", "0.0.0.0:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "zeroops"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if *configFile != "" {
		if err := loadFromFile(cfg, *configFile); err != nil {
			log.Err(err)
			return nil, err
		}
	}

	return cfg, nil
}

func loadFromFile(cfg *Config, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
