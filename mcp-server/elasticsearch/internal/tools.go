package internal

import "github.com/mark3labs/mcp-go/mcp"

// GetServiceByHostIDTool 根据host_id获取对应service的工具
func GetServiceByHostIDTool() mcp.Tool {
	return mcp.NewTool(
		"elasticsearch_get_service",
		mcp.WithDescription("根据主机ID获取对应的服务名称，用于确定主机运行的服务"),
		mcp.WithString("host_id",
			mcp.Required(),
			mcp.Description("主机ID，用于定位具体的服务器节点")),
		mcp.WithString("start_time",
			mcp.Required(),
			mcp.Description("查询开始时间，格式为 '2025-08-20T00:00:00Z'")),
		mcp.WithString("end_time",
			mcp.Required(),
			mcp.Description("查询结束时间，格式为 '2025-08-20T23:59:59Z'")),
	)
}

// FetchLogsByServiceAndHostTool 根据服务和host_id获取时间段内所有日志的工具
func FetchLogsByServiceAndHostTool() mcp.Tool {
	return mcp.NewTool(
		"elasticsearch_fetch_logs",
		mcp.WithDescription("根据服务名称和主机ID获取指定时间段内的所有日志，用于分析服务运行状态。如果service参数为空，将自动根据host_id获取对应的service"),
		mcp.WithString("service",
			mcp.Description("服务名称，如 'storage-service'，如果为空将自动根据host_id获取")),
		mcp.WithString("host_id",
			mcp.Required(),
			mcp.Description("主机ID，用于定位具体的服务器节点")),
		mcp.WithString("start_time",
			mcp.Required(),
			mcp.Description("查询开始时间，格式为 '2025-08-20T00:00:00Z'")),
		mcp.WithString("end_time",
			mcp.Required(),
			mcp.Description("查询结束时间，格式为 '2025-08-20T23:59:59Z'")),
	)
}

// FetchRequestTraceTool 根据request_id追踪请求经过服务的工具
func FetchRequestTraceTool() mcp.Tool {
	return mcp.NewTool(
		"elasticsearch_request_trace",
		mcp.WithDescription("根据请求ID追踪该请求在指定时间段内经过的所有服务，用于请求链路分析"),
		mcp.WithString("request_id",
			mcp.Required(),
			mcp.Description("请求ID，用于追踪特定的请求")),
		mcp.WithString("start_time",
			mcp.Required(),
			mcp.Description("查询开始时间，格式为 '2025-08-20T00:00:00Z'")),
		mcp.WithString("end_time",
			mcp.Required(),
			mcp.Description("查询结束时间，格式为 '2025-08-20T23:59:59Z'")),
	)
}
