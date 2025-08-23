package main

import (
	"context"
	"mocks3/services/mock-error/internal/config"
	"mocks3/services/mock-error/internal/handler"
	"mocks3/services/mock-error/internal/repository"
	"mocks3/services/mock-error/internal/service"
	"mocks3/shared/bootstrap"
	"mocks3/shared/models"
	logger "mocks3/shared/observability/log"
	"time"
)

// MockErrorServiceInitializer 实现服务初始化接口
type MockErrorServiceInitializer struct {
	cfg          *config.Config
	errorService *service.ErrorInjectorService
}

// Initialize 初始化错误注入服务的特定组件
func (m *MockErrorServiceInitializer) Initialize(bootstrap *bootstrap.ServiceBootstrap) error {
	// 验证配置
	if err := m.cfg.Validate(); err != nil {
		return err
	}

	// 初始化仓库
	ruleRepo := repository.NewRuleRepository()
	statsRepo := repository.NewStatsRepository(10000, m.cfg.ErrorEngine.StatRetentionHours)

	// 初始化规则引擎
	ruleEngine := service.NewRuleEngine(bootstrap.GetLogger())

	// 初始化错误注入服务
	m.errorService = service.NewErrorInjectorService(m.cfg, ruleRepo, statsRepo, ruleEngine, bootstrap.GetLogger())

	// 初始化处理器
	errorHandler := handler.NewErrorHandler(m.errorService, bootstrap.GetLogger())

	// 注册路由
	errorHandler.RegisterRoutes(bootstrap.GetRouter())

	// 显示启动信息
	bootstrap.GetLogger().Info("Service configuration",
		"max_rules", m.cfg.ErrorEngine.MaxRules,
		"default_probability", m.cfg.ErrorEngine.DefaultProbability,
		"enable_statistics", m.cfg.ErrorEngine.EnableStatistics,
		"global_probability", m.cfg.Injection.GlobalProbability)

	// 添加一些示例规则（仅在开发环境）
	if m.cfg.Server.Environment == "development" {
		addSampleRules(context.Background(), m.errorService, bootstrap.GetLogger())
	}

	return nil
}

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建服务配置
	serviceConfig := bootstrap.ServiceConfig{
		ServiceName: "mock-error-service",
		ServicePort: 8085,
		Version:     cfg.Server.Version,
		Environment: cfg.Server.Environment,
		LogLevel:    cfg.LogLevel,
		ConsulAddr:  "consul:8500",
	}

	// 创建初始化器
	initializer := &MockErrorServiceInitializer{cfg: cfg}

	// 运行服务
	bootstrap.RunService(serviceConfig, initializer)
}

// addSampleRules 添加示例规则
func addSampleRules(ctx context.Context, service *service.ErrorInjectorService, logger *logger.Logger) {
	logger.Info("Adding sample error injection rules for development")

	// 示例规则1: 存储服务随机错误
	delay1 := 500 * time.Millisecond
	rule1 := &models.ErrorRule{
		Name:        "Storage Service Random Error",
		Description: "Randomly inject 500 errors into storage service operations",
		Service:     "storage-service",
		Enabled:     true,
		Priority:    1,
		Conditions: []models.ErrorCondition{
			{
				Type:     models.ErrorConditionTypeProbability,
				Operator: "eq",
				Value:    0.1, // 10% 概率
			},
		},
		Action: models.ErrorAction{
			Type:     models.ErrorActionTypeHTTPError,
			HTTPCode: 500,
			Message:  "Internal server error injected for testing",
			Delay:    &delay1,
		},
	}

	// 示例规则2: 元数据服务延迟
	delay2 := 2 * time.Second
	rule2 := &models.ErrorRule{
		Name:        "Metadata Service Delay",
		Description: "Add delay to metadata service operations",
		Service:     "metadata-service",
		Operation:   "GetMetadata",
		Enabled:     true,
		Priority:    2,
		Conditions: []models.ErrorCondition{
			{
				Type:     models.ErrorConditionTypeProbability,
				Operator: "eq",
				Value:    0.2, // 20% 概率
			},
		},
		Action: models.ErrorAction{
			Type:  models.ErrorActionTypeDelay,
			Delay: &delay2,
		},
	}

	// 示例规则3: 队列服务网络错误
	rule3 := &models.ErrorRule{
		Name:        "Queue Service Network Error",
		Description: "Inject network errors into queue service",
		Service:     "queue-service",
		Enabled:     false, // 默认禁用
		Priority:    3,
		MaxTriggers: 10, // 最多触发10次
		Conditions: []models.ErrorCondition{
			{
				Type:     models.ErrorConditionTypeProbability,
				Operator: "eq",
				Value:    0.05, // 5% 概率
			},
		},
		Action: models.ErrorAction{
			Type:    models.ErrorActionTypeNetworkError,
			Message: "Network timeout injected",
		},
	}

	// 添加规则
	rules := []*models.ErrorRule{rule1, rule2, rule3}
	for _, rule := range rules {
		if err := service.AddErrorRule(ctx, rule); err != nil {
			logger.Warn("Failed to add sample rule", "rule_name", rule.Name, "error", err)
		} else {
			logger.Info("Added sample rule", "rule_name", rule.Name, "enabled", rule.Enabled)
		}
	}
}
