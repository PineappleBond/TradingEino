package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/PineappleBond/TradingEino/backend/internal/agent"
	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/server"
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

	err = agent.InitAgents(ctx, svcCtx)
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

	serve := server.NewServer(svcCtx)

	// Start HTTP server in a goroutine (ListenAndServe is blocking)
	go func() {
		if err := serve.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "failed to start server", err)
			os.Exit(1)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info(ctx, "server started, waiting for shutdown signal")

	// Wait for shutdown signal
	<-sigChan
	logger.Info(ctx, "shutdown signal received")

	// Ordered shutdown: Server -> Scheduler -> Agents -> DB -> Logger

	// 1. Stop HTTP server
	if err := serve.Shutdown(ctx); err != nil {
		logger.Error(ctx, "failed to shutdown server", err)
	}

	// 2. Stop scheduler
	if err := sch.Stop(); err != nil {
		logger.Error(ctx, "failed to stop scheduler", err)
	}

	// 3. Close agents
	if err := agent.Agents().Close(); err != nil {
		logger.Error(ctx, "failed to close agents", err)
	}

	// 4. Close database
	db, _ := svcCtx.DB.DB()
	if err := db.Close(); err != nil {
		logger.Error(ctx, "failed to close database", err)
	}

	// 5. Close logger (must be last to allow logging during shutdown)
	logger.Info(ctx, "graceful shutdown completed")

	if err := logger.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", err)
	}
}
