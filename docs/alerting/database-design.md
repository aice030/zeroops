# 数据库设计 - Monitoring & Alerting Service

## 概述

本文档为最新数据库设计，总计包含 7 张表：

- alert_issues
- alert_issue_comments
- metric_alert_changes
- alert_rules
- service_alert_metas
- service_metrics
- service_states

## 数据表设计

### 1) alert_issues（告警问题表）

存储告警问题的主要信息。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | varchar(64) PK | 告警 issue ID |
| state | enum(Closed, Open) | 问题状态 |
| level | varchar(32) | 告警等级：如 P0/P1/Px/Warning |
| alert_state | enum(Pending, Restored, AutoRestored, InProcessing) | 处理状态 |
| title | varchar(255) | 告警标题 |
| labels | json | 标签，格式：[{key, value}] |
| alert_since | TIMESTAMP(6) | 告警首次发生时间 |

**索引建议：**
- PRIMARY KEY: `id`
- INDEX: `(state, level, alert_since)`
- INDEX: `(alert_state, alert_since)`

---

### 2) alert_issue_comments（告警评论/处理记录表）

记录 AI/系统/人工在处理告警过程中的动作与备注。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| issue_id | varchar(64) FK | 对应 `alert_issues.id` |
| create_at | TIMESTAMP(6) | 评论创建时间 |
| content | text | Markdown 内容 |

**索引建议：**
- PRIMARY KEY: `(issue_id, create_at)`
- FOREIGN KEY: `issue_id` REFERENCES `alert_issues(id)`

---

### 3) metric_alert_changes（指标告警规则变更记录表）

用于追踪指标类告警规则或参数的变更历史。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | varchar(64) PK | 变更记录 ID |
| change_time | TIMESTAMP(6) | 变更时间 |
| alert_name | varchar(255) | 告警名称/规则名 |
| change_items | json | 变更项数组：[{key, old_value, new_value}] |

**索引建议：**
- PRIMARY KEY: `id`
- INDEX: `(change_time)`
- INDEX: `(alert_name, change_time)`

---

### 4) alert_rules（告警规则表）

定义可复用的规则表达式，支持作用域绑定。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | varchar(255) PK | 规则 ID（可与 K8s 资源 ID 对应或做映射） |
| name | varchar(255) | 规则名称，表达式可读的名称 |
| scopes | varchar(255) | 作用域，例："services:svc1,svc2" |
| expr | text | 规则表达式（可含占位符） |

**索引建议：**
- PRIMARY KEY: `id`
- INDEX: `(name)`
- INDEX: `(scopes)`

---

### 5) service_alert_metas（服务告警元数据表）

按服务维度存放参数化配置，用于渲染具体规则。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| service | varchar(255) | 服务名 |
| key | varchar(255) | 参数名（如 `apitime_threshold`） |
| value | varchar(255) | 参数值（如 `50`） |

**索引建议：**
- PRIMARY KEY: `(service, key)`
- INDEX: `(service)`

---

### 6) service_metrics（服务指标清单表）

记录服务所关注的指标清单（可用于 UI 侧展示或校验）。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| service | varchar(255) PK | 服务名 |
| metrics | json | 指标名数组：["metric1", "metric2", ...] |

**索引建议：**
- PRIMARY KEY: `service`

---

### 7) service_states（服务异常状态表）

追踪服务在某一版本上的健康状态与处置进度。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| service | varchar(255) PK | 服务名 |
| version | varchar(255) PK | 版本号 |
<!-- | level | varchar(32) | 影响等级：如 P0/P1/Px/Warning  | -->
| detail | text | 异常详情（可为 JSON 文本）（可空） |
| report_at | TIMESTAMP(6) | 首次报告时间 |
| resolved_at | TIMESTAMP(6) | 解决时间（可空） |
| health_state | enum(Normal,Processing,Error) | 处置阶段 |
| correlation_id | varchar(255) | 关联 ID（用于跨系统联动/串联事件）（可空） |

**索引建议：**
- PRIMARY KEY: `(service, version)`
- INDEX: `(health_state, report_at)`
- INDEX: `(correlation_id)`

## 数据关系（ER）

```mermaid
erDiagram
    alert_issues ||--o{ alert_issue_comments : "has comments"

    alert_rules {
        varchar id PK
        varchar name
        varchar scopes
        text expr
    }

    service_alert_metas {
        varchar service PK
        varchar key PK
        varchar value
    }

    service_metrics {
        varchar service PK
        json metrics
    }

    service_states {
        varchar service PK
        varchar version PK
        <!-- enum level -->
        text detail
        timestamp report_at
        timestamp resolved_at
        varchar health_state
        varchar correlation_id
    }

    alert_issues {
        varchar id PK
        enum state
        varchar level
        enum alert_state
        varchar title
        json labels
        timestamp alert_since
    }

    alert_issue_comments {
        varchar issue_id FK
        timestamp create_at
        text content
    }

    %% 通过 service 逻辑关联
    service_alert_metas ||..|| service_metrics : "by service"
    service_states ||..|| service_alert_metas : "by service"
```

## 数据流转

1. 以 `alert_rules` 为模版，结合 `service_alert_metas` 渲染出面向具体服务的规则。
2. 指标或规则参数发生调整时，记录到 `metric_alert_changes`。
3. 规则触发创建 `alert_issues`；处理过程中的动作写入 `alert_issue_comments`。
4. 面向服务的整体健康态以 `service_states` 记录和推进（new → analyzing → processing → resolved）。