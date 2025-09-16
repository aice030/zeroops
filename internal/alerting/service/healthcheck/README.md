# healthcheck — Pending 告警扫描与分发任务

本包提供一个定时任务：
- 周期性扫描 Pending 状态的告警
- 将告警投递到 channel（供下游处理器消费），后续再接入消息队列
- 成功投递后，原子地把缓存中的状态更新：
  - `alert:issue:{id}` 的 `alertState`：Pending → InProcessing
  - `service_state:{service}:{version}` 的 `health_state`：由告警等级推导（P0→Error；P1/P2→Warning）

此任务默认只更新缓存，不直接更新数据库，以降低耦合与避免与业务处理竞争。数据库状态可由下游处理器在处理开始时回写，或由后续补偿任务兜底。

——

## 1. 触发与频率

- 间隔：默认每 10s 扫描一次（可配置）
- 批量：每次最多处理 200 条 Pending（可配置）
- 并发：串行或小并发（<= 4），避免重复投递

环境变量建议：
```
HC_SCAN_INTERVAL=10s
HC_SCAN_BATCH=200
HC_WORKERS=1
```

——

## 2. 数据来源与过滤

优先以数据库为准，结合缓存加速：

- 数据库查询（推荐）
  ```sql
  SELECT id, level, title, labels, alert_since
  FROM alert_issues
  WHERE alert_state = 'Pending'
  ORDER BY alert_since ASC
  LIMIT $1;
  ```

当告警切换为 InProcessing 时，需要更新对应 `service_states.report_at` 为该 service/version 关联的 `alert_issue_ids` 中，所有 alert_issues 里 alert_state=InProcessing 的 `alert_since` 最早时间（min）。可通过下游处理器或本任务的补充逻辑回填：

```sql
UPDATE service_states ss
SET report_at = sub.min_since
FROM (
  SELECT si.service, si.version, MIN(ai.alert_since) AS min_since
  FROM service_states si
  JOIN alert_issues ai ON ai.id = ANY(si.alert_issue_ids)
  WHERE ai.alert_state = 'InProcessing'
  GROUP BY si.service, si.version
) AS sub
WHERE ss.service = sub.service AND ss.version = sub.version;
```

- 或仅用缓存（可选）：
  - 维护集合 `alert:index:alert_state:Pending`（若未维护，可临时 SCAN `alert:issue:*` 并过滤 JSON 中的 `alertState`，但不推荐在大规模下使用 SCAN）。

——

## 3. 通道（channel）

现阶段通过进程内 channel 向下游处理器传递告警消息，后续再平滑切换为消息队列（Kafka/NATS 等）。

消息格式保留为 `AlertMessage`：
```go
type AlertMessage struct {
    ID         string            `json:"id"`
    Service    string            `json:"service"`
    Version    string            `json:"version,omitempty"`
    Level      string            `json:"level"`
    Title      string            `json:"title"`
    AlertSince time.Time         `json:"alert_since"`
    Labels     map[string]string `json:"labels"`
}
```

发布样例（避免阻塞可用非阻塞写）：
```go
func publishToChannel(ctx context.Context, ch chan<- AlertMessage, m AlertMessage) error {
    select {
    case ch <- m:
        return nil
    default:
        return fmt.Errorf("alert channel full")
    }
}
```

配置：当前无需队列相关配置。未来切换到消息队列时，可启用以下配置项：
```
# ALERT_QUEUE_KIND=redis_stream|kafka|nats
# ALERT_QUEUE_DSN=redis://localhost:6379/0
# ALERT_QUEUE_TOPIC=alerts.pending
```

——

## 4. 缓存键与原子更新

现有（或建议）键：
- 告警：`alert:issue:{id}` → JSON，字段包含 `alertState`
- 指数（可选）：`alert:index:alert_state:{Pending|InProcessing|...}`
- 服务态：`service_state:{service}:{version}` → JSON，字段包含 `health_state`
- 指数：`service_state:index:health:{Error|Warning|...}`

为避免并发写冲突，建议使用 Lua CAS（Compare-And-Set）脚本原子修改值与索引：

```lua
-- KEYS[1] = alert key, ARGV[1] = expected, ARGV[2] = next, KEYS[2] = idx:old, KEYS[3] = idx:new, ARGV[3] = id
local v = redis.call('GET', KEYS[1])
if not v then return 0 end
local obj = cjson.decode(v)
if obj.alertState ~= ARGV[1] then return -1 end
obj.alertState = ARGV[2]
redis.call('SET', KEYS[1], cjson.encode(obj), 'KEEPTTL')
if KEYS[2] ~= '' then redis.call('SREM', KEYS[2], ARGV[3]) end
if KEYS[3] ~= '' then redis.call('SADD', KEYS[3], ARGV[3]) end
return 1
```

