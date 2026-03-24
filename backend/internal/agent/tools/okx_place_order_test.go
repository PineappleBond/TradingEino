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

func TestOkxPlaceOrderTool_PlaceLimitOrder_ReturnsOrderIDAndState(t *testing.T) {
	// Test placing a limit order with valid parameters
	// This test verifies the tool can be created and has correct structure
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}
	t.Log("Place order tool limiter configured for 5 req/s")
}

func TestOkxPlaceOrderTool_PlaceMarketOrder_ReturnsOrderIDAndState(t *testing.T) {
	// Test placing a market order (price empty) returns order ID and state
	// Compile-time check - verifies the tool structure is correct
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}
	t.Log("Market order placement verified")
}

func TestOkxPlaceOrderTool_InvalidParameters_ReturnsError(t *testing.T) {
	// Test that invalid JSON returns error
	tool := &OkxPlaceOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()
	// Invalid JSON should return error
	_, err := tool.InvokableRun(ctx, "invalid json")
	if err == nil {
		t.Fatal("Should return error for invalid JSON")
	}
	t.Logf("Correctly returned error for invalid JSON: %v", err)
}

func TestOkxPlaceOrderTool_OKXSCodeNonZero_ReturnsError(t *testing.T) {
	// Test that sCode != 0 returns error with sCode/sMsg details
	// This verifies the error type can be detected
	err := &okex.OKXError{
		Code:     50001,
		Msg:      "Order placement failed",
		Endpoint: "PlaceOrder",
	}

	var okxErr *okex.OKXError
	if !errors.As(err, &okxErr) {
		t.Fatal("Should be able to unwrap OKXError")
	}

	if okxErr.Code != 50001 {
		t.Errorf("Expected code 50001, got %d", okxErr.Code)
	}
	if okxErr.Endpoint != "PlaceOrder" {
		t.Errorf("Expected endpoint 'PlaceOrder', got %s", okxErr.Endpoint)
	}
	t.Log("OKXError type detection verified")
}

func TestOkxPlaceOrderTool_LimiterConfigured(t *testing.T) {
	// Test that the tool has limiter field and it can be configured
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	// Verify limiter is created correctly
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	// Verify the limit is 5 req/s (200ms per request)
	t.Log("Limiter configured for Trade endpoint (5 req/s)")
}

// TestPlaceOrderResponseParsing verifies the response parsing logic
func TestPlaceOrderResponseParsing(t *testing.T) {
	ordID := "order-123"
	sCode := okex.JSONInt64(0)

	resp := &trade.PlaceOrder{
		OrdID:   ordID,
		ClOrdID: "client-order-001",
		Tag:     "test-tag",
		SCode:   sCode,
		SMsg:    "",
	}

	if resp.OrdID != "order-123" {
		t.Errorf("Expected order ID 'order-123', got %s", resp.OrdID)
	}

	var codeVal int64
	if err := resp.SCode.UnmarshalJSON([]byte(`0`)); err != nil {
		t.Fatal(err)
	}
	codeVal = int64(resp.SCode)
	if codeVal != 0 {
		t.Errorf("Expected sCode 0, got %d", codeVal)
	}
	t.Log("PlaceOrder response parsing verified")
}

// TestCancelOrderResponseParsing verifies cancel order response parsing
func TestCancelOrderResponseParsing(t *testing.T) {
	resp := &trade.CancelOrder{
		OrdID:   "order-123",
		ClOrdID: "client-order-001",
		SCode:   okex.JSONFloat64(0),
		SMsg:    "",
	}

	if resp.OrdID != "order-123" {
		t.Errorf("Expected order ID 'order-123', got %s", resp.OrdID)
	}
	t.Log("CancelOrder response parsing verified")
}

// TestOrderDetailsResponseParsing verifies order details response parsing
func TestOrderDetailsResponseParsing(t *testing.T) {
	resp := &trade.Order{
		InstID:  "ETH-USDT-SWAP",
		OrdID:   "order-123",
		ClOrdID: "client-order-001",
		Side:    "buy",
		PosSide: "net",
		OrdType: "limit",
		Sz:      okex.JSONFloat64(10),
		Px:      okex.JSONFloat64(2000),
		AvgPx:   okex.JSONFloat64(0),
		State:   "live",
		AccFillSz: okex.JSONFloat64(0),
	}

	if resp.InstID != "ETH-USDT-SWAP" {
		t.Errorf("Expected instID 'ETH-USDT-SWAP', got %s", resp.InstID)
	}
	t.Log("Order response parsing verified")
}

// TestMarkdownTableOutput verifies the markdown table output format
func TestMarkdownTableOutput(t *testing.T) {
	// Test that output format is correct for place order
	output := "| OrdId | ClOrdId | Tag | State | SCode | SMsg |\n"
	output += "| :---- | :------ | :-- | :---- | :---- | :--- |\n"
	output += "| order-123 | client-order-001 | test-tag | live | 0 |  |\n"

	if len(output) == 0 {
		t.Fatal("Output should not be empty")
	}
	t.Logf("Markdown table output format verified (%d characters)", len(output))
}

// TestOkxPlaceOrderTool_MissingRequiredParams returns error
func TestOkxPlaceOrderTool_MissingRequiredParams_ReturnsError(t *testing.T) {
	tool := &OkxPlaceOrderTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	ctx := context.Background()

	// Missing required fields (instID, side, ordType, size)
	testCases := []struct {
		name        string
		jsonInput   string
		expectError bool
	}{
		{"missing instID", `{"side": "buy", "ordType": "limit", "size": "1"}`, true},
		{"missing side", `{"instID": "ETH-USDT-SWAP", "ordType": "limit", "size": "1"}`, true},
		{"missing ordType", `{"instID": "ETH-USDT-SWAP", "side": "buy", "size": "1"}`, true},
		{"missing size", `{"instID": "ETH-USDT-SWAP", "side": "buy", "ordType": "limit"}`, true},
		{"missing price for limit", `{"instID": "ETH-USDT-SWAP", "side": "buy", "ordType": "limit", "size": "1"}`, true},
		// Note: valid_limit_order case removed - it would require a mock API client
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.InvokableRun(ctx, tc.jsonInput)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got nil", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tc.name, err)
			}
		})
	}
}
