package scheduler

import (
	"github.com/PineappleBond/TradingEino/backend/internal/service/scheduler/handlers"
)

// RegisterDefaultHandlers 注册所有默认的 Handler
// 可以根据需要添加或移除
func (s *Scheduler) RegisterDefaultHandlers() {
	s.RegisterHandler(handlers.NewOKXWatcherHandler(s.svcCtx))
}

// GetAvailableHandlers 返回所有已注册的 Handler 名称
func (s *Scheduler) GetAvailableHandlers() []string {
	// 这个方法主要用于调试和管理界面
	return nil
}
