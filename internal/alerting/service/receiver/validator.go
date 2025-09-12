package receiver

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateAMWebhook(w *AMWebhook) error {
	if w == nil {
		return ErrInvalidPayload
	}
	if len(w.Alerts) == 0 {
		return errors.New("alerts empty")
	}
	for i := range w.Alerts {
		a := &w.Alerts[i]
		if a.StartsAt.IsZero() {
			return fmt.Errorf("alerts[%d].startsAt empty", i)
		}
		if a.Status == "" {
			a.Status = "firing"
		}
	}
	return nil
}

var allowedLevels = map[string]bool{"P0": true, "P1": true, "P2": true, "WARNING": true}

func NormalizeLevel(sev string) string {
	s := strings.ToUpper(strings.TrimSpace(sev))
	if allowedLevels[s] {
		return s
	}
	return "Warning"
}
