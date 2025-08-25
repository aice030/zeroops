package middleware

import (
	"context"
	"fmt"
	"mocks3/shared/client"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"mocks3/shared/utils"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// MetricInjectorConfig 指标异常注入器配置
type MetricInjectorConfig struct {
	MockErrorService MockErrorServiceConfig `yaml:"mock_error_service"`
	Cache            CacheConfig            `yaml:"cache"`
}

// MockErrorServiceConfig Mock Error Service配置
type MockErrorServiceConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL time.Duration `yaml:"ttl"`
}

// MetricInjector 指标异常注入器
type MetricInjector struct {
	mockErrorClient *client.BaseHTTPClient
	serviceName     string
	logger          *observability.Logger

	// 缓存
	cache    map[string]*CachedAnomaly
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
}

// CachedAnomaly 缓存的异常配置
type CachedAnomaly struct {
	Anomaly   map[string]any
	ExpiresAt time.Time
}

// NewMetricInjector 从YAML配置创建指标异常注入器
func NewMetricInjector(configPath string, serviceName string, logger *observability.Logger) (*MetricInjector, error) {
	// 加载配置文件
	var config MetricInjectorConfig
	if err := utils.LoadConfig(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to load metric injector config: %w", err)
	}

	// 设置默认值
	if config.MockErrorService.URL == "" {
		config.MockErrorService.URL = "http://localhost:8085" // 默认Mock Error Service地址
	}
	if config.MockErrorService.Timeout == 0 {
		config.MockErrorService.Timeout = 5 * time.Second // 默认5秒超时
	}
	if config.Cache.TTL == 0 {
		config.Cache.TTL = 30 * time.Second // 默认30秒缓存TTL
	}

	client := client.NewBaseHTTPClient(
		config.MockErrorService.URL,
		config.MockErrorService.Timeout,
		"metric-injector",
		logger,
	)

	injector := &MetricInjector{
		mockErrorClient: client,
		serviceName:     serviceName,
		logger:          logger,
		cache:           make(map[string]*CachedAnomaly),
		cacheTTL:        config.Cache.TTL,
	}

	// 启动缓存清理协程
	injector.StartCacheCleanup()

	return injector, nil
}

// NewMetricInjectorWithDefaults 使用默认配置创建指标异常注入器
func NewMetricInjectorWithDefaults(mockErrorServiceURL string, serviceName string, logger *observability.Logger) *MetricInjector {
	client := client.NewBaseHTTPClient(mockErrorServiceURL, 5*time.Second, "metric-injector", logger)

	injector := &MetricInjector{
		mockErrorClient: client,
		serviceName:     serviceName,
		logger:          logger,
		cache:           make(map[string]*CachedAnomaly),
		cacheTTL:        30 * time.Second,
	}

	injector.StartCacheCleanup()
	return injector
}

// InjectMetricAnomaly 检查并注入指标异常
func (mi *MetricInjector) InjectMetricAnomaly(ctx context.Context, metricName string, originalValue float64) float64 {
	// 检查缓存
	cacheKey := mi.serviceName + ":" + metricName
	mi.cacheMu.RLock()
	if cached, exists := mi.cache[cacheKey]; exists && time.Now().Before(cached.ExpiresAt) {
		mi.cacheMu.RUnlock()
		if cached.Anomaly != nil {
			return mi.applyAnomaly(ctx, cached.Anomaly, originalValue, metricName)
		}
		return originalValue
	}
	mi.cacheMu.RUnlock()

	// 查询Mock Error Service
	request := map[string]string{
		"service":     mi.serviceName,
		"metric_name": metricName,
	}

	var response struct {
		ShouldInject bool           `json:"should_inject"`
		Service      string         `json:"service"`
		MetricName   string         `json:"metric_name"`
		Anomaly      map[string]any `json:"anomaly,omitempty"`
	}

	// 使用较短的超时时间避免影响正常指标收集
	opts := client.RequestOptions{
		Method: "POST",
		Path:   "/api/v1/metric-inject/check",
		Body:   request,
	}

	err := mi.mockErrorClient.DoRequestWithJSON(ctx, opts, &response)
	if err != nil {
		mi.logger.Debug(ctx, "Failed to check metric injection",
			observability.Error(err),
			observability.String("metric_name", metricName))
		// 失败时缓存空结果，避免频繁请求
		mi.updateCache(cacheKey, nil)
		return originalValue
	}

	// 更新缓存
	var anomaly map[string]any
	if response.ShouldInject {
		anomaly = response.Anomaly
	}
	mi.updateCache(cacheKey, anomaly)

	// 应用异常
	if response.ShouldInject && response.Anomaly != nil {
		return mi.applyAnomaly(ctx, response.Anomaly, originalValue, metricName)
	}

	return originalValue
}

