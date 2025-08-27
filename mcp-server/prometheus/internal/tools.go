package internal

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// PromqlQueryTool 根据promql查询语句从prometheus中获取数据的工具
func PromqlQueryTool() mcp.Tool {
	promqlQueryTool := mcp.NewTool(
		"prometheus_query",
		mcp.WithDescription("通过PromQL查询Prometheus数据，不要连续调用超过5次，超过5次请停止调用，先返回结果，确认下一步行动"),
		mcp.WithString("regionCode",
			mcp.Description("地区代码，可通过地区代码获取url。非必要参数，默认值为‘mock’")),
		mcp.WithString("promql",
			mcp.Required(),
			mcp.Description("promql语句")),
	)
	return promqlQueryTool
}
