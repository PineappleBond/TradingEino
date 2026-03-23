package scheduler

// RegisterDefaultHandlers 注册所有默认的 Handler
// 可以根据需要添加或移除
func (s *Scheduler) RegisterDefaultHandlers() {
	// 示例：注册示例 Handler
	// s.RegisterHandler(handlers.NewExampleHandler())

	// TODO: 在这里注册实际的 Handler
	// 例如:
	// s.RegisterHandler(handlers.NewOkexSyncHandler())
	// s.RegisterHandler(handlers.NewDataCleanupHandler())
}

// GetAvailableHandlers 返回所有已注册的 Handler 名称
func (s *Scheduler) GetAvailableHandlers() []string {
	// 这个方法主要用于调试和管理界面
	return nil
}
