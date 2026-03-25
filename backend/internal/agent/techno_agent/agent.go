package techno_agent

import (
	"context"
	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// TechnoAgent 技术分析 Agent（普通 ChatModelAgent，不是 DeepAgent）
type TechnoAgent struct {
	agent adk.Agent
}

// TechnicalIndicatorsHeaders 技术指标表头（从 tools 包继承）
var TechnicalIndicatorsHeaders = tools.TechnicalIndicatorsHeaders

func NewTechnoAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*TechnoAgent, error) {
	baseTools := []tool.BaseTool{
		tools.NewOkxCandlesticksTool(svcCtx),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TechnoAgent",
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

	return &TechnoAgent{agent: agent}, nil
}

func (t *TechnoAgent) Agent() adk.Agent {
	return t.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
