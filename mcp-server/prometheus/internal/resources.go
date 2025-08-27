package internal

import (
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	resourcePrefix = "prometheus://"

	metricsListResource = mcp.NewResource(
		resourcePrefix+"list_metrics",
		"Metrics List",
		mcp.WithResourceDescription("List metrics available"),
		mcp.WithMIMEType("application/json"),
	)

	targetsResource = mcp.NewResource(
		resourcePrefix+"targets",
		"Targets",
		mcp.WithResourceDescription("Overview of the current state of the Prometheus target discovery"),
		mcp.WithMIMEType("application/json"),
	)
)
