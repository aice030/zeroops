package main

import (
	"mocks3/services/storage/internal/config"
	"mocks3/services/storage/internal/handler"
	"mocks3/services/storage/internal/service"
	"mocks3/shared/bootstrap"
)

// StorageServiceInitializer 实现服务初始化接口
type StorageServiceInitializer struct {
	cfg *config.Config
}

// Initialize 初始化存储服务的特定组件
func (s *StorageServiceInitializer) Initialize(bootstrap *bootstrap.ServiceBootstrap) error {
	// 初始化存储服务
	storageService, err := service.NewStorageService(s.cfg, bootstrap.GetLogger())
	if err != nil {
		return err
	}

	// 初始化处理器
	storageHandler := handler.NewStorageHandler(storageService, bootstrap.GetLogger())

	// 注册路由
	storageHandler.RegisterRoutes(bootstrap.GetRouter())

	return nil
}

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建服务配置
	serviceConfig := bootstrap.ServiceConfig{
		ServiceName: "storage-service",
		ServicePort: 8082,
		Version:     cfg.Server.Version,
		Environment: cfg.Server.Environment,
		LogLevel:    cfg.LogLevel,
		ConsulAddr:  "consul:8500",
	}

	// 创建初始化器
	initializer := &StorageServiceInitializer{cfg: cfg}

	// 运行服务
	bootstrap.RunService(serviceConfig, initializer)
}
