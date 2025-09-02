package internal

import (
	"fmt"
	"strings"
	"time"

	"qiniu1024-mcp-server/pkg/common/config"
)

// IndexManager Elasticsearch索引管理器
type IndexManager struct {
	pattern         string // 索引模式
	dateFormat      string // 日期格式
	wildcardPattern string // 通配符模式
}

// NewIndexManager 创建索引管理器
func NewIndexManager() *IndexManager {
	// 从全局配置获取索引配置
	esConfig := config.GlobalConfig.ElasticSearch.Index

	// 设置默认值
	pattern := esConfig.Pattern
	if pattern == "" {
		pattern = "mock-{service}-logs-{date}"
	}

	dateFormat := esConfig.DateFormat
	if dateFormat == "" {
		dateFormat = "2006.01.02"
	}

	wildcardPattern := esConfig.WildcardPattern
	if wildcardPattern == "" {
		wildcardPattern = "mock-*-logs-{date}"
	}

	return &IndexManager{
		pattern:         pattern,
		dateFormat:      dateFormat,
		wildcardPattern: wildcardPattern,
	}
}

// GenerateIndexName 生成具体的索引名称
// service: 服务名称
// date: 日期，如果为空则使用当前日期
func (im *IndexManager) GenerateIndexName(service string, date time.Time) string {
	if date.IsZero() {
		date = time.Now()
	}

	dateStr := date.Format(im.dateFormat)

	// 替换模式中的占位符
	indexName := im.pattern
	indexName = strings.ReplaceAll(indexName, "{service}", service)
	indexName = strings.ReplaceAll(indexName, "{date}", dateStr)

	return indexName
}

// GenerateWildcardPattern 生成通配符模式
// date: 日期，如果为空则使用当前日期
func (im *IndexManager) GenerateWildcardPattern(date time.Time) string {
	if date.IsZero() {
		date = time.Now()
	}

	dateStr := date.Format(im.dateFormat)

	// 替换模式中的占位符
	pattern := im.wildcardPattern
	pattern = strings.ReplaceAll(pattern, "{date}", dateStr)

	return pattern
}

// GenerateIndexPatternsForDateRange 为日期范围生成索引模式列表
// startDate: 开始日期
// endDate: 结束日期
// service: 服务名称，如果为空则使用通配符
func (im *IndexManager) GenerateIndexPatternsForDateRange(startDate, endDate time.Time, service string) []string {
	var patterns []string

	// 计算日期范围内的所有日期
	currentDate := startDate
	for !currentDate.After(endDate) {
		if service == "" {
			// 使用通配符模式
			pattern := im.GenerateWildcardPattern(currentDate)
			patterns = append(patterns, pattern)
		} else {
			// 使用具体服务模式
			indexName := im.GenerateIndexName(service, currentDate)
			patterns = append(patterns, indexName)
		}

		// 移动到下一天
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return patterns
}

// GenerateIndexPatternsForTimeRange 为时间范围生成索引模式列表
// startTime: 开始时间字符串，格式为 "2025-08-20T00:00:00Z"
// endTime: 结束时间字符串，格式为 "2025-08-20T23:59:59Z"
// service: 服务名称，如果为空则使用通配符
func (im *IndexManager) GenerateIndexPatternsForTimeRange(startTime, endTime string, service string) ([]string, error) {
	// 解析时间字符串
	startDate, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("解析开始时间失败: %w", err)
	}

	endDate, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("解析结束时间失败: %w", err)
	}

	// 提取日期部分（去掉时间）
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	return im.GenerateIndexPatternsForDateRange(startDate, endDate, service), nil
}

// GetCurrentIndexPattern 获取当前日期的索引模式
// service: 服务名称，如果为空则使用通配符
func (im *IndexManager) GetCurrentIndexPattern(service string) string {
	if service == "" {
		return im.GenerateWildcardPattern(time.Time{}) // 使用当前日期
	}
	return im.GenerateIndexName(service, time.Time{}) // 使用当前日期
}

// ValidateIndexPattern 验证索引模式是否有效
func (im *IndexManager) ValidateIndexPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("索引模式不能为空")
	}

	// 检查是否包含必要的占位符
	if !strings.Contains(pattern, "{date}") {
		return fmt.Errorf("索引模式必须包含{date}占位符")
	}

	return nil
}

// GetIndexConfig 获取索引配置信息
func (im *IndexManager) GetIndexConfig() map[string]interface{} {
	return map[string]interface{}{
		"pattern":          im.pattern,
		"date_format":      im.dateFormat,
		"wildcard_pattern": im.wildcardPattern,
	}
}
