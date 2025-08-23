package main

import (
	"mocks3/services/metadata/internal/config"
	"mocks3/services/metadata/internal/handler"
	"mocks3/services/metadata/internal/repository"
	"mocks3/services/metadata/internal/service"
	"mocks3/shared/bootstrap"
	"mocks3/shared/client"
	"time"
)

// MetadataServiceInitializer 实现服务初始化接口
type MetadataServiceInitializer struct {
	cfg *config.Config
}

// Initialize 初始化元数据服务的特定组件
func (m *MetadataServiceInitializer) Initialize(bootstrap *bootstrap.ServiceBootstrap) error {
	// 初始化数据库
	db, err := repository.NewDatabase(m.cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	// 初始化仓库
	metadataRepo := repository.NewMetadataRepository(db)

	// 初始化队列客户端
	_ = client.NewQueueClient("http://localhost:8083", 30*time.Second)

	// 初始化服务
	metadataService := service.NewMetadataService(metadataRepo, bootstrap.GetLogger())

	// 初始化处理器
	metadataHandler := handler.NewMetadataHandler(metadataService, bootstrap.GetLogger())

	// 注册路由
	metadataHandler.RegisterRoutes(bootstrap.GetRouter())

	return nil
}

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建服务配置
	serviceConfig := bootstrap.ServiceConfig{
		ServiceName: "metadata-service",
		ServicePort: 8081,
		Version:     cfg.Server.Version,
		Environment: cfg.Server.Environment,
		LogLevel:    cfg.LogLevel,
		ConsulAddr:  "consul:8500",
	}

	// 创建初始化器
	initializer := &MetadataServiceInitializer{cfg: cfg}

	// 运行服务
	bootstrap.RunService(serviceConfig, initializer)
}
