package receiver

import (
	"net/http"
	"strings"

	"github.com/fox-gonic/fox"
)

type Handler struct {
	dao AlertIssueDAO
}

func NewHandler(dao AlertIssueDAO) *Handler { return &Handler{dao: dao} }

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
		MarkSeen(key)
		created++
	}

	c.JSON(http.StatusOK, map[string]any{"ok": true, "created": created})
}
