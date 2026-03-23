package handlers

import (
	"context"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
)

type OKXWatcherHandler struct {
	svcCtx *svc.ServiceContext
}

// NewOKXWatcherHandler 创建示例执行器
func NewOKXWatcherHandler(svcCtx *svc.ServiceContext) *OKXWatcherHandler {
	return &OKXWatcherHandler{svcCtx: svcCtx}
}

// Name 返回执行器名称（必须唯一，与 CronTask.ExecType 匹配）
func (h *OKXWatcherHandler) Name() string {
	return "OKXWatcher"
}

// Execute 执行任务
// task: 任务定义，包含 Raw 字段（JSON 格式的执行参数）
// execution: 执行记录
func (h *OKXWatcherHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
	// TODO: 实现具体的执行逻辑

	// 1. 解析 task.Raw 中的 JSON 参数
	// var params YourParamsStruct
	// if err := json.Unmarshal([]byte(task.Raw), &params); err != nil {
	//     return scheduler.NewNonRetryableError(fmt.Errorf("failed to parse params: %w", err))
	// }

	// 2. 执行具体业务逻辑
	// ...

	// 3. 如果需要重试，返回实现 RetryableError 接口的错误
	// if isTemporaryError(err) {
	//     return scheduler.NewRetryableError(err)
	// }

	// 4. 成功返回 nil
	return nil
}

// OKXWatcherRetryableError 示例：可重试错误实现
type OKXWatcherRetryableError struct {
	err error
}

func (e *OKXWatcherRetryableError) Error() string     { return e.err.Error() }
func (e *OKXWatcherRetryableError) IsRetryable() bool { return true }

// OKXWatcherNonRetryableError 示例：不可重试错误实现
type OKXWatcherNonRetryableError struct {
	err error
}

func (e *OKXWatcherNonRetryableError) Error() string     { return e.err.Error() }
func (e *OKXWatcherNonRetryableError) IsRetryable() bool { return false }
