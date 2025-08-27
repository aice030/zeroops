package cmd

import (
	"path/filepath"
	"qiniu1024-mcp-server/pkg/common/config"
	"qiniu1024-mcp-server/prometheus/internal"
)

func Run() {
	// 启动服务前，先加载配置文件 configs/config.yaml
	// 获取项目根目录
	projectRoot := config.GetProjectRoot()
	// 构建配置文件的绝对路径
	configPath := filepath.Join(projectRoot, "configs", "config.yaml")

	if err := config.LoadConfig(configPath); err != nil {
		panic("配置文件加载失败: " + err.Error())
	}

	// 启动服务
	internal.StartPrometheusMcpServer()
}
