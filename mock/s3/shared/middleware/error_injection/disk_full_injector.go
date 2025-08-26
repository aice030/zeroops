package error_injection

import (
	"context"
	"fmt"
	"mocks3/shared/observability"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DiskFullInjector 磁盘满载异常注入器
type DiskFullInjector struct {
	logger    *observability.Logger
	isActive  bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	tempFiles []string
	targetGB  int64
	currentGB int64
	tempDir   string
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
func (d *DiskFullInjector) StartDiskFull(ctx context.Context, targetDiskGB int64, duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isActive {
		d.logger.Warn(ctx, "Disk full injection already active")
		return
	}

	d.isActive = true
	d.targetGB = targetDiskGB
	d.logger.Info(ctx, "Starting disk full injection",
		observability.Int64("target_disk_gb", targetDiskGB),
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
		observability.Int64("created_files_gb", d.currentGB))
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
	d.currentGB = 0

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

// GetCurrentDiskGB 获取当前占用的磁盘空间（GB）
func (d *DiskFullInjector) GetCurrentDiskGB() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentGB
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

			// 检查是否达到目标磁盘使用量
			if d.currentGB >= d.targetGB {
				d.mu.Unlock()
				continue
			}

			// 每次创建500MB文件
			fileSizeGB := int64(1) // 1GB per file
			if d.currentGB+fileSizeGB > d.targetGB {
				fileSizeGB = d.targetGB - d.currentGB
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
			d.currentGB += fileSizeGB

			d.logger.Info(ctx, "Disk file created",
				observability.String("filename", filename),
				observability.Int64("file_size_gb", fileSizeGB),
				observability.Int64("total_disk_usage_gb", d.currentGB),
				observability.Int64("target_gb", d.targetGB))

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

// Cleanup 清理资源
func (d *DiskFullInjector) Cleanup() {
	close(d.stopChan)
	d.StopDiskFull(context.Background())
}
