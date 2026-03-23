package svc

import (
	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/api"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config    config.Config
	DB        *gorm.DB
	ChatModel *openai.ChatModel
	OKXClient *api.Client
}

func NewServiceContext(cfg config.Config) *ServiceContext {
	log := logger.New(config.LoggerConfig{
		Level:     cfg.Logger.Level,
		Output:    cfg.Logger.Output,
		FilePath:  cfg.Logger.DBLogPath(),
		AddSource: cfg.Logger.AddSource,
	}, 5)

	s := &ServiceContext{
		DB:        mustInitDB(cfg, log),
		Config:    cfg,
		ChatModel: mustInitChatModel(cfg),
		OKXClient: mustInitOKXClient(cfg),
	}
	return s
}
