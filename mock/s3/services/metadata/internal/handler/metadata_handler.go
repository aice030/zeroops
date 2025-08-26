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

// MetadataHandler Metadata HTTP处理器
type MetadataHandler struct {
	service interfaces.MetadataService
	logger  *observability.Logger
}

// NewMetadataHandler 创建处理器
func NewMetadataHandler(service interfaces.MetadataService, logger *observability.Logger) *MetadataHandler {
	return &MetadataHandler{
		service: service,
		logger:  logger,
	}
}

// SetupRoutes 设置路由
func (h *MetadataHandler) SetupRoutes(router *gin.Engine) {
	// API路由组
	api := router.Group("/api/v1")
	{
		// 元数据CRUD
		api.POST("/metadata", h.SaveMetadata)
		api.GET("/metadata/:bucket/:key", h.GetMetadata)
		api.PUT("/metadata/:bucket/:key", h.UpdateMetadata)
		api.DELETE("/metadata/:bucket/:key", h.DeleteMetadata)

		// 列表和搜索
		api.GET("/metadata", h.ListMetadata)
		api.GET("/metadata/search", h.SearchMetadata)

		// 统计信息
		api.GET("/stats", h.GetStats)
	}

	// 健康检查
	router.GET("/health", h.HealthCheck)
}

// SaveMetadata 保存元数据 POST /api/v1/metadata
func (h *MetadataHandler) SaveMetadata(c *gin.Context) {
	var metadata models.Metadata
	if err := c.ShouldBindJSON(&metadata); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid request body", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.SaveMetadata(c.Request.Context(), &metadata)
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Metadata saved successfully",
	})
}

// GetMetadata 获取元数据 GET /api/v1/metadata/:bucket/:key
func (h *MetadataHandler) GetMetadata(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	metadata, err := h.service.GetMetadata(c.Request.Context(), bucket, key)
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusNotFound, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, metadata)
}

// UpdateMetadata 更新元数据 PUT /api/v1/metadata/:bucket/:key
func (h *MetadataHandler) UpdateMetadata(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	var metadata models.Metadata
	if err := c.ShouldBindJSON(&metadata); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid request body", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 确保URL中的bucket和key与请求体一致
	metadata.Bucket = bucket
	metadata.Key = key

	err := h.service.UpdateMetadata(c.Request.Context(), &metadata)
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, map[string]any{
		"success": true,
		"message": "Metadata updated successfully",
	})
}

// DeleteMetadata 删除元数据 DELETE /api/v1/metadata/:bucket/:key
func (h *MetadataHandler) DeleteMetadata(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket and key are required")
		return
	}

	err := h.service.DeleteMetadata(c.Request.Context(), bucket, key)
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMetadata 列出元数据 GET /api/v1/metadata
func (h *MetadataHandler) ListMetadata(c *gin.Context) {
	bucket := c.Query("bucket")
	prefix := c.Query("prefix")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if bucket == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "bucket parameter is required")
		return
	}

	metadataList, err := h.service.ListMetadata(c.Request.Context(), bucket, prefix, limit, offset)
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, map[string]any{
		"metadata": metadataList,
		"count":    len(metadataList),
		"bucket":   bucket,
		"prefix":   prefix,
		"limit":    limit,
		"offset":   offset,
	})
}

// SearchMetadata 搜索元数据 GET /api/v1/metadata/search
func (h *MetadataHandler) SearchMetadata(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	metadataList, err := h.service.SearchMetadata(c.Request.Context(), query, limit)
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, map[string]any{
		"query":    query,
		"metadata": metadataList,
		"count":    len(metadataList),
		"limit":    limit,
	})
}

// GetStats 获取统计信息 GET /api/v1/stats
func (h *MetadataHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, stats)
}

// HealthCheck 健康检查 GET /health
func (h *MetadataHandler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		utils.SetErrorResponse(c.Writer, http.StatusServiceUnavailable, "Service unhealthy")
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, map[string]any{
		"status":    "healthy",
		"service":   "metadata-service",
		"timestamp": "now",
	})
}
