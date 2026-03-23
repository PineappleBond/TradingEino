package agent

import (
	"context"
	"sync"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

type AgentsModel struct {
	svcCtx     *svc.ServiceContext
	OkxWatcher adk.Agent
	mux        sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

var _agents *AgentsModel

func Agents() *AgentsModel {
	return _agents
}

func InitAgents(svcCtx *svc.ServiceContext) error {
	ctx, cancel := context.WithCancel(context.Background())
	// TODO 多维度分析Agent
	subAgents := make([]adk.Agent, 0)
	baseTools := make([]tool.BaseTool, 0)
	baseTools = append(baseTools, tools.NewOkxCandlesticksTool(svcCtx))
	okxWatcher, err := deep.New(ctx, &deep.Config{
		Name:        "OKXWatcher",
		Description: "定时任务将唤醒OKXWatcher，作为盯盘手，你会获取多周期K线和指标数据，进行分析，最终给出交易建议。",
		ChatModel:   svcCtx.ChatModel,
		Instruction: "定时任务将唤醒OKXWatcher，作为盯盘手，你会获取多周期K线和指标数据，进行分析，最终给出交易建议。",
		SubAgents:   subAgents,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:                baseTools,
				UnknownToolsHandler:  nil,
				ExecuteSequentially:  false,
				ToolArgumentsHandler: nil,
				ToolCallMiddlewares:  nil,
			},
			ReturnDirectly:     nil,
			EmitInternalEvents: false,
		},
		MaxIteration:                 0,
		WithoutWriteTodos:            false,
		WithoutGeneralSubAgent:       false,
		TaskToolDescriptionGenerator: nil,
		Middlewares:                  nil,
	})
	if err != nil {
		cancel()
		return err
	}

	tmp := &AgentsModel{
		svcCtx:     svcCtx,
		OkxWatcher: okxWatcher,
		mux:        sync.Mutex{},
		ctx:        ctx,
		cancel:     cancel,
	}
	_agents = tmp
	return nil
}
