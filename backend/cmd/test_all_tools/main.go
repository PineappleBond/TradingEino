package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
)

var configPath = flag.String("c", "etc/config.yaml", "path to config file")

var svcCtx *svc.ServiceContext
var ctx context.Context

// Test instrument to use
const (
	testInstID   = "ETH-USDT-SWAP"
	testInstID2  = "BTC-USDT-SWAP"
	testSize     = "0.1"
	testLeverage = 1
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	svcCtx = svc.NewServiceContext(*cfg)
	ctx = context.Background()

	passed := 0
	failed := 0
	skipped := 0

	fmt.Println("========================================")
	fmt.Println("  TradingEino OKX Tools Test Suite")
	fmt.Println("========================================")
	fmt.Printf("Config: %s\n", *configPath)
	fmt.Printf("Sandbox: %v\n", cfg.OKX.Sandbox)
	fmt.Printf("Test Instrument: %s\n", testInstID)
	fmt.Println("========================================\n")

	// ========== Phase 1: Read-Only Tools (No dependencies) ==========
	fmt.Println("\n=== Phase 1: Read-Only Tools ===\n")

	// Test 1: okx-account-balance-tool
	fmt.Println("=== Test 1: okx-account-balance-tool ===")
	result, err := testAccountBalance()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 2: okx-candlesticks
	fmt.Println("=== Test 2: okx-candlesticks ===")
	result, err = testCandlesticks()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 3: okx-get-fundingrate
	fmt.Println("=== Test 3: okx-get-fundingrate ===")
	result, err = testGetFundingRate()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 4: okx-orderbook
	fmt.Println("=== Test 4: okx-orderbook ===")
	result, err = testOrderbook()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 5: okx-trades-history
	fmt.Println("=== Test 5: okx-trades-history ===")
	result, err = testTradesHistory()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 6: okx-get-order-history
	fmt.Println("=== Test 6: okx-get-order-history ===")
	result, err = testGetOrderHistory()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// ========== Phase 2: Position Tools ==========
	fmt.Println("\n=== Phase 2: Position Tools ===\n")

	// Test 7: okx-get-positions-tool
	fmt.Println("=== Test 7: okx-get-positions-tool ===")
	result, _, err = testGetPositions()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// ========== Phase 3: Order Placement Tools ==========
	fmt.Println("\n=== Phase 3: Order Placement Tools ===\n")

	// Test 8: okx-place-order - Place a limit order (won't fill immediately)
	fmt.Println("=== Test 8: okx-place-order (Limit Order) ===")
	var orderID string
	result, orderID, err = testPlaceLimitOrder()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 9: okx-get-order - Query the placed order
	fmt.Println("=== Test 9: okx-get-order ===")
	if orderID != "" {
		result, err = testGetOrder(orderID)
		if err != nil {
			fmt.Printf("FAILED: %v\n\n", err)
			failed++
		} else {
			fmt.Printf("PASSED:\n%s\n\n", result)
			passed++
		}
	} else {
		fmt.Println("SKIPPED: No order ID to query\n")
		skipped++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 10: okx-cancel-order - Cancel the limit order
	fmt.Println("=== Test 10: okx-cancel-order ===")
	if orderID != "" {
		result, err = testCancelOrder(orderID)
		if err != nil {
			fmt.Printf("FAILED: %v\n\n", err)
			failed++
		} else {
			fmt.Printf("PASSED:\n%s\n\n", result)
			passed++
		}
	} else {
		fmt.Println("SKIPPED: No order ID to cancel\n")
		skipped++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 11: okx-batch-place-order - Place multiple orders
	fmt.Println("=== Test 11: okx-batch-place-order ===")
	var batchOrderIDs []string
	result, batchOrderIDs, err = testBatchPlaceOrder()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 12: okx-batch-cancel-order - Cancel multiple orders
	fmt.Println("=== Test 12: okx-batch-cancel-order ===")
	if len(batchOrderIDs) > 0 {
		result, err = testBatchCancelOrder(batchOrderIDs)
		if err != nil {
			fmt.Printf("FAILED: %v\n\n", err)
			failed++
		} else {
			fmt.Printf("PASSED:\n%s\n\n", result)
			passed++
		}
	} else {
		fmt.Println("SKIPPED: No order IDs to cancel\n")
		skipped++
	}
	time.Sleep(500 * time.Millisecond)

	// ========== Phase 4: Advanced Order Tools ==========
	fmt.Println("\n=== Phase 4: Advanced Order Tools ===\n")

	// Test 13: okx-place-order-with-sl-tp
	fmt.Println("=== Test 13: okx-place-order-with-sl-tp ===")
	result, err = testPlaceOrderWithSLTP()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 14: okx-attach-sl-tp
	fmt.Println("=== Test 14: okx-attach-sl-tp ===")
	result, err = testAttachSLTP()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// ========== Phase 5: Position Management Tools ==========
	fmt.Println("\n=== Phase 5: Position Management Tools ===\n")

	// Test 15: okx-close-position - Close a single position
	fmt.Println("=== Test 15: okx-close-position ===")
	result, err = testClosePosition()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 16: okx-close-all-positions
	fmt.Println("=== Test 16: okx-close-all-positions ===")
	result, err = testCloseAllPositions()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// Test 17: okx-cancel-all-orders
	fmt.Println("=== Test 17: okx-cancel-all-orders ===")
	result, err = testCancelAllOrders()
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
		failed++
	} else {
		fmt.Printf("PASSED:\n%s\n\n", result)
		passed++
	}
	time.Sleep(500 * time.Millisecond)

	// ========== Summary ==========
	fmt.Println("\n========================================")
	fmt.Printf("  Test Results: %d passed, %d failed, %d skipped\n", passed, failed, skipped)
	fmt.Println("========================================")

	if failed > 0 {
		os.Exit(1)
	}
}

// ========== Test Functions ==========

func testAccountBalance() (string, error) {
	tool := tools.NewOkxAccountBalanceTool(svcCtx)
	args := map[string]interface{}{}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testCandlesticks() (string, error) {
	tool := tools.NewOkxCandlesticksTool(svcCtx)
	args := map[string]interface{}{
		"instID": testInstID,
		"bar":    "1m",
		"limit":  10,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testGetFundingRate() (string, error) {
	tool := tools.NewOkxGetFundingRateTool(svcCtx)
	args := map[string]interface{}{
		"symbol": testInstID,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testOrderbook() (string, error) {
	tool := tools.NewOkxOrderbookTool(svcCtx)
	args := map[string]interface{}{
		"symbol": testInstID,
		"depth":  10,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testTradesHistory() (string, error) {
	tool := tools.NewOkxTradesHistoryTool(svcCtx)
	args := map[string]interface{}{
		"symbol": testInstID,
		"limit":  10,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testGetOrderHistory() (string, error) {
	tool := tools.NewOkxGetOrderHistoryTool(svcCtx)
	args := map[string]interface{}{
		"instType": "SWAP",
		"instID":   testInstID,
		"ordType":  "limit",
		"limit":    10,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testGetPositions() (string, bool, error) {
	tool := tools.NewOkxGetPositionsTool(svcCtx)
	args := map[string]interface{}{
		"symbol":   testInstID,
		"leverage": testLeverage,
	}
	argsJSON, _ := json.Marshal(args)
	result, err := tool.InvokableRun(ctx, string(argsJSON))
	hasPositions := strings.Contains(result, "|") && !strings.Contains(result, "无仓位")
	return result, hasPositions, err
}

func testPlaceLimitOrder() (string, string, error) {
	tool := tools.NewOkxPlaceOrderTool(svcCtx)

	// Get current price from orderbook to place a limit order far away
	currentPrice := 3000.0 // Default fallback
	orderbookResult, err := testOrderbook()
	if err == nil {
		// Try to extract price from orderbook result
		lines := strings.Split(orderbookResult, "\n")
		for _, line := range lines {
			if strings.Contains(line, "|") && strings.Contains(line, ".") {
				parts := strings.Split(line, "|")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.Contains(part, ".") {
						fmt.Sscanf(part, "%f", &currentPrice)
						break
					}
				}
				break
			}
		}
	}

	// Place a limit order at a price that won't fill (5% below current)
	limitPrice := currentPrice * 0.95
	limitPriceStr := fmt.Sprintf("%.2f", limitPrice)

	args := map[string]interface{}{
		"instID":  testInstID,
		"side":    "buy",
		"posSide": "net",
		"ordType": "limit",
		"size":    testSize,
		"price":   limitPriceStr,
	}
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("Placing limit order at %s (current ~%.2f)...\n", limitPriceStr, currentPrice)

	result, err := tool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", "", err
	}

	// Extract order ID - skip header row (looks for numeric order IDs)
	var orderID string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "| ") && strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				candidate := strings.TrimSpace(parts[1])
				// Skip header row and empty values - order ID is a large number
				if candidate != "" && candidate != "OrdId" && len(candidate) > 10 {
					orderID = candidate
					break
				}
			}
		}
	}

	return result, orderID, nil
}

func testGetOrder(ordID string) (string, error) {
	tool := tools.NewOkxGetOrderTool(svcCtx)
	args := map[string]interface{}{
		"instID": testInstID,
		"ordID":  ordID,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testCancelOrder(ordID string) (string, error) {
	tool := tools.NewOkxCancelOrderTool(svcCtx)
	args := map[string]interface{}{
		"instID": testInstID,
		"ordID":  ordID,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testBatchPlaceOrder() (string, []string, error) {
	tool := tools.NewOkxBatchPlaceOrderTool(svcCtx)

	// Get current price for limit orders
	currentPrice := 3000.0
	orderbookResult, err := testOrderbook()
	if err == nil {
		lines := strings.Split(orderbookResult, "\n")
		for _, line := range lines {
			if strings.Contains(line, "|") && strings.Contains(line, ".") {
				parts := strings.Split(line, "|")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.Contains(part, ".") {
						fmt.Sscanf(part, "%f", &currentPrice)
						break
					}
				}
				break
			}
		}
	}

	// Place 2 limit orders at different prices
	limitPrice1 := currentPrice * 0.90 // 10% below
	limitPrice2 := currentPrice * 0.85 // 15% below

	args := map[string]interface{}{
		"orders": []map[string]interface{}{
			{
				"instID":  testInstID,
				"side":    "buy",
				"posSide": "net",
				"ordType": "limit",
				"size":    "0.1",
				"price":   fmt.Sprintf("%.2f", limitPrice1),
			},
			{
				"instID":  testInstID,
				"side":    "sell",
				"posSide": "net",
				"ordType": "limit",
				"size":    "0.1",
				"price":   fmt.Sprintf("%.2f", limitPrice2),
			},
		},
	}
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("Placing batch orders: buy@%.2f, sell@%.2f\n", limitPrice1, limitPrice2)

	result, err := tool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", nil, err
	}

	// Extract order IDs from result - skip header rows
	var orderIDs []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "| ") && strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				ordID := strings.TrimSpace(parts[1])
				// Skip header row and empty values - order ID is a large number
				if ordID != "" && ordID != "OrdId" && len(ordID) > 10 {
					orderIDs = append(orderIDs, ordID)
				}
			}
		}
	}

	return result, orderIDs, nil
}

func testBatchCancelOrder(orderIDs []string) (string, error) {
	tool := tools.NewOkxBatchCancelOrderTool(svcCtx)

	orders := make([]map[string]interface{}, len(orderIDs))
	for i, ordID := range orderIDs {
		orders[i] = map[string]interface{}{
			"instID": testInstID,
			"ordID":  ordID,
		}
	}

	args := map[string]interface{}{
		"orders": orders,
	}
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("Cancelling %d batch orders...\n", len(orderIDs))

	return tool.InvokableRun(ctx, string(argsJSON))
}

func testPlaceOrderWithSLTP() (string, error) {
	tool := tools.NewOkxPlaceOrderWithSlTpTool(svcCtx)

	// Get current price for SL/TP triggers
	currentPrice := 3000.0
	orderbookResult, err := testOrderbook()
	if err == nil {
		lines := strings.Split(orderbookResult, "\n")
		for _, line := range lines {
			if strings.Contains(line, "|") && strings.Contains(line, ".") {
				parts := strings.Split(line, "|")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.Contains(part, ".") {
						fmt.Sscanf(part, "%f", &currentPrice)
						break
					}
				}
				break
			}
		}
	}

	slTrigger := currentPrice * 0.95 // 5% below
	tpTrigger := currentPrice * 1.05 // 5% above

	// Use larger size to meet minimum order amount requirement (min ~$100 USD)
	// For ETH-USDT-SWAP at ~2150, 0.1 contract = ~$215, should be enough
	args := map[string]interface{}{
		"instID":      testInstID,
		"side":        "buy",
		"posSide":     "net",
		"ordType":     "market",
		"size":        "1", // 1 contract to ensure minimum order amount
		"slTriggerPx": fmt.Sprintf("%.2f", slTrigger),
		"tpTriggerPx": fmt.Sprintf("%.2f", tpTrigger),
		"slOrderPx":   "-1",
		"tpOrderPx":   "-1",
	}
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("Placing market order with SL@%.2f / TP@%.2f...\n", slTrigger, tpTrigger)

	result, err := tool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", err
	}

	return result, nil
}

func testAttachSLTP() (string, error) {
	tool := tools.NewOkxAttachSlTpTool(svcCtx)

	// Get current price for SL/TP triggers
	currentPrice := 3000.0
	orderbookResult, err := testOrderbook()
	if err == nil {
		lines := strings.Split(orderbookResult, "\n")
		for _, line := range lines {
			if strings.Contains(line, "|") && strings.Contains(line, ".") {
				parts := strings.Split(line, "|")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.Contains(part, ".") {
						fmt.Sscanf(part, "%f", &currentPrice)
						break
					}
				}
				break
			}
		}
	}

	slTrigger := currentPrice * 0.95 // 5% below
	tpTrigger := currentPrice * 1.05 // 5% above

	// Use larger size to meet minimum order amount requirement
	args := map[string]interface{}{
		"instID":      testInstID,
		"side":        "sell",
		"posSide":     "net",
		"slTriggerPx": fmt.Sprintf("%.2f", slTrigger),
		"tpTriggerPx": fmt.Sprintf("%.2f", tpTrigger),
		"slOrderPx":   "-1",
		"tpOrderPx":   "-1",
		"sz":          "1", // 1 contract to ensure minimum order amount
	}
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("Attaching SL@%.2f / TP@%.2f...\n", slTrigger, tpTrigger)

	return tool.InvokableRun(ctx, string(argsJSON))
}

func testClosePosition() (string, error) {
	tool := tools.NewOkxClosePositionTool(svcCtx)
	args := map[string]interface{}{
		"instID":  testInstID,
		"posSide": "net",
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testCloseAllPositions() (string, error) {
	tool := tools.NewOkxCloseAllPositionsTool(svcCtx)
	args := map[string]interface{}{}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}

func testCancelAllOrders() (string, error) {
	tool := tools.NewOkxCancelAllOrdersTool(svcCtx)
	args := map[string]interface{}{
		"instID": testInstID,
	}
	argsJSON, _ := json.Marshal(args)
	return tool.InvokableRun(ctx, string(argsJSON))
}
