package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
)

var configPath = flag.String("c", "etc/config.yaml", "path to config file")

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create and set default logger
	log := logger.New(cfg.Logger, 4)
	logger.SetDefault(log)
	logger.SetGlobalDefault(log)

	// Log startup info
	ctx := context.Background()
	logger.Info(ctx, "application starting",
		"config_path", *configPath,
		"log_level", cfg.Logger.Level,
		"log_format", cfg.Logger.Format,
		"log_output", cfg.Logger.Output,
	)

	svcCtx := svc.NewServiceContext(*cfg)

	_ = svcCtx

	logger.Info(ctx, "server initialized successfully")
}
