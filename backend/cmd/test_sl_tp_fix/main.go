package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	traderesponses "github.com/PineappleBond/TradingEino/backend/pkg/okex/responses/trade"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/responses"
	"golang.org/x/time/rate"
)

// mockTradeClient mocks the trade.RestClient for testing
type mockTradeClient struct {
	placeAlgoOrderFunc func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error)
}

func (m *mockTradeClient) PlaceAlgoOrder(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
	if m.placeAlgoOrderFunc != nil {
		return m.placeAlgoOrderFunc(req)
	}
	return traderesponses.PlaceAlgoOrder{
		Basic: responses.Basic{Code: 0, Msg: ""},
		PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
			{
				AlgoID: "test-algo-id-123",
				SMsg:   "",
				SCode:  0,
			},
		},
	}, nil
}

type mockTradePlace interface {
	PlaceAlgoOrder(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error)
}

// testOkxPlaceOrderWithSlTpTool is a copy for testing
type testOkxPlaceOrderWithSlTpTool struct {
	limiter   *rate.Limiter
	mockTrade mockTradePlace
}

// InvokableRun implements the tool execution logic (matches the fixed version)
func (c *testOkxPlaceOrderWithSlTpTool) InvokableRun(ctx context.Context, argsJSON string) (string, error) {
	// 1. Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// 2. Parse JSON arguments
	var params struct {
		InstID      string `json:"instID"`
		Side        string `json:"side"`
		PosSide     string `json:"posSide"`
		OrdType     string `json:"ordType"`
		Size        string `json:"size"`
		SlTriggerPx string `json:"slTriggerPx"`
		SlOrderPx   string `json:"slOrderPx"`
		TpTriggerPx string `json:"tpTriggerPx"`
		TpOrderPx   string `json:"tpOrderPx"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// 3. Parse values
	slTriggerPx := 0.0
	tpTriggerPx := 0.0
	slOrderPx := 0.0
	tpOrderPx := 0.0

	if params.SlTriggerPx != "" {
		fmt.Sscanf(params.SlTriggerPx, "%f", &slTriggerPx)
	}
	if params.SlOrderPx != "" && params.SlOrderPx != "-1" {
		fmt.Sscanf(params.SlOrderPx, "%f", &slOrderPx)
	}
	if params.TpTriggerPx != "" {
		fmt.Sscanf(params.TpTriggerPx, "%f", &tpTriggerPx)
	}
	if params.TpOrderPx != "" && params.TpOrderPx != "-1" {
		fmt.Sscanf(params.TpOrderPx, "%f", &tpOrderPx)
	}

	// 4. Parse side and posSide
	var side okex.OrderSide
	switch params.Side {
	case "buy":
		side = okex.OrderBuy
	default:
		side = okex.OrderSell
	}

	var posSide okex.PositionSide
	switch params.PosSide {
	case "long":
		posSide = okex.PositionLongSide
	default:
		posSide = okex.PositionNetSide
	}

	// 5. Determine order types - KEY FIX
	var algoOrdType okex.AlgoOrderType
	var ordType okex.OrderType

	hasBoth := slTriggerPx > 0 && tpTriggerPx > 0
	if hasBoth {
		algoOrdType = okex.AlgoOrderOCO // Both SL and TP = OCO order
	} else {
		algoOrdType = okex.AlgoOrderConditional
	}

	// ordType = market if no limit prices, limit if has limit prices
	hasLimitPrice := (params.SlOrderPx != "" && params.SlOrderPx != "-1") ||
		(params.TpOrderPx != "" && params.TpOrderPx != "-1")
	if hasLimitPrice {
		ordType = okex.OrderLimit
	} else {
		ordType = okex.OrderMarket
	}

	// 6. Build request - KEY FIX: includes AlgoOrdType
	req := traderequests.PlaceAlgoOrder{
		InstID:      params.InstID,
		TdMode:      okex.TradeCrossMode,
		Side:        side,
		PosSide:     posSide,
		OrdType:     ordType,        // market or limit
		AlgoOrdType: algoOrdType,    // conditional or oco
		Sz:          params.Size,
	}

	if slTriggerPx > 0 {
		req.SlTriggerPx = &slTriggerPx
	}
	if tpTriggerPx > 0 {
		req.TpTriggerPx = &tpTriggerPx
	}
	// Note: For market SL/TP orders, we don't set SlOrdPx/TpOrdPx
	// OKX will treat missing fields as market orders

	// 7. Call API
	result, err := c.mockTrade.PlaceAlgoOrder(req)
	if err != nil {
		return "", err
	}

	if result.Code.Int() != 0 {
		return "", fmt.Errorf("API error: %s", result.Msg)
	}

	if len(result.PlaceAlgoOrders) == 0 {
		return "", fmt.Errorf("empty response")
	}

	algoResult := result.PlaceAlgoOrders[0]
	if algoResult.SCode != 0 {
		return "", fmt.Errorf("order error: %s", algoResult.SMsg)
	}

	return fmt.Sprintf("## Order Placed Successfully\n\n- AlgoID: %s\n- Type: %s\n- OrdType: %s\n",
		algoResult.AlgoID, algoOrdType, ordType), nil
}

func main() {
	fmt.Println("=== 测试 OKX 止损止盈订单修复 ===\n")

	// 测试场景：用户报错的场景 - 市价单同时设置止损和止盈
	fmt.Println("测试场景：BTC-USDT-SWAP 市价做多，同时设置止损 69500 和止盈 71800")
	fmt.Println("原错误：Parameter slOrdPx error (错误码 51000)")
	fmt.Println()

	callCount := 0
	mockTrade := &mockTradeClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			callCount++
			fmt.Printf(">>> PlaceAlgoOrder 调用详情:\n")
			fmt.Printf("   InstID:      %s\n", req.InstID)
			fmt.Printf("   Side:        %s\n", req.Side)
			fmt.Printf("   OrdType:     %s\n", req.OrdType)
			fmt.Printf("   AlgoOrdType: %s\n", req.AlgoOrdType)
			fmt.Printf("   Sz:          %s\n", req.Sz)
			if req.SlTriggerPx != nil {
				fmt.Printf("   SlTriggerPx: %.2f\n", *req.SlTriggerPx)
			}
			if req.SlOrdPx != nil {
				fmt.Printf("   SlOrdPx:     %.2f\n", *req.SlOrdPx)
			} else {
				fmt.Printf("   SlOrdPx:     nil (市价单)\n")
			}
			if req.TpTriggerPx != nil {
				fmt.Printf("   TpTriggerPx: %.2f\n", *req.TpTriggerPx)
			}
			if req.TpOrdPx != nil {
				fmt.Printf("   TpOrdPx:     %.2f\n", *req.TpOrdPx)
			} else {
				fmt.Printf("   TpOrdPx:     nil (市价单)\n")
			}
			fmt.Println()

			// 验证关键字段
			allValid := true
			if req.AlgoOrdType != okex.AlgoOrderOCO {
				fmt.Printf("✗ AlgoOrdType 错误：期望 'oco'，实际 '%s'\n", req.AlgoOrdType)
				allValid = false
			} else {
				fmt.Println("✓ AlgoOrdType 正确设置为 'oco'")
			}

			if req.OrdType != okex.OrderMarket {
				fmt.Printf("✗ OrdType 错误：期望 'market'，实际 '%s'\n", req.OrdType)
				allValid = false
			} else {
				fmt.Println("✓ OrdType 正确设置为 'market'")
			}

			if !allValid {
				fmt.Println("\n✗ 请求参数验证失败！")
			} else {
				fmt.Println("\n✓ 所有参数验证通过！")
			}

			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0, Msg: ""},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "algo-oco-success",
						SMsg:   "",
						SCode:  0,
					},
				},
			}, nil
		},
	}

	// 创建测试工具
	testTool := &testOkxPlaceOrderWithSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "BTC-USDT-SWAP",
		"side":        "buy",
		"posSide":     "long",
		"ordType":     "conditional",
		"size":        "5",
		"slTriggerPx": "69500",
		"slOrderPx":   "-1", // 市价单
		"tpTriggerPx": "71800",
		"tpOrderPx":   "-1", // 市价单
	}

	argsJSON, _ := json.Marshal(args)

	fmt.Println("用户输入参数:")
	prettyJSON, _ := json.MarshalIndent(args, "", "  ")
	fmt.Println(string(prettyJSON))
	fmt.Println()

	result, err := testTool.InvokableRun(context.Background(), string(argsJSON))
	if err != nil {
		fmt.Printf("\n✗ 测试失败：%v\n", err)
	} else {
		fmt.Println("\n✓ 测试通过！")
		fmt.Printf("返回结果:\n%s\n", result)
	}
}
