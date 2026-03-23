package svc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// mustInitDB 初始化数据库连接
func mustInitDB(cfg config.Config, log *logger.Logger) *gorm.DB {
	if cfg.DB.Type != "sqlite" {
		fmt.Fprintf(os.Stderr, "db type %s not supported\n", cfg.DB.Type)
		os.Exit(1)
		return nil
	}

	dir := filepath.Dir(cfg.DB.DBPath)
	_ = os.MkdirAll(dir, os.ModePerm)

	// 使用无 CGO 的 sqlite3 驱动 (ncruces/go-sqlite3)
	logLevel := gormlogger.Info
	if cfg.Logger.Level == "warn" {
		logLevel = gormlogger.Warn
	} else if cfg.Logger.Level == "error" {
		logLevel = gormlogger.Error
	}
	db, err := gorm.Open(gormlite.Open(cfg.DB.DBPath), &gorm.Config{
		Logger: newGormLogger(log, logLevel),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init gorm: %v\n", err)
		os.Exit(1)
		return nil
	}

	err = db.AutoMigrate(
		&model.CronExecution{},
		&model.CronExecutionLog{},
		&model.CronTask{},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate: %v\n", err)
		os.Exit(1)
	}

	return db
}
