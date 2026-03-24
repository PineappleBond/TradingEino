package risk_officer

import (
	"context"
	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// RiskOfficerAgent 风控专家 Agent（普通 ChatModelAgent，不是 DeepAgent）
type RiskOfficerAgent struct {
	agent adk.Agent
}

func NewRiskOfficerAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*RiskOfficerAgent, error) {
	baseTools := []tool.BaseTool{
		tools.NewOkxGetPositionsTool(svcCtx),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "RiskOfficer",
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

	return &RiskOfficerAgent{agent: agent}, nil
}

func (r *RiskOfficerAgent) Agent() adk.Agent {
	return r.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
