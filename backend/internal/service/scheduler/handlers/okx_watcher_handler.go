package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/agent"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/repository"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

type OKXWatcherHandler struct {
	svcCtx                     *svc.ServiceContext
	cronTaskRepository         *repository.CronTaskRepository
	cronExecutionRepository    *repository.CronExecutionRepository
	cronExecutionLogRepository *repository.CronExecutionLogRepository
}

// NewOKXWatcherHandler 创建示例执行器
func NewOKXWatcherHandler(svcCtx *svc.ServiceContext) *OKXWatcherHandler {
	return &OKXWatcherHandler{
		svcCtx:                     svcCtx,
		cronTaskRepository:         repository.NewCronTaskRepository(svcCtx),
		cronExecutionRepository:    repository.NewCronExecutionRepository(svcCtx),
		cronExecutionLogRepository: repository.NewCronExecutionLogRepository(svcCtx),
	}
}

// Name 返回执行器名称（必须唯一，与 CronTask.ExecType 匹配）
func (h *OKXWatcherHandler) Name() string {
	return "OKXWatcher"
}

// Execute 执行任务
// task: 任务定义，包含 Raw 字段（JSON 格式的执行参数）
// execution: 执行记录
func (h *OKXWatcherHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
	agents := agent.Agents()
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agents.OkxWatcher})
	var (
		lastMessage       adk.Message
		lastMessageStream *schema.StreamReader[adk.Message]
	)
	iter := runner.Query(ctx, fmt.Sprintf("现在时间是 %s, 开始对ETH-USDT-SWAP进行分析", time.Now().Format("2006年01月02日15时04分")))
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if lastMessageStream != nil {
				lastMessageStream.Close()
			}
			if event.Output.MessageOutput.IsStreaming {
				cpStream := event.Output.MessageOutput.MessageStream.Copy(2)
				event.Output.MessageOutput.MessageStream = cpStream[0]
				lastMessage = nil
				lastMessageStream = cpStream[1]
			} else {
				lastMessage = event.Output.MessageOutput.Message
				lastMessageStream = nil
				_ = h.cronExecutionLogRepository.Create(ctx, &model.CronExecutionLog{
					ExecutionID: execution.ID,
					From:        event.AgentName,
					Level:       "info",
					Message:     lastMessage.String(),
				})
			}
		}
	}

	if lastMessage == nil && lastMessageStream != nil {
		msg, _ := schema.ConcatMessageStream(lastMessageStream)
		_ = h.cronExecutionLogRepository.Create(ctx, &model.CronExecutionLog{
			ExecutionID: execution.ID,
			From:        "unknown",
			Level:       "info",
			Message:     msg.String(),
		})
	}
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
