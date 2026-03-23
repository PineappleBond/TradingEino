package request

import (
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

// ListTasksRequest 获取任务列表请求
type ListTasksRequest struct {
	Page     int               `form:"page"`
	PageSize int               `form:"pageSize"`
	Status   *model.TaskStatus `form:"status"`
	Type     *model.TaskType   `form:"type"`
	Enabled  *bool             `form:"enabled"`
}

// GetTaskRequest 获取任务详情请求
type GetTaskRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Name            string         `json:"name" binding:"required,max=100"`
	Spec            string         `json:"spec" binding:"omitempty,max=50"`
	Type            model.TaskType `json:"type" binding:"required,oneof=once recurring"`
	ExecType        string         `json:"exec_type" binding:"required,max=50"`
	Raw             string         `json:"raw" binding:"omitempty"`
	ValidFrom       *time.Time     `json:"valid_from"`
	ValidUntil      *time.Time     `json:"valid_until"`
	Enabled         bool           `json:"enabled"`
	MaxRetries      int            `json:"max_retries"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	NextExecutionAt *time.Time     `json:"next_execution_at"`
}

// UpdateTaskRequest 更新任务请求（URI 参数）
type UpdateTaskRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// UpdateTaskBody 更新任务请求体
type UpdateTaskBody struct {
	Name            string         `json:"name" binding:"omitempty,max=100"`
	Spec            string         `json:"spec" binding:"omitempty,max=50"`
	Type            model.TaskType `json:"type" binding:"omitempty,oneof=once recurring"`
	ExecType        string         `json:"exec_type" binding:"omitempty,max=50"`
	Raw             string         `json:"raw" binding:"omitempty"`
	ValidFrom       *time.Time     `json:"valid_from"`
	ValidUntil      *time.Time     `json:"valid_until"`
	Enabled         *bool          `json:"enabled"`
	MaxRetries      int            `json:"max_retries"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	NextExecutionAt *time.Time     `json:"next_execution_at"`
}

// DeleteTaskRequest 删除任务请求
type DeleteTaskRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// TaskActionRequest 任务操作请求
type TaskActionRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// StartTaskRequest 启动任务请求
type StartTaskRequest struct {
	ID                uint   `uri:"id" binding:"required,min=1"`
	NextExecutionTime string `json:"next_execution_time" binding:"required"`
}
