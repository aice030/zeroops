package receiver

import (
	"net/http"
	"os"

	"github.com/fox-gonic/fox"
)

func authEnabled() bool {
	return os.Getenv("ALERT_WEBHOOK_BASIC_USER") != "" ||
		os.Getenv("ALERT_WEBHOOK_BASIC_PASS") != "" ||
		os.Getenv("ALERT_WEBHOOK_BEARER") != ""
}

// AuthMiddleware returns false if unauthorized and writes a 401 response.
func AuthMiddleware(c *fox.Context) bool {
	if !authEnabled() {
		return true
	}

	user := os.Getenv("ALERT_WEBHOOK_BASIC_USER")
	pass := os.Getenv("ALERT_WEBHOOK_BASIC_PASS")
	bearer := os.Getenv("ALERT_WEBHOOK_BEARER")

	if user != "" || pass != "" {
		u, p, ok := c.Request.BasicAuth()
		if !ok || u != user || p != pass {
			c.JSON(http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return false
		}
		return true
	}

	if bearer != "" {
		if c.GetHeader("Authorization") != "Bearer "+bearer {
			c.JSON(http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return false
		}
	}
	return true
}
