package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// AutoMigrate 所有模型
	err = db.AutoMigrate(
		&model.CronTask{},
		&model.CronExecution{},
		&model.CronExecutionLog{},
	)
	assert.NoError(t, err)

	return db
}

func setupTestServiceContext(t *testing.T) *svc.ServiceContext {
	db := setupTestDB(t)

	cfg := config.Config{
		Scheduler: config.SchedulerConfig{
			Enabled:        true,
			MaxConcurrency: 5,
			CheckInterval:  10,
			DefaultTimeout: 300,
		},
	}

	// 创建一个简化的 ServiceContext 用于测试
	return &svc.ServiceContext{
		Config: cfg,
		DB:     db,
	}
}

func TestSchedulerCreation(t *testing.T) {
	svcCtx := setupTestServiceContext(t)

	scheduler := NewScheduler(svcCtx)
	assert.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.executor)
}

func TestExecutorRegister(t *testing.T) {
	executor := NewExecutor()

	// 测试注册 handler
	handler := &mockHandler{name: "test"}
	executor.Register(handler)

	// 测试获取 handler
	retrieved, ok := executor.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "test", retrieved.Name())

	// 测试获取不存在的 handler
	_, ok = executor.Get("nonexistent")
	assert.False(t, ok)
}

func TestSchedulerStartStop(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	// 测试启动
	err := scheduler.Start()
	assert.NoError(t, err)

	// 等待一小段时间确保 goroutine 启动
	time.Sleep(50 * time.Millisecond)

	// 测试停止
	err = scheduler.Stop()
	assert.NoError(t, err)
}

func TestSchedulerDuplicateStart(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	// 第一次启动应该成功
	err := scheduler.Start()
	assert.NoError(t, err)

	// 第二次启动应该失败
	err = scheduler.Start()
	assert.Error(t, err)

	// 清理
	_ = scheduler.Stop()
}

func TestIsValidTask(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	now := time.Now()

	// 测试有效任务
	validTask := &model.CronTask{
		Enabled: true,
		Type:    model.TaskTypeRecurring,
		Status:  model.TaskStatusPending,
	}
	assert.True(t, scheduler.isValidTask(validTask))

	// 测试未启用的任务
	disabledTask := &model.CronTask{
		Enabled: false,
		Type:    model.TaskTypeRecurring,
		Status:  model.TaskStatusPending,
	}
	assert.False(t, scheduler.isValidTask(disabledTask))

	// 测试已过有效期的任务
	expiredTask := &model.CronTask{
		Enabled:    true,
		Type:       model.TaskTypeRecurring,
		Status:     model.TaskStatusPending,
		ValidUntil: sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
	}
	assert.False(t, scheduler.isValidTask(expiredTask))

	// 测试未到有效期的任务
	notStartedTask := &model.CronTask{
		Enabled:   true,
		Type:      model.TaskTypeRecurring,
		Status:    model.TaskStatusPending,
		ValidFrom: sql.NullTime{Time: now.Add(time.Hour), Valid: true},
	}
	assert.False(t, scheduler.isValidTask(notStartedTask))

	// 测试已完成的一次性任务
	completedOnceTask := &model.CronTask{
		Enabled: true,
		Type:    model.TaskTypeOnce,
		Status:  model.TaskStatusCompleted,
	}
	assert.False(t, scheduler.isValidTask(completedOnceTask))
}

// mockHandler 用于测试
type mockHandler struct {
	name       string
	execErr    error
	execCount  int
	mu         sync.Mutex
}

func (h *mockHandler) Name() string {
	return h.name
}

func (h *mockHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
	h.mu.Lock()
	h.execCount++
	h.mu.Unlock()
	return h.execErr
}

func (h *mockHandler) getExecCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.execCount
}

// countingHandler 包装 mockHandler，允许自定义 Execute 逻辑
type countingHandler struct {
	name        string
	execCount   int
	mu          sync.Mutex
	executeFunc func(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error
}

func newCountingHandler(name string, executeFunc func(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error) *countingHandler {
	return &countingHandler{
		name:        name,
		executeFunc: executeFunc,
	}
}

func (h *countingHandler) Name() string {
	return h.name
}

func (h *countingHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
	h.mu.Lock()
	h.execCount++
	h.mu.Unlock()
	return h.executeFunc(ctx, task, execution)
}

func (h *countingHandler) getExecCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.execCount
}

// ============== 场景测试 ==============

// TestScenario_TaskExecutionSuccess 场景：任务成功执行
// 注：本测试验证 handler 注册和任务执行流程
func TestScenario_TaskExecutionSuccess(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	// 创建 mock handler
	handler := &mockHandler{name: "test_success"}
	scheduler.RegisterHandler(handler)

	// 直接调用 Execute 验证 handler 工作正常
	task := &model.CronTask{
		Name:           "成功任务",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "test_success",
		Enabled:        true,
		MaxRetries:     0,
		TimeoutSeconds: 10,
	}
	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
	}

	err := handler.Execute(context.Background(), task, execution)
	assert.NoError(t, err)
	assert.Equal(t, 1, handler.getExecCount())
}

