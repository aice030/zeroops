package service

import "errors"

// 业务错误定义
var (
	ErrDeploymentConflict = errors.New("deployment conflict: service version already in deployment")
	ErrServiceNotFound    = errors.New("service not found")
	ErrDeploymentNotFound = errors.New("deployment not found")
	ErrInvalidDeployState = errors.New("invalid deployment state")
)
