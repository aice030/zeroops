package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 通用的YAML配置加载函数
func LoadConfig(configPath string, config any) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

// SaveConfigToYAML 将配置保存为YAML文件
func SaveConfigToYAML(configPath string, config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}
