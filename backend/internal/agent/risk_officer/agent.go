package risk_officer

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

var riskOfficer adk.Agent

func RiskOfficer() adk.Agent {
	return riskOfficer
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string

func Init(ctx context.Context, svcCtx *svc.ServiceContext, subAgents ...adk.Agent) error {
	var err error
	baseTools := make([]tool.BaseTool, 0)
	baseTools = append(baseTools, tools.NewOkxGetPositionsTool(svcCtx))
	riskOfficer, err = deep.New(ctx, &deep.Config{
		Name:        "RiskOfficer",
		Description: DESCRIPTION,
		ChatModel:   svcCtx.ChatModel,
		Instruction: SOUL,
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
			EmitInternalEvents: true,
		},
		MaxIteration:                 0,
		WithoutWriteTodos:            false,
		WithoutGeneralSubAgent:       false,
		TaskToolDescriptionGenerator: nil,
		Middlewares:                  nil,
	})
	if err != nil {
		logger.Error(ctx, "InitRiskOfficer error", err)
	}
	return err
}
