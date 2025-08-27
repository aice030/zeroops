package cmd

import (
	"qiniu1024-mcp-server/elasticsearch/internal"
	"qiniu1024-mcp-server/pkg/common/config"
)

func Run() {
	// 启动服务前，先加载配置文件 configs/config.yaml
	if err := config.LoadConfig("configs/config.yaml"); err != nil {
		panic("配置文件加载失败: " + err.Error())
	}

	internal.StartElasticsearchMcpServer()
}
