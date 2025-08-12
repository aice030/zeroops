package httpclient

import (
	"context"
	"net/http"
	"shared/config"
)

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	// Do 执行HTTP请求
	Do(req *http.Request) (*http.Response, error)

	// DoWithContext 执行带上下文的HTTP请求
	DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error)

	// Get 执行GET请求
	Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error)

	// Post 执行POST请求
	Post(ctx context.Context, url, contentType string, body []byte, headers map[string]string) (*http.Response, error)

	// Put 执行PUT请求
	Put(ctx context.Context, url, contentType string, body []byte, headers map[string]string) (*http.Response, error)

	// Delete 执行DELETE请求
	Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error)

	// Close 关闭客户端
	Close() error
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(config config.HTTPClientConfig) HTTPClient {
	// TODO: 实现HTTP客户端
	return nil
}
