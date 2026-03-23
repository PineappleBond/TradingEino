package okx_watcher

import (
	"context"
	"sync"

	_ "embed"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

var okxWatcher adk.Agent
var mux sync.Mutex

func OkxWatcher() adk.Agent {
	return okxWatcher
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string

func Init(ctx context.Context, svcCtx *svc.ServiceContext, subAgents ...adk.Agent) error {
	var err error
	baseTools := make([]tool.BaseTool, 0)
	baseTools = append(baseTools, tools.NewOkxCandlesticksTool(svcCtx))
	okxWatcher, err = deep.New(ctx, &deep.Config{
		Name:        "OKXWatcher",
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
			EmitInternalEvents: false,
		},
		MaxIteration:                 0,
		WithoutWriteTodos:            false,
		WithoutGeneralSubAgent:       false,
		TaskToolDescriptionGenerator: nil,
		Middlewares:                  nil,
	})
	if err != nil {
		logger.Error(ctx, "InitOkxWatcher error", err)
	}
	return err
}
