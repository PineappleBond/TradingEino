package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
)

type CronExecutionRepository struct {
	svcCtx *svc.ServiceContext
}

func NewCronExecutionRepository(svcCtx *svc.ServiceContext) *CronExecutionRepository {
	return &CronExecutionRepository{svcCtx: svcCtx}
}

// Create 创建新的 CronExecution 记录
func (r *CronExecutionRepository) Create(ctx context.Context, execution *model.CronExecution) error {
	return r.svcCtx.DB.WithContext(ctx).Create(execution).Error
}

// GetByID 根据 ID 获取 CronExecution 记录
func (r *CronExecutionRepository) GetByID(ctx context.Context, id uint) (*model.CronExecution, error) {
	var execution model.CronExecution
	err := r.svcCtx.DB.WithContext(ctx).First(&execution, id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// GetByTaskID 根据 TaskID 获取 CronExecution 记录列表
func (r *CronExecutionRepository) GetByTaskID(ctx context.Context, taskID uint) ([]*model.CronExecution, error) {
	var executions []*model.CronExecution
	err := r.svcCtx.DB.WithContext(ctx).Where("task_id = ?", taskID).Order("scheduled_at DESC").Find(&executions).Error
	return executions, err
}

// GetPendingExecutions 获取所有待执行的 CronExecution 记录
func (r *CronExecutionRepository) GetPendingExecutions(ctx context.Context, before time.Time) ([]*model.CronExecution, error) {
	var executions []*model.CronExecution
	err := r.svcCtx.DB.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", model.ExecutionStatusPending, before).
		Order("scheduled_at ASC").
		Find(&executions).Error
	return executions, err
}

// GetRunningExecutions 获取所有执行中的 CronExecution 记录
func (r *CronExecutionRepository) GetRunningExecutions(ctx context.Context) ([]*model.CronExecution, error) {
	var executions []*model.CronExecution
	err := r.svcCtx.DB.WithContext(ctx).
		Where("status = ?", model.ExecutionStatusRunning).
		Order("started_at ASC").
		Find(&executions).Error
	return executions, err
}

// Update 更新 CronExecution 记录
func (r *CronExecutionRepository) Update(ctx context.Context, execution *model.CronExecution) error {
	return r.svcCtx.DB.WithContext(ctx).Save(execution).Error
}

// MarkAsRunning 标记为执行中
func (r *CronExecutionRepository) MarkAsRunning(ctx context.Context, id uint) error {
	now := time.Now()
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.ExecutionStatusRunning,
			"started_at": now,
		}).Error
}

// MarkAsSuccess 标记为成功
func (r *CronExecutionRepository) MarkAsSuccess(ctx context.Context, id uint, completedAt time.Time) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ExecutionStatusSuccess,
			"completed_at": completedAt,
		}).Error
}

// MarkAsFailed 标记为失败
func (r *CronExecutionRepository) MarkAsFailed(ctx context.Context, id uint, completedAt time.Time, errMsg string) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ExecutionStatusFailed,
			"completed_at": completedAt,
			"error":        errMsg,
		}).Error
}

// IncrementRetryCount 增加重试次数
func (r *CronExecutionRepository) IncrementRetryCount(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + ?", 1)).Error
}

// MarkAsRetried 标记为已重试
func (r *CronExecutionRepository) MarkAsRetried(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": model.ExecutionStatusRetried,
		}).Error
}

// MarkAsCancelled 标记为已取消
func (r *CronExecutionRepository) MarkAsCancelled(ctx context.Context, id uint) error {
	now := time.Now()
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ExecutionStatusCancelled,
			"completed_at": now,
		}).Error
}

// Delete 删除 CronExecution 记录 (软删除)
func (r *CronExecutionRepository) Delete(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Delete(&model.CronExecution{}, id).Error
}

// DeleteByTaskID 根据 TaskID 删除所有 CronExecution 记录 (软删除)
func (r *CronExecutionRepository) DeleteByTaskID(ctx context.Context, taskID uint) error {
	return r.svcCtx.DB.WithContext(ctx).Where("task_id = ?", taskID).Delete(&model.CronExecution{}).Error
}

// GetOverdueRunning 获取超时未完成的执行记录 (超过指定时间仍处于 running 状态)
func (r *CronExecutionRepository) GetOverdueRunning(ctx context.Context, timeout time.Duration) ([]*model.CronExecution, error) {
	cutoff := time.Now().Add(-timeout)
	var executions []*model.CronExecution
	err := r.svcCtx.DB.WithContext(ctx).
		Where("status = ? AND started_at < ? AND started_at IS NOT NULL", model.ExecutionStatusRunning, cutoff).
		Find(&executions).Error
	return executions, err
}

// GetPagedExecutions 分页获取执行记录列表，支持任务 ID、状态、时间范围过滤
func (r *CronExecutionRepository) GetPagedExecutions(ctx context.Context, page, pageSize int, taskID *uint, status *model.ExecutionStatus, startTime, endTime *time.Time) ([]*model.CronExecution, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := r.svcCtx.DB.WithContext(ctx).Model(&model.CronExecution{})

	// Apply filters
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if startTime != nil {
		query = query.Where("scheduled_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("scheduled_at <= ?", *endTime)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paged data
	var executions []*model.CronExecution
	offset := (page - 1) * pageSize
	err := query.Order("scheduled_at DESC").Offset(offset).Limit(pageSize).Find(&executions).Error
	return executions, total, err
}

// GetByExecutionID 根据 ID 获取单条执行记录（带关联日志）
func (r *CronExecutionRepository) GetByExecutionID(ctx context.Context, id uint) (*model.CronExecution, error) {
	var execution model.CronExecution
	err := r.svcCtx.DB.WithContext(ctx).First(&execution, id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}
