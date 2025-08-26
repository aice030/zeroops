package handler

import (
	"mocks3/shared/interfaces"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"mocks3/shared/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// QueueHandler Queue HTTP处理器
type QueueHandler struct {
	service interfaces.QueueService
	logger  *observability.Logger
}

// NewQueueHandler 创建Queue处理器
func NewQueueHandler(service interfaces.QueueService, logger *observability.Logger) *QueueHandler {
	return &QueueHandler{
		service: service,
		logger:  logger,
	}
}

// SetupRoutes 设置路由
func (h *QueueHandler) SetupRoutes(router *gin.Engine) {
	// API路由组
	api := router.Group("/api/v1")
	{
		// 删除任务操作
		api.POST("/delete-tasks", h.EnqueueDeleteTask)
		api.GET("/delete-tasks/dequeue", h.DequeueDeleteTask)
		api.PUT("/delete-tasks/:taskId/status", h.UpdateDeleteTaskStatus)

		// 保存任务操作
		api.POST("/save-tasks", h.EnqueueSaveTask)
		api.GET("/save-tasks/dequeue", h.DequeueSaveTask)
		api.PUT("/save-tasks/:taskId/status", h.UpdateSaveTaskStatus)

		// 统计信息
		api.GET("/stats", h.GetStats)
	}

	// 健康检查
	router.GET("/health", h.HealthCheck)
}

// 删除任务相关处理器

// EnqueueDeleteTask 入队删除任务 POST /api/v1/delete-tasks
func (h *QueueHandler) EnqueueDeleteTask(c *gin.Context) {
	var task models.DeleteTask
	if err := c.ShouldBindJSON(&task); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid delete task request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.EnqueueDeleteTask(c.Request.Context(), &task)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to enqueue delete task", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]any{
		"success": true,
		"task_id": task.ID,
		"message": "Delete task enqueued successfully",
	}

	utils.SetJSONResponse(c.Writer, http.StatusCreated, response)
}

// DequeueDeleteTask 出队删除任务 GET /api/v1/delete-tasks/dequeue
func (h *QueueHandler) DequeueDeleteTask(c *gin.Context) {
	task, err := h.service.DequeueDeleteTask(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to dequeue delete task", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	if task == nil {
		// 队列为空，返回204
		c.Status(http.StatusNoContent)
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, task)
}

// UpdateDeleteTaskStatus 更新删除任务状态 PUT /api/v1/delete-tasks/:taskId/status
func (h *QueueHandler) UpdateDeleteTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "task_id is required")
		return
	}

	var req struct {
		Status models.TaskStatus `json:"status" binding:"required"`
		Error  string            `json:"error,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid status update request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.UpdateDeleteTaskStatus(c.Request.Context(), taskID, req.Status, req.Error)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to update delete task status",
			observability.String("task_id", taskID),
			observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]any{
		"success": true,
		"task_id": taskID,
		"status":  req.Status,
		"message": "Task status updated successfully",
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, response)
}

// 保存任务相关处理器

// EnqueueSaveTask 入队保存任务 POST /api/v1/save-tasks
func (h *QueueHandler) EnqueueSaveTask(c *gin.Context) {
	var task models.SaveTask
	if err := c.ShouldBindJSON(&task); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid save task request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.EnqueueSaveTask(c.Request.Context(), &task)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to enqueue save task", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]any{
		"success": true,
		"task_id": task.ID,
		"message": "Save task enqueued successfully",
	}

	utils.SetJSONResponse(c.Writer, http.StatusCreated, response)
}

// DequeueSaveTask 出队保存任务 GET /api/v1/save-tasks/dequeue
func (h *QueueHandler) DequeueSaveTask(c *gin.Context) {
	task, err := h.service.DequeueSaveTask(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to dequeue save task", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	if task == nil {
		// 队列为空，返回204
		c.Status(http.StatusNoContent)
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, task)
}

// UpdateSaveTaskStatus 更新保存任务状态 PUT /api/v1/save-tasks/:taskId/status
func (h *QueueHandler) UpdateSaveTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "task_id is required")
		return
	}

	var req struct {
		Status models.TaskStatus `json:"status" binding:"required"`
		Error  string            `json:"error,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid status update request", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.UpdateSaveTaskStatus(c.Request.Context(), taskID, req.Status, req.Error)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to update save task status",
			observability.String("task_id", taskID),
			observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]any{
		"success": true,
		"task_id": taskID,
		"status":  req.Status,
		"message": "Task status updated successfully",
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, response)
}

// 统计信息和健康检查

// GetStats 获取统计信息 GET /api/v1/stats
func (h *QueueHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to get stats", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, stats)
}

// HealthCheck 健康检查 GET /health
func (h *QueueHandler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "Health check failed", observability.Error(err))
		utils.SetErrorResponse(c.Writer, http.StatusServiceUnavailable, "Service unhealthy")
		return
	}

	utils.SetJSONResponse(c.Writer, http.StatusOK, map[string]any{
		"status":    "healthy",
		"service":   "queue-service",
		"timestamp": "now",
	})
}
