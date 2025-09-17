package receiver

import (
	"encoding/json"
	"time"
)

type KV map[string]string

// AMAlert represents a single alert from Alertmanager
type AMAlert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}

// AMWebhook is the root webhook payload from Alertmanager
type AMWebhook struct {
	Receiver          string    `json:"receiver"`
	Status            string    `json:"status"`
	Alerts            []AMAlert `json:"alerts"`
	GroupLabels       KV        `json:"groupLabels"`
	CommonLabels      KV        `json:"commonLabels"`
	CommonAnnotations KV        `json:"commonAnnotations"`
	ExternalURL       string    `json:"externalURL"`
	Version           string    `json:"version"`
	GroupKey          string    `json:"groupKey"`
}

// AlertIssueRow represents the row to insert into alert_issues
type AlertIssueRow struct {
	ID         string
	State      string
	Level      string
	AlertState string
	Title      string
	LabelJSON  json.RawMessage
	AlertSince time.Time
}
