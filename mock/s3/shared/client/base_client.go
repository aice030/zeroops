package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mocks3/shared/observability"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// BaseHTTPClient 基础HTTP客户端，封装通用的HTTP操作
type BaseHTTPClient struct {
	baseURL     string
	httpClient  *http.Client
	timeout     time.Duration
	serviceName string
	logger      *observability.Logger
}

// NewBaseHTTPClient 创建基础HTTP客户端
func NewBaseHTTPClient(baseURL string, timeout time.Duration, serviceName string, logger *observability.Logger) *BaseHTTPClient {
	client := &http.Client{
		Timeout:   timeout,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	return &BaseHTTPClient{
		baseURL:     baseURL,
		httpClient:  client,
		timeout:     timeout,
		serviceName: serviceName,
		logger:      logger,
	}
}

// RequestOptions 请求选项
type RequestOptions struct {
	Method      string
	Path        string
	Body        any
	QueryParams map[string]string
	Headers     map[string]string
}

// DoRequest 执行HTTP请求
func (c *BaseHTTPClient) DoRequest(ctx context.Context, opts RequestOptions) (*http.Response, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("http.method", opts.Method),
		attribute.String("http.url", opts.Path),
		attribute.String("service.name", c.serviceName),
	)

	// 构建URL
	requestURL, err := c.buildURL(opts.Path, opts.QueryParams)
	if err != nil {
		c.logger.Error(ctx, "Failed to build URL",
			observability.Error(err),
			observability.String("path", opts.Path))
		return nil, fmt.Errorf("build url: %w", err)
	}

	// 构建请求体
	var bodyReader io.Reader
	if opts.Body != nil {
		bodyBytes, err := json.Marshal(opts.Body)
		if err != nil {
			c.logger.Error(ctx, "Failed to marshal request body", observability.Error(err))
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, opts.Method, requestURL, bodyReader)
	if err != nil {
		c.logger.Error(ctx, "Failed to create HTTP request", observability.Error(err))
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置默认头部
	if opts.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义头部
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// 执行请求
	c.logger.Debug(ctx, "Sending HTTP request",
		observability.String("method", opts.Method),
		observability.String("url", requestURL))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(ctx, "HTTP request failed",
			observability.Error(err),
			observability.String("url", requestURL))
		return nil, fmt.Errorf("do request: %w", err)
	}

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	c.logger.Debug(ctx, "HTTP request completed",
		observability.String("url", requestURL),
		observability.Int("status_code", resp.StatusCode))

	return resp, nil
}

// DoRequestWithJSON 执行请求并解析JSON响应
func (c *BaseHTTPClient) DoRequestWithJSON(ctx context.Context, opts RequestOptions, result any) error {
	resp, err := c.DoRequest(ctx, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !isSuccessStatus(resp.StatusCode) {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error(ctx, "HTTP request failed",
			observability.Int("status_code", resp.StatusCode),
			observability.String("response_body", string(body)))
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			c.logger.Error(ctx, "Failed to decode JSON response", observability.Error(err))
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// DoRequestExpectStatus 执行请求并检查期望的状态码
func (c *BaseHTTPClient) DoRequestExpectStatus(ctx context.Context, opts RequestOptions, expectedStatus ...int) error {
	resp, err := c.DoRequest(ctx, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	for _, status := range expectedStatus {
		if resp.StatusCode == status {
			return nil
		}
	}

	return fmt.Errorf("unexpected status code: %d, expected: %v", resp.StatusCode, expectedStatus)
}

// Get 执行GET请求
func (c *BaseHTTPClient) Get(ctx context.Context, path string, queryParams map[string]string, result any) error {
	opts := RequestOptions{
		Method:      "GET",
		Path:        path,
		QueryParams: queryParams,
	}
	return c.DoRequestWithJSON(ctx, opts, result)
}

// Post 执行POST请求
func (c *BaseHTTPClient) Post(ctx context.Context, path string, body any, result any) error {
	opts := RequestOptions{
		Method: "POST",
		Path:   path,
		Body:   body,
	}
	return c.DoRequestWithJSON(ctx, opts, result)
}

// PostExpectStatus 执行POST请求并检查状态码
func (c *BaseHTTPClient) PostExpectStatus(ctx context.Context, path string, body any, expectedStatus ...int) error {
	opts := RequestOptions{
		Method: "POST",
		Path:   path,
		Body:   body,
	}
	return c.DoRequestExpectStatus(ctx, opts, expectedStatus...)
}

// Put 执行PUT请求
func (c *BaseHTTPClient) Put(ctx context.Context, path string, body any, result any) error {
	opts := RequestOptions{
		Method: "PUT",
		Path:   path,
		Body:   body,
	}
	return c.DoRequestWithJSON(ctx, opts, result)
}

// PutExpectStatus 执行PUT请求并检查状态码
func (c *BaseHTTPClient) PutExpectStatus(ctx context.Context, path string, body any, expectedStatus ...int) error {
	opts := RequestOptions{
		Method: "PUT",
		Path:   path,
		Body:   body,
	}
	return c.DoRequestExpectStatus(ctx, opts, expectedStatus...)
}

// Delete 执行DELETE请求
func (c *BaseHTTPClient) Delete(ctx context.Context, path string, expectedStatus ...int) error {
	opts := RequestOptions{
		Method: "DELETE",
		Path:   path,
	}
	if len(expectedStatus) == 0 {
		expectedStatus = []int{http.StatusNoContent, http.StatusOK}
	}
	return c.DoRequestExpectStatus(ctx, opts, expectedStatus...)
}

// HealthCheck 健康检查
func (c *BaseHTTPClient) HealthCheck(ctx context.Context) error {
	opts := RequestOptions{
		Method: "GET",
		Path:   "/health",
	}
	return c.DoRequestExpectStatus(ctx, opts, http.StatusOK)
}

// BuildQueryParams 构建查询参数的辅助函数
func BuildQueryParams(params map[string]any) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		switch val := v.(type) {
		case string:
			if val != "" {
				result[k] = val
			}
		case int:
			if val > 0 {
				result[k] = strconv.Itoa(val)
			}
		case int64:
			if val > 0 {
				result[k] = strconv.FormatInt(val, 10)
			}
		case bool:
			result[k] = strconv.FormatBool(val)
		}
	}
	return result
}

// PathEscape URL路径转义的便利函数
func PathEscape(segment string) string {
	return url.PathEscape(segment)
}

// buildURL 构建请求URL
func (c *BaseHTTPClient) buildURL(path string, queryParams map[string]string) (string, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", err
	}

	if len(queryParams) > 0 {
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// isSuccessStatus 检查是否为成功状态码
func isSuccessStatus(status int) bool {
	return status >= 200 && status < 300
}
