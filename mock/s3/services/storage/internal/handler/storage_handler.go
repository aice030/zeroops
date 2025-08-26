package handler

import (
	"mocks3/shared/interfaces"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"mocks3/shared/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StorageHandler Storage HTTP处理器
type StorageHandler struct {
	service interfaces.StorageService
	logger  *observability.Logger
}

// NewStorageHandler 创建Storage处理器
func NewStorageHandler(service interfaces.StorageService, logger *observability.Logger) *StorageHandler {
	return &StorageHandler{
		service: service,
		logger:  logger,
	}
}

// SetupRoutes 设置路由
func (h *StorageHandler) SetupRoutes(router *gin.Engine) {
	// 公共API路由组
	api := router.Group("/api/v1")
	{
		// 对象操作
		api.POST("/objects", h.CreateObject)
		api.GET("/objects/:bucket/:key", h.GetObject)
		api.PUT("/objects/:bucket/:key", h.UpdateObject)
		api.DELETE("/objects/:bucket/:key", h.DeleteObject)

		// 对象列表
		api.GET("/objects", h.ListObjects)

		// 统计信息
		api.GET("/stats", h.GetStats)
	}

	// 内部API路由组（供Queue Service等系统内部服务使用）
	internal := router.Group("/api/v1/internal")
	{
		// 仅操作存储层的内部接口
		internal.POST("/objects", h.WriteObjectToStorage)
		internal.DELETE("/objects/:bucket/:key", h.DeleteObjectFromStorage)
	}

	// 健康检查
	router.GET("/health", h.HealthCheck)
}

// CreateObject 创建对象 POST /api/v1/objects
func (h *StorageHandler) CreateObject(c *gin.Context) {
	var req models.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid upload request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 转换为Object模型
	object := &models.Object{
		Bucket:      req.Bucket,
		Key:         req.Key,
		Data:        req.Data,
		Size:        int64(len(req.Data)),
		ContentType: req.ContentType,
		Headers:     req.Headers,
		Tags:        req.Tags,
	}

	err := h.service.WriteObject(c.Request.Context(), object)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to write object", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.UploadResponse{
		Success:  true,
		ObjectID: object.ID,
		Key:      object.Key,
		Bucket:   object.Bucket,
		Size:     object.Size,
		MD5Hash:  object.MD5Hash,
		ETag:     object.MD5Hash,
		Message:  "Object uploaded successfully",
	}

	utils.SetJSONResponse(c.Writer, http.StatusCreated, response)
}

// GetObject 获取对象 GET /api/v1/objects/:bucket/:key
func (h *StorageHandler) GetObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	object, err := h.service.ReadObject(c.Request.Context(), bucket, key)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to read object",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusNotFound, err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", object.ContentType)
	c.Header("Content-Length", strconv.FormatInt(object.Size, 10))
	c.Header("Content-MD5", object.MD5Hash)
	c.Header("ETag", object.MD5Hash)

	// 设置自定义头部
	for k, v := range object.Headers {
		c.Header(k, v)
	}

	// 返回文件数据
	c.Data(http.StatusOK, object.ContentType, object.Data)
}

// UpdateObject 更新对象 PUT /api/v1/objects/:bucket/:key
func (h *StorageHandler) UpdateObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	var req models.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid upload request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 确保URL参数与请求体一致
	req.Bucket = bucket
	req.Key = key

	// 转换为Object模型
	object := &models.Object{
		Bucket:      req.Bucket,
		Key:         req.Key,
		Data:        req.Data,
		Size:        int64(len(req.Data)),
		ContentType: req.ContentType,
		Headers:     req.Headers,
		Tags:        req.Tags,
	}

	err := h.service.WriteObject(c.Request.Context(), object)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to update object", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.UploadResponse{
		Success:  true,
		ObjectID: object.ID,
		Key:      object.Key,
		Bucket:   object.Bucket,
		Size:     object.Size,
		MD5Hash:  object.MD5Hash,
		ETag:     object.MD5Hash,
		Message:  "Object updated successfully",
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, response)
}

// DeleteObject 删除对象 DELETE /api/v1/objects/:bucket/:key
func (h *StorageHandler) DeleteObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	err := h.service.DeleteObject(c.Request.Context(), bucket, key)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to delete object",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListObjects 列出对象 GET /api/v1/objects
func (h *StorageHandler) ListObjects(c *gin.Context) {
	var req models.ListObjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid list request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if req.Bucket == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket parameter is required")
		return
	}

	// 设置默认值
	if req.MaxKeys <= 0 {
		req.MaxKeys = 1000
	}

	response, err := h.service.ListObjects(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to list objects",
			observability.String("bucket", req.Bucket),
			observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, response)
}

// GetStats 获取统计信息 GET /api/v1/stats
func (h *StorageHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to get stats", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, stats)
}

// HealthCheck 健康检查 GET /health
func (h *StorageHandler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Health check failed", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusServiceUnavailable, "Service unhealthy")
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, map[string]any{
		"status":    "healthy",
		"service":   "storage-service",
		"timestamp": "now",
	})
}

// 内部API处理器

// WriteObjectToStorage 仅写入到存储节点 POST /api/v1/internal/objects
func (h *StorageHandler) WriteObjectToStorage(c *gin.Context) {
	var req models.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid internal upload request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 转换为Object模型
	object := &models.Object{
		Bucket:      req.Bucket,
		Key:         req.Key,
		Data:        req.Data,
		Size:        int64(len(req.Data)),
		ContentType: req.ContentType,
		Headers:     req.Headers,
		Tags:        req.Tags,
	}

	err := h.service.WriteObjectToStorage(c.Request.Context(), object)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to write object to storage (internal)", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.UploadResponse{
		Success:  true,
		ObjectID: object.ID,
		Key:      object.Key,
		Bucket:   object.Bucket,
		Size:     object.Size,
		MD5Hash:  object.MD5Hash,
		ETag:     object.MD5Hash,
		Message:  "Object written to storage successfully (internal)",
	}

	utils.SetJSONResponse(c.Writer, http.StatusCreated, response)
}

// DeleteObjectFromStorage 仅从存储节点删除文件 DELETE /api/v1/internal/objects/:bucket/:key
func (h *StorageHandler) DeleteObjectFromStorage(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	err := h.service.DeleteObjectFromStorage(c.Request.Context(), bucket, key)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to delete object from storage (internal)",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
