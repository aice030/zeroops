package healthcheck

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	adb "github.com/qiniu/zeroops/internal/alerting/database"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Deps struct {
	DB       *adb.Database
	Redis    *redis.Client
	AlertCh  chan<- AlertMessage
	Batch    int
	Interval time.Duration
}

// NewRedisClientFromEnv constructs a redis client from env.
func NewRedisClientFromEnv() *redis.Client {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
}

func StartScheduler(ctx context.Context, deps Deps) {
	if deps.Interval <= 0 {
		deps.Interval = 10 * time.Second
	}
	if deps.Batch <= 0 {
		deps.Batch = 200
	}
	t := time.NewTicker(deps.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := runOnce(ctx, deps.DB, deps.Redis, deps.AlertCh, deps.Batch); err != nil {
				log.Error().Err(err).Msg("healthcheck runOnce failed")
			}
		}
	}
}

type pendingRow struct {
	ID         string
	Level      string
	Title      string
	AlertSince time.Time
	LabelsJSON string
}

func runOnce(ctx context.Context, db *adb.Database, rdb *redis.Client, ch chan<- AlertMessage, batch int) error {
	rows, err := queryPendingFromDB(ctx, db, batch)
	if err != nil {
		return err
	}
	for _, it := range rows {
		labels := parseLabels(it.LabelsJSON)
		svc := labels["service"]
		ver := labels["service_version"]
		// 1) publish to channel (non-blocking)
		if ch != nil {
			select {
			case ch <- AlertMessage{ID: it.ID, Service: svc, Version: ver, Level: it.Level, Title: it.Title, AlertSince: it.AlertSince, Labels: labels}:
			default:
				// channel full, skip state change
				continue
			}
		}
		// 2) alert state CAS: Pending -> InProcessing
		_ = alertStateCAS(ctx, rdb, it.ID, "Pending", "InProcessing")
		// 3) service state CAS by derived level
		if svc != "" {
			target := deriveHealth(it.Level)
			_ = serviceStateCAS(ctx, rdb, svc, ver, target)
		}
	}
	return nil
}

func queryPendingFromDB(ctx context.Context, db *adb.Database, limit int) ([]pendingRow, error) {
	if db == nil {
		return []pendingRow{}, nil
	}
	const q = `SELECT id, level, title, labels, alert_since
FROM alert_issues
WHERE alert_state = 'Pending'
ORDER BY alert_since ASC
LIMIT $1`
	rows, err := db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]pendingRow, 0, limit)
	for rows.Next() {
		var it pendingRow
		if err := rows.Scan(&it.ID, &it.Level, &it.Title, &it.LabelsJSON, &it.AlertSince); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func alertStateCAS(ctx context.Context, rdb *redis.Client, id, expected, next string) error {
	if rdb == nil {
		return nil
	}
	key := "alert:issue:" + id
	script := redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then return 0 end
local obj = cjson.decode(v)
if obj.alertState ~= ARGV[1] then return -1 end
obj.alertState = ARGV[2]
redis.call('SET', KEYS[1], cjson.encode(obj), 'KEEPTTL')
redis.call('SREM', KEYS[2], ARGV[3])
redis.call('SADD', KEYS[3], ARGV[3])
return 1
`)
	_, _ = script.Run(ctx, rdb, []string{key, "alert:index:alert_state:Pending", "alert:index:alert_state:InProcessing"}, expected, next, id).Result()
	return nil
}

func serviceStateCAS(ctx context.Context, rdb *redis.Client, service, version, target string) error {
	if rdb == nil {
		return nil
	}
	key := "service_state:" + service + ":" + version
	script := redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then v = '{}'; end
local obj = cjson.decode(v)
obj.health_state = ARGV[2]
redis.call('SET', KEYS[1], cjson.encode(obj), 'KEEPTTL')
if ARGV[2] ~= '' then redis.call('SADD', KEYS[3], KEYS[1]) end
return 1
`)
	_, _ = script.Run(ctx, rdb, []string{key, "", "service_state:index:health:" + target}, "", target, key).Result()
	return nil
}

// parseLabels supports either flat map {"k":"v"} or array [{"key":"k","value":"v"}]
func parseLabels(s string) map[string]string {
	m := map[string]string{}
	if s == "" {
		return m
	}
	// try map form
	if json.Unmarshal([]byte(s), &m) == nil && len(m) > 0 {
		return m
	}
	// try array form
	var arr []struct{ Key, Value string }
	if json.Unmarshal([]byte(s), &arr) == nil {
		out := make(map[string]string, len(arr))
		for _, kv := range arr {
			out[kv.Key] = kv.Value
		}
		return out
	}
	return map[string]string{}
}
