package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/cronutil"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/repository"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/robfig/cron/v3"
)

// TaskHandler 任务执行器接口
type TaskHandler interface {
	// Name 返回执行器名称（用于匹配 CronTask.ExecType）
	Name() string
	// Execute 执行任务
	Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error
}

// RetryableError 可重试错误接口
type RetryableError interface {
	error
	IsRetryable() bool
}

// Executor 执行器，管理所有 TaskHandler
type Executor struct {
	mu       sync.RWMutex
	handlers map[string]TaskHandler
}

// NewExecutor 创建执行器
func NewExecutor() *Executor {
	return &Executor{
		handlers: make(map[string]TaskHandler),
	}
}

// Register 注册执行器
func (e *Executor) Register(handler TaskHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[handler.Name()] = handler
}

// Get 获取执行器
func (e *Executor) Get(execType string) (TaskHandler, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	handler, ok := e.handlers[execType]
	return handler, ok
}

// Scheduler 调度器
type Scheduler struct {
	svcCtx         *svc.ServiceContext
	cronTaskRepo   *repository.CronTaskRepository
	executionRepo  *repository.CronExecutionRepository
	logRepo        *repository.CronExecutionLogRepository
	executor       *Executor
	cron           *cron.Cron
	running        bool
	mu             sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	semaphore      chan struct{}
	checkInterval  time.Duration
	defaultTimeout time.Duration
}

// NewScheduler 创建调度器
func NewScheduler(svcCtx *svc.ServiceContext) *Scheduler {
	cfg := svcCtx.Config.Scheduler

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		svcCtx:        svcCtx,
		cronTaskRepo:  repository.NewCronTaskRepository(svcCtx),
		executionRepo: repository.NewCronExecutionRepository(svcCtx),
		logRepo:       repository.NewCronExecutionLogRepository(svcCtx),
		executor:      NewExecutor(),
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		))),
		ctx:            ctx,
		cancel:         cancel,
		semaphore:      make(chan struct{}, cfg.MaxConcurrency),
		checkInterval:  time.Duration(cfg.CheckInterval) * time.Second,
		defaultTimeout: time.Duration(cfg.DefaultTimeout) * time.Second,
	}
}

// RegisterHandler 注册执行器
func (s *Scheduler) RegisterHandler(handler TaskHandler) {
	s.executor.Register(handler)
	logger.Info(s.ctx, "Scheduler: handler registered", "name", handler.Name())
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("scheduler is already running")
	}

	if !s.svcCtx.Config.Scheduler.Enabled {
		logger.Info(s.ctx, "Scheduler: disabled in config")
		return nil
	}

	s.running = true

	// 加载所有启用的任务
	if err := s.loadTasks(); err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// 启动调度循环
	s.wg.Add(1)
	go s.scheduleLoop()

	s.cron.Start()

	logger.Info(s.ctx, "Scheduler: started")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return errors.New("scheduler is not running")
	}
	s.running = false
	s.mu.Unlock()

	// 取消所有正在执行的任务
	s.cancel()

	// 等待所有 goroutine 完成
	s.wg.Wait()

	// 停止 cron 调度器
	s.cron.Stop()

	logger.Info(s.ctx, "Scheduler: stopped")
	return nil
}

// loadTasks 从数据库加载所有启用的任务
func (s *Scheduler) loadTasks() error {
	ctx := context.Background()

	// 获取所有启用的重复执行任务
	tasks, err := s.cronTaskRepo.GetRecurringTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get recurring tasks: %w", err)
	}

	for _, task := range tasks {
		// 过滤掉 once 类型已完成的任务（GetRecurringTasks 已经过滤了）
		// 检查是否在有效期内
		if !s.isValidTask(task) {
			continue
		}

		// 检查执行器是否已注册
		if _, ok := s.executor.Get(task.ExecType); !ok {
			logger.Warn(ctx, "Scheduler: handler not found, skipping task",
				"task_id", task.ID, "task_name", task.Name, "exec_type", task.ExecType)
			continue
		}

		// 添加到 cron 调度器
		if err := s.addToCron(task); err != nil {
			logger.Error(ctx, "Scheduler: failed to add task to cron", err,
				"task_id", task.ID, "task_name", task.Name)
		}
	}

	return nil
}

