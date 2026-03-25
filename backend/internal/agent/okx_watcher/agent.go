package okx_watcher

import (
	"context"

	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// OkxWatcherAgent OKX 盯盘 Agent（DeepAgent，作为顶层编排器）
type OkxWatcherAgent struct {
	agent adk.Agent
}

func NewOkxWatcherAgent(ctx context.Context, svcCtx *svc.ServiceContext, subAgents ...adk.Agent) (*OkxWatcherAgent, error) {
	baseTools := []tool.BaseTool{
		tools.NewOkxCandlesticksTool(svcCtx),
	}

	agent, err := deep.New(ctx, &deep.Config{
		Name:        "OKXWatcher",
		Description: DESCRIPTION,
		ChatModel:   svcCtx.ChatModel,
		Instruction: SOUL,
		SubAgents:   subAgents,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: baseTools,
			},
			EmitInternalEvents: true,
		},
		MaxIteration: 100,
	})
	if err != nil {
		logger.Error(ctx, "NewOkxWatcher error", err)
		return nil, err
	}

	return &OkxWatcherAgent{agent: agent}, nil
}

func (o *OkxWatcherAgent) Agent() adk.Agent {
	return o.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
