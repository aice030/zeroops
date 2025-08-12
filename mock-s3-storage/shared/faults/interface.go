package faults

type Fault interface {
	// Name 故障名称
	Name() string

	// Start 启动故障
	Start() error

	// Stop 停止故障
	Stop() error

	// Status 故障状态
	Status() string
}
