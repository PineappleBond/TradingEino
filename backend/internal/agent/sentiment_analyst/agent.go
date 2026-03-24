package sentiment_analyst

import (
	"context"
	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// SentimentAnalystAgent 情绪分析师 Agent（普通 ChatModelAgent，不是 DeepAgent）
type SentimentAnalystAgent struct {
	agent adk.Agent
}

func NewSentimentAnalystAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*SentimentAnalystAgent, error) {
	baseTools := []tool.BaseTool{
		tools.NewOkxGetFundingRateTool(svcCtx),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SentimentAnalyst",
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

	return &SentimentAnalystAgent{agent: agent}, nil
}

func (s *SentimentAnalystAgent) Agent() adk.Agent {
	return s.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
