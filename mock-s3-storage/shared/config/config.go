package config

import "context"

// Loader 配置加载器 - 唯一核心接口
type Loader interface {
	// Load 加载配置到目标结构体
	Load(ctx context.Context, target any) error

	// Get 获取单个配置值
	Get(key string) (any, bool)

	// Close 关闭加载器
	Close() error
}

// Source 配置源类型
type Source string

const (
	SourceEnv    Source = "env"
	SourceFile   Source = "file"
	SourceConsul Source = "consul"
)

// SourceConfig 配置源配置
type SourceConfig struct {
	Type   Source         `json:"type" yaml:"type"`
	Config map[string]any `json:"config" yaml:"config"`
}

// NewLoader 创建配置加载器的工厂函数
func NewLoader(sources ...SourceConfig) Loader {
	// TODO: 具体实现
	return nil
}

// NewEnvLoader 创建环境变量配置加载器的便利函数
func NewEnvLoader(prefix string) Loader {
	return NewLoader(SourceConfig{
		Type: SourceEnv,
		Config: map[string]any{
			"prefix": prefix,
		},
	})
}

// NewFileLoader 创建文件配置加载器的便利函数
func NewFileLoader(filePath string) Loader {
	return NewLoader(SourceConfig{
		Type: SourceFile,
		Config: map[string]any{
			"path": filePath,
		},
	})
}

// NewConsulLoader 创建Consul配置加载器的便利函数
func NewConsulLoader(address, datacenter, token string) Loader {
	return NewLoader(SourceConfig{
		Type: SourceConsul,
		Config: map[string]any{
			"address":    address,
			"datacenter": datacenter,
			"token":      token,
		},
	})
}
