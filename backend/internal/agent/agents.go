package agent

import (
	"context"
	"sync"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/okx_watcher"
	"github.com/PineappleBond/TradingEino/backend/internal/agent/risk_officer"
	"github.com/PineappleBond/TradingEino/backend/internal/agent/sentiment_analyst"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
)

type AgentsModel struct {
	svcCtx           *svc.ServiceContext
	OkxWatcher       adk.Agent
	RiskOfficer      adk.Agent
	SentimentAnalyst adk.Agent
	mux              sync.Mutex
	ctx              context.Context
	cancel           context.CancelFunc
}

var _agents *AgentsModel

func Agents() *AgentsModel {
	return _agents
}

func InitAgents(svcCtx *svc.ServiceContext) error {
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化子 Agent（普通 ChatModelAgent）
	riskOfficerAgent, err := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
	if err != nil {
		cancel()
		return err
	}

	sentimentAnalystAgent, err := sentiment_analyst.NewSentimentAnalystAgent(ctx, svcCtx)
	if err != nil {
		cancel()
		return err
	}

	// 初始化顶层 DeepAgent（编排器）
	okxWatcherAgent, err := okx_watcher.NewOkxWatcherAgent(ctx, svcCtx, riskOfficerAgent.Agent(), sentimentAnalystAgent.Agent())
	if err != nil {
		cancel()
		return err
	}

	_agents = &AgentsModel{
		svcCtx:           svcCtx,
		OkxWatcher:       okxWatcherAgent.Agent(),
		RiskOfficer:      riskOfficerAgent.Agent(),
		SentimentAnalyst: sentimentAnalystAgent.Agent(),
		mux:              sync.Mutex{},
		ctx:              ctx,
		cancel:           cancel,
	}
	return nil
}
