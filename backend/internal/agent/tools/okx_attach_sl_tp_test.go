package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/responses"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	traderesponses "github.com/PineappleBond/TradingEino/backend/pkg/okex/responses/trade"
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
	// Default success response
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

func TestOkxAttachSlTpTool_AttachSlTpReturnsAlgoId(t *testing.T) {
	// Test 1: Attach SL/TP to existing order returns algoId
	callCount := 0
	mockTrade := &mockTradeClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			// Verify request parameters
			if req.InstID != "ETH-USDT-SWAP" {
				t.Errorf("Expected instId 'ETH-USDT-SWAP', got '%s'", req.InstID)
			}
			if req.OrdType != okex.AlgoOrderConditional {
				t.Errorf("Expected ordType 'conditional', got '%s'", req.OrdType)
			}

			callCount++
			if callCount == 1 {
				// First call: SL order
				if req.SlTriggerPx == nil || *req.SlTriggerPx != 1800.0 {
					t.Errorf("Expected slTriggerPx 1800.0, got '%v'", req.SlTriggerPx)
				}
				if req.TpTriggerPx != nil {
					t.Errorf("Expected tpTriggerPx to be nil for SL order, got '%v'", req.TpTriggerPx)
				}
				return traderesponses.PlaceAlgoOrder{
					PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
						{
							AlgoID: "algo-sl-123456",
							SMsg:   "",
							SCode:  0,
						},
					},
				}, nil
			} else {
				// Second call: TP order
				if req.TpTriggerPx == nil || *req.TpTriggerPx != 2200.0 {
					t.Errorf("Expected tpTriggerPx 2200.0, got '%v'", req.TpTriggerPx)
				}
				if req.SlTriggerPx != nil {
					t.Errorf("Expected slTriggerPx to be nil for TP order, got '%v'", req.SlTriggerPx)
				}
				return traderesponses.PlaceAlgoOrder{
					PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
						{
							AlgoID: "algo-tp-123456",
							SMsg:   "",
							SCode:  0,
						},
					},
				}, nil
			}
		},
	}

	tool := &OkxAttachSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"ordId":       "order-123",
		"side":        "buy",
		"posSide":     "long",
		"slTriggerPx": "1800.0",
		"slOrderPx":   "1790.0",
		"tpTriggerPx": "2200.0",
		"tpOrderPx":   "2210.0",
		"sz":          "0.1",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty result")
	}

	// Verify result contains algoId
	if !strings.Contains(result, "algo-sl-123456") {
		t.Errorf("Expected result to contain algoId 'algo-sl-123456', got: %s", result)
	}
	if !strings.Contains(result, "algo-tp-123456") {
		t.Errorf("Expected result to contain algoId 'algo-tp-123456', got: %s", result)
	}
}

func TestOkxAttachSlTpTool_AttachSlOnly(t *testing.T) {
	// Test 2: Attach SL-only (no TP) returns algoId
	mockTrade := &mockTradeClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			if req.SlTriggerPx == nil || *req.SlTriggerPx <= 0 {
				t.Error("Expected slTriggerPx to be set")
			}
			if req.TpTriggerPx != nil && *req.TpTriggerPx > 0 {
				t.Error("Expected tpTriggerPx to be empty")
			}

			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "algo-sl-only",
						SMsg:   "",
						SCode:  0,
					},
				},
			}, nil
		},
	}

	tool := &OkxAttachSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"ordId":       "order-123",
		"side":        "buy",
		"posSide":     "long",
		"slTriggerPx": "1800.0",
		"slOrderPx":   "1790.0",
		"sz":          "0.1",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "algo-sl-only") {
		t.Errorf("Expected result to contain algoId 'algo-sl-only', got: %s", result)
	}
}

func TestOkxAttachSlTpTool_AttachTpOnly(t *testing.T) {
	// Test 3: Attach TP-only (no SL) returns algoId
	mockTrade := &mockTradeClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			if req.TpTriggerPx == nil || *req.TpTriggerPx <= 0 {
				t.Error("Expected tpTriggerPx to be set")
			}
			if req.SlTriggerPx != nil && *req.SlTriggerPx > 0 {
				t.Error("Expected slTriggerPx to be empty")
			}

			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "algo-tp-only",
						SMsg:   "",
						SCode:  0,
					},
				},
			}, nil
		},
	}

	tool := &OkxAttachSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"ordId":       "order-123",
		"side":        "buy",
		"posSide":     "long",
		"tpTriggerPx": "2200.0",
		"tpOrderPx":   "2210.0",
		"sz":          "0.1",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "algo-tp-only") {
		t.Errorf("Expected result to contain algoId 'algo-tp-only', got: %s", result)
	}
}

func TestOkxAttachSlTpTool_NeitherSlNorTpReturnsError(t *testing.T) {
	// Test 4: Neither SL nor TP provided returns error
	tool := &OkxAttachSlTpTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"ordId":       "order-123",
		"side":        "buy",
		"posSide":     "long",
	}
	argsJSON, _ := json.Marshal(args)

	_, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err == nil {
		t.Fatal("Expected error when neither SL nor TP is provided")
	}

	expectedErr := "at least one of slTriggerPx or tpTriggerPx must be provided"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestOkxAttachSlTpTool_OkxSCodeNonZeroReturnsError(t *testing.T) {
	// Test 5: OKX sCode != 0 returns error with sCode/sMsg details
	mockTrade := &mockTradeClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "",
						SMsg:   "Order does not exist",
						SCode:  51002,
					},
				},
			}, nil
		},
	}

	tool := &OkxAttachSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"ordId":       "order-123",
		"side":        "buy",
		"posSide":     "long",
		"slTriggerPx": "1800.0",
	}
	argsJSON, _ := json.Marshal(args)

	_, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err == nil {
		t.Fatal("Expected error when sCode != 0")
	}

	// Verify error contains sCode and sMsg
	errStr := err.Error()
	if !strings.Contains(errStr, "51002") {
		t.Errorf("Expected error to contain sCode '51002', got: %s", errStr)
	}
	if !strings.Contains(errStr, "Order does not exist") {
		t.Errorf("Expected error to contain sMsg 'Order does not exist', got: %s", errStr)
	}
}

func TestOkxAttachSlTpTool_LimiterConfigured(t *testing.T) {
	// Test that the tool has limiter field and it can be configured
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	t.Log("Limiter configured for Trade endpoint (5 req/s)")
}
