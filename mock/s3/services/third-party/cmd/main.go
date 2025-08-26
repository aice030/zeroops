package main

import (
	"context"
	"fmt"
	"mocks3/services/third-party/internal/handler"
	"mocks3/services/third-party/internal/service"
	"mocks3/shared/observability"
	"mocks3/shared/server"
	"os"
)

func main() {
	// 1. 加载配置
	config, err := service.LoadConfig("config/third-party-config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 创建服务启动器
	bootstrap := server.NewServiceBootstrap("Third-Party Service", config)

	// 3. 设置自定义初始化逻辑
	bootstrap.WithCustomInit(func(ctx context.Context, logger *observability.Logger) error {
		// 初始化第三方服务
		thirdPartyService := service.NewThirdPartyService(config, logger)

		// 初始化HTTP处理器
		thirdPartyHandler := handler.NewThirdPartyHandler(thirdPartyService, logger)

		// 设置处理器到启动器
		bootstrap.WithHandler(thirdPartyHandler)

		// 记录配置信息
		logger.Info(ctx, "Third-Party Service initialized",
			observability.String("mock_enabled", fmt.Sprintf("%v", config.Mock.Enabled)),
			observability.Int("data_sources", len(config.DataSources)))

		return nil
	})

	// 4. 启动服务
	if err := bootstrap.Start(); err != nil {
		fmt.Printf("Failed to start service: %v\n", err)
		os.Exit(1)
	}
}
