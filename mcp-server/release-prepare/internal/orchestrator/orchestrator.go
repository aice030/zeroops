package orchestrator

import (
	"qiniu1024-mcp-server/pkg/models"
	"qiniu1024-mcp-server/release-prepare/internal/issue"
	"qiniu1024-mcp-server/release-prepare/internal/metrics"
	"qiniu1024-mcp-server/release-prepare/internal/planner"
	"qiniu1024-mcp-server/release-prepare/internal/risk"
	"qiniu1024-mcp-server/release-prepare/internal/version"
)

func RunReleasePreparation(service, candidate string) map[string]interface{} {
	steps := []models.StepResult{}

	// 依次记录版本冲突、待监测指标列表、灰度策略和风险预案
	steps = append(steps, version.CheckVersion(service, candidate))
	steps = append(steps, metrics.GenerateMetricList(service))
	steps = append(steps, planner.ReleasePlan(service))
	steps = append(steps, risk.PredictRisk(service))

	// 根据记录结果生成发布issue
	return issue.GenerateIssue(service, candidate, steps)
}
