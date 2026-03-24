package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	tradeModels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxBatchPlaceOrderTool places multiple orders (max 20) via OKX REST API
type OkxBatchPlaceOrderTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxBatchPlaceOrderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-batch-place-order",
		Desc:  "Place multiple orders (max 20) in a single batch call on OKX exchange",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"orders": {
				Type: schema.Array,
				Desc: "Array of order objects. Each order must contain: instID (string, required), side (buy/sell, required), posSide (long/short/net, required), ordType (market/limit/post_only/fok/ioc, required), size (string, required), price (string, optional, required for limit/post_only orders)",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun executes the batch place order tool
func (c *OkxBatchPlaceOrderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type OrderRequest struct {
		InstID  string `json:"instID"`
		Side    string `json:"side"`
		PosSide string `json:"posSide"`
		OrdType string `json:"ordType"`
		Size    string `json:"size"`
		Price   string `json:"price,omitempty"`
	}

	type Request struct {
		Orders []OrderRequest `json:"orders"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate orders array is not empty
	if len(req.Orders) == 0 {
		return "", fmt.Errorf("orders array cannot be empty")
	}

	// Validate max 20 orders (OKX limit)
	if len(req.Orders) > 20 {
		return "", fmt.Errorf("maximum 20 orders allowed per batch call, got %d", len(req.Orders))
	}

	// Validate each order and convert to OKX request format
	orders := make([]tradeRequests.PlaceOrder, 0, len(req.Orders))
	for i, order := range req.Orders {
		// Validate required fields
		if order.InstID == "" {
			return "", fmt.Errorf("order[%d]: instID is required", i)
		}
		if order.Side == "" {
			return "", fmt.Errorf("order[%d]: side is required (buy/sell)", i)
		}
		if order.OrdType == "" {
			return "", fmt.Errorf("order[%d]: ordType is required (market/limit/post_only/fok/ioc)", i)
		}
		if order.Size == "" {
			return "", fmt.Errorf("order[%d]: size is required", i)
		}

		// Validate price for limit/post_only orders
		if (order.OrdType == "limit" || order.OrdType == "post_only") && order.Price == "" {
			return "", fmt.Errorf("order[%d]: price is required for limit/post_only orders", i)
		}

		// Convert to OKX request format
		ordRequest := tradeRequests.PlaceOrder{
			InstID:  order.InstID,
			Side:    okex.OrderSide(order.Side),
			PosSide: okex.PositionSide(order.PosSide),
			OrdType: okex.OrderType(order.OrdType),
			Sz:      order.Size,
			Px:      order.Price,
			TdMode:  okex.TradeCrossMode, // Default to cross margin mode
		}
		orders = append(orders, ordRequest)
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Place batch orders
	resp, err := c.svcCtx.OKXClient.Rest.Trade.PlaceMultipleOrders(orders)
	if err != nil {
		return "", err
	}

	// Check response code (EXEC-06: sCode/sMsg validation)
	if resp.Code != 0 {
		return "", &okex.OKXError{
			Code:     resp.Code,
			Msg:      resp.Msg,
			Endpoint: "PlaceMultipleOrders",
		}
	}

	// Check for empty response
	if len(resp.PlaceOrders) == 0 {
		return "", fmt.Errorf("batch place order failed: empty response")
	}

	// Categorize results by sCode
	successes := make([]*tradeModels.PlaceOrder, 0)
	failures := make([]*tradeModels.PlaceOrder, 0)

	for _, result := range resp.PlaceOrders {
		if int64(result.SCode) == 0 {
			successes = append(successes, result)
		} else {
			failures = append(failures, result)
		}
	}

	// Format output as Markdown tables
	output := c.formatBatchPlaceOrderOutput(successes, failures, len(req.Orders))
	return output, nil
}

// formatBatchPlaceOrderOutput formats the batch place order result as Markdown tables
func (c *OkxBatchPlaceOrderTool) formatBatchPlaceOrderOutput(successes, failures []*tradeModels.PlaceOrder, total int) string {
	output := ""
	output += fmt.Sprintf("# Batch Order Placement (%d orders)\n\n", total)

	// Success section
	if len(successes) > 0 {
		output += "## 成功订单\n\n"
		output += "```markdown\n"
		output += "| OrdId | ClOrdId | Tag | SCode | SMsg |\n"
		output += "| :---- | :------ | :-- | :---- | :--- |\n"
		for _, s := range successes {
			sCodeStr := fmt.Sprintf("%d", int64(s.SCode))
			output += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				s.OrdID,
				s.ClOrdID,
				s.Tag,
				sCodeStr,
				s.SMsg,
			)
		}
		output += "\n```\n\n"
	}

	// Failure section
	if len(failures) > 0 {
		output += "## 失败订单\n\n"
		output += "```markdown\n"
		output += "| 请求索引 | SCode | SMsg |\n"
		output += "| :------- | :---- | :--- |\n"
		for i, f := range failures {
			sCodeStr := fmt.Sprintf("%d", int64(f.SCode))
			output += fmt.Sprintf("| %d | %s | %s |\n",
				i+1,
				sCodeStr,
				f.SMsg,
			)
		}
		output += "\n```\n\n"
	}

	output += "---\n\n"
	return output
}

// NewOkxBatchPlaceOrderTool creates a new OkxBatchPlaceOrderTool instance
func NewOkxBatchPlaceOrderTool(svcCtx *svc.ServiceContext) *OkxBatchPlaceOrderTool {
	return &OkxBatchPlaceOrderTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
