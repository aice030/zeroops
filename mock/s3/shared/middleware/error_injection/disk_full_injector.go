package error_injection

import (
	"context"
	"fmt"
	"mocks3/shared/observability"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// DiskFullInjector 磁盘满载异常注入器
type DiskFullInjector struct {
	logger         *observability.Logger
	isActive       bool
	mu             sync.RWMutex
	stopChan       chan struct{}
	tempFiles      []string
	targetPercent  float64 // 目标磁盘使用率百分比
	baseUsage      float64 // 基础磁盘使用率
	estimatedTotal int64   // 估算的总容量
	tempDir        string
}

// NewDiskFullInjector 创建磁盘满载异常注入器
func NewDiskFullInjector(logger *observability.Logger, tempDir string) *DiskFullInjector {
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	return &DiskFullInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
		tempDir:  tempDir,
	}
}

// StartDiskFull 开始磁盘满载异常注入
func (d *DiskFullInjector) StartDiskFull(ctx context.Context, targetPercent float64, duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isActive {
		d.logger.Warn(ctx, "Disk full injection already active")
		return
	}

	d.isActive = true
	d.targetPercent = targetPercent
	// 估算容器总磁盘容量为10GB
	d.estimatedTotal = 10 * 1024 * 1024 * 1024
	// 获取当前基础使用率
	d.baseUsage = d.getCurrentDiskUsage()
	d.logger.Info(ctx, "Starting disk full injection",
		observability.Float64("target_percent", targetPercent),
		observability.Float64("base_usage", d.baseUsage),
		observability.String("duration", duration.String()),
		observability.String("temp_dir", d.tempDir))

	// 创建临时目录
	injectorDir := filepath.Join(d.tempDir, "metric_injector")
	if err := os.MkdirAll(injectorDir, 0755); err != nil {
		d.logger.Error(ctx, "Failed to create temp directory", observability.Error(err))
		d.isActive = false
		return
	}
	d.tempDir = injectorDir

	// 启动磁盘填充协程
	go d.diskFillTask(ctx)

	// 设置定时器自动停止
	go func() {
		select {
		case <-time.After(duration):
			d.StopDiskFull(ctx)
		case <-d.stopChan:
			return
		}
	}()
}

// StopDiskFull 停止磁盘满载异常注入
func (d *DiskFullInjector) StopDiskFull(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isActive {
		return
	}

	d.logger.Info(ctx, "Stopping disk full injection",
		observability.Float64("target_percent", d.targetPercent))
	d.isActive = false

	// 删除所有创建的临时文件
	for _, filename := range d.tempFiles {
		if err := os.Remove(filename); err != nil {
			d.logger.Warn(ctx, "Failed to remove temp file",
				observability.String("filename", filename),
				observability.Error(err))
		}
	}
	d.tempFiles = nil
	d.targetPercent = 0
	d.baseUsage = 0

	// 清理临时目录
	if err := os.RemoveAll(d.tempDir); err != nil {
		d.logger.Warn(ctx, "Failed to remove temp directory",
			observability.String("temp_dir", d.tempDir),
			observability.Error(err))
	}
}

// IsActive 检查磁盘满载注入是否活跃
func (d *DiskFullInjector) IsActive() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isActive
}

// GetCurrentDiskUsage 获取当前磁盘使用率百分比
func (d *DiskFullInjector) GetCurrentDiskUsage() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getCurrentDiskUsage()
}

// getCurrentDiskUsage 获取当前真实磁盘使用率
func (d *DiskFullInjector) getCurrentDiskUsage() float64 {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "/app"
	}

	// 使用syscall.Statfs获取真实的文件系统统计信息
	var stat syscall.Statfs_t
	if err := syscall.Statfs(workDir, &stat); err != nil {
		d.logger.Error(context.Background(), "Failed to get disk stats", observability.Error(err))
		return 0.0
	}

	// 计算磁盘使用率
	// stat.Blocks: 总块数
	// stat.Bavail: 可用块数
	// stat.Bsize: 块大小
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	usedBytes := totalBytes - availableBytes

	usagePercent := float64(usedBytes) / float64(totalBytes) * 100.0

	d.logger.Debug(context.Background(), "Real disk usage calculated",
		observability.String("total_bytes", fmt.Sprintf("%d", totalBytes)),
		observability.String("used_bytes", fmt.Sprintf("%d", usedBytes)),
		observability.String("available_bytes", fmt.Sprintf("%d", availableBytes)),
		observability.Float64("usage_percent", usagePercent))

	return usagePercent
}

// diskFillTask 磁盘填充任务
func (d *DiskFullInjector) diskFillTask(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second) // 每2秒创建一次文件
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			d.mu.Lock()
			if !d.isActive {
				d.mu.Unlock()
				return
			}

			// 检查当前使用率是否达到目标
			currentUsage := d.getCurrentDiskUsage()
			if currentUsage >= d.targetPercent {
				d.mu.Unlock()
				continue
			}

			// 计算需要创建的文件大小
			neededPercent := d.targetPercent - currentUsage
			fileSizeGB := int64(1)   // 每次创建1GB文件
			if neededPercent < 5.0 { // 如果差距小于5%，创建较小文件
				fileSizeGB = 1
			}

			filename := filepath.Join(d.tempDir, fmt.Sprintf("disk_fill_%d_%d.tmp",
				time.Now().UnixNano(), fileSizeGB))

			if err := d.createLargeFile(filename, fileSizeGB); err != nil {
				d.logger.Error(ctx, "Failed to create large file",
					observability.String("filename", filename),
					observability.Error(err))
				d.mu.Unlock()
				continue
			}

			d.tempFiles = append(d.tempFiles, filename)
			newUsage := d.getCurrentDiskUsage()

			d.logger.Info(ctx, "Disk file created",
				observability.String("filename", filename),
				observability.Int64("file_size_gb", fileSizeGB),
				observability.Float64("current_usage_percent", newUsage),
				observability.Float64("target_percent", d.targetPercent))

			d.mu.Unlock()
		}
	}
}

// createLargeFile 创建大文件
func (d *DiskFullInjector) createLargeFile(filename string, sizeGB int64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建1MB的数据块
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	// 写入数据直到达到目标大小
	totalMB := sizeGB * 1024
	for i := int64(0); i < totalMB; i++ {
		if _, err := file.Write(data); err != nil {
			return err
		}
	}

	return file.Sync()
}

// calculateDirectorySize 计算目录总大小
func (d *DiskFullInjector) calculateDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略无法访问的文件
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// Cleanup 清理资源
func (d *DiskFullInjector) Cleanup() {
	close(d.stopChan)
	d.StopDiskFull(context.Background())
}
