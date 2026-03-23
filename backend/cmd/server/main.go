package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/agent"
	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/service/scheduler"
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
		"log_output", cfg.Logger.Output,
	)

	svcCtx := svc.NewServiceContext(*cfg)

	err = agent.InitAgents(svcCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init agents: %v\n", err)
		os.Exit(1)
	}

	sch := scheduler.NewScheduler(svcCtx)
	sch.RegisterDefaultHandlers()
	err = sch.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start scheduler: %v\n", err)
		os.Exit(1)
	}

	logger.Info(ctx, "server initialized successfully")

	time.Sleep(time.Hour * 24 * 360)
}
