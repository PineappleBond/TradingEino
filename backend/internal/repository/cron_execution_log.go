package repository

import (
	"context"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
)

type CronExecutionLogRepository struct {
	svcCtx *svc.ServiceContext
}

func NewCronExecutionLogRepository(svcCtx *svc.ServiceContext) *CronExecutionLogRepository {
	return &CronExecutionLogRepository{svcCtx: svcCtx}
}

// Create 创建新的 CronExecutionLog 记录
func (r *CronExecutionLogRepository) Create(ctx context.Context, log *model.CronExecutionLog) error {
	return r.svcCtx.DB.WithContext(ctx).Create(log).Error
}

// GetByID 根据 ID 获取 CronExecutionLog 记录
func (r *CronExecutionLogRepository) GetByID(ctx context.Context, id uint) (*model.CronExecutionLog, error) {
	var log model.CronExecutionLog
	err := r.svcCtx.DB.WithContext(ctx).First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByExecutionID 根据 ExecutionID 获取 CronExecutionLog 记录列表
func (r *CronExecutionLogRepository) GetByExecutionID(ctx context.Context, executionID uint) ([]*model.CronExecutionLog, error) {
	var logs []*model.CronExecutionLog
	err := r.svcCtx.DB.WithContext(ctx).Where("execution_id = ?", executionID).Order("created_at ASC").Find(&logs).Error
	return logs, err
}

// GetByLevel 根据日志级别获取 CronExecutionLog 记录列表
func (r *CronExecutionLogRepository) GetByLevel(ctx context.Context, level string, limit int) ([]*model.CronExecutionLog, error) {
	var logs []*model.CronExecutionLog
	query := r.svcCtx.DB.WithContext(ctx).Where("level = ?", level)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetRecentLogs 获取最近的日志记录
func (r *CronExecutionLogRepository) GetRecentLogs(ctx context.Context, limit int) ([]*model.CronExecutionLog, error) {
	var logs []*model.CronExecutionLog
	if limit <= 0 {
		limit = 100
	}
	err := r.svcCtx.DB.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// GetErrorLogs 获取错误日志
func (r *CronExecutionLogRepository) GetErrorLogs(ctx context.Context, since time.Time) ([]*model.CronExecutionLog, error) {
	var logs []*model.CronExecutionLog
	err := r.svcCtx.DB.WithContext(ctx).
		Where("level IN ? AND created_at >= ?", []string{"error", "warn"}, since).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// CreateBatch 批量创建日志记录
func (r *CronExecutionLogRepository) CreateBatch(ctx context.Context, logs []*model.CronExecutionLog) error {
	return r.svcCtx.DB.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// Delete 删除 CronExecutionLog 记录 (软删除)
func (r *CronExecutionLogRepository) Delete(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Delete(&model.CronExecutionLog{}, id).Error
}

// DeleteByExecutionID 根据 ExecutionID 删除所有日志记录 (软删除)
func (r *CronExecutionLogRepository) DeleteByExecutionID(ctx context.Context, executionID uint) error {
	return r.svcCtx.DB.WithContext(ctx).Where("execution_id = ?", executionID).Delete(&model.CronExecutionLog{}).Error
}

// DeleteOlderThan 删除指定时间之前的日志记录
func (r *CronExecutionLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	return r.svcCtx.DB.WithContext(ctx).Where("created_at < ?", before).Delete(&model.CronExecutionLog{}).Error
}

// GetLogsByFrom 根据日志来源获取日志记录
func (r *CronExecutionLogRepository) GetLogsByFrom(ctx context.Context, from string, limit int) ([]*model.CronExecutionLog, error) {
	var logs []*model.CronExecutionLog
	query := r.svcCtx.DB.WithContext(ctx).Where("`from` = ?", from)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}
