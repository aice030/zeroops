package faults

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// FaultManager 统一管理所有故障实例
type FaultManager struct {
	mu       sync.RWMutex
	faultMap map[string]Fault
	rnd      *rand.Rand // 随机数生成器，用于概率注入
}

func NewFaultManager() *FaultManager {
	return &FaultManager{
		faultMap: make(map[string]Fault),
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Register 注册故障
func (m *FaultManager) Register(f Fault) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.faultMap[f.Name()] = f
}

// Start 启动指定故障
func (m *FaultManager) Start(name string) error {
	m.mu.RLock()
	f, ok := m.faultMap[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("fault %s not found", name)
	}
	return f.Start()
}

// Stop 停止指定故障
func (m *FaultManager) Stop(name string) error {
	m.mu.RLock()
	f, ok := m.faultMap[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("fault %s not found", name)
	}
	return f.Stop()
}

// Status 查询故障状态
func (m *FaultManager) Status(name string) (string, error) {
	m.mu.RLock()
	f, ok := m.faultMap[name]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("fault %s not found", name)
	}
	return f.Status(), nil
}

// List 返回所有故障名
func (m *FaultManager) List() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.faultMap))
	for k := range m.faultMap {
		keys = append(keys, k)
	}
	return keys, nil
}

// ShouldInject 根据概率判断是否应该注入故障
func (m *FaultManager) ShouldInject(rate float64) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 1.0 {
		return true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rnd.Float64() <= rate
}
