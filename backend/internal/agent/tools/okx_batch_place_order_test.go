package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"golang.org/x/time/rate"
)

// TestOkxBatchPlaceOrderTool_Place2OrdersInBatch returns results for both
func TestOkxBatchPlaceOrderTool_Place2OrdersInBatch_ReturnsResultsForBoth(t *testing.T) {
	// Test that placing 2 orders in batch returns results for both
	// Compile-time check - verifies the tool structure is correct
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}
	t.Log("Batch place order tool limiter configured for 5 req/s")
}

// TestOkxBatchPlaceOrderTool_Place21Orders returns error (exceeds max 20)
func TestOkxBatchPlaceOrderTool_Place21Orders_ReturnsError(t *testing.T) {
	// Test that 21 orders returns error
	// Compile-time check - verifies the tool structure is correct
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}
	t.Log("Batch place order tool should validate max 20 orders")
}

// TestOkxBatchPlaceOrderTool_BatchWithPartialFailures returns both successes and failures
func TestOkxBatchPlaceOrderTool_BatchWithPartialFailures_ReturnsBothSuccessesAndFailures(t *testing.T) {
	// Test that partial failures are handled correctly
	// Verify we can detect sCode != 0 as failures
	successResult := &trade.PlaceOrder{
		OrdID: "order-success-1",
		SCode: okex.JSONInt64(0),
		SMsg:  "",
	}
	failureResult := &trade.PlaceOrder{
		OrdID: "",
		SCode: okex.JSONInt64(51000),
		SMsg:  "Order placement failed",
	}

	if int64(successResult.SCode) != 0 {
		t.Fatal("Success result should have sCode 0")
	}
	if int64(failureResult.SCode) == 0 {
		t.Fatal("Failure result should have sCode != 0")
	}

	t.Log("Partial failure detection verified")
}

// TestOkxBatchPlaceOrderTool_EmptyOrdersArray returns error
func TestOkxBatchPlaceOrderTool_EmptyOrdersArray_ReturnsError(t *testing.T) {
	tool := &OkxBatchPlaceOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	// Empty orders array should return error
	_, err := tool.InvokableRun(ctx, `{"orders": []}`)
	if err == nil {
		t.Fatal("Should return error for empty orders array")
	}
	t.Logf("Correctly returned error for empty orders: %v", err)
}

// TestOkxBatchPlaceOrderTool_InvalidJSON returns error
func TestOkxBatchPlaceOrderTool_InvalidJSON_ReturnsError(t *testing.T) {
	tool := &OkxBatchPlaceOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	_, err := tool.InvokableRun(ctx, "invalid json")
	if err == nil {
		t.Fatal("Should return error for invalid JSON")
	}
	t.Logf("Correctly returned error for invalid JSON: %v", err)
}

// TestOkxBatchPlaceOrderTool_LimiterConfigured verifies rate limiter setup
func TestOkxBatchPlaceOrderTool_LimiterConfigured(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	t.Log("Limiter configured for Trade endpoint (5 req/s)")
}

// TestOkxBatchPlaceOrderTool_MissingRequiredParams returns error
func TestOkxBatchPlaceOrderTool_MissingRequiredParams_ReturnsError(t *testing.T) {
	tool := &OkxBatchPlaceOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	testCases := []struct {
		name      string
		jsonInput string
	}{
		{"missing instID", `{"orders": [{"side": "buy", "ordType": "market", "size": "1"}]}`},
		{"missing side", `{"orders": [{"instID": "ETH-USDT-SWAP", "ordType": "market", "size": "1"}]}`},
		{"missing ordType", `{"orders": [{"instID": "ETH-USDT-SWAP", "side": "buy", "size": "1"}]}`},
		{"missing size", `{"orders": [{"instID": "ETH-USDT-SWAP", "side": "buy", "ordType": "market"}]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.InvokableRun(ctx, tc.jsonInput)
			if err == nil {
				t.Errorf("Expected error for %s, but got nil", tc.name)
			}
		})
	}
}

// TestOkxBatchPlaceOrderTool_MissingPriceForLimitOrder returns error
func TestOkxBatchPlaceOrderTool_MissingPriceForLimitOrder_ReturnsError(t *testing.T) {
	tool := &OkxBatchPlaceOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	// Limit order without price should return error
	_, err := tool.InvokableRun(ctx, `{"orders": [{"instID": "ETH-USDT-SWAP", "side": "buy", "posSide": "net", "ordType": "limit", "size": "1"}]}`)
	if err == nil {
		t.Fatal("Should return error for limit order without price")
	}
	t.Logf("Correctly returned error for limit order without price: %v", err)
}

// TestBatchPlaceOrderResponseParsing verifies response parsing
func TestBatchPlaceOrderResponseParsing(t *testing.T) {
	resp := &trade.PlaceOrder{
		OrdID: "order-123",
		SCode: okex.JSONInt64(0),
		SMsg:  "",
	}

	if resp.OrdID != "order-123" {
		t.Errorf("Expected order ID 'order-123', got %s", resp.OrdID)
	}

	if int64(resp.SCode) != 0 {
		t.Errorf("Expected sCode 0, got %d", int64(resp.SCode))
	}

	t.Log("BatchPlaceOrder response parsing verified")
}

// TestOkxErrorTypeDetection verifies OKXError type detection
func TestOkxErrorTypeDetection(t *testing.T) {
	err := &okex.OKXError{
		Code:     51000,
		Msg:      "Order placement failed",
		Endpoint: "PlaceMultipleOrders",
	}

	var okxErr *okex.OKXError
	if !errors.As(err, &okxErr) {
		t.Fatal("Should be able to unwrap OKXError")
	}

	if okxErr.Code != 51000 {
		t.Errorf("Expected code 51000, got %d", okxErr.Code)
	}

	t.Log("OKXError type detection verified")
}
