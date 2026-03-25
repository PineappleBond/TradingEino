package position_manager

import (
	"context"
	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// PositionManagerAgent 持仓管理专家 Agent（普通 ChatModelAgent，不是 DeepAgent）
type PositionManagerAgent struct {
	agent adk.Agent
}

// NewPositionManagerAgent 创建持仓管理 Agent
func NewPositionManagerAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*PositionManagerAgent, error) {
	baseTools := []tool.BaseTool{
		tools.NewOkxGetPositionsTool(svcCtx),
		tools.NewOkxAccountBalanceTool(svcCtx),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "PositionManager",
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

	return &PositionManagerAgent{agent: agent}, nil
}

// Agent 返回内部 agent 实例
func (r *PositionManagerAgent) Agent() adk.Agent {
	return r.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
