package agent

import (
	"context"
	"sync"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/executor_agent"
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
	Executor         adk.Agent
	mux              sync.Mutex
	ctx              context.Context
	cancel           context.CancelFunc
}

var (
	agentsOnce sync.Once
	_agents    *AgentsModel
)

// Agents returns the global AgentsModel instance
func Agents() *AgentsModel {
	return _agents
}

// InitAgents initializes all agents with proper context propagation
// Uses sync.Once to ensure single initialization
func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
	var initErr error
	agentsOnce.Do(func() {
		// Derive child context from parent ctx (not context.Background())
		ctx, cancel := context.WithCancel(ctx)

		// Initialize RiskOfficer agent (ChatModelAgent)
		riskOfficerAgent, err := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
		if err != nil {
			initErr = err
			cancel()
			return
		}

		// Initialize SentimentAnalyst agent (ChatModelAgent)
		sentimentAnalystAgent, err := sentiment_analyst.NewSentimentAnalystAgent(ctx, svcCtx)
		if err != nil {
			initErr = err
			cancel()
			return
		}

		// Initialize OKXWatcher agent (DeepAgent orchestrator)
		okxWatcherAgent, err := okx_watcher.NewOkxWatcherAgent(ctx, svcCtx, riskOfficerAgent.Agent(), sentimentAnalystAgent.Agent())
		if err != nil {
			initErr = err
			cancel()
			return
		}

		// Initialize Executor agent (ChatModelAgent for trade execution)
		executorAgent, err := executor_agent.NewExecutorAgent(ctx, svcCtx)
		if err != nil {
			initErr = err
			cancel()
			return
		}

		_agents = &AgentsModel{
			svcCtx:           svcCtx,
			OkxWatcher:       okxWatcherAgent.Agent(),
			RiskOfficer:      riskOfficerAgent.Agent(),
			SentimentAnalyst: sentimentAnalystAgent.Agent(),
			Executor:         executorAgent.Agent(),
			mux:              sync.Mutex{},
			ctx:              ctx,
			cancel:           cancel,
		}
	})
	return initErr
}

// Close cleans up resources held by AgentsModel
func (a *AgentsModel) Close() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}
