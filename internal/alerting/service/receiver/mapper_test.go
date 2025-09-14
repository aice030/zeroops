package receiver

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMapToAlertIssueRow(t *testing.T) {
	w := &AMWebhook{GroupKey: "gk"}
	a := &AMAlert{
		Annotations:  KV{"summary": "a very very very long title that should be accepted"},
		Labels:       KV{"severity": "P1", "alertname": "X", "service": "svc", "idc": "idc"},
		StartsAt:     time.Now(),
		Fingerprint:  "fp",
		GeneratorURL: "http://gen",
	}
	row, err := MapToAlertIssueRow(w, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.State != "Open" || row.AlertState != "Pending" {
		t.Fatal("unexpected state mapping")
	}
	var flat []map[string]string
	if err := json.Unmarshal(row.LabelJSON, &flat); err != nil {
		t.Fatalf("invalid label json: %v", err)
	}
	// ensure am_fingerprint present
	found := false
	for _, kv := range flat {
		if kv["key"] == "am_fingerprint" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("am_fingerprint not found in label json")
	}
}
