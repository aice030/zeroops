package internal

import (
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	resourcePrefix = "prometheus://"

	metricsListResource = mcp.NewResource(
		resourcePrefix+"metricsList",
		"Prometheus Metrics List",
		mcp.WithResourceDescription("加载Prometheus平台监测的全部指标列表"),
		mcp.WithMIMEType("application/json"),
	)

	targetsResource = mcp.NewResource(
		resourcePrefix+"targets",
		"Targets",
		mcp.WithResourceDescription("Overview of the current state of the Prometheus target discovery"),
		mcp.WithMIMEType("application/json"),
	)
)
