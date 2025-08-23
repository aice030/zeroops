package main

import (
	"context"
	"mocks3/services/third-party/internal/config"
	"mocks3/services/third-party/internal/handler"
	"mocks3/services/third-party/internal/repository"
	"mocks3/services/third-party/internal/service"
	"mocks3/shared/bootstrap"
)

// ThirdPartyServiceInitializer 实现服务初始化接口
type ThirdPartyServiceInitializer struct {
	cfg *config.Config
}

// Initialize 初始化第三方服务的特定组件
func (t *ThirdPartyServiceInitializer) Initialize(bootstrap *bootstrap.ServiceBootstrap) error {
	// 初始化仓库
	dataSourceRepo := repository.NewDataSourceRepository(t.cfg.DataSources)
	cacheRepo := repository.NewCacheRepository(&t.cfg.Cache)

	// 初始化服务
	thirdPartyService := service.NewThirdPartyService(dataSourceRepo, cacheRepo, bootstrap.GetLogger())

	// 初始化处理器
	thirdPartyHandler := handler.NewThirdPartyHandler(thirdPartyService, bootstrap.GetLogger())

	// 注册路由
	thirdPartyHandler.RegisterRoutes(bootstrap.GetRouter())

	// 打印数据源信息
	ctx := context.Background()
	dataSources, _ := dataSourceRepo.GetAll(ctx)
	for _, ds := range dataSources {
		bootstrap.GetLogger().Info("Configured data source",
			"name", ds.Name,
			"type", ds.Type,
			"enabled", ds.Enabled,
			"priority", ds.Priority)
	}

	return nil
}

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建服务配置
	serviceConfig := bootstrap.ServiceConfig{
		ServiceName: "third-party-service",
		ServicePort: 8084,
		Version:     cfg.Server.Version,
		Environment: cfg.Server.Environment,
		LogLevel:    cfg.LogLevel,
		ConsulAddr:  "consul:8500",
	}

	// 创建初始化器
	initializer := &ThirdPartyServiceInitializer{cfg: cfg}

	// 运行服务
	bootstrap.RunService(serviceConfig, initializer)
}
