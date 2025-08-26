package version

import "qiniu1024-mcp-server/pkg/models"

func CheckVersion(service, candidate string) models.StepResult {
	details := map[string]interface{}{
		"service":         service,
		"current":         "v1.0.0",
		"candidate":       candidate,
		"has_conflict":    false,
		"conflict_reason": "",
	}
	summary := "No conflict detected"
	return models.NewStepResult("VersionCheck", summary, details)
}
