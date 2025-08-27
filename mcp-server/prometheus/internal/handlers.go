package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"qiniu1024-mcp-server/pkg/formatter"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func MetricsListResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	log.Printf("开始加载prometheus指标")
	// 创建Prometheus客户端
	client, err := NewPrometheusClient("mock") // 先只支持单实例
	if err != nil {
		return nil, fmt.Errorf("error create prometheus client: %w", err)
	}

	// 拉取数据
	metrics, err := client.FetchMetricsList(ctx, "__name__", nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("error getting metric names: %w", err)
	}

	metricsJSON, _ := json.Marshal(metrics)
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      resourcePrefix + "metricsList",
			MIMEType: "application/json",
			Text:     string(metricsJSON),
		},
	}, nil
}

// PromqlQueryHandler 处理PromQL查询请求
func PromqlQueryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取请求参数
	regionCode := req.GetString("regionCode", "mock")
	promql := req.GetString("promql", "")

	// 参数验证
	if promql == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent("PromQL查询语句不能为空")},
		}, nil
	}

	log.Printf("开始执行PromQL查询: regionCode=%s, promql=%s", regionCode, promql)

	// 创建Prometheus客户端
	client, err := NewPrometheusClient(regionCode)
	if err != nil {
		log.Printf("创建Prometheus客户端失败: %v", err)
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("创建Prometheus客户端失败: %v", err))},
		}, nil
	}

	// 执行查询
	data, err := client.FetchByPromQl(promql)

	if err != nil {
		log.Printf("获取Prometheus数据失败: %v", err)
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取Prometheus数据失败: %v", err))},
		}, nil
	}

	log.Printf("成功获取原始数据，长度: %d", len(data))

	// 格式化数据
	formatter := formatter.NewPrometheusDataFormatter("Asia/Shanghai")
	formattedData, err := formatter.FormatPrometheusData(data, true)

	if err != nil {
		log.Printf("格式化数据失败: %v", err)
		// 即使格式化失败，也返回原始数据
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("数据格式化失败，返回原始数据: %v\n\n原始数据:\n%s", err, data))},
		}, nil
	}

	//log.Printf("数据格式化成功，返回格式化后的数据")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(formattedData)},
	}, nil
}
