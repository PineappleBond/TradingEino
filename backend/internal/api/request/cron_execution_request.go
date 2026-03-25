package request

import (
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

// ListExecutionsRequest 获取执行记录列表请求
type ListExecutionsRequest struct {
	Page      int                    `form:"page"`
	PageSize  int                    `form:"pageSize"`
	TaskID    *uint                  `form:"task_id"`
	Status    *model.ExecutionStatus `form:"status"`
	StartTime *time.Time             `form:"start_time"`
	EndTime   *time.Time             `form:"end_time"`
}

// GetExecutionRequest 获取执行记录详情请求
type GetExecutionRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// GetByTaskIDRequest 获取任务执行记录列表请求
type GetByTaskIDRequest struct {
	TaskID   uint `uri:"task_id" binding:"required,min=1"`
	Page     int  `form:"page"`
	PageSize int  `form:"pageSize"`
}
