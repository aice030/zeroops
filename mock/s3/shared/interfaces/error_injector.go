package interfaces

import (
	"context"
	"mocks3/shared/models"
	"net/http"
)

// ErrorInjectorService 错误注入服务接口
type ErrorInjectorService interface {
	// 错误规则管理
	CreateRule(ctx context.Context, rule *models.ErrorRule) error
	DeleteRule(ctx context.Context, ruleID string) error
	GetRule(ctx context.Context, ruleID string) (*models.ErrorRule, error)
	ListRules(ctx context.Context) ([]*models.ErrorRule, error)

	// 错误注入核心功能
	ShouldInjectError(ctx context.Context, service, operation string) (map[string]any, bool)
}

// ErrorInjector HTTP错误注入器接口
type ErrorInjector interface {
	// HTTP 中间件 - 核心功能
	HTTPMiddleware() func(http.Handler) http.Handler

	// 错误注入 - 统一入口
	InjectError(w http.ResponseWriter, r *http.Request, action map[string]any) bool
}

// ErrorRuleEngine 错误规则引擎接口
type ErrorRuleEngine interface {
	// 规则评估 - 核心功能
	EvaluateRules(ctx context.Context, service, operation string) (map[string]any, bool)

	// 规则管理 - 基础CRUD
	AddRule(rule *models.ErrorRule) error
	RemoveRule(ruleID string) error
	ListRules() []*models.ErrorRule
}
