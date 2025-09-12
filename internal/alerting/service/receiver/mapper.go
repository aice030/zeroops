package receiver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func MapToAlertIssueRow(w *AMWebhook, a *AMAlert) (*AlertIssueRow, error) {
	// Title
	title := strings.TrimSpace(a.Annotations["summary"])
	if title == "" {
		title = strings.TrimSpace(fmt.Sprintf("%s %s %s", a.Labels["idc"], a.Labels["service"], a.Labels["alertname"]))
		if title == "" {
			title = "Alert from Alertmanager"
		}
	}
	if len(title) > 255 {
		title = title[:255]
	}

	level := NormalizeLevel(a.Labels["severity"])

	flat := make([]map[string]string, 0, len(a.Labels)+3)
	for k, v := range a.Labels {
		flat = append(flat, map[string]string{"key": k, "value": v})
	}
	if a.Fingerprint != "" {
		flat = append(flat, map[string]string{"key": "am_fingerprint", "value": a.Fingerprint})
	}
	if g := strings.TrimSpace(a.GeneratorURL); g != "" {
		flat = append(flat, map[string]string{"key": "generatorURL", "value": g})
	}
	if w.GroupKey != "" {
		flat = append(flat, map[string]string{"key": "groupKey", "value": w.GroupKey})
	}
	b, _ := json.Marshal(flat)

	return &AlertIssueRow{
		ID:         uuid.NewString(),
		State:      "Open",
		AlertState: "InProcessing",
		Level:      level,
		Title:      title,
		LabelJSON:  b,
		AlertSince: a.StartsAt.UTC().Truncate(time.Second),
	}, nil
}
