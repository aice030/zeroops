package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 部署服务配置结构
type Config struct {
	Database   DatabaseConfig `yaml:"database"`   // 数据库配置
	PrivateKey string         `yaml:"privateKey"` // RSA私钥（用于Floyd认证）
}

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Host     string `yaml:"host"`     // 数据库主机地址
	Port     int    `yaml:"port"`     // 数据库端口
	User     string `yaml:"user"`     // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
	DBName   string `yaml:"dbname"`   // 数据库名称
	SSLMode  string `yaml:"sslmode"`  // SSL模式
}

// GetDSN 生成PostgreSQL数据库连接字符串（DSN - Data Source Name）
// 返回格式：host=xxx port=xxx user=xxx password=xxx dbname=xxx sslmode=xxx
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// LoadConfig 从指定路径加载配置文件
// 参数：
//   - configPath: 配置文件路径（相对路径或绝对路径）
//
// 返回：
//   - *Config: 配置对象指针
//   - error: 错误信息
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 解析YAML配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// 验证必要的配置项
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig 验证配置的有效性
func validateConfig(cfg *Config) error {
	// 验证数据库配置
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", cfg.Database.Port)
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user cannot be empty")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	// 验证私钥配置
	if cfg.PrivateKey == "" {
		return fmt.Errorf("private key cannot be empty")
	}

	return nil
}
