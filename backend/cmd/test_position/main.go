package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	accountRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
)

var configPath = flag.String("c", "etc/config.yaml", "path to config file")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	svcCtx := svc.NewServiceContext(*cfg)
	ctx := context.Background()

	// Get positions
	fmt.Println("=== Getting Current Positions ===")
	resp, err := svcCtx.OKXClient.Rest.Account.GetPositions(accountRequests.GetPositions{
		InstID: []string{"ETH-USDT-SWAP"},
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if len(resp.Positions) == 0 {
		fmt.Println("No positions found")
		return
	}

	for i, pos := range resp.Positions {
		fmt.Printf("\n--- Position %d ---\n", i+1)
		data, _ := json.MarshalIndent(pos, "", "  ")
		fmt.Printf("%s\n", string(data))
	}

	_ = ctx
}
