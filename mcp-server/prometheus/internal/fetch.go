package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"qiniu1024-mcp-server/pkg/common/config"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var (
	apiTimeout = 10 * time.Second
)

type queryApiResponse struct {
	Result   string      `json:"result"`
	Warnings v1.Warnings `json:"warnings"`
}

// matchURL 根据 regionCode 返回对应的 Prometheus URL
func matchURL(regionCode string) (string, error) {
	// 根据regionCode全局配置 config.GlobalConfig 读取映射关系，获取url地址，提升灵活性和安全性
	url, ok := config.GlobalConfig.Prometheus.Regions[regionCode]
	if !ok {
		return "", fmt.Errorf("不支持的 regionCode: %s", regionCode)
	}
	return url, nil
}

// FetchByPromQl 执行PromQL查询并返回JSON格式的数据
func (p *PrometheusClient) FetchByPromQl(promql string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	// 执行PromQL查询
	result, warnings, err := p.Client.Query(ctx, promql, time.Now())
	if err != nil {
		return "", fmt.Errorf("PromQL 查询失败: %w", err)
	}

	// 记录警告信息
	if len(warnings) > 0 {
		fmt.Printf("Prometheus 查询警告: %v\n", warnings)
	}

	// 构建标准的Prometheus响应格式
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": result.Type().String(),
			"result":     result,
		},
	}

	// 将结果转换为JSON格式
	jsonData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %w", err)
	}

	return string(jsonData), nil
}

func (p *PrometheusClient) FetchMetricsList(ctx context.Context, label string, matches []string, start, end time.Time) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	result, warnings, err := p.Client.LabelValues(ctx, label, matches, start, end)
	if err != nil {
		return "", fmt.Errorf("error getting label values: %w", err)
	}

	lvals := make([]string, len(result))
	for i, lval := range result {
		lvals[i] = string(lval)
	}

	res := queryApiResponse{
		Result:   strings.Join(lvals, "\n"),
		Warnings: warnings,
	}

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return "", fmt.Errorf("error converting label values response to JSON: %w", err)
	}

	return string(jsonBytes), nil
}
