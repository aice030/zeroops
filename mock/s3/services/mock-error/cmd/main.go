package main

import (
	"context"
	"fmt"
	"mocks3/services/mock-error/internal/handler"
	"mocks3/services/mock-error/internal/service"
	"mocks3/shared/observability"
	"mocks3/shared/server"
	"os"
)

func main() {
	// 1. 加载配置
	config, err := service.LoadConfig("config/mock-error-config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 创建服务启动器
	bootstrap := server.NewServiceBootstrap("Mock Error Service", config)

	// 3. 设置自定义初始化逻辑
	bootstrap.WithCustomInit(func(ctx context.Context, logger *observability.Logger) error {
		// 初始化错误注入服务
		errorService := service.NewMockErrorService(config, logger)

		// 初始化HTTP处理器
		errorHandler := handler.NewMockErrorHandler(errorService, logger)

		// 设置处理器到启动器
		bootstrap.WithHandler(errorHandler)

		// 记录配置信息
		logger.Info(ctx, "Mock Error Service initialized",
			observability.String("data_dir", config.Storage.DataDir),
			observability.String("consul_addr", config.Consul.Address))

		return nil
	})

	// 4. 启动服务
	if err := bootstrap.Start(); err != nil {
		fmt.Printf("Failed to start service: %v\n", err)
		os.Exit(1)
	}
}
