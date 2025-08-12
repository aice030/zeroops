package service

import (
	"context"
	"io"
)

// FileInfo 文件信息结构体
type FileInfo struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// StorageService 存储服务接口
// 这个接口定义了存储服务的基本操作，支持多种存储后端
type StorageService interface {
	// UploadFile 上传文件
	// fileID: 文件唯一标识符
	// fileName: 文件名
	// contentType: 文件类型
	// reader: 文件内容读取器
	// 返回文件信息和错误
	UploadFile(ctx context.Context, fileID, fileName, contentType string, reader io.Reader) (*FileInfo, error)

	// DownloadFile 下载文件
	// fileID: 文件唯一标识符
	// 返回文件内容读取器和文件信息
	DownloadFile(ctx context.Context, fileID string) (io.Reader, *FileInfo, error)

	// DeleteFile 删除文件
	// fileID: 文件唯一标识符
	// 返回删除是否成功和错误
	DeleteFile(ctx context.Context, fileID string) error

	// GetFileInfo 获取文件信息
	// fileID: 文件唯一标识符
	// 返回文件信息
	GetFileInfo(ctx context.Context, fileID string) (*FileInfo, error)

	// ListFiles 列出所有文件
	// 返回文件信息列表
	ListFiles(ctx context.Context) ([]*FileInfo, error)

	// Close 关闭存储连接
	// 用于清理资源
	Close() error
}