服务态类似（示例将态切换到推导的新态）：
```lua
-- KEYS[1] = service_state key, ARGV[1] = expected(optional), ARGV[2] = next, KEYS[2] = idx:old(optional), KEYS[3] = idx:new, ARGV[3] = member
local v = redis.call('GET', KEYS[1])
if not v then return 0 end
local obj = cjson.decode(v)
if ARGV[1] ~= '' and obj.health_state ~= ARGV[1] then return -1 end
obj.health_state = ARGV[2]
redis.call('SET', KEYS[1], cjson.encode(obj), 'KEEPTTL')
if KEYS[2] ~= '' then redis.call('SREM', KEYS[2], ARGV[3]) end
if KEYS[3] ~= '' then redis.call('SADD', KEYS[3], ARGV[3]) end
return 1
```

——

## 5. 任务流程（伪代码）

```go
func runOnce(ctx context.Context, db *Database, rdb *redis.Client, ch chan<- AlertMessage, batch int) error {
    rows := queryPendingFromDB(ctx, db, batch) // id, level, title, labels(JSON), alert_since
    for _, it := range rows {
        svc := it.Labels["service"]
        ver := it.Labels["service_version"]
        // 1) 投递消息到 channel（非阻塞）
        select {
        case ch <- AlertMessage{ID: it.ID, Service: svc, Version: ver, Level: it.Level, Title: it.Title, AlertSince: it.AlertSince, Labels: it.Labels}:
            // ok
        default:
            // 投递失败：通道已满，跳过状态切换，计数并继续
            continue
        }
        // 2) 缓存状态原子切换（告警）
        alertKey := "alert:issue:" + it.ID
        rdb.Eval(ctx, alertCAS, []string{alertKey, "alert:index:alert_state:Pending", "alert:index:alert_state:InProcessing"}, "Pending", "InProcessing", it.ID)
        // 3) 缓存状态原子切换（服务态：按告警等级推导）
        if svc != "" { // version 可空
            target := deriveHealth(it.Level) // P0->Error; P1/P2->Warning; else Warning
            svcKey := "service_state:" + svc + ":" + ver
            -- 可按需指定旧态索引，否则留空
            localOld := ''
            newIdx := "service_state:index:health:" + target
            member := svcKey
            rdb.Eval(ctx, svcCAS, []string{svcKey, localOld, newIdx}, '', target, member)
        }
    }
    return nil
}

func StartScheduler(ctx context.Context, deps Deps) {
    t := time.NewTicker(deps.Interval)
    defer t.Stop()
    for {
        select {
        case <-ctx.Done(): return
        case <-t.C:
            _ = runOnce(ctx, deps.DB, deps.Redis, deps.AlertCh, deps.Batch)
        }
    }
}
```

——

## 6. 可观测与重试

- 指标：扫描次数、选出数量、成功投递数量、CAS 成功/失败数量、用时分位
- 日志：每批开始/结束、首尾 ID、错误明细
- 重试：
  - 消息投递失败：不更改缓存状态，等待下次扫描重试
  - CAS 返回 -1（状态被他处更改）：记录并跳过

——

## 7. 本地验证

1) 准备 Redis 与 DB（见 receiver/README.md）

2) 造数据：插入一条 `alert_issues.alert_state='Pending'` 且缓存中存在 `alert:issue:{id}` 的 JSON。

3) 启动任务：观察日志/指标。

4) 验证缓存：
```bash
redis-cli --raw GET alert:issue:<id> | jq
redis-cli --raw SMEMBERS alert:index:alert_state:InProcessing | head -n 20
redis-cli --raw GET service_state:<service>:<version> | jq
redis-cli --raw SMEMBERS service_state:index:health:Processing | head -n 20
```

5) 验证 channel：在消费端确认是否收到消息。

——

## 8. 配置汇总

```
# 扫描任务
HC_SCAN_INTERVAL=10s
HC_SCAN_BATCH=200
HC_WORKERS=1

# 通道
# 当前无需额外配置
# 预留（未来切换到消息队列时启用）：
# ALERT_QUEUE_KIND=redis_stream|kafka|nats
# ALERT_QUEUE_DSN=redis://localhost:6379/0
# ALERT_QUEUE_TOPIC=alerts.pending
```

——


