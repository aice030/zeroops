package receiver

import "github.com/fox-gonic/fox"

func RegisterReceiverRoutes(r *fox.Engine, h *Handler) {
	r.POST("/v1/integrations/alertmanager/webhook", h.AlertmanagerWebhook)
}