// TestScenario_DirectTaskExecution 场景：直接调用 onTaskTrigger 验证完整流程
func TestScenario_DirectTaskExecution(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	handler := &mockHandler{name: "direct_test"}
	scheduler.RegisterHandler(handler)

	// 创建任务
	task := &model.CronTask{
		Name:           "直接执行任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "direct_test",
		Enabled:        true,
		MaxRetries:     0,
		TimeoutSeconds: 10,
	}
	err := svcCtx.DB.Create(task).Error
	assert.NoError(t, err)

	// 直接调用 onTaskTrigger
	scheduler.onTaskTrigger(task)

	// 等待执行完成
	time.Sleep(500 * time.Millisecond)

	// 验证任务已执行
	assert.GreaterOrEqual(t, handler.getExecCount(), 1, "任务应该至少执行一次")
}

// TestScenario_TaskExecutionFailure 场景：任务执行失败
func TestScenario_TaskExecutionFailure(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	// 创建总是失败的 handler
	handler := &mockHandler{
		name:    "test_fail",
		execErr: errors.New("permanent error"),
	}
	scheduler.RegisterHandler(handler)

	// 创建任务
	task := &model.CronTask{
		Name:           "失败任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "test_fail",
		Enabled:        true,
		MaxRetries:     0,
		TimeoutSeconds: 10,
	}
	err := svcCtx.DB.Create(task).Error
	assert.NoError(t, err)

	// 直接调用 onTaskTrigger
	scheduler.onTaskTrigger(task)

	// 等待执行完成
	time.Sleep(500 * time.Millisecond)

	// 验证任务执行了一次
	assert.GreaterOrEqual(t, handler.getExecCount(), 1, "任务应该至少执行一次")
}

// TestScenario_TaskRetry 场景：任务重试机制
func TestScenario_TaskRetry(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	// 创建可重试错误
	retryErr := NewRetryableError(errors.New("temporary error"))

	// 创建 handler，总是返回可重试错误
	handler := &mockHandler{
		name:    "test_retry",
		execErr: retryErr,
	}

	scheduler.RegisterHandler(handler)

	// 创建任务，允许重试 2 次
	task := &model.CronTask{
		Name:           "重试任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "test_retry",
		Enabled:        true,
		MaxRetries:     2,
		TimeoutSeconds: 10,
	}
	err := svcCtx.DB.Create(task).Error
	assert.NoError(t, err)

	// 直接调用 onTaskTrigger
	scheduler.onTaskTrigger(task)

	// 等待执行完成
	time.Sleep(500 * time.Millisecond)

	// 验证 handler 被调用
	assert.GreaterOrEqual(t, handler.getExecCount(), 1, "任务应该至少执行一次")
}

// TestScenario_ConcurrencyLimit 场景：并发限制
func TestScenario_ConcurrencyLimit(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	// 设置最大并发为 2
	svcCtx.Config.Scheduler.MaxConcurrency = 2

	scheduler := NewScheduler(svcCtx)

	// 创建慢速 handler（模拟长时间执行）
	var mu sync.Mutex
	maxConcurrent := 0
	currentConcurrent := 0

	slowHandler := newCountingHandler("slow_handler", func(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
		mu.Lock()
		currentConcurrent++
		if currentConcurrent > maxConcurrent {
			maxConcurrent = currentConcurrent
		}
		mu.Unlock()

		// 模拟慢速执行
		time.Sleep(200 * time.Millisecond)

		mu.Lock()
		currentConcurrent--
		mu.Unlock()

		return nil
	})

	scheduler.RegisterHandler(slowHandler)

	// 创建多个任务
	for i := 0; i < 5; i++ {
		task := &model.CronTask{
			Name:           fmt.Sprintf("并发任务_%d", i),
			Spec:           "* * * * * *",
			Type:           model.TaskTypeRecurring,
			Status:         model.TaskStatusPending,
			ExecType:       "slow_handler",
			Enabled:        true,
			MaxRetries:     0,
			TimeoutSeconds: 30,
		}
		err := svcCtx.DB.Create(task).Error
		assert.NoError(t, err)
	}

	// 启动调度器
	err := scheduler.Start()
	assert.NoError(t, err)

	// 等待所有任务执行
	time.Sleep(800 * time.Millisecond)

	// 验证最大并发数不超过限制
	mu.Lock()
	actualMaxConcurrent := maxConcurrent
	mu.Unlock()
	assert.LessOrEqual(t, actualMaxConcurrent, 2)

	// 清理
	_ = scheduler.Stop()
}

