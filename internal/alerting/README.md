监控告警模块下也可拆分为多个文件夹：
alerting/
├── receiver/       # 告警接收与处理
├── rules/          # 告警规则调整
├── metadata/       # 监控与告警元数据
├── healthcheck/    # 周期体检任务
├── severity/       # 告警等级计算
└── remediation/    # 自动化治愈行为