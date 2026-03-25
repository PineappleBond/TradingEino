package model

import (
	"time"

	"gorm.io/gorm"
)

// CronExecutionLog 定时任务执行日志
type CronExecutionLog struct {
	ID uint `gorm:"primarykey"`

	ExecutionID uint   `gorm:"not null;index;comment:'关联执行记录 ID'"`
	From        string `gorm:"type:varchar(100);not null;comment:'日志来源阶段'"`
	Level       string `gorm:"type:varchar(20);not null;comment:'日志级别 (info/warn/error/debug)'"`
	Message     string `gorm:"type:text;not null;comment:'日志内容'"`

	CreatedAt time.Time      `gorm:"index:,sort:desc;"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CronExecutionLog) TableName() string {
	return "cron_execution_log"
}