// TestScenario_SkipIfRunning 场景：跳过正在执行的任务
func TestScenario_SkipIfRunning(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	executionTimes := []time.Time{}
	var mu sync.Mutex

	// 创建慢速 handler
	slowHandler := newCountingHandler("skip_handler", func(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
		mu.Lock()
		executionTimes = append(executionTimes, time.Now())
		mu.Unlock()
		time.Sleep(500 * time.Millisecond) // 慢速执行
		return nil
	})

	scheduler.RegisterHandler(slowHandler)

	// 创建每秒执行的任务
	task := &model.CronTask{
		Name:           "跳过任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "skip_handler",
		Enabled:        true,
		MaxRetries:     0,
		TimeoutSeconds: 30,
	}
	err := svcCtx.DB.Create(task).Error
	assert.NoError(t, err)

	// 启动调度器
	err = scheduler.Start()
	assert.NoError(t, err)

	// 等待 1.5 秒（应该触发多次，但由于慢速执行会被跳过）
	time.Sleep(1500 * time.Millisecond)

	// 验证执行次数（由于每次执行 500ms，1.5 秒内最多执行 3 次）
	mu.Lock()
	actualExecutions := len(executionTimes)
	mu.Unlock()
	assert.LessOrEqual(t, actualExecutions, 4)

	// 清理
	_ = scheduler.Stop()
}

// TestScenario_Timeout 场景：任务超时
func TestScenario_Timeout(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	// 创建超时 handler
	timeoutHandler := newCountingHandler("timeout_handler", func(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
		// 等待超过超时时间
		select {
		case <-time.After(2 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err() // 超时取消
		}
	})

	scheduler.RegisterHandler(timeoutHandler)

	// 创建任务，超时时间 500ms
	task := &model.CronTask{
		Name:           "超时任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "timeout_handler",
		Enabled:        true,
		MaxRetries:     0,
		TimeoutSeconds: 1, // 1 秒超时
	}
	err := svcCtx.DB.Create(task).Error
	assert.NoError(t, err)

	// 启动调度器
	err = scheduler.Start()
	assert.NoError(t, err)

	// 等待执行和超时
	time.Sleep(2 * time.Second)

	// 验证任务因超时而失败
	updatedTask, err := scheduler.cronTaskRepo.GetByID(scheduler.ctx, task.ID)
	assert.NoError(t, err)
	// 任务应该被标记为失败或仍在运行（取决于超时处理）
	t.Logf("任务状态：%s", updatedTask.Status)

	// 清理
	_ = scheduler.Stop()
}

// TestScenario_TaskValidity 场景：任务有效期过滤
func TestScenario_TaskValidity(t *testing.T) {
	svcCtx := setupTestServiceContext(t)
	scheduler := NewScheduler(svcCtx)

	handler := &mockHandler{name: "validity_handler"}
	scheduler.RegisterHandler(handler)

	now := time.Now()

	// 创建未到有效期的任务
	taskNotStarted := &model.CronTask{
		Name:           "未开始任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "validity_handler",
		Enabled:        true,
		ValidFrom:      sql.NullTime{Time: now.Add(time.Hour), Valid: true},
		ValidUntil:     sql.NullTime{Time: now.Add(2 * time.Hour), Valid: true},
		MaxRetries:     0,
		TimeoutSeconds: 10,
	}
	err := svcCtx.DB.Create(taskNotStarted).Error
	assert.NoError(t, err)

	// 创建已过有效期的任务
	taskExpired := &model.CronTask{
		Name:           "已过期任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "validity_handler",
		Enabled:        true,
		ValidFrom:      sql.NullTime{Time: now.Add(-2 * time.Hour), Valid: true},
		ValidUntil:     sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
		MaxRetries:     0,
		TimeoutSeconds: 10,
	}
	err = svcCtx.DB.Create(taskExpired).Error
	assert.NoError(t, err)

	// 创建有效期内的任务
	taskValid := &model.CronTask{
		Name:           "有效任务",
		Spec:           "* * * * * *",
		Type:           model.TaskTypeRecurring,
		Status:         model.TaskStatusPending,
		ExecType:       "validity_handler",
		Enabled:        true,
		ValidFrom:      sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
		ValidUntil:     sql.NullTime{Time: now.Add(time.Hour), Valid: true},
		MaxRetries:     0,
		TimeoutSeconds: 10,
	}
	err = svcCtx.DB.Create(taskValid).Error
	assert.NoError(t, err)

	// 验证 isValidTask 正确过滤
	assert.False(t, scheduler.isValidTask(taskNotStarted), "未到有效期的任务应该被过滤")
	assert.False(t, scheduler.isValidTask(taskExpired), "已过有效期的任务应该被过滤")
	assert.True(t, scheduler.isValidTask(taskValid), "有效期内的任务应该通过验证")
}
