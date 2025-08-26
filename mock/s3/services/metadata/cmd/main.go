package main

import (
	"context"
	"fmt"
	"mocks3/services/metadata/internal/handler"
	"mocks3/services/metadata/internal/repository"
	"mocks3/services/metadata/internal/service"
	"mocks3/shared/observability"
	"mocks3/shared/server"
	"os"
)

// 全局变量用于在初始化和清理时共享
var globalRepo *repository.PostgreSQLRepository

func main() {
	// 1. 加载配置
	config, err := service.LoadConfig("config/metadata-config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 创建服务启动器
	bootstrap := server.NewServiceBootstrap("Metadata Service", config)

	// 3. 设置自定义初始化逻辑
	bootstrap.WithCustomInit(func(ctx context.Context, logger *observability.Logger) error {
		// 初始化数据库仓库
		repo, err := repository.NewPostgreSQLRepository(config.GetDSN())
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
		globalRepo = repo // 保存供清理使用

		logger.Info(ctx, "Database connection established")

		// 初始化业务服务
		metadataService := service.NewMetadataService(repo, logger)

		// 初始化HTTP处理器
		metadataHandler := handler.NewMetadataHandler(metadataService, logger)

		// 设置处理器到启动器
		bootstrap.WithHandler(metadataHandler)

		return nil
	})

	// 4. 设置自定义清理逻辑
	bootstrap.WithCustomCleanup(func(ctx context.Context, logger *observability.Logger) error {
		// 关闭数据库连接
		if globalRepo != nil {
			globalRepo.Close()
			logger.Info(ctx, "Database connection closed")
		}
		return nil
	})

	// 5. 启动服务
	if err := bootstrap.Start(); err != nil {
		fmt.Printf("Failed to start service: %v\n", err)
		os.Exit(1)
	}
}
