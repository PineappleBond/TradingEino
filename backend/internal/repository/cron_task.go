package repository

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
)

type CronTaskRepository struct {
	svcCtx *svc.ServiceContext
}

func NewCronTaskRepository(svcCtx *svc.ServiceContext) *CronTaskRepository {
	return &CronTaskRepository{svcCtx: svcCtx}
}

// Create 创建新的 CronTask 记录
func (r *CronTaskRepository) Create(ctx context.Context, task *model.CronTask) error {
	return r.svcCtx.DB.WithContext(ctx).Create(task).Error
}

// GetByID 根据 ID 获取 CronTask 记录
func (r *CronTaskRepository) GetByID(ctx context.Context, id uint) (*model.CronTask, error) {
	var task model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByName 根据名称获取 CronTask 记录
func (r *CronTaskRepository) GetByName(ctx context.Context, name string) (*model.CronTask, error) {
	var task model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).Where("name = ?", name).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetAll 获取所有 CronTask 记录
func (r *CronTaskRepository) GetAll(ctx context.Context) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetEnabledTasks 获取所有已启用的任务
func (r *CronTaskRepository) GetEnabledTasks(ctx context.Context) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).
		Where("enabled = ? AND (status = ? OR status = ?)", true, model.TaskStatusPending, model.TaskStatusRunning).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetRecurringTasks 获取所有重复执行的任务
func (r *CronTaskRepository) GetRecurringTasks(ctx context.Context) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).
		Where("type = ? AND enabled = ?", model.TaskTypeRecurring, true).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetPendingTasks 获取待执行的任务
func (r *CronTaskRepository) GetPendingTasks(ctx context.Context) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).
		Where("status = ? AND enabled = ?", model.TaskStatusPending, true).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetRunningTasks 获取执行中的任务
func (r *CronTaskRepository) GetRunningTasks(ctx context.Context) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).
		Where("status = ?", model.TaskStatusRunning).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// Update 更新 CronTask 记录
func (r *CronTaskRepository) Update(ctx context.Context, task *model.CronTask) error {
	return r.svcCtx.DB.WithContext(ctx).Save(task).Error
}

// UpdateStatus 更新任务状态
func (r *CronTaskRepository) UpdateStatus(ctx context.Context, id uint, status model.TaskStatus) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// MarkAsRunning 标记任务为执行中
func (r *CronTaskRepository) MarkAsRunning(ctx context.Context, id uint) error {
	now := sql.NullTime{Time: time.Now(), Valid: true}
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           model.TaskStatusRunning,
			"last_executed_at": now,
		}).Error
}

// MarkAsCompleted 标记任务为已完成
func (r *CronTaskRepository) MarkAsCompleted(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("status", model.TaskStatusCompleted).Error
}

// MarkAsStopped 标记任务为已停止
func (r *CronTaskRepository) MarkAsStopped(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("status", model.TaskStatusStopped).Error
}

// MarkAsFailed 标记任务为失败
func (r *CronTaskRepository) MarkAsFailed(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("status", model.TaskStatusFailed).Error
}

// UpdateNextExecution 更新下次执行时间
func (r *CronTaskRepository) UpdateNextExecution(ctx context.Context, id uint, nextAt time.Time) error {
	next := sql.NullTime{Time: nextAt, Valid: true}
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("next_execution_at", next).Error
}

// IncrementTotalExecutions 增加累计执行次数
func (r *CronTaskRepository) IncrementTotalExecutions(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		UpdateColumn("total_executions", gorm.Expr("total_executions + ?", 1)).Error
}

// Enable 启用任务
func (r *CronTaskRepository) Enable(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("enabled", true).Error
}

// Disable 禁用任务
func (r *CronTaskRepository) Disable(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Model(&model.CronTask{}).
		Where("id = ?", id).
		Update("enabled", false).Error
}

// Delete 删除 CronTask 记录 (软删除)
func (r *CronTaskRepository) Delete(ctx context.Context, id uint) error {
	return r.svcCtx.DB.WithContext(ctx).Delete(&model.CronTask{}, id).Error
}

// GetDueTasks 获取到期需要执行的任务 (一次性任务)
func (r *CronTaskRepository) GetDueTasks(ctx context.Context, now time.Time) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).
		Where("type = ? AND status = ? AND enabled = ? AND next_execution_at <= ? AND next_execution_at IS NOT NULL",
			model.TaskTypeOnce, model.TaskStatusPending, true, now).
		Find(&tasks).Error
	return tasks, err
}

// GetTasksDueForExecution 获取需要执行的重复任务
func (r *CronTaskRepository) GetTasksDueForExecution(ctx context.Context, now time.Time) ([]*model.CronTask, error) {
	var tasks []*model.CronTask
	err := r.svcCtx.DB.WithContext(ctx).
		Where("type = ? AND (status = ? OR status = ?) AND enabled = ? AND next_execution_at <= ? AND next_execution_at IS NOT NULL",
			model.TaskTypeRecurring, model.TaskStatusPending, model.TaskStatusRunning, true, now).
		Find(&tasks).Error
	return tasks, err
}
