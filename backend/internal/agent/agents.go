package agent

import (
	"context"
	"sync"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/okx_watcher"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
)

type AgentsModel struct {
	svcCtx     *svc.ServiceContext
	OkxWatcher adk.Agent
	mux        sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

var _agents *AgentsModel

func Agents() *AgentsModel {
	return _agents
}

func InitAgents(svcCtx *svc.ServiceContext) error {
	ctx, cancel := context.WithCancel(context.Background())
	// TODO 多维度分析Agent
	subAgents := make([]adk.Agent, 0)
	err := okx_watcher.Init(ctx, svcCtx, subAgents...)
	if err != nil {
		cancel()
		return err
	}

	tmp := &AgentsModel{
		svcCtx:     svcCtx,
		OkxWatcher: okx_watcher.OkxWatcher(),
		mux:        sync.Mutex{},
		ctx:        ctx,
		cancel:     cancel,
	}
	_agents = tmp
	return nil
}
