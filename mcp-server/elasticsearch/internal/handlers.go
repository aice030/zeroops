package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetServiceByHostIDHandler 处理根据host_id获取service的请求
func GetServiceByHostIDHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hostID := req.GetString("host_id", "")
	startTime := req.GetString("start_time", "")
	endTime := req.GetString("end_time", "")

	// 创建Elasticsearch客户端
	client, err := NewElasticsearchClient()
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("创建Elasticsearch客户端失败: %v", err))},
		}, nil
	}

	// 调用获取service的方法
	service, err := client.GetServiceByHostID(hostID, startTime, endTime)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取service失败: %v", err))},
		}, nil
	}

	// 格式化返回结果
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"host_id":    hostID,
			"service":    service,
			"start_time": startTime,
			"end_time":   endTime,
		},
	}

	responseData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("响应JSON序列化失败: %v", err))},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(responseData))},
	}, nil
}

// FetchLogsByServiceAndHostHandler 处理根据服务和host_id获取日志的请求
func FetchLogsByServiceAndHostHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	service := req.GetString("service", "")
	hostID := req.GetString("host_id", "")
	startTime := req.GetString("start_time", "")
	endTime := req.GetString("end_time", "")

	// 创建Elasticsearch客户端
	client, err := NewElasticsearchClient()
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("创建Elasticsearch客户端失败: %v", err))},
		}, nil
	}

	// 调用获取日志的方法
	data, err := client.FetchLogsByServiceAndHost(service, hostID, startTime, endTime)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取日志失败: %v", err))},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(data)},
	}, nil
}

// FetchRequestTraceHandler 处理根据request_id追踪请求的请求
func FetchRequestTraceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestID := req.GetString("request_id", "")
	startTime := req.GetString("start_time", "")
	endTime := req.GetString("end_time", "")

	// 创建Elasticsearch客户端
	client, err := NewElasticsearchClient()
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("创建Elasticsearch客户端失败: %v", err))},
		}, nil
	}

	// 调用追踪请求的方法
	data, err := client.FetchRequestTrace(requestID, startTime, endTime)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("追踪请求失败: %v", err))},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(data)},
	}, nil
}
