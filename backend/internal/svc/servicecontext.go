package svc

import (
	"path/filepath"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"gorm.io/gorm"
)

type ServiceContext struct {
	config config.Config
	DB     *gorm.DB
}

func NewServiceContext(cfg config.Config) *ServiceContext {
	filePath := cfg.Logger.FilePath
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	dbLogFile := filepath.Join(dir, "db.log"+ext)
	log := logger.New(config.LoggerConfig{
		Level:     cfg.Logger.Level,
		Format:    cfg.Logger.Format,
		Output:    cfg.Logger.Output,
		FilePath:  dbLogFile,
		AddSource: cfg.Logger.AddSource,
	}, 5)
	s := &ServiceContext{
		DB:     mustInitDB(cfg, log),
		config: cfg,
	}
	return s
}
