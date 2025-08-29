package metrics

import (
	"fmt"
	"qiniu1024-mcp-server/pkg/models"
)

// 指标模板和对应的阈值
var metricTemplates = map[string]struct {
	templates  []string
	thresholds []string
}{
	"CPU": {
		templates:  []string{"%s_service_cpu_user_seconds_total", "%s_service_cpu_system_seconds_total"},
		thresholds: []string{"3s", "5s"}, // 用户CPU时间最高3s，系统CPU最高5s
	},
	"Memory": {
		templates:  []string{"%s_service_memory_used_bytes", "%s_service_memory_active_bytes"},
		thresholds: []string{"1GB", "2GB"}, // 内存使用最高1GB，活跃内存最高2GB
	},
	"Error Rate": {
		templates:  []string{"%s_service_errors_total"},
		thresholds: []string{"0.1%"}, // 错误率最高0.1%
	},
}

// BuildMetricsForService 根据服务名和指标模板拼接各问题对应的Prometheus指标列表
func BuildMetricsForService(service string) []map[string]interface{} {
	metrics := []map[string]interface{}{}
	for name, data := range metricTemplates {
		candidates := []string{}
		for _, tpl := range data.templates {
			candidates = append(candidates, fmt.Sprintf(tpl, service))
		}
		metrics = append(metrics, map[string]interface{}{
			"name":       name,
			"candidates": candidates,
			"thresholds": data.thresholds, // 添加阈值字段，与candidates一一对应
		})
	}
	return metrics
}

func GenerateMetricList(service string) models.StepResult {
	details := map[string]interface{}{
		"service": service,
		"metrics": BuildMetricsForService(service),
	}
	return models.NewStepResult("MetricAnalysis", fmt.Sprintf("%d group metrics identified with thresholds", len(metricTemplates)), details)
}
