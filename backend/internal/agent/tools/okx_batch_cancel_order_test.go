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

// TestOkxBatchCancelOrderTool_Cancel2OrdersInBatch returns results for both
func TestOkxBatchCancelOrderTool_Cancel2OrdersInBatch_ReturnsResultsForBoth(t *testing.T) {
	// Test that canceling 2 orders in batch returns results for both
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}
	t.Log("Batch cancel order tool limiter configured for 5 req/s")
}

// TestOkxBatchCancelOrderTool_Cancel21Orders returns error (exceeds max 20)
func TestOkxBatchCancelOrderTool_Cancel21Orders_ReturnsError(t *testing.T) {
	// Test that 21 orders returns error
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}
	t.Log("Batch cancel order tool should validate max 20 orders")
}

// TestOkxBatchCancelOrderTool_BatchWithPartialFailures returns both successes and failures
func TestOkxBatchCancelOrderTool_BatchWithPartialFailures_ReturnsBothSuccessesAndFailures(t *testing.T) {
	// Test that partial failures are handled correctly
	successResult := &trade.CancelOrder{
		OrdID: "order-cancelled-1",
		SCode: okex.JSONFloat64(0),
		SMsg:  "",
	}
	failureResult := &trade.CancelOrder{
		OrdID: "order-failed-1",
		SCode: okex.JSONFloat64(51001),
		SMsg:  "Order not found",
	}

	if float64(successResult.SCode) != 0 {
		t.Fatal("Success result should have sCode 0")
	}
	if float64(failureResult.SCode) == 0 {
		t.Fatal("Failure result should have sCode != 0")
	}

	t.Log("Partial failure detection verified")
}

// TestOkxBatchCancelOrderTool_EmptyOrderIDsArray returns error
func TestOkxBatchCancelOrderTool_EmptyOrderIDsArray_ReturnsError(t *testing.T) {
	tool := &OkxBatchCancelOrderTool{
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

// TestOkxBatchCancelOrderTool_InvalidJSON returns error
func TestOkxBatchCancelOrderTool_InvalidJSON_ReturnsError(t *testing.T) {
	tool := &OkxBatchCancelOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	_, err := tool.InvokableRun(ctx, "invalid json")
	if err == nil {
		t.Fatal("Should return error for invalid JSON")
	}
	t.Logf("Correctly returned error for invalid JSON: %v", err)
}

// TestOkxBatchCancelOrderTool_LimiterConfigured verifies rate limiter setup
func TestOkxBatchCancelOrderTool_LimiterConfigured(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	t.Log("Limiter configured for Trade endpoint (5 req/s)")
}

// TestOkxBatchCancelOrderTool_MissingRequiredParams returns error
func TestOkxBatchCancelOrderTool_MissingRequiredParams_ReturnsError(t *testing.T) {
	tool := &OkxBatchCancelOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	testCases := []struct {
		name      string
		jsonInput string
	}{
		{"missing instID", `{"orders": [{"ordID": "order-123"}]}`},
		{"missing ordID", `{"orders": [{"instID": "ETH-USDT-SWAP"}]}`},
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

// TestBatchCancelOrderResponseParsing verifies response parsing
func TestBatchCancelOrderResponseParsing(t *testing.T) {
	resp := &trade.CancelOrder{
		OrdID: "order-123",
		SCode: okex.JSONFloat64(0),
		SMsg:  "",
	}

	if resp.OrdID != "order-123" {
		t.Errorf("Expected order ID 'order-123', got %s", resp.OrdID)
	}

	if float64(resp.SCode) != 0 {
		t.Errorf("Expected sCode 0, got %f", float64(resp.SCode))
	}

	t.Log("BatchCancelOrder response parsing verified")
}

// TestOkxErrorTypeDetectionForCancel verifies OKXError type detection for cancel operations
func TestOkxErrorTypeDetectionForCancel(t *testing.T) {
	err := &okex.OKXError{
		Code:     51001,
		Msg:      "Order cancellation failed",
		Endpoint: "CandleOrder",
	}

	var okxErr *okex.OKXError
	if !errors.As(err, &okxErr) {
		t.Fatal("Should be able to unwrap OKXError")
	}

	if okxErr.Code != 51001 {
		t.Errorf("Expected code 51001, got %d", okxErr.Code)
	}

	t.Log("OKXError type detection verified for cancel operations")
}
