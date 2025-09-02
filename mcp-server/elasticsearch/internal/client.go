package internal

import (
	"time"

	"qiniu1024-mcp-server/pkg/common/config"
)

// ElasticsearchClient Elasticsearch客户端
type ElasticsearchClient struct {
	BaseURL      string        // Elasticsearch基础URL
	Timeout      time.Duration // 请求超时时间
	MaxRetries   int           // 最大重试次数
	IndexManager *IndexManager // 索引管理器
}

// NewElasticsearchClient 创建Elasticsearch客户端
func NewElasticsearchClient() (*ElasticsearchClient, error) {
	// 从配置文件获取Elasticsearch配置
	esConfig := config.GlobalConfig.ElasticSearch.Connection

	// 设置默认值
	baseURL := esConfig.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:9200"
	}

	timeout := time.Duration(esConfig.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	maxRetries := esConfig.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	// 创建索引管理器
	indexManager := NewIndexManager()

	return &ElasticsearchClient{
		BaseURL:      baseURL,
		Timeout:      timeout,
		MaxRetries:   maxRetries,
		IndexManager: indexManager,
	}, nil
}

// GetIndexManager 获取索引管理器
func (e *ElasticsearchClient) GetIndexManager() *IndexManager {
	return e.IndexManager
}

// GetConnectionConfig 获取连接配置信息
func (e *ElasticsearchClient) GetConnectionConfig() map[string]interface{} {
	return map[string]interface{}{
		"base_url":    e.BaseURL,
		"timeout":     e.Timeout.String(),
		"max_retries": e.MaxRetries,
	}
}
