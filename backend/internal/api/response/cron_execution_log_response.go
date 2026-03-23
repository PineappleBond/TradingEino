package response

import (
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

// CronExecutionLogResponse CronExecutionLog 的 API 响应结构体
type CronExecutionLogResponse struct {
	ID          uint      `json:"id"`
	ExecutionID uint      `json:"execution_id"`
	From        string    `json:"from"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToCronExecutionLogResponse 将 model.CronExecutionLog 转换为 CronExecutionLogResponse
func ToCronExecutionLogResponse(log *model.CronExecutionLog) *CronExecutionLogResponse {
	return &CronExecutionLogResponse{
		ID:          log.ID,
		ExecutionID: log.ExecutionID,
		From:        log.From,
		Level:       log.Level,
		Message:     log.Message,
		CreatedAt:   log.CreatedAt,
		UpdatedAt:   log.UpdatedAt,
	}
}

// ToCronExecutionLogListResponse 批量转换 CronExecutionLog 为响应格式
func ToCronExecutionLogListResponse(logs []*model.CronExecutionLog) []*CronExecutionLogResponse {
	result := make([]*CronExecutionLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, ToCronExecutionLogResponse(log))
	}
	return result
}
