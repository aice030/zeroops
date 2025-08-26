package models

import (
	"time"
)

// StepResult 定义通用步骤结果结构
type StepResult struct {
	StepName  string                 `json:"step_name"`
	Status    string                 `json:"status"` // success / failed / warning
	Summary   string                 `json:"summary"`
	Details   map[string]interface{} `json:"details"`
	Timestamp string                 `json:"timestamp"`
}

// NewStepResult 工具函数：创建成功结果
func NewStepResult(step string, summary string, details map[string]interface{}) StepResult {
	return StepResult{
		StepName:  step,
		Status:    "success",
		Summary:   summary,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
