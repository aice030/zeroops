package issue

import (
	"qiniu1024-mcp-server/pkg/models"
	"time"
)

func GenerateIssue(service, candidate string, steps []models.StepResult) map[string]interface{} {
	return map[string]interface{}{
		"title":      "Release Preparation - " + service + " " + candidate,
		"service":    service,
		"candidate":  candidate,
		"steps":      steps,
		"created_at": time.Now().Format(time.RFC3339),
	}
}
