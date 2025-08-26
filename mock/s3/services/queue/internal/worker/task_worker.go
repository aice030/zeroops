package worker

import (
	"context"
	"mocks3/shared/interfaces"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"sync"
	"time"
)

// TaskWorker 任务工作者
type TaskWorker struct {
	queueService    interfaces.QueueService
	deleteProcessor interfaces.DeleteTaskProcessor
	saveProcessor   interfaces.SaveTaskProcessor
	logger          *observability.Logger

	// 控制相关
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 配置
	workerCount  int
	pollInterval time.Duration
}

// WorkerConfig Worker配置
type WorkerConfig struct {
	WorkerCount  int           `yaml:"worker_count"`
	PollInterval time.Duration `yaml:"poll_interval"`
}

// NewTaskWorker 创建任务工作者
func NewTaskWorker(
	queueService interfaces.QueueService,
	deleteProcessor interfaces.DeleteTaskProcessor,
	saveProcessor interfaces.SaveTaskProcessor,
	config *WorkerConfig,
	logger *observability.Logger,
) *TaskWorker {
	ctx, cancel := context.WithCancel(context.Background())

	// 设置默认值
	if config.WorkerCount <= 0 {
		config.WorkerCount = 3
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 5 * time.Second
	}

	return &TaskWorker{
		queueService:    queueService,
		deleteProcessor: deleteProcessor,
		saveProcessor:   saveProcessor,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		workerCount:     config.WorkerCount,
		pollInterval:    config.PollInterval,
	}
}

// Start 启动任务工作者
func (w *TaskWorker) Start() {
	w.logger.Info(w.ctx, "Starting task workers",
		observability.Int("worker_count", w.workerCount),
		observability.String("poll_interval", w.pollInterval.String()))

	// 启动删除任务工作者
	for i := 0; i < w.workerCount/2+1; i++ {
		w.wg.Add(1)
		go w.runDeleteTaskWorker(i)
	}

	// 启动保存任务工作者
	for i := 0; i < w.workerCount/2; i++ {
		w.wg.Add(1)
		go w.runSaveTaskWorker(i)
	}

	w.logger.Info(w.ctx, "All task workers started successfully")
}

// Stop 停止任务工作者
func (w *TaskWorker) Stop() {
	w.logger.Info(w.ctx, "Stopping task workers...")

	w.cancel()
	w.wg.Wait()

	w.logger.Info(w.ctx, "All task workers stopped")
}

// runDeleteTaskWorker 运行删除任务工作者
func (w *TaskWorker) runDeleteTaskWorker(workerID int) {
	defer w.wg.Done()

	w.logger.Info(w.ctx, "Delete task worker started",
		observability.Int("worker_id", workerID))

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info(w.ctx, "Delete task worker stopping",
				observability.Int("worker_id", workerID))
			return

		case <-ticker.C:
			w.processDeleteTasks(workerID)
		}
	}
}

// runSaveTaskWorker 运行保存任务工作者
func (w *TaskWorker) runSaveTaskWorker(workerID int) {
	defer w.wg.Done()

	w.logger.Info(w.ctx, "Save task worker started",
		observability.Int("worker_id", workerID))

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info(w.ctx, "Save task worker stopping",
				observability.Int("worker_id", workerID))
			return

		case <-ticker.C:
			w.processSaveTasks(workerID)
		}
	}
}

// processDeleteTasks 处理删除任务
func (w *TaskWorker) processDeleteTasks(workerID int) {
	task, err := w.queueService.DequeueDeleteTask(w.ctx)
	if err != nil {
		w.logger.Error(w.ctx, "Failed to dequeue delete task",
			observability.Int("worker_id", workerID),
			observability.Error(err))
		return
	}

	if task == nil {
		// 队列为空，继续轮询
		return
	}

	w.logger.Info(w.ctx, "Processing delete task",
		observability.Int("worker_id", workerID),
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	// 处理任务
	err = w.deleteProcessor.ProcessDeleteTask(w.ctx, task)

	// 更新任务状态
	var status models.TaskStatus
	var errorMsg string

	if err != nil {
		status = models.TaskStatusFailed
		errorMsg = err.Error()
		w.logger.Error(w.ctx, "Delete task processing failed",
			observability.Int("worker_id", workerID),
			observability.String("task_id", task.ID),
			observability.Error(err))
	} else {
		status = models.TaskStatusCompleted
		w.logger.Info(w.ctx, "Delete task processed successfully",
			observability.Int("worker_id", workerID),
			observability.String("task_id", task.ID))
	}

	// 更新任务状态
	if updateErr := w.queueService.UpdateDeleteTaskStatus(w.ctx, task.ID, status, errorMsg); updateErr != nil {
		w.logger.Error(w.ctx, "Failed to update delete task status",
			observability.String("task_id", task.ID),
			observability.Error(updateErr))
	}
}

// processSaveTasks 处理保存任务
func (w *TaskWorker) processSaveTasks(workerID int) {
	task, err := w.queueService.DequeueSaveTask(w.ctx)
	if err != nil {
		w.logger.Error(w.ctx, "Failed to dequeue save task",
			observability.Int("worker_id", workerID),
			observability.Error(err))
		return
	}

	if task == nil {
		// 队列为空，继续轮询
		return
	}

	w.logger.Info(w.ctx, "Processing save task",
		observability.Int("worker_id", workerID),
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	// 处理任务
	err = w.saveProcessor.ProcessSaveTask(w.ctx, task)

	// 更新任务状态
	var status models.TaskStatus
	var errorMsg string

	if err != nil {
		status = models.TaskStatusFailed
		errorMsg = err.Error()
		w.logger.Error(w.ctx, "Save task processing failed",
			observability.Int("worker_id", workerID),
			observability.String("task_id", task.ID),
			observability.Error(err))
	} else {
		status = models.TaskStatusCompleted
		w.logger.Info(w.ctx, "Save task processed successfully",
			observability.Int("worker_id", workerID),
			observability.String("task_id", task.ID))
	}

	// 更新任务状态
	if updateErr := w.queueService.UpdateSaveTaskStatus(w.ctx, task.ID, status, errorMsg); updateErr != nil {
		w.logger.Error(w.ctx, "Failed to update save task status",
			observability.String("task_id", task.ID),
			observability.Error(updateErr))
	}
}

// GetStats 获取Worker统计信息
func (w *TaskWorker) GetStats() map[string]any {
	return map[string]any{
		"worker_count":  w.workerCount,
		"poll_interval": w.pollInterval.String(),
		"running":       w.ctx.Err() == nil,
	}
}