// updateCache 更新缓存
func (mi *MetricInjector) updateCache(key string, anomaly map[string]any) {
	mi.cacheMu.Lock()
	defer mi.cacheMu.Unlock()

	mi.cache[key] = &CachedAnomaly{
		Anomaly:   anomaly,
		ExpiresAt: time.Now().Add(mi.cacheTTL),
	}
}

// applyAnomaly 应用指标异常
func (mi *MetricInjector) applyAnomaly(ctx context.Context, anomaly map[string]any, originalValue float64, metricName string) float64 {
	anomalyType, ok := anomaly["anomaly_type"].(string)
	if !ok {
		return originalValue
	}

	targetValueRaw, ok := anomaly["target_value"]
	if !ok {
		return originalValue
	}

	var targetValue float64
	switch v := targetValueRaw.(type) {
	case float64:
		targetValue = v
	case float32:
		targetValue = float64(v)
	case int:
		targetValue = float64(v)
	case int64:
		targetValue = float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			targetValue = parsed
		} else {
			return originalValue
		}
	default:
		return originalValue
	}

	ruleID, _ := anomaly["rule_id"].(string)

	mi.logger.Info(ctx, "Applying metric anomaly",
		observability.String("metric_name", metricName),
		observability.String("anomaly_type", anomalyType),
		observability.Float64("original_value", originalValue),
		observability.Float64("target_value", targetValue),
		observability.String("rule_id", ruleID))

	// 根据异常类型应用不同的策略
	switch anomalyType {
	case models.AnomalyCPUSpike:
		// CPU峰值：直接设置为目标值
		return targetValue

	case models.AnomalyMemoryLeak:
		// 内存泄露：逐渐增长到目标值
		if originalValue < targetValue {
			increment := (targetValue - originalValue) * 0.1 // 每次增长10%
			return originalValue + increment
		}
		return targetValue

	case models.AnomalyDiskFull:
		// 磁盘满载：设置为目标值
		return targetValue

	case models.AnomalyNetworkFlood:
		// 网络风暴：设置为目标值的随机波动
		variation := targetValue * 0.1 // 10%的波动
		randomFactor := float64(time.Now().UnixNano()%100) / 100.0
		return targetValue + (variation * (2*randomFactor - 1))

	case models.AnomalyMachineDown:
		// 机器宕机：设置为0
		return 0

	default:
		return targetValue
	}
}

// HTTPMiddleware HTTP中间件（占位实现）
func (mi *MetricInjector) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 这里可以添加HTTP级别的错误注入逻辑
			// 目前主要通过InjectMetricAnomaly方法在指标收集时注入异常
			next.ServeHTTP(w, r)
		})
	}
}

// CleanupCache 清理过期缓存
func (mi *MetricInjector) CleanupCache() {
	mi.cacheMu.Lock()
	defer mi.cacheMu.Unlock()

	now := time.Now()
	for key, cached := range mi.cache {
		if now.After(cached.ExpiresAt) {
			delete(mi.cache, key)
		}
	}
}

// StartCacheCleanup 启动缓存清理协程
func (mi *MetricInjector) StartCacheCleanup() {
	go func() {
		ticker := time.NewTicker(mi.cacheTTL)
		defer ticker.Stop()

		for range ticker.C {
			mi.CleanupCache()
		}
	}()
}
