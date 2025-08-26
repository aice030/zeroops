package main

import (
	"context"
	"fmt"
	"mocks3/services/storage/internal/handler"
	"mocks3/services/storage/internal/repository"
	"mocks3/services/storage/internal/service"
	"mocks3/shared/client"
	"mocks3/shared/observability"
	"mocks3/shared/server"
	"os"
)

func main() {
	// 1. 加载配置
	config, err := service.LoadConfig("config/storage-config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 创建服务启动器
	bootstrap := server.NewServiceBootstrap("Storage Service", config)

	// 3. 设置自定义初始化逻辑
	bootstrap.WithCustomInit(func(ctx context.Context, logger *observability.Logger) error {
		// 初始化文件存储仓库
		nodes := make([]repository.NodeInfo, 0, len(config.Storage.Nodes))
		for _, nodeConfig := range config.Storage.Nodes {
			nodes = append(nodes, repository.NodeInfo{
				ID:   nodeConfig.ID,
				Path: nodeConfig.Path,
			})
		}

		repo, err := repository.NewFileStorageRepository(nodes, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize storage repository: %w", err)
		}

		logger.Info(ctx, "Storage repository initialized",
			observability.Int("nodes", len(nodes)))

		// 初始化外部服务客户端
		metadataClient := client.NewMetadataClient(
			config.Services.Metadata.URL,
			config.GetMetadataTimeout(),
			logger,
		)

		queueClient := client.NewQueueClient(
			config.Services.Queue.URL,
			config.GetQueueTimeout(),
			logger,
		)

		thirdPartyClient := client.NewThirdPartyClient(
			config.Services.ThirdParty.URL,
			config.GetThirdPartyTimeout(),
			logger,
		)

		logger.Info(ctx, "External service clients initialized")

		// 初始化业务服务
		storageService := service.NewStorageService(repo, metadataClient, queueClient, thirdPartyClient, logger)

		// 初始化HTTP处理器
		storageHandler := handler.NewStorageHandler(storageService, logger)

		// 设置处理器到启动器
		bootstrap.WithHandler(storageHandler)

		return nil
	})

	// 4. 启动服务
	if err := bootstrap.Start(); err != nil {
		fmt.Printf("Failed to start service: %v\n", err)
		os.Exit(1)
	}
}
