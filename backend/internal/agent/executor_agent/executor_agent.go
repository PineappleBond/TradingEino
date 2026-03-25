package executor_agent

import (
	"context"
	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// ExecutorAgent 交易执行代理（ChatModelAgent，Level 1 自主性）
type ExecutorAgent struct {
	agent adk.Agent
}

// NewExecutorAgent creates a new ExecutorAgent with Level 1 autonomy constraints
// The executor only executes trades on explicit OKXWatcher commands
func NewExecutorAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*ExecutorAgent, error) {
	// P0 tools for order execution and management
	baseTools := []tool.BaseTool{
		tools.NewOkxPlaceOrderTool(svcCtx),
		tools.NewOkxCancelOrderTool(svcCtx),
		tools.NewOkxGetOrderTool(svcCtx),
		tools.NewOkxGetOrderListTool(svcCtx),
		tools.NewOkxGetAllPositionsTool(svcCtx),
		tools.NewOkxAttachSlTpTool(svcCtx),
		tools.NewOkxPlaceOrderWithSlTpTool(svcCtx),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ExecutorAgent",
		Description: DESCRIPTION,
		Model:       svcCtx.ChatModel,
		Instruction: SOUL,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: baseTools,
			},
			EmitInternalEvents: true,
		},
		MaxIterations: 100,
	})
	if err != nil {
		return nil, err
	}

	return &ExecutorAgent{agent: agent}, nil
}

// Agent returns the underlying adk.Agent
func (e *ExecutorAgent) Agent() adk.Agent {
	return e.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
