package receiver

import (
	"sync"
	"time"
)

var (
	idempMu  sync.Mutex
	idempMap = make(map[string]time.Time)
)

func BuildIdempotencyKey(a AMAlert) string {
	return a.Fingerprint + "|" + a.StartsAt.UTC().Format(time.RFC3339Nano)
}

func AlreadySeen(key string) bool {
	idempMu.Lock()
	defer idempMu.Unlock()
	if t, ok := idempMap[key]; ok {
		if time.Since(t) < 30*time.Minute {
			return true
		}
		delete(idempMap, key)
	}
	return false
}

func MarkSeen(key string) {
	idempMu.Lock()
	defer idempMu.Unlock()
	idempMap[key] = time.Now()
}
