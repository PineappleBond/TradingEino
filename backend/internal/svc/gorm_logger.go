package svc

import (
	"context"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	gormlogger "gorm.io/gorm/logger"
)

// gormLogger 实现了 gorm.logger.Interface 接口，使用项目自定的 logger
type gormLogger struct {
	logger   *logger.Logger
	logLevel gormlogger.LogLevel
}

// newGormLogger 创建一个新的 GORM logger
func newGormLogger(logger *logger.Logger, logLevel gormlogger.LogLevel) *gormLogger {
	return &gormLogger{
		logger:   logger,
		logLevel: logLevel,
	}
}

// LogMode 设置日志级别模式
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormLogger{
		logger:   l.logger,
		logLevel: level,
	}
}

// Info 打印 info 日志
func (l *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.logLevel >= gormlogger.Info {
		l.logger.Infof(ctx, "[GORM] "+msg, args...)
	}
}

// Warn 打印 warning 日志
func (l *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.logLevel >= gormlogger.Warn {
		l.logger.Warnf(ctx, "[GORM] "+msg, args...)
	}
}

// Error 打印 error 日志
func (l *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.logLevel >= gormlogger.Error {
		l.logger.Errorf(ctx, "[GORM] "+msg, nil, args...)
	}
}

// Trace 打印 SQL 跟踪日志
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)

	if l.logLevel >= gormlogger.Error && err != nil {
		sql, rows := fc()
		l.logger.Errorf(ctx, "[GORM] SQL error: %v, duration: %v, rows: %d, sql: %s", err, elapsed, rows, sql)
		return
	}

	if l.logLevel >= gormlogger.Warn && elapsed > 1*time.Second {
		sql, rows := fc()
		l.logger.Warnf(ctx, "[GORM] slow SQL: duration: %v, rows: %d, sql: %s", elapsed, rows, sql)
		return
	}

	if l.logLevel >= gormlogger.Info {
		sql, rows := fc()
		l.logger.Infof(ctx, "[GORM] SQL: duration: %v, rows: %d, sql: %s", elapsed, rows, sql)
	}
}
