package response

import (
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/cronutil"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

// CronTaskResponse CronTask 的 API 响应结构体
type CronTaskResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Spec            string     `json:"spec"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	ExecType        string     `json:"exec_type"`
	Raw             string     `json:"raw"`
	ValidFrom       *time.Time `json:"valid_from,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	Enabled         bool       `json:"enabled"`
	MaxRetries      int        `json:"max_retries"`
	TimeoutSeconds  int        `json:"timeout_seconds"`
	LastExecutedAt  *time.Time `json:"last_executed_at,omitempty"`
	NextExecutionAt *time.Time `json:"next_execution_at,omitempty"`
	TotalExecutions uint       `json:"total_executions"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToCronTaskResponse 将 model.CronTask 转换为 CronTaskResponse
// 对于周期性任务 (recurring)，会自动计算下次执行时间
func ToCronTaskResponse(task *model.CronTask) *CronTaskResponse {
	resp := &CronTaskResponse{
		ID:              task.ID,
		Name:            task.Name,
		Spec:            task.Spec,
		Type:            string(task.Type),
		Status:          string(task.Status),
		ExecType:        task.ExecType,
		Raw:             task.Raw,
		Enabled:         task.Enabled,
		MaxRetries:      task.MaxRetries,
		TimeoutSeconds:  task.TimeoutSeconds,
		TotalExecutions: task.TotalExecutions,
		CreatedAt:       task.CreatedAt,
		UpdatedAt:       task.UpdatedAt,
	}

	if task.ValidFrom.Valid {
		resp.ValidFrom = &task.ValidFrom.Time
	}
	if task.ValidUntil.Valid {
		resp.ValidUntil = &task.ValidUntil.Time
	}
	if task.LastExecutedAt.Valid {
		resp.LastExecutedAt = &task.LastExecutedAt.Time
	}
	if task.NextExecutionAt.Valid {
		resp.NextExecutionAt = &task.NextExecutionAt.Time
	}

	// 对于周期性任务，如果数据库中未设置下次执行时间，计算下次执行时间
	if task.Type == model.TaskTypeRecurring && resp.NextExecutionAt == nil {
		nextExec, err := cronutil.GetNextExecutionTime(task.Spec)
		if err == nil {
			resp.NextExecutionAt = &nextExec
		}
	}

	return resp
}

// ToCronTaskListResponse 批量转换 CronTask 为响应格式
func ToCronTaskListResponse(tasks []*model.CronTask) []*CronTaskResponse {
	result := make([]*CronTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, ToCronTaskResponse(task))
	}
	return result
}
