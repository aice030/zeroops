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
					"batch_id": 1,
					"nodes": []map[string]interface{}{
						{"node_id": "bj1-node-001", "host_id": "127.0.0.1"},
					},
				},
				{
					"batch_id": 2,
					"nodes": []map[string]interface{}{
						{"node_id": "sh1-node-001", "host_id": "192.168.1.11"},
						{"node_id": "sh2-node-001", "host_id": "192.168.2.11"},
					},
				},
				{
					"batch_id": 3,
					"nodes": []map[string]interface{}{
						{"node_id": "sh1-node-002", "host_id": "192.168.1.12"},
						{"node_id": "sh1-node-003", "host_id": "192.168.1.13"},
						{"node_id": "sh2-node-002", "host_id": "192.168.2.12"},
						{"node_id": "sh2-node-003", "host_id": "192.168.2.13"},
					},
				},
			},
		},
	}
	return models.NewStepResult("ReleasePlan", "2 batches created", details)
}
