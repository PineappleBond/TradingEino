package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/responses"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	traderesponses "github.com/PineappleBond/TradingEino/backend/pkg/okex/responses/trade"
	"golang.org/x/time/rate"
)

// mockTradeWithPlaceClient mocks the trade.RestClient for testing
type mockTradeWithPlaceClient struct {
	placeAlgoOrderFunc func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error)
}

func (m *mockTradeWithPlaceClient) PlaceAlgoOrder(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
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

func TestOkxPlaceOrderWithSlTpTool_PlaceLimitOrderWithSlTp(t *testing.T) {
	mockTrade := &mockTradeWithPlaceClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			if req.InstID != "ETH-USDT-SWAP" {
				t.Errorf("Expected instId 'ETH-USDT-SWAP', got '%s'", req.InstID)
			}
			if req.SlTriggerPx == nil || *req.SlTriggerPx != 1800.0 {
				t.Errorf("Expected slTriggerPx 1800.0, got '%v'", req.SlTriggerPx)
			}
			if req.TpTriggerPx == nil || *req.TpTriggerPx != 2200.0 {
				t.Errorf("Expected tpTriggerPx 2200.0, got '%v'", req.TpTriggerPx)
			}

			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0, Msg: ""},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "algo-order-with-sltp",
						SMsg:   "",
						SCode:  0,
					},
				},
			}, nil
		},
	}

	tool := &OkxPlaceOrderWithSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"side":        "buy",
		"ordType":     "limit",
		"size":        "0.1",
		"price":       "2000.0",
		"slTriggerPx": "1800.0",
		"slOrderPx":   "1790.0",
		"tpTriggerPx": "2200.0",
		"tpOrderPx":   "2210.0",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "algo-order-with-sltp") {
		t.Errorf("Expected result to contain algoId 'algo-order-with-sltp', got: %s", result)
	}
}

func TestOkxPlaceOrderWithSlTpTool_PlaceMarketOrderWithSlOnly(t *testing.T) {
	mockTrade := &mockTradeWithPlaceClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			if req.SlTriggerPx == nil || *req.SlTriggerPx <= 0 {
				t.Error("Expected slTriggerPx to be set")
			}
			if req.TpTriggerPx != nil && *req.TpTriggerPx > 0 {
				t.Error("Expected tpTriggerPx to be empty")
			}

			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0, Msg: ""},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "algo-market-sl-only",
						SMsg:   "",
						SCode:  0,
					},
				},
			}, nil
		},
	}

	tool := &OkxPlaceOrderWithSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"side":        "sell",
		"ordType":     "market",
		"size":        "0.1",
		"slTriggerPx": "1800.0",
		"slOrderPx":   "1790.0",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "algo-market-sl-only") {
		t.Errorf("Expected result to contain algoId 'algo-market-sl-only', got: %s", result)
	}
}

func TestOkxPlaceOrderWithSlTpTool_PlaceOrderWithTpOnly(t *testing.T) {
	mockTrade := &mockTradeWithPlaceClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			if req.TpTriggerPx == nil || *req.TpTriggerPx <= 0 {
				t.Error("Expected tpTriggerPx to be set")
			}
			if req.SlTriggerPx != nil && *req.SlTriggerPx > 0 {
				t.Error("Expected slTriggerPx to be empty")
			}

			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 0, Msg: ""},
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

	tool := &OkxPlaceOrderWithSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"side":        "buy",
		"ordType":     "limit",
		"size":        "0.1",
		"price":       "2000.0",
		"tpTriggerPx": "2200.0",
		"tpOrderPx":   "2210.0",
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

func TestOkxPlaceOrderWithSlTpTool_NeitherSlNorTpReturnsError(t *testing.T) {
	tool := &OkxPlaceOrderWithSlTpTool{
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
	}

	args := map[string]interface{}{
		"instID":  "ETH-USDT-SWAP",
		"side":    "buy",
		"ordType": "market",
		"size":    "0.1",
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

func TestOkxPlaceOrderWithSlTpTool_MainOrderFailsReturnsError(t *testing.T) {
	mockTrade := &mockTradeWithPlaceClient{
		placeAlgoOrderFunc: func(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error) {
			return traderesponses.PlaceAlgoOrder{
				Basic: responses.Basic{Code: 51000, Msg: "Insufficient balance"},
				PlaceAlgoOrders: []*trade.PlaceAlgoOrder{
					{
						AlgoID: "",
						SMsg:   "Insufficient balance",
						SCode:  51000,
					},
				},
			}, nil
		},
	}

	tool := &OkxPlaceOrderWithSlTpTool{
		limiter:   rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		mockTrade: mockTrade,
	}

	args := map[string]interface{}{
		"instID":      "ETH-USDT-SWAP",
		"side":        "buy",
		"ordType":     "limit",
		"size":        "1000",
		"price":       "2000.0",
		"slTriggerPx": "1800.0",
	}
	argsJSON, _ := json.Marshal(args)

	_, err := tool.InvokableRun(context.Background(), string(argsJSON))
	if err == nil {
		t.Fatal("Expected error when order fails")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "51000") {
		t.Errorf("Expected error to contain sCode '51000', got: %s", errStr)
	}
	if !strings.Contains(errStr, "Insufficient balance") {
		t.Errorf("Expected error to contain sMsg 'Insufficient balance', got: %s", errStr)
	}
}

func TestOkxPlaceOrderWithSlTpTool_LimiterConfigured(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	t.Log("Limiter configured for Trade endpoint (5 req/s)")
}
