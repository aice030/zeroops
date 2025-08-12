package handler

import (
	"net/http"
	"strings"
)

// Router 路由处理器
type Router struct {
	fileHandler *FileHandler
}

// NewRouter 创建路由处理器
func NewRouter(fileHandler *FileHandler) *Router {
	return &Router{
		fileHandler: fileHandler,
	}
}

// ServeHTTP 实现http.Handler接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理预检请求
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 路由分发
	path := req.URL.Path
	method := req.Method

	switch {
	// 健康检查
	case path == "/api/health" && method == "GET":
		r.fileHandler.HealthCheck(w, req)

	// 文件上传
	case path == "/api/files/upload" && method == "POST":
		r.fileHandler.UploadFile(w, req)

	// 文件下载
	case strings.HasPrefix(path, "/api/files/download/") && method == "GET":
		r.fileHandler.DownloadFile(w, req)

	// 文件删除
	case strings.HasPrefix(path, "/api/files/") && strings.Count(path, "/") == 3 && method == "DELETE":
		r.fileHandler.DeleteFile(w, req)

	// 获取文件信息
	case strings.HasSuffix(path, "/info") && method == "GET":
		r.fileHandler.GetFileInfo(w, req)

	// 文件列表
	case path == "/api/files" && method == "GET":
		r.fileHandler.ListFiles(w, req)

	// 默认路由
	default:
		http.NotFound(w, req)
	}
}
