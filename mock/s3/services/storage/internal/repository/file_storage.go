package repository

import (
	"context"
	"fmt"
	"io"
	"mocks3/shared/interfaces"
	"mocks3/shared/observability"
	"os"
	"path/filepath"
	"sync"
)

// FileStorageRepository 文件存储仓库实现
type FileStorageRepository struct {
	nodes  []NodeInfo
	logger *observability.Logger
	mu     sync.RWMutex
}

// NodeInfo 存储节点信息
type NodeInfo struct {
	ID   string
	Path string
}

// NewFileStorageRepository 创建文件存储仓库
func NewFileStorageRepository(nodes []NodeInfo, logger *observability.Logger) (*FileStorageRepository, error) {
	repo := &FileStorageRepository{
		nodes:  nodes,
		logger: logger,
	}

	// 初始化存储节点目录
	if err := repo.initializeNodes(); err != nil {
		return nil, fmt.Errorf("initialize nodes: %w", err)
	}

	return repo, nil
}

// initializeNodes 初始化存储节点目录
func (r *FileStorageRepository) initializeNodes() error {
	for _, node := range r.nodes {
		if err := os.MkdirAll(node.Path, 0755); err != nil {
			return fmt.Errorf("create node directory %s: %w", node.Path, err)
		}
		r.logger.Info(context.Background(), "Storage node initialized",
			observability.String("node_id", node.ID),
			observability.String("path", node.Path))
	}
	return nil
}

// WriteObject 写入对象到所有节点
func (r *FileStorageRepository) WriteObject(ctx context.Context, bucket, key string, data io.Reader, size int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 读取数据到内存
	content, err := io.ReadAll(data)
	if err != nil {
		r.logger.Error(ctx, "Failed to read object data", observability.Error(err))
		return fmt.Errorf("read data: %w", err)
	}

	if int64(len(content)) != size {
		r.logger.Warn(ctx, "Object size mismatch",
			observability.Int64("expected", size),
			observability.Int("actual", len(content)))
	}

	// 写入到所有节点
	var writeErrors []error
	successCount := 0

	for _, node := range r.nodes {
		if err := r.writeToNode(ctx, node, bucket, key, content); err != nil {
			r.logger.Error(ctx, "Failed to write to node",
				observability.String("node_id", node.ID),
				observability.String("bucket", bucket),
				observability.String("key", key),
				observability.Error(err))
			writeErrors = append(writeErrors, fmt.Errorf("node %s: %w", node.ID, err))
		} else {
			successCount++
		}
	}

	// 至少要有一个节点写入成功
	if successCount == 0 {
		return fmt.Errorf("failed to write to any node: %v", writeErrors)
	}

	r.logger.Info(ctx, "Object written to storage nodes",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.Int("success_nodes", successCount),
		observability.Int("total_nodes", len(r.nodes)))

	return nil
}

// writeToNode 写入数据到指定节点
func (r *FileStorageRepository) writeToNode(_ context.Context, node NodeInfo, bucket, key string, data []byte) error {
	objectPath := r.getObjectPath(node.Path, bucket, key)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(objectPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// 写入文件
	file, err := os.Create(objectPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

// ReadObject 从节点读取对象
func (r *FileStorageRepository) ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 尝试从各个节点读取，返回第一个成功的
	for _, node := range r.nodes {
		objectPath := r.getObjectPath(node.Path, bucket, key)

		file, err := os.Open(objectPath)
		if err != nil {
			r.logger.Debug(ctx, "Failed to read from node",
				observability.String("node_id", node.ID),
				observability.String("path", objectPath),
				observability.Error(err))
			continue
		}

		stat, err := file.Stat()
		if err != nil {
			file.Close()
			continue
		}

		r.logger.Info(ctx, "Object read from storage node",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.String("node_id", node.ID))

		return file, stat.Size(), nil
	}

	return nil, 0, fmt.Errorf("object not found in any node: %s/%s", bucket, key)
}

// DeleteObject 从所有节点删除对象
func (r *FileStorageRepository) DeleteObject(ctx context.Context, bucket, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var deleteErrors []error
	successCount := 0

	for _, node := range r.nodes {
		objectPath := r.getObjectPath(node.Path, bucket, key)

		if err := os.Remove(objectPath); err != nil {
			if !os.IsNotExist(err) {
				r.logger.Error(ctx, "Failed to delete from node",
					observability.String("node_id", node.ID),
					observability.String("path", objectPath),
					observability.Error(err))
				deleteErrors = append(deleteErrors, fmt.Errorf("node %s: %w", node.ID, err))
			}
		} else {
			successCount++
		}
	}

	r.logger.Info(ctx, "Object deletion attempted",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.Int("success_nodes", successCount),
		observability.Int("total_nodes", len(r.nodes)))

	// 如果所有节点都报错（且不是文件不存在），返回错误
	if len(deleteErrors) == len(r.nodes) {
		return fmt.Errorf("failed to delete from all nodes: %v", deleteErrors)
	}

	return nil
}

// ObjectExists 检查对象是否存在
func (r *FileStorageRepository) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查是否至少有一个节点存在该对象
	for _, node := range r.nodes {
		objectPath := r.getObjectPath(node.Path, bucket, key)
		if _, err := os.Stat(objectPath); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// GetNodeStats 获取节点统计信息
func (r *FileStorageRepository) GetNodeStats(ctx context.Context) ([]interfaces.NodeStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make([]interfaces.NodeStatus, len(r.nodes))

	for i, node := range r.nodes {
		stat, err := r.getNodeStat(node)
		if err != nil {
			r.logger.Error(ctx, "Failed to get node stats",
				observability.String("node_id", node.ID),
				observability.Error(err))
			stats[i] = interfaces.NodeStatus{
				ID:     node.ID,
				Status: "error",
				Error:  err.Error(),
			}
		} else {
			stats[i] = *stat
		}
	}

	return stats, nil
}

// getNodeStat 获取节点状态
func (r *FileStorageRepository) getNodeStat(node NodeInfo) (*interfaces.NodeStatus, error) {
	// 检查目录是否可访问
	info, err := os.Stat(node.Path)
	if err != nil {
		return &interfaces.NodeStatus{
			ID:     node.ID,
			Status: "unreachable",
			Error:  err.Error(),
		}, nil
	}

	if !info.IsDir() {
		return &interfaces.NodeStatus{
			ID:     node.ID,
			Status: "error",
			Error:  "path is not a directory",
		}, nil
	}

	// 计算目录实际使用空间
	usedSpace, err := r.calculateDirectorySize(node.Path)
	if err != nil {
		r.logger.Warn(context.Background(), "Failed to calculate directory size",
			observability.String("node_id", node.ID),
			observability.Error(err))
		usedSpace = 0
	}

	return &interfaces.NodeStatus{
		ID:        node.ID,
		Status:    "healthy",
		UsedSpace: usedSpace,
	}, nil
}

// calculateDirectorySize 计算目录总大小
func (r *FileStorageRepository) calculateDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// getObjectPath 构建对象在节点中的完整路径
func (r *FileStorageRepository) getObjectPath(nodePath, bucket, key string) string {
	return filepath.Join(nodePath, bucket, key)
}
