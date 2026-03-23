package svc

import (
	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"gorm.io/gorm"
)

type ServiceContext struct {
	config config.Config
	DB     *gorm.DB
}

func NewServiceContext(cfg config.Config) *ServiceContext {
	log := logger.New(config.LoggerConfig{
		Level:     cfg.Logger.Level,
		Output:    cfg.Logger.Output,
		FilePath:  cfg.Logger.DBLogPath(),
		AddSource: cfg.Logger.AddSource,
	}, 5)
	s := &ServiceContext{
		DB:     mustInitDB(cfg, log),
		config: cfg,
	}
	return s
}
