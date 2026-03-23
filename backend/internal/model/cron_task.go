package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// TaskStatus 定时任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待执行
	TaskStatusRunning   TaskStatus = "running"   // 执行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成（once 类型）
	TaskStatusStopped   TaskStatus = "stopped"   // 已停止
	TaskStatusFailed    TaskStatus = "failed"    // 失败
)

// TaskType 定时任务类型
type TaskType string

const (
	TaskTypeOnce      TaskType = "once"      // 一次性任务
	TaskTypeRecurring TaskType = "recurring" // 重复执行任务
)

// CronTask 定时任务
type CronTask struct {
	ID uint `gorm:"primarykey"`

	Name            string     `gorm:"type:varchar(100);not null;comment:'任务名称'"`
	Spec            string     `gorm:"type:varchar(50);comment:'cron 表达式'"`
	Type            TaskType   `gorm:"type:varchar(20);not null;comment:'任务类型 (once/recurring)'"`
	Status          TaskStatus `gorm:"type:varchar(20);not null;default:pending;comment:'任务状态'"`
	ExecType        string     `gorm:"type:varchar(50);not null;comment:'执行类型'"`
	Raw             string     `gorm:"type:text;comment:'原始数据 (JSON 格式)'"`
	ValidFrom       sql.NullTime `gorm:"comment:'有效期开始'"`
	ValidUntil      sql.NullTime `gorm:"comment:'有效期结束'"`
	Enabled         bool       `gorm:"not null;default:true;comment:'是否启用'"`
	MaxRetries      int        `gorm:"not null;default:0;comment:'最大重试次数'"`
	LastExecutedAt  sql.NullTime `gorm:"comment:'上次执行时间'"`
	NextExecutionAt sql.NullTime `gorm:"comment:'下次执行时间'"`
	TotalExecutions uint       `gorm:"not null;default:0;comment:'累计执行次数'"`

	CreatedAt time.Time `gorm:"index:,sort:desc;"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CronTask) TableName() string {
	return "cron_task"
}
