package impl

import (
	"shared/faults"
	"shared/faults/memory"
	"storage-service/internal/service"
	"time"
)

type faultServiceImpl struct {
	manager *faults.FaultManager
}

func NewFaultServiceImpl() service.FaultService {
	m := faults.NewFaultManager()
	// 注册所有故障实例
	m.Register(memory.NewMemLeakFault(100*1024*1024, 1000*time.Millisecond))
	return &faultServiceImpl{
		manager: m,
	}
}

func (f *faultServiceImpl) StartFault(name string) error {
	return f.manager.Start(name)
}

func (f *faultServiceImpl) StopFault(name string) error {
	return f.manager.Stop(name)
}

func (f *faultServiceImpl) GetFaultStatus(name string) (string, error) {
	return f.manager.Status(name)
}

func (f *faultServiceImpl) ListFaults() ([]string, error) {
	return f.manager.List()
}
