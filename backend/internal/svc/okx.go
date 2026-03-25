package svc

import (
	"context"
	"fmt"
	"os"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/api"
)

func mustInitOKXClient(cfg config.Config) *api.Client {
	dest := okex.NormalServer
	if cfg.OKX.Sandbox {
		dest = okex.DemoServer
	}
	client, err := api.NewClient(
		context.Background(),
		cfg.OKX.ApiKey,
		cfg.OKX.SecretKey,
		cfg.OKX.Passphrase,
		dest,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init okx client: %v\n", err)
		os.Exit(1)
	}
	return client
}
