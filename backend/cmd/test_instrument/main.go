package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"
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

	// Get instrument info for ETH-USDT-SWAP
	fmt.Println("=== Getting ETH-USDT-SWAP Instrument Info ===")
	resp, err := svcCtx.OKXClient.Rest.PublicData.GetInstruments(traderequests.GetInstruments{
		InstType: "SWAP",
		InstID:   "ETH-USDT-SWAP",
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if len(resp.Instruments) == 0 {
		fmt.Println("No instrument found")
		return
	}

	inst := resp.Instruments[0]
	data, _ := json.MarshalIndent(inst, "", "  ")
	fmt.Printf("Instrument Details:\n%s\n", string(data))

	fmt.Println("\n=== Key Fields ===")
	fmt.Printf("InstID: %s\n", inst.InstID)
	fmt.Printf("CtVal (合约面值): %g\n", float64(inst.CtVal))
	fmt.Printf("CtMult (合约乘数): %g\n", float64(inst.CtMult))
	fmt.Printf("LotSz (下单精度): %g\n", float64(inst.LotSz))
	fmt.Printf("TickSz (价格精度): %g\n", float64(inst.TickSz))
	fmt.Printf("MinSz (最小下单数量): %g\n", float64(inst.MinSz))
	fmt.Printf("State (状态): %s\n", inst.State)

	_ = ctx
}
