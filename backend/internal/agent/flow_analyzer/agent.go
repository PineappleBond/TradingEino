package flow_analyzer

import (
	"context"
	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// FlowAnalyzerAgent - 订单流分析师 Agent (ChatModelAgent，不是 DeepAgent)
type FlowAnalyzerAgent struct {
	agent adk.Agent
}

func NewFlowAnalyzerAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*FlowAnalyzerAgent, error) {
	baseTools := []tool.BaseTool{
		tools.NewOkxOrderbookTool(svcCtx),
		tools.NewOkxTradesHistoryTool(svcCtx),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "FlowAnalyzer",
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

	return &FlowAnalyzerAgent{agent: agent}, nil
}

func (f *FlowAnalyzerAgent) Agent() adk.Agent {
	return f.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
