package receiver

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// AlertIssueCache defines the minimal cache contract used by the handler.
type AlertIssueCache interface {
	WriteIssue(ctx context.Context, r *AlertIssueRow, a AMAlert) error
	TryMarkIdempotent(ctx context.Context, a AMAlert) (bool, error)
}

// NoopCache is a no-op implementation of AlertIssueCache.
type NoopCache struct{}

func (NoopCache) WriteIssue(ctx context.Context, r *AlertIssueRow, a AMAlert) error { return nil }
func (NoopCache) TryMarkIdempotent(ctx context.Context, a AMAlert) (bool, error)    { return true, nil }

// Cache implements AlertIssueCache using Redis.
type Cache struct{ R *redis.Client }

// NewCacheFromEnv constructs a Redis client using environment variables.
// REDIS_ADDR, REDIS_PASSWORD, REDIS_DB
func NewCacheFromEnv() *Cache {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	c := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	return &Cache{R: c}
}

// WriteIssue writes the alert issue into Redis as a JSON blob and updates a few indices.
// Best-effort: failure should not block the main flow.
func (c *Cache) WriteIssue(ctx context.Context, r *AlertIssueRow, a AMAlert) error {
	if c == nil || c.R == nil {
		return nil
	}
	key := "alert:issue:" + r.ID
	payload := map[string]any{
		"id":          r.ID,
		"state":       r.State,
		"level":       r.Level,
		"alertState":  r.AlertState,
		"title":       r.Title,
		"labels":      json.RawMessage(r.LabelJSON),
		"alertSince":  r.AlertSince,
		"fingerprint": a.Fingerprint,
		"service":     a.Labels["service"],
		"alertname":   a.Labels["alertname"],
	}
	b, _ := json.Marshal(payload)
	svc := strings.TrimSpace(a.Labels["service"])
	pipe := c.R.Pipeline()
	pipe.Set(ctx, key, b, 72*time.Hour)
	pipe.SAdd(ctx, "alert:index:open", r.ID)
	if svc != "" {
		pipe.SAdd(ctx, "alert:index:svc:"+svc+":open", r.ID)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// TryMarkIdempotent marks an alert event as processed using Redis SETNX + TTL.
// Returns false if the key already exists (duplicate).
func (c *Cache) TryMarkIdempotent(ctx context.Context, a AMAlert) (bool, error) {
	if c == nil || c.R == nil {
		return true, nil
	}
	k := "alert:idemp:" + a.Fingerprint + "|" + a.StartsAt.UTC().Format(time.RFC3339Nano)
	ok, err := c.R.SetNX(ctx, k, "1", 10*time.Minute).Result()
	return ok, err
}
