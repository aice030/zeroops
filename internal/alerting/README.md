监控/告警处理模块（Alerting）

目标
- 统一接收、聚合与去重来自 Prometheus/ES/第三方的告警事件
- 将事件归并为告警问题（Issue），支持生命周期/状态机管理
- 结合服务元数据计算影响面，产出最终告警等级（P0/P1/P2/Warning）
- 支持自动/半自动治愈与回填评论，形成可追溯处置记录
- 提供查询、检索与统计 API，服务控制台与自动化流程

模块边界
- 输入：告警事件流（Webhook/轮询）、指标查询、服务元数据
- 输出：告警问题（Issue）数据、评论记录、通知、治愈执行
- 不做：指标采集、底层存储运维（交由公共组件）

目录结构（Hexagonal/Clean）
internal/alerting/
- domain/                 领域模型与端口（接口）
  - types.go              Level/Issue/Comment/Event 等
  - ports.go              IssueRepository/RuleCalculator/Notifier/Healer 等
- usecase/                应用服务（聚合、状态机、等级计算、治愈编排）
  - service.go            New(repo, rules, notifiers, healers)
- adapter/                适配器：传输/存储/规则/通知/治愈/接入
  - httpapi/              HTTP 路由与 DTO
  - repository/memory/    内存仓储（示例）
  - rules/default/        默认等级计算器
  - notifier/feishu/      飞书通知（示例）
  - ingest/prometheus/    Prometheus 事件接入（示例）
- api/                    便捷装配（示例项目直接调用）
- scheduler/              （可选）体检/巡检定时任务
- README.md               本文档

数据模型（MySQL）
1) alert_issues（告警问题表）
- 主键：id（字符串或自增，推荐字符串以便跨源唯一）
- 字段：
- state：问题状态（Open/Closed）
- level：告警等级（P0/P1/P2/Warning）
- alertState：处理状态（InProcessing/AutoRestored/Restored）
- title：标题
- label：JSON（单一标签对象 {key,value}）
- alertSince：DATETIME（首次告警时间）
- json：JSON 扩展（原始维度/链接等）

2) alert_issue_comments（告警问题评论表）
- 主键：无单独主键，按业务以 issueID+createAt 唯一（或可加自增列）
- issueID：外键关联 alert_issues.id
- createAt：DATETIME
- content：TEXT（Markdown，记录AI/系统/人工动作）

建表示例
```sql
CREATE TABLE alert_issues (
  id VARCHAR(64) PRIMARY KEY,
  state VARCHAR(16) NOT NULL,
  level VARCHAR(16) NOT NULL,
  alertState VARCHAR(32) NOT NULL,
  title VARCHAR(255) NOT NULL,
  label JSON NULL,
  alertSince DATETIME(3) NOT NULL,
  json JSON NULL,
  KEY idx_state_level (state, level),
  KEY idx_alert_since (alertSince)
);

CREATE TABLE alert_issue_comments (
  issueID VARCHAR(64) NOT NULL,
  createAt DATETIME(3) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  KEY idx_issue (issueID),
  CONSTRAINT fk_issue FOREIGN KEY (issueID) REFERENCES alert_issues(id)
);
```

状态机
- Issue.state：Open → Closed（单向闭环）
- Issue.alertState：
- InProcessing（处理中）
- AutoRestored（自然恢复）
- Restored（已恢复）

告警等级计算
- 输入：原始告警等级（来自源头）+ 服务影响面（流量、租户数、区域、核心度）
- 输出：最终 level（P0/P1/P2/Warning）
- 计算器放置于 `rules/`，通过接口可热插拔与单元测试

API 接口
1) 列表
GET /v1/issues?start=xxx&limit=10&state=Closed|Open
响应：
{
    "items": [ 
        {
            "id": "xxx", // 告警 issue ID
            "state": "Closed", // 告警条目的状态。Closed处理完成、Open处理中
            "level": "P0", // 枚举值：P0严重、P1重要、P2、Warning需要关注但不是线上异常
            "alertState": "Restored", // 告警处理状态。Restored 已恢复、AutoRestored 系统自动恢复、InProcessing 处理中
            "title": "yzh S3APIV2s3apiv2.putobject 0_64K上传响应时间95值:50012ms > 450ms", // 告警标题
            "labels": [
                {
                    "key": "api",
                    "value: "s3apiv2.putobject"
                },
                {
                    "key": "idc",
                    "value": "yzh"
                }
            ],
            "alertSince": "2025-05-05 11:00:00.0000Z"
        }
    ]
}

2) 详情
GET /v1/issues/:issueID
响应：
{
  "id": "xxx",
  "state": "Closed",
  "level": "P0",
  "alertState": "Restored",
  "title": "...",
  "label": {"key":"api","value":"s3apiv2.putobject"},
  "alertSince": "2025-05-05T11:00:00Z",
  "json": {"k":"v"},
  "comments": [
    {"issueID": "xxx", "createAt": "2024-01-03T03:00:00Z", "content": "markdown content"}
  ]
}

4) 手动关闭（标记恢复正常）
POST /v1/issues/:issueID/close
响应：200


摄入（Ingress）
- Prometheus Webhook：/v1/ingest/prometheus
- Elastic/Logs：定制 handler 于 `ingest/`
- 每个接入负责标准化为内部 Event，交由 service 层聚合

治愈（Healing）
- `healing/` 定义动作（如重启、扩容、清缓存），由编排器串联
- 执行结果写入 `alert_issue_comments`，并可更新 `alertState`

通知（Notifier）
- 在 state 变化或等级升级时触发
- 通过 `notifier/` 适配钉钉/飞书/邮件，支持静默窗口与去重

定时体检（Scheduler）
- 周期巡检 SLO/关键链路，将异常转化为 Issue 流入统一闭环

安全与审计
- API 走网关鉴权；重要动作（关闭/忽略/治愈）记录评论与审计日志

代码组织建议（Go）
- domain：领域模型与端口接口；无外部依赖
- usecase：只依赖 domain；通过端口调用适配器
- adapter：实现端口；彼此解耦，可替换
- api：http 仅做 DTO/编排，依赖 usecase

测试建议
- rules：基于样例数据的表驱动测试
- service：状态机与聚合流程的单元测试
- api：handler 层的端到端（通过 fake service）