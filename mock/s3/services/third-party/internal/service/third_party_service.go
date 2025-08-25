package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"

	"github.com/google/uuid"
)

// ThirdPartyService 第三方服务实现
type ThirdPartyService struct {
	config *Config
	logger *observability.Logger
	rand   *rand.Rand
}

// NewThirdPartyService 创建第三方服务
func NewThirdPartyService(config *Config, logger *observability.Logger) *ThirdPartyService {
	return &ThirdPartyService{
		config: config,
		logger: logger,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetObject 从第三方获取对象
func (s *ThirdPartyService) GetObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	s.logger.Info(ctx, "Getting object from third-party service",
		observability.String("bucket", bucket),
		observability.String("key", key))

	// 模拟延迟
	if s.config.Mock.LatencyMs > 0 {
		time.Sleep(time.Duration(s.config.Mock.LatencyMs) * time.Millisecond)
	}

	// 模拟失败率
	if s.rand.Float64() > s.config.Mock.SuccessRate {
		err := fmt.Errorf("third-party service temporarily unavailable")
		s.logger.Warn(ctx, "Simulated third-party service failure",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		return nil, err
	}

	// 如果启用Mock模式，返回Mock数据
	if s.config.Mock.Enabled {
		return s.generateMockObject(ctx, bucket, key)
	}

	// 实际第三方API调用（此处为示例，实际应调用真实API）
	return s.callRealThirdPartyAPI(ctx, bucket, key)
}

// generateMockObject 生成Mock对象数据
func (s *ThirdPartyService) generateMockObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	// 生成Mock数据
	mockData := s.generateMockData(key)
	
	// 计算MD5哈希
	hash := md5.Sum(mockData)
	md5Hash := fmt.Sprintf("%x", hash)

	object := &models.Object{
		ID:          uuid.New().String(),
		Key:         key,
		Bucket:      bucket,
		Size:        int64(len(mockData)),
		ContentType: s.config.Mock.DefaultContentType,
		MD5Hash:     md5Hash,
		Data:        mockData,
		Headers:     map[string]string{
			"X-Third-Party-Source": "mock-service",
			"X-Generated-At":       time.Now().Format(time.RFC3339),
		},
		Tags: map[string]string{
			"source": "third-party",
			"mock":   "true",
		},
		CreatedAt: time.Now(),
	}

	s.logger.Info(ctx, "Generated mock object from third-party service",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.Int64("size", object.Size),
		observability.String("md5", md5Hash),
		observability.String("source", "mock"))

	return object, nil
}

// generateMockData 生成Mock数据内容
func (s *ThirdPartyService) generateMockData(key string) []byte {
	// 根据key生成不同类型的Mock数据
	switch {
	case len(key) > 10 && key[len(key)-4:] == ".txt":
		return []byte(fmt.Sprintf("This is mock text content for key: %s\nGenerated at: %s\nContent from third-party mock service.", 
			key, time.Now().Format(time.RFC3339)))
	case len(key) > 10 && key[len(key)-4:] == ".json":
		return []byte(fmt.Sprintf(`{"key": "%s", "generated_at": "%s", "source": "third-party-mock", "data": {"message": "Mock JSON content"}}`, 
			key, time.Now().Format(time.RFC3339)))
	case len(key) > 10 && key[len(key)-4:] == ".xml":
		return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><root><key>%s</key><generated_at>%s</generated_at><source>third-party-mock</source><message>Mock XML content</message></root>`, 
			key, time.Now().Format(time.RFC3339)))
	default:
		// 生成二进制数据
		size := 100 + s.rand.Intn(900) // 100-1000字节的随机数据
		data := make([]byte, size)
		
		// 生成可识别的模式数据而不是纯随机数据
		pattern := []byte(fmt.Sprintf("MOCK-DATA-FOR-%s-", key))
		for i := 0; i < size; i++ {
			data[i] = pattern[i%len(pattern)]
		}
		
		return data
	}
}

// callRealThirdPartyAPI 调用真实的第三方API（示例实现）
func (s *ThirdPartyService) callRealThirdPartyAPI(ctx context.Context, bucket, key string) (*models.Object, error) {
	s.logger.Info(ctx, "Calling real third-party API",
		observability.String("bucket", bucket),
		observability.String("key", key))

	// 遍历配置的数据源
	for _, dataSource := range s.config.GetEnabledDataSources() {
		s.logger.Debug(ctx, "Trying data source",
			observability.String("source", dataSource.Name),
			observability.String("url", dataSource.URL))

		// 这里应该实现真实的HTTP客户端调用
		// 示例：使用HTTP客户端从 dataSource.URL 获取数据
		
		// 暂时返回错误，表示未实现
		s.logger.Debug(ctx, "Real API call not implemented, returning error",
			observability.String("source", dataSource.Name))
	}

	return nil, fmt.Errorf("real third-party API not implemented, bucket=%s, key=%s", bucket, key)
}

// HealthCheck 健康检查
func (s *ThirdPartyService) HealthCheck(ctx context.Context) error {
	s.logger.Debug(ctx, "Performing health check")

	// 如果是Mock模式，直接返回健康
	if s.config.Mock.Enabled {
		s.logger.Debug(ctx, "Health check passed (mock mode)")
		return nil
	}

	// 检查配置的数据源是否可用
	enabledSources := s.config.GetEnabledDataSources()
	if len(enabledSources) == 0 {
		return fmt.Errorf("no enabled data sources configured")
	}

	// 这里应该检查真实数据源的连通性
	// 暂时认为健康
	s.logger.Debug(ctx, "Health check passed",
		observability.Int("enabled_sources", len(enabledSources)))

	return nil
}

// GetStats 获取服务统计信息
func (s *ThirdPartyService) GetStats(ctx context.Context) (map[string]any, error) {
	enabledSources := s.config.GetEnabledDataSources()
	
	stats := map[string]any{
		"service":          "third-party-service",
		"mode":             map[string]any{"mock_enabled": s.config.Mock.Enabled},
		"data_sources":     len(s.config.DataSources),
		"enabled_sources":  len(enabledSources),
		"config": map[string]any{
			"success_rate": s.config.Mock.SuccessRate,
			"latency_ms":   s.config.Mock.LatencyMs,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.logger.Debug(ctx, "Retrieved service stats",
		observability.Int("data_sources", len(s.config.DataSources)),
		observability.Int("enabled_sources", len(enabledSources)),
		observability.String("mock_enabled", fmt.Sprintf("%v", s.config.Mock.Enabled)))

	return stats, nil
}