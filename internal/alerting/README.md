# Monitoring & Alerting Service

## 概述

本模块提供 监控/告警处理能力，实现从告警接收到自动治愈的完整生命周期管理。

**核心思想：**
- 告警规则采用 模版 + 服务元数据 (Metadata) 的方式定义，便于统一管理和灵活扩展。
- 支持自动化处理、AI 辅助分析和多级告警分级。

---

## 目录结构

```
alerting/
├── api/            # API 接口层，提供 RESTful 访问
├── database/       # 数据库定义和 migration
├── model/          # 数据模型 (ORM / DTO)
├── service/        # 核心业务逻辑
│   ├── receiver/      # 告警接收与处理
│   ├── ruleset/       # 告警规则模版 + metadata
│   ├── healthcheck/   # 周期体检任务
│   ├── severity/      # 告警等级计算
│   └── remediation/   # 自动化治愈行为
└── README.md       # 项目说明文档
```


---

## 数据库设计

数据库采用4张核心表设计：告警问题表、评论表、模版表和服务元数据表。

详细的表结构设计、索引建议和性能优化方案请参考：**[数据库设计文档](../../docs/alerting/database-design.md)**


---

## 告警规则机制

告警规则由两部分组成：

### 1. 模版 (Template)

定义规则逻辑，带占位符：

```json
{
  "id": "tmpl_apitime",
  "expr": "apitime > {apitime_threshold}",
  "level": "P1"
}
```

### 2. 服务 Metadata (Service Config)

不同服务定义不同的阈值：

```json
{
  "serviceA": { "apitime_threshold": 100 },
  "serviceB": { "apitime_threshold": 50 }
}
```

### 3. 展开后的实际规则 (Resolved Rule)

模版 + metadata → 生成实际规则：

```json
{
  "service": "serviceA",
  "expr": "apitime > 100",
  "level": "P1"
}
```

```json
{
  "service": "serviceB",
  "expr": "apitime > 50",
  "level": "P1"
}
```


---

## 流程图

### 规则生成流程

```mermaid
flowchart TD
    TPL[告警模版<br/>(Templates)<br/>例: apitime > {apitime_threshold}] --> META
    META[服务 Metadata<br/>(Service Config)<br/>例: serviceA=100, serviceB=50] --> RESOLVED
    RESOLVED[实际告警规则<br/>(Resolved Rules)<br/>例: serviceA: apitime>100<br/>serviceB: apitime>50] --> ALERT
    ALERT[告警触发<br/>(Alert Trigger)<br/>生成 Issue & 进入处理流程]
```


---

## API 接口

服务提供 RESTful API 接口，支持告警列表查询、详情获取等核心功能。

主要接口包括：
- `GET /v1/issues` - 获取告警列表
- `GET /v1/issues/{issueID}` - 获取告警详情

完整的接口文档、请求参数、响应格式和使用示例请参考：**[API 文档](../../docs/alerting/api.md)**


---

## 模块功能说明

- **receiver/**：统一接收告警，写入数据库，触发处理流程
- **ruleset/**：管理告警模版 & 服务 metadata，生成实际规则
- **healthcheck/**：周期性体检，提前发现潜在问题
- **severity/**：计算告警等级 = 原始等级 + 影响范围
- **remediation/**：自动化治愈动作（回滚），并记录处理日志

## 相关文档

- [数据库设计文档](../../docs/alerting/database-design.md) - 详细的表结构和索引设计
- [API 文档](../../docs/alerting/api.md) - RESTful API 接口规范