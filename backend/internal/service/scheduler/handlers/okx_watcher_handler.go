package handlers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/agent"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/repository"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

//go:embed templates/OKXWatcherSystem.md
var OKXWatcherSystemTemplateText string

//go:embed templates/OKXWatcherUser.md
var OKXWatcherUserTemplateText string

type OKXWatcherHandler struct {
	svcCtx                     *svc.ServiceContext
	cronTaskRepository         *repository.CronTaskRepository
	cronExecutionRepository    *repository.CronExecutionRepository
	cronExecutionLogRepository *repository.CronExecutionLogRepository

	okxWatcherSystemTemplate *template.Template
	okxWatcherUserTemplate   *template.Template
}

// NewOKXWatcherHandler 创建示例执行器
func NewOKXWatcherHandler(svcCtx *svc.ServiceContext) *OKXWatcherHandler {
	okxWatcherSystemTemplate, err := template.New("OKXWatcherSystem").Parse(OKXWatcherSystemTemplateText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse okxWatcherSystemTemplate: %v\n", err)
		os.Exit(1)
		return nil
	}
	okxWatcherUserTemplate, err := template.New("OKXWatcherUser").Parse(OKXWatcherUserTemplateText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse okxWatcherUserTemplate: %v\n", err)
		os.Exit(1)
		return nil
	}
	return &OKXWatcherHandler{
		svcCtx:                     svcCtx,
		cronTaskRepository:         repository.NewCronTaskRepository(svcCtx),
		cronExecutionRepository:    repository.NewCronExecutionRepository(svcCtx),
		cronExecutionLogRepository: repository.NewCronExecutionLogRepository(svcCtx),
		okxWatcherSystemTemplate:   okxWatcherSystemTemplate,
		okxWatcherUserTemplate:     okxWatcherUserTemplate,
	}
}

// Name 返回执行器名称（必须唯一，与 CronTask.ExecType 匹配）
func (h *OKXWatcherHandler) Name() string {
	return "OKXWatcher"
}

type OkxWatcherRawModel struct {
	Symbol string `json:"symbol"`
}

func getActualAgentName(event *adk.AgentEvent) string {
	if len(event.RunPath) == 0 {
		return event.AgentName
	}
	// RunPath 的最后一个元素是实际产生事件的 Agent
	return event.RunPath[len(event.RunPath)-1].String()
}

func formatEventMessage(event *adk.AgentEvent) string {
	if event.Output == nil || event.Output.MessageOutput == nil {
		return ""
	}
	msg := event.Output.MessageOutput.Message
	if msg == nil {
		return ""
	}

	role := string(msg.Role)
	content := msg.Content
	reasoningContent := msg.ReasoningContent
	callFunctions := make([]schema.FunctionCall, 0)
	for _, toolCall := range msg.ToolCalls {
		if toolCall.Type == "function" {
			callFunctions = append(callFunctions, toolCall.Function)
		}
	}

	output := "# 角色：" + role + "\n---\n"
	output += content + "\n---\n"
	output += "<details>\n<summary>推理过程</summary>\n" + reasoningContent + "\n</details>\n"
	if len(callFunctions) > 0 {
		output += "\n## 调用工具\n"
		output += "| 工具名 | 参数 |\n"
		output += "| :--- | :--- |\n"
		for _, f := range callFunctions {
			output += fmt.Sprintf("| %s | %s |\n", f.Name, f.Arguments)
		}
		output += "\n\n"
	}
	return output
}

// Execute 执行任务
// task: 任务定义，包含 Raw 字段（JSON 格式的执行参数）
// execution: 执行记录
func (h *OKXWatcherHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
	if task == nil {
		return nil
	}
	if task.ExecType != h.Name() {
		return nil
	}
	okxWatcherRawModel := &OkxWatcherRawModel{}
	err := json.Unmarshal([]byte(task.Raw), okxWatcherRawModel)
	if err != nil {
		return &OKXWatcherNonRetryableError{
			err: err,
		}
	}
	agents := agent.Agents()
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agents.OkxWatcher})

	var (
		lastMessage       adk.Message
		lastMessageStream *schema.StreamReader[adk.Message]
	)

	//以前是这样写的
	//queryText := fmt.Sprintf("现在时间是`%s`, 开始对`%s`进行分析.\n> 如果有必要，你不必询问我，你可以调度不同维度的 Agent 共同讨论问题，最终整理你们的建议并告知原因.", time.Now().Format("2006 年 01 月 02 日 15 时 04 分"), okxWatcherRawModel.Symbol)
	//iter := runner.Query(ctx, queryText)

	promptMessages := make([]adk.Message, 0)
	systemMessageTextWriter := &strings.Builder{}
	userMessageTextWriter := &strings.Builder{}
	localTimezone := "Asia/ShangHai"
	location, err := time.LoadLocation("Local")
	if err == nil && location.String() != "Local" {
		localTimezone = location.String()
	}
	applyData := map[string]any{
		"Symbol":   okxWatcherRawModel.Symbol,
		"Now":      time.Now().Format("2006-01-02T15:04"),
		"Timezone": localTimezone,
	}
	err = h.okxWatcherSystemTemplate.Execute(systemMessageTextWriter, applyData)
	if err != nil {
		logger.Error(ctx, "failed to execute okxWatcher system template: %v", err)
		return err
	}
	err = h.okxWatcherUserTemplate.Execute(userMessageTextWriter, applyData)
	if err != nil {
		logger.Error(ctx, "failed to execute okxWatcher user template: %v", err)
		return err
	}
	systemMessage := schema.SystemMessage(systemMessageTextWriter.String())
	promptMessages = append(promptMessages, systemMessage)
	userMessage := schema.UserMessage(userMessageTextWriter.String())
	promptMessages = append(promptMessages, userMessage)

	iter := runner.Run(ctx, promptMessages)
	_ = h.cronExecutionLogRepository.Create(ctx, &model.CronExecutionLog{
		ExecutionID: execution.ID,
		From:        "system",
		Level:       "info",
		Message:     systemMessage.Content,
	})
	_ = h.cronExecutionLogRepository.Create(ctx, &model.CronExecutionLog{
		ExecutionID: execution.ID,
		From:        "user",
		Level:       "info",
		Message:     userMessage.Content,
	})

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		// 添加调试输出
		if false {
			debugBytes, _ := json.Marshal(event)
			fmt.Printf("DEBUG: event.AgentName=%s, RunPath=%v, JSON=%s\n", event.AgentName, event.RunPath, string(debugBytes))
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
				agentName := getActualAgentName(event)
				text := formatEventMessage(event)
				if text != "" {
					_ = h.cronExecutionLogRepository.Create(ctx, &model.CronExecutionLog{
						ExecutionID: execution.ID,
						From:        agentName,
						Level:       "info",
						Message:     text,
					})
				}
			}
		}
	}

	if lastMessage == nil && lastMessageStream != nil {
		msg, _ := schema.ConcatMessageStream(lastMessageStream)
		_ = h.cronExecutionLogRepository.Create(ctx, &model.CronExecutionLog{
			ExecutionID: execution.ID,
			From:        "agent",
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
