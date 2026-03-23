package handlers

import (
	"context"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/service/scheduler"
)

// ExampleHandler 示例任务执行器
// 复制此模板来实现具体的执行器
type ExampleHandler struct {
	// 添加执行器需要的依赖
	// 例如：svcCtx *svc.ServiceContext
}

// NewExampleHandler 创建示例执行器
func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{}
}

// Name 返回执行器名称（必须唯一，与 CronTask.ExecType 匹配）
func (h *ExampleHandler) Name() string {
	return "example"
}

// Execute 执行任务
// task: 任务定义，包含 Raw 字段（JSON 格式的执行参数）
// execution: 执行记录
func (h *ExampleHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
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

// 示例：可重试错误实现
type exampleRetryableError struct {
	err error
}

func (e *exampleRetryableError) Error() string   { return e.err.Error() }
func (e *exampleRetryableError) IsRetryable() bool { return true }

// 示例：不可重试错误实现
type exampleNonRetryableError struct {
	err error
}

func (e *exampleNonRetryableError) Error() string   { return e.err.Error() }
func (e *exampleNonRetryableError) IsRetryable() bool { return false }

// 辅助函数（可根据需要修改）
func isTemporaryError(err error) bool {
	// 判断是否是临时错误（可重试）
	return false
}

// 确保编译时检查接口实现
var _ scheduler.TaskHandler = (*ExampleHandler)(nil)
