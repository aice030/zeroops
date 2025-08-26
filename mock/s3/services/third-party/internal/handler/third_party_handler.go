package handler

import (
	"mocks3/services/third-party/internal/service"
	"mocks3/shared/observability"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ThirdPartyHandler 第三方服务HTTP处理器
type ThirdPartyHandler struct {
	service *service.ThirdPartyService
	logger  *observability.Logger
}

// NewThirdPartyHandler 创建第三方服务处理器
func NewThirdPartyHandler(service *service.ThirdPartyService, logger *observability.Logger) *ThirdPartyHandler {
	return &ThirdPartyHandler{
		service: service,
		logger:  logger,
	}
}

// SetupRoutes 设置路由
func (h *ThirdPartyHandler) SetupRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", h.HealthCheck)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 获取对象
		api.GET("/objects/:bucket/:key", h.GetObject)

		// 统计信息
		api.GET("/stats", h.GetStats)
	}
}

// GetObject 获取对象
func (h *ThirdPartyHandler) GetObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	h.logger.Info(c.Request.Context(), "Handling get object request",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.String("client_ip", c.ClientIP()))

	// 调用服务层
	object, err := h.service.GetObject(c.Request.Context(), bucket, key)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to get object from third-party service",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "object not found",
			"message": err.Error(),
			"bucket":  bucket,
			"key":     key,
		})
		return
	}

	h.logger.Info(c.Request.Context(), "Successfully retrieved object from third-party service",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.Int64("size", object.Size),
		observability.String("content_type", object.ContentType))

	// 设置响应头
	c.Header("Content-Type", object.ContentType)
	c.Header("Content-Length", string(rune(object.Size)))
	c.Header("ETag", object.MD5Hash)
	c.Header("X-Third-Party-Source", "true")

	// 设置自定义头
	for key, value := range object.Headers {
		c.Header(key, value)
	}

	// 返回对象数据
	c.Data(http.StatusOK, object.ContentType, object.Data)
}

// GetStats 获取统计信息
func (h *ThirdPartyHandler) GetStats(c *gin.Context) {
	h.logger.Debug(c.Request.Context(), "Handling get stats request")

	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to get stats",
			observability.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get stats",
			"message": err.Error(),
		})
		return
	}

	h.logger.Debug(c.Request.Context(), "Successfully retrieved stats")
	c.JSON(http.StatusOK, stats)
}

// HealthCheck 健康检查
func (h *ThirdPartyHandler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Warn(c.Request.Context(), "Health check failed",
			observability.Error(err))

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "third-party-service",
	})
}
