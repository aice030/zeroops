package receiver

import (
	"testing"
	"time"
)

func TestBuildIdempotencyKey(t *testing.T) {
	a := AMAlert{Fingerprint: "fp", StartsAt: time.Unix(0, 123).UTC()}
	key := BuildIdempotencyKey(a)
	if key == "" || key[:2] != "fp" {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestAlreadySeenAndMarkSeen(t *testing.T) {
	key := "k|t"
	if AlreadySeen(key) {
		t.Fatal("should not be seen initially")
	}
	MarkSeen(key)
	if !AlreadySeen(key) {
		t.Fatal("should be seen after MarkSeen")
	}
}
