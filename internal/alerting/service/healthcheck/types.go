package healthcheck

import (
	"time"
)

// AlertMessage is the payload sent to downstream processors.
type AlertMessage struct {
	ID         string            `json:"id"`
	Service    string            `json:"service"`
	Version    string            `json:"version,omitempty"`
	Level      string            `json:"level"`
	Title      string            `json:"title"`
	AlertSince time.Time         `json:"alert_since"`
	Labels     map[string]string `json:"labels"`
}

// deriveHealth maps alert level to service health state.
func deriveHealth(level string) string {
	switch level {
	case "P0":
		return "Error"
	case "P1", "P2":
		return "Warning"
	default:
		return "Warning"
	}
}
