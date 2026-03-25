package main

import (
	"encoding/json"
	"fmt"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/api/rest"
	accountrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
	marketrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
)

func main() {
	fmt.Println("=== TradingEino OKX 配置测试 ===\n")

	// 1. 读取配置文件
	fmt.Println("1. 读取配置文件...")
	cfg, err := config.Load("etc/config.example.yaml")
	if err != nil {
		fmt.Printf("读取配置文件失败：%v\n", err)
		return
	}
	fmt.Printf("   配置文件读取成功!\n")
	fmt.Printf("   OKX 沙盒模式：%v\n", cfg.OKX.Sandbox)
	fmt.Printf("   OKX API Key: %s\n", maskString(cfg.OKX.ApiKey, 8))
	fmt.Println()

	// 2. 初始化 OKX 客户端
	fmt.Println("2. 初始化 OKX 客户端...")
	var destination okex.Destination
	var baseURL okex.BaseURL

	if cfg.OKX.Sandbox {
		destination = okex.DemoServer
		baseURL = okex.DemoRestURL
		fmt.Println("   使用 Demo 沙盒环境")
	} else {
		destination = okex.NormalServer
		baseURL = okex.RestURL
		fmt.Println("   使用生产环境")
	}

	client := rest.NewClient(cfg.OKX.ApiKey, cfg.OKX.SecretKey, cfg.OKX.Passphrase, baseURL, destination)
	fmt.Println("   OKX 客户端初始化成功!")
	fmt.Println()

	// 3. 测试账户 API - 获取余额
	fmt.Println("3. 测试账户 API - 获取账户余额...")
	balance, err := client.Account.GetBalance(accountrequests.GetBalance{})
	if err != nil {
		fmt.Printf("   获取余额失败：%v\n", err)
	} else {
		fmt.Printf("   获取余额成功!\n")
		fmt.Printf("   响应码：%d\n", balance.Code.Int())
		if balance.Code.Int() == 0 && len(balance.Balances) > 0 {
			fmt.Println("   账户资产:")
			for _, b := range balance.Balances[0].Details {
				if float64(b.AvailBal) > 0 || float64(b.Eq) > 0 {
					fmt.Printf("      %s: 可用=%s, 总计=%s\n", b.Ccy, b.AvailBal, b.Eq)
				}
			}
		} else {
			fmt.Printf("   响应消息：%s\n", balance.Msg)
		}
	}
	fmt.Println()

	// 4. 测试市场 API - 获取 Ticker
	fmt.Println("4. 测试市场 API - 获取 BTC-USDT-SWAP Ticker...")
	ticker, err := client.Market.GetTicker(marketrequests.GetTickers{InstID: "BTC-USDT-SWAP"})
	if err != nil {
		fmt.Printf("   获取 Ticker 失败：%v\n", err)
	} else {
		fmt.Printf("   获取 Ticker 成功!\n")
		fmt.Printf("   响应码：%d\n", ticker.Code.Int())
		if ticker.Code.Int() == 0 && len(ticker.Tickers) > 0 {
			t := ticker.Tickers[0]
			fmt.Printf("      最新价：%.2f\n", float64(t.Last))
			fmt.Printf("      24h 最高：%.2f\n", float64(t.High24h))
			fmt.Printf("      24h 最低：%.2f\n", float64(t.Low24h))
			fmt.Printf("      24h 成交量：%.2f\n", float64(t.Vol24h))
		} else {
			fmt.Printf("   响应消息：%s\n", ticker.Msg)
		}
	}
	fmt.Println()

	// 5. 测试持仓查询
	fmt.Println("5. 测试账户 API - 获取当前持仓...")
	positions, err := client.Account.GetPositions(accountrequests.GetPositions{})
	if err != nil {
		fmt.Printf("   获取持仓失败：%v\n", err)
	} else {
		fmt.Printf("   获取持仓成功!\n")
		fmt.Printf("   响应码：%d\n", positions.Code.Int())
		if positions.Code.Int() == 0 && len(positions.Positions) > 0 {
			fmt.Println("      当前持仓:")
			for _, p := range positions.Positions {
				fmt.Printf("         %s %s: 持仓量=%s, 强平价=%s, 未实现盈亏=%s\n",
					p.InstID, p.PosSide, p.Pos, p.LiqPx, p.Upl)
			}
		} else {
			fmt.Println("      当前无持仓")
		}
	}
	fmt.Println()

	// 6. 测试算法订单 API - 获取当前算法订单
	fmt.Println("6. 测试交易 API - 获取当前算法订单...")
	algoOrders, err := client.Trade.GetAlgoOrderList(traderequests.AlgoOrderList{
		InstType: "SWAP",
	}, false)
	if err != nil {
		fmt.Printf("   获取算法订单失败：%v\n", err)
	} else {
		fmt.Printf("   获取算法订单成功!\n")
		fmt.Printf("   响应码：%d\n", algoOrders.Code.Int())
		if algoOrders.Code.Int() == 0 && len(algoOrders.AlgoOrders) > 0 {
			fmt.Println("      当前算法订单:")
			for _, o := range algoOrders.AlgoOrders {
				fmt.Printf("         %s %s: AlgoID=%s, 状态=%s\n",
					o.InstID, o.Side, o.AlgoID, o.State)
			}
		} else {
			fmt.Println("      当前无算法订单")
		}
	}
	fmt.Println()

	// 7. 测试止损止盈订单请求结构
	fmt.Println("7. 测试 PlaceAlgoOrder 请求结构...")
	testPlaceAlgoOrderRequest()
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
}

func maskString(s string, visible int) string {
	if len(s) <= visible {
		return "***"
	}
	return s[:visible] + "..."
}

func testPlaceAlgoOrderRequest() {
	// 测试请求结构序列化
	slTriggerPx := 70500.0
	tpTriggerPx := 72000.0

	req := traderequests.PlaceAlgoOrder{
		InstID:      "BTC-USDT-SWAP",
		TdMode:      okex.TradeCrossMode,
		Side:        okex.OrderBuy,
		PosSide:     okex.PositionNetSide,
		OrdType:     okex.OrderMarket,
		AlgoOrdType: okex.AlgoOrderConditional,
		Sz:          "5",
		ReduceOnly:  true,
		StopOrder: traderequests.StopOrder{
			SlTriggerPx: &slTriggerPx,
			TpTriggerPx: &tpTriggerPx,
		},
	}

	data, err := json.MarshalIndent(req, "   ", "  ")
	if err != nil {
		fmt.Printf("   序列化失败：%v\n", err)
		return
	}

	fmt.Println("   PlaceAlgoOrder 请求 JSON:")
	fmt.Println(string(data))
	fmt.Println("   ✓ 请求结构序列化成功")
}
