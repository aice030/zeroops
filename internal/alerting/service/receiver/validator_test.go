package receiver

import (
	"testing"
	"time"
)

func TestValidateAMWebhook(t *testing.T) {
	if err := ValidateAMWebhook(&AMWebhook{}); err == nil {
		t.Fatal("expected error for empty alerts")
	}

	w := &AMWebhook{Alerts: []AMAlert{{}}}
	if err := ValidateAMWebhook(w); err == nil {
		t.Fatal("expected error for empty startsAt")
	}

	w = &AMWebhook{Alerts: []AMAlert{{StartsAt: time.Now()}}}
	if err := ValidateAMWebhook(w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeLevel(t *testing.T) {
	if NormalizeLevel("p1") != "P1" && NormalizeLevel("p1") != "Warning" {
		t.Fatal("unexpected level normalization")
	}
}
