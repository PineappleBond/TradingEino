package response

import (
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

// CronExecutionResponse CronExecution 的 API 响应结构体
type CronExecutionResponse struct {
	ID          uint       `json:"id"`
	TaskID      uint       `json:"task_id"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      string     `json:"status"`
	RetryCount  int        `json:"retry_count"`
	Error       string     `json:"error"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToCronExecutionResponse 将 model.CronExecution 转换为 CronExecutionResponse
func ToCronExecutionResponse(execution *model.CronExecution) *CronExecutionResponse {
	resp := &CronExecutionResponse{
		ID:          execution.ID,
		TaskID:      execution.TaskID,
		ScheduledAt: execution.ScheduledAt,
		Status:      string(execution.Status),
		RetryCount:  execution.RetryCount,
		Error:       execution.Error,
		CreatedAt:   execution.CreatedAt,
		UpdatedAt:   execution.UpdatedAt,
	}

	if execution.StartedAt.Valid {
		resp.StartedAt = &execution.StartedAt.Time
	}
	if execution.CompletedAt.Valid {
		resp.CompletedAt = &execution.CompletedAt.Time
	}

	return resp
}

// ToCronExecutionListResponse 批量转换 CronExecution 为响应格式
func ToCronExecutionListResponse(executions []*model.CronExecution) []*CronExecutionResponse {
	result := make([]*CronExecutionResponse, 0, len(executions))
	for _, execution := range executions {
		result = append(result, ToCronExecutionResponse(execution))
	}
	return result
}
