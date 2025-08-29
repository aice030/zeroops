package planner

import (
	"qiniu1024-mcp-server/pkg/models"
)

func ReleasePlan(service string) models.StepResult {
	details := map[string]interface{}{
		"service": service,
		"gray_strategy": map[string]interface{}{
			"batches": []map[string]interface{}{
				{
					"batch_id":         1,
					"observation_time": "600s", // 第一批次：单主机，快速验证，观察时间较短
					"hosts": []map[string]interface{}{
						{"host_id": "bj1-node-001", "host_ip": "127.0.0.1"},
					},
				},
				{
					"batch_id":         2,
					"observation_time": "1800s", // 第二批次：双主机，稳定性验证，观察时间中等
					"hosts": []map[string]interface{}{
						{"host_id": "sh1-node-001", "host_ip": "192.168.1.11"},
						{"host_id": "sh2-node-001", "host_ip": "192.168.2.11"},
					},
				},
				{
					"batch_id":         3,
					"observation_time": "3600s", // 第三批次：多主机，全面验证，观察时间较长
					"hosts": []map[string]interface{}{
						{"host_id": "sh1-node-002", "host_ip": "192.168.1.12"},
						{"host_id": "sh1-node-003", "host_ip": "192.168.1.13"},
						{"host_id": "sh2-node-002", "host_ip": "192.168.2.12"},
						{"host_id": "sh2-node-003", "host_ip": "192.168.2.13"},
					},
				},
			},
		},
	}
	return models.NewStepResult("ReleasePlan", "3 batches created with observation times", details)
}