// isValidTask 检查任务是否在有效期内
func (s *Scheduler) isValidTask(task *model.CronTask) bool {
	now := time.Now()

	if task.ValidFrom.Valid && now.Before(task.ValidFrom.Time) {
		return false
	}
	if task.ValidUntil.Valid && now.After(task.ValidUntil.Time) {
		return false
	}

	// 对于 once 类型，检查是否已经执行过
	if task.Type == model.TaskTypeOnce && task.Status == model.TaskStatusCompleted {
		return false
	}

	return task.Enabled
}

// addToCron 添加任务到 cron 调度器
func (s *Scheduler) addToCron(task *model.CronTask) error {
	// 计算下次执行时间
	nextAt, err := s.getNextExecutionTime(task.Spec)
	if err != nil {
		return fmt.Errorf("invalid cron spec: %w", err)
	}

	// 更新任务的下次执行时间
	if err := s.cronTaskRepo.UpdateNextExecution(s.ctx, task.ID, nextAt); err != nil {
		return fmt.Errorf("failed to update next execution time: %w", err)
	}

	// 创建 cron 任务
	_, err = s.cron.AddFunc(task.Spec, func() {
		s.onTaskTrigger(task)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	logger.Info(s.ctx, "Scheduler: task added to cron",
		"task_id", task.ID, "task_name", task.Name, "next_execution", nextAt)
	return nil
}

// getNextExecutionTime 计算下次执行时间（支持 6 位 cron 表达式）
func (s *Scheduler) getNextExecutionTime(spec string) (time.Time, error) {
	return cronutil.GetNextExecutionTime(spec)
}

// onTaskTrigger 任务触发时的回调
func (s *Scheduler) onTaskTrigger(task *model.CronTask) {
	ctx := s.ctx

	logger.Info(ctx, "Scheduler: task triggered", "task_id", task.ID, "task_name", task.Name)

	// 检查是否有正在执行的相同任务（跳过策略）
	runningExecutions, err := s.executionRepo.GetRunningExecutions(ctx)
	if err != nil {
		logger.Error(ctx, "Scheduler: failed to get running executions", err)
		return
	}

	for _, exec := range runningExecutions {
		if exec.TaskID == task.ID {
			logger.Warn(ctx, "Scheduler: task already running, skipping",
				"task_id", task.ID, "task_name", task.Name)
			return
		}
	}

	// 获取超时时间
	timeout := s.defaultTimeout
	if task.TimeoutSeconds > 0 {
		timeout = time.Duration(task.TimeoutSeconds) * time.Second
	}

	// 创建执行记录
	execution := &model.CronExecution{
		TaskID:      task.ID,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
	}
	if err := s.executionRepo.Create(ctx, execution); err != nil {
		logger.Error(ctx, "Scheduler: failed to create execution record", err,
			"task_id", task.ID)
		return
	}

	// 获取信号量（并发控制）
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	// 启动执行 goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.executeTask(task, execution, timeout)
	}()
}

// executeTask 执行任务
func (s *Scheduler) executeTask(task *model.CronTask, execution *model.CronExecution, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	logger.Info(ctx, "Executor: starting execution",
		"task_id", task.ID, "task_name", task.Name, "execution_id", execution.ID)

	// 标记为执行中
	if err := s.executionRepo.MarkAsRunning(ctx, execution.ID); err != nil {
		logger.Error(ctx, "Executor: failed to mark as running", err)
		return
	}

	// 更新任务状态
	if err := s.cronTaskRepo.MarkAsRunning(ctx, task.ID); err != nil {
		logger.Error(ctx, "Executor: failed to update task status", err)
		return
	}

	// 获取执行器
	handler, ok := s.executor.Get(task.ExecType)
	if !ok {
		s.markExecutionFailed(ctx, execution.ID, task.ID, fmt.Errorf("handler not found: %s", task.ExecType))
		return
	}

	// 执行任务
	execErr := handler.Execute(ctx, task, execution)

	// 标记执行完成
	if execErr != nil {
		s.handleExecutionError(ctx, task, execution, execErr)
	} else {
		s.markExecutionSuccess(ctx, execution.ID, task.ID)
	}
}

// markExecutionSuccess 标记执行为成功
func (s *Scheduler) markExecutionSuccess(ctx context.Context, executionID uint, taskID uint) {
	now := time.Now()

	if err := s.executionRepo.MarkAsSuccess(ctx, executionID, now); err != nil {
		logger.Error(ctx, "Executor: failed to mark execution as success", err)
	}

	if err := s.cronTaskRepo.IncrementTotalExecutions(ctx, taskID); err != nil {
		logger.Error(ctx, "Executor: failed to increment execution count", err)
	}

	// 对于 once 类型任务，标记为已完成
	task, err := s.cronTaskRepo.GetByID(ctx, taskID)
	if err == nil && task.Type == model.TaskTypeOnce {
		if err := s.cronTaskRepo.MarkAsCompleted(ctx, taskID); err != nil {
			logger.Error(ctx, "Executor: failed to mark task as completed", err)
		}
	}

	logger.Info(ctx, "Executor: execution completed successfully",
		"execution_id", executionID, "task_id", taskID)
}

// markExecutionFailed 标记执行为失败
func (s *Scheduler) markExecutionFailed(ctx context.Context, executionID uint, taskID uint, err error) {
	now := time.Now()

	if err := s.executionRepo.MarkAsFailed(ctx, executionID, now, err.Error()); err != nil {
		logger.Error(ctx, "Executor: failed to mark execution as failed", err)
	}

	logger.Error(ctx, "Executor: execution failed", err,
		"execution_id", executionID, "task_id", taskID)
}

// handleExecutionError 处理执行错误（包括重试逻辑）
func (s *Scheduler) handleExecutionError(ctx context.Context, task *model.CronTask, execution *model.CronExecution, execErr error) {
	// 检查是否可重试
	var retryableErr RetryableError
	isRetryable := errors.As(execErr, &retryableErr) && retryableErr.IsRetryable()

	if isRetryable && execution.RetryCount < task.MaxRetries {
		// 标记为已重试
		if err := s.executionRepo.IncrementRetryCount(ctx, execution.ID); err != nil {
			logger.Error(ctx, "Executor: failed to increment retry count", err)
		}

		logger.Warn(ctx, "Executor: scheduling retry",
			"task_id", task.ID, "execution_id", execution.ID,
			"retry_count", execution.RetryCount+1, "max_retries", task.MaxRetries)

		// 创建新的执行记录用于重试
		retryExecution := &model.CronExecution{
			TaskID:      task.ID,
			ScheduledAt: time.Now().Add(time.Minute), // 1 分钟后重试
			Status:      model.ExecutionStatusPending,
			RetryCount:  execution.RetryCount + 1,
		}
		if err := s.executionRepo.Create(ctx, retryExecution); err != nil {
			logger.Error(ctx, "Executor: failed to create retry execution", err)
		}
	} else {
		// 不再重试，标记为失败
		s.markExecutionFailed(ctx, execution.ID, task.ID, execErr)

		// 更新任务状态为失败
		if err := s.cronTaskRepo.MarkAsFailed(ctx, task.ID); err != nil {
			logger.Error(ctx, "Executor: failed to mark task as failed", err)
		}
	}
}

// scheduleLoop 调度循环（用于检查到期任务）
func (s *Scheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkDueTasks()
		}
	}
}

// checkDueTasks 检查并执行到期任务
func (s *Scheduler) checkDueTasks() {
	ctx := s.ctx
	now := time.Now()

	// 使用 SELECT FOR UPDATE SKIP LOCKED 防止多实例部署时的竞态条件
	// 被其他实例锁定的行会被跳过，不会阻塞
	tasks, err := s.cronTaskRepo.GetDueTasksWithLock(ctx, now)
	if err != nil {
		logger.Error(ctx, "Scheduler: failed to get due tasks", err)
		return
	}

	for _, task := range tasks {
		s.onTaskTrigger(task)
	}
}
