# receiver 变更记录（与数据库/API 文档同步）

时间：2025-09-12

## 变更摘要

- 对接最新 `docs/alerting/database-design.md` 与 `docs/alerting/api.md`，统一列命名与时间精度。
- 更新 DAO 插入语句、测试建表语句与 README 中的 SQL 片段。

## 具体修改

1) `dao.go`
- 将 `INSERT INTO alert_issues` 的列名由旧版驼峰改为蛇形：
  - `alertState` → `alert_state`
  - `label` → `labels`
  - `alertSince` → `alert_since`

2) `dao_integration_test.go`
- 初始化表结构同步到最新设计：
  - `id varchar(64)`（原为 varchar(255)）
  - 列名改为 `alert_state`、`labels`、`alert_since`；`alert_since` 使用 `timestamp(6)`。
  - 索引列同步：`(state, level, alert_since)` 与 `(alert_state, alert_since)`。

3) `README.md`
- 所有 SQL 示例与查询示例同步上述列名与类型调整，避免误导联调。
- 注明查询样例中选择列为 `alert_state` 与 `alert_since`。

## 变更原因

- 数据库设计文件已经更新为 7 张表版本，并统一命名规范为 snake_case；为避免字段不匹配导致插入失败，代码与文档需保持一致。
- API 文档已统一时间格式为 ISO 8601，数据库侧采用 `timestamp(6)` 存储精度，与接收端 `time.Time` 保持纳秒到毫秒的合理折衷。

## 验证

- 运行集成测试（需 Postgres 实例与 `-tags=integration`）可验证插入成功：
  - `go test ./internal/alerting/service/receiver -tags=integration -run TestPgDAO_InsertAlertIssue -v`
- 按 `README.md` 的 Docker 步骤进行端到端验证，可看到 `alert_issues` 正确落库且列名匹配。

## 兼容性

- 若已有旧表结构，需要执行迁移（ALTER TABLE 重命名列，或重建表）。本次改动不改变业务语义，仅为命名与精度统一。


