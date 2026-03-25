package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"   // 待执行
	ExecutionStatusRunning   ExecutionStatus = "running"   // 执行中
	ExecutionStatusSuccess   ExecutionStatus = "success"   // 成功
	ExecutionStatusFailed    ExecutionStatus = "failed"    // 失败
	ExecutionStatusRetried   ExecutionStatus = "retried"   // 已重试
	ExecutionStatusCancelled ExecutionStatus = "cancelled" // 已取消
)

// CronExecution 定时任务执行记录
type CronExecution struct {
	ID uint `gorm:"primarykey"`

	TaskID      uint            `gorm:"not null;index;comment:'关联任务 ID'"`
	ScheduledAt time.Time       `gorm:"not null;index;comment:'计划执行时间'"`
	StartedAt   sql.NullTime    `gorm:"comment:'实际开始时间'"`
	CompletedAt sql.NullTime    `gorm:"comment:'完成时间'"`
	Status      ExecutionStatus `gorm:"type:varchar(20);not null;default:pending;comment:'执行状态'"`
	RetryCount  int             `gorm:"not null;default:0;comment:'重试次数'"`
	Error       string          `gorm:"type:text;comment:'错误信息'"`

	CreatedAt time.Time      `gorm:"index:,sort:desc;"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CronExecution) TableName() string {
	return "cron_execution"
}
