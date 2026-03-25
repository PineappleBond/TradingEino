package request

// ListLogsRequest 获取日志记录列表请求
type ListLogsRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	ExecutionID *uint `form:"execution_id"`
	Level      *string `form:"level"`
	From       *string `form:"from"`
}

// GetLogRequest 获取日志记录详情请求
type GetLogRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// GetByExecutionIDRequest 根据执行 ID 获取日志列表请求
type GetByExecutionIDRequest struct {
	ExecutionID uint `uri:"execution_id" binding:"required,min=1"`
	Page        int  `form:"page"`
	PageSize    int  `form:"pageSize"`
}
