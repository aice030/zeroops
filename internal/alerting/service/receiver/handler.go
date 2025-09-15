package receiver

import (
	"net/http"
	"strings"
	"time"

	"github.com/fox-gonic/fox"
)

type Handler struct {
	dao   AlertIssueDAO
	cache AlertIssueCache
}

// NewHandler keeps backward compatibility and uses a NoopCache by default.
func NewHandler(dao AlertIssueDAO) *Handler { return &Handler{dao: dao, cache: NoopCache{}} }

// NewHandlerWithCache allows injecting a real cache implementation.
func NewHandlerWithCache(dao AlertIssueDAO, cache AlertIssueCache) *Handler {
	if cache == nil {
		cache = NoopCache{}
	}
	return &Handler{dao: dao, cache: cache}
}

func (h *Handler) AlertmanagerWebhook(c *fox.Context) {
	if !AuthMiddleware(c) {
		return
	}
	var req AMWebhook
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid JSON"})
		return
	}

	if err := ValidateAMWebhook(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if strings.ToLower(req.Status) != "firing" {
		c.JSON(http.StatusOK, map[string]any{"ok": true, "msg": "ignored (not firing)"})
		return
	}

	created := 0
	for _, a := range req.Alerts {
		key := BuildIdempotencyKey(a)
		// Distributed idempotency (best-effort). If key exists, skip.
		if ok, _ := h.cache.TryMarkIdempotent(c.Request.Context(), a); !ok {
			continue
		}
		if AlreadySeen(key) {
			continue
		}
		row, err := MapToAlertIssueRow(&req, &a)
		if err != nil {
			continue
		}
		if err := h.dao.InsertAlertIssue(c.Request.Context(), row); err != nil {
			continue
		}

		if w, ok := h.dao.(ServiceStateWriter); ok {
			service := strings.TrimSpace(a.Labels["service"])
			version := strings.TrimSpace(a.Labels["service_version"]) // optional
			if service != "" {
				derived := "Warning"
				if row.Level == "P0" {
					derived = "Error"
				} else if row.Level == "P1" || row.Level == "P2" {
					derived = "Warning"
				}
				_ = w.UpsertServiceState(c.Request.Context(), service, version, nil, derived, row.ID)
				_ = h.cache.WriteServiceState(c.Request.Context(), service, version, time.Time{}, derived)
			}
		}
		// Write-through to cache. Errors are ignored to avoid impacting webhook ack.
		_ = h.cache.WriteIssue(c.Request.Context(), row, a)
		MarkSeen(key)
		created++
	}

	c.JSON(http.StatusOK, map[string]any{"ok": true, "created": created})
}
