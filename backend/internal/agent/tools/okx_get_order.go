package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxGetOrderTool queries order details via OKX REST API
type OkxGetOrderTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxGetOrderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-order",
		Desc:  "Query order status and details from OKX exchange",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP)",
				Required: true,
			},
			"ordID": {
				Type:     schema.String,
				Desc:     "OKX order ID to query",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun executes the get order tool
func (c *OkxGetOrderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstID string `json:"instID"`
		OrdID  string `json:"ordID"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if req.InstID == "" {
		return "", fmt.Errorf("instID is required")
	}
	if req.OrdID == "" {
		return "", fmt.Errorf("ordID is required")
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Get order details
	resp, err := c.svcCtx.OKXClient.Rest.Trade.GetOrderDetail(tradeRequests.OrderDetails{
		InstID: req.InstID,
		OrdID:  req.OrdID,
	})
	if err != nil {
		return "", err
	}

	// Check response code
	if resp.Code != 0 {
		return "", &okex.OKXError{
			Code:     resp.Code,
			Msg:      resp.Msg,
			Endpoint: "GetOrderDetail",
		}
	}

	// Check for empty response
	if len(resp.Orders) == 0 {
		return "", fmt.Errorf("order not found: %s", req.OrdID)
	}

	order := resp.Orders[0]

	// Format output as Markdown table with comprehensive order details
	output := c.formatGetOrderOutput(order)
	return output, nil
}

// formatGetOrderOutput formats the order details as a Markdown table
func (c *OkxGetOrderTool) formatGetOrderOutput(order *trade.Order) string {
	output := ""
	output += "# Order Details\n\n"
	output += "```markdown\n"
	output += "| Field | Value |\n"
	output += "| :---- | :---- |\n"
	output += fmt.Sprintf("| Order ID | %s |\n", order.OrdID)
	output += fmt.Sprintf("| Instrument ID | %s |\n", order.InstID)
	output += fmt.Sprintf("| Side | %s |\n", order.Side)
	output += fmt.Sprintf("| Position Side | %s |\n", order.PosSide)
	output += fmt.Sprintf("| Order Type | %s |\n", order.OrdType)
	output += fmt.Sprintf("| State | %s |\n", order.State)

	// Size and price
	sz := float64(order.Sz)
	accFillSz := float64(order.AccFillSz)
	unfillSz := sz - accFillSz

	output += fmt.Sprintf("| Size | %g |\n", sz)
	output += fmt.Sprintf("| Price | %g |\n", float64(order.Px))
	output += fmt.Sprintf("| Average Price | %g |\n", float64(order.AvgPx))
	output += fmt.Sprintf("| Filled Size | %g |\n", accFillSz)
	output += fmt.Sprintf("| Unfilled Size | %g |\n", unfillSz)

	// Additional details
	output += fmt.Sprintf("| Fee Currency | %s |\n", order.FeeCcy)
	output += fmt.Sprintf("| Fee | %g |\n", float64(order.Fee))
	output += fmt.Sprintf("| Leverage | %g |\n", float64(order.Lever))
	output += fmt.Sprintf("| Trade Mode | %s |\n", order.TdMode)

	// Timestamps
	if !time.Time(order.UTime).IsZero() {
		output += fmt.Sprintf("| Update Time | %s |\n", time.Time(order.UTime).Format("2006-01-02 15:04:05"))
	}
	if !time.Time(order.CTime).IsZero() {
		output += fmt.Sprintf("| Create Time | %s |\n", time.Time(order.CTime).Format("2006-01-02 15:04:05"))
	}

	output += "\n```\n---\n\n"
	return output
}

// NewOkxGetOrderTool creates a new OkxGetOrderTool instance
func NewOkxGetOrderTool(svcCtx *svc.ServiceContext) *OkxGetOrderTool {
	return &OkxGetOrderTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
