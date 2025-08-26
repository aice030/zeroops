package metrics

import (
	"fmt"
	"qiniu1024-mcp-server/pkg/models"
)

// 指标模板
var metricTemplates = map[string][]string{
	"CPU":        {"%s_service_cpu_user_seconds_total", "%s_service_cpu_system_seconds_total"},
	"Memory":     {"%s_service_memory_used_bytes", "%s_service_memory_active_bytes"},
	"Error Rate": {"%s_service_errors_total"},
}

// BuildMetricsForService 根据服务名和指标模板拼接各问题对应的Prometheus指标列表
func BuildMetricsForService(service string) []map[string]interface{} {
	metrics := []map[string]interface{}{}
	for name, templates := range metricTemplates {
		candidates := []string{}
		for _, tpl := range templates {
			candidates = append(candidates, fmt.Sprintf(tpl, service))
		}
		metrics = append(metrics, map[string]interface{}{
			"name":       name,
			"candidates": candidates,
		})
	}
	return metrics
}

func GenerateMetricList(service string) models.StepResult {
	details := map[string]interface{}{
		"service": service,
		"metrics": BuildMetricsForService(service),
	}
	return models.NewStepResult("MetricAnalysis", fmt.Sprintf("%d group metrics identified", len(metricTemplates)), details)
}
