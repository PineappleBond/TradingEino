package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	tradeModels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxCancelAllOrdersTool cancels all pending orders via OKX REST API
type OkxCancelAllOrdersTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxCancelAllOrdersTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-cancel-all-orders",
		Desc:  "Cancel all pending orders for a specific instrument or all instruments. First queries all pending orders, then cancels them in batches.",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP), leave empty to cancel orders for all instruments",
				Required: false,
			},
			"instType": {
				Type:     schema.String,
				Desc:     "Instrument type: SPOT/SWAP/FUTURES/OPTION/MARGIN, default SWAP",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun executes the cancel all orders tool
func (c *OkxCancelAllOrdersTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstID   string `json:"instID,omitempty"`
		InstType string `json:"instType,omitempty"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Default to SWAP if not specified
	instType := req.InstType
	if instType == "" {
		instType = "SWAP"
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// 1. Query all pending orders
	orderListReq := tradeRequests.OrderList{
		InstType: okex.InstrumentType(instType),
	}
	if req.InstID != "" {
		orderListReq.InstID = req.InstID
	}

	resp, err := c.svcCtx.OKXClient.Rest.Trade.GetOrderList(orderListReq)
	if err != nil {
		return "", err
	}

	// Check response code
	if resp.Code.Int() != 0 {
		return "", &okex.OKXError{
			Code:     resp.Code.Int(),
			Msg:      resp.Msg,
			Endpoint: "GetOrderList",
		}
	}

	// Check if there are any pending orders
	if len(resp.Orders) == 0 {
		return c.formatNoOrdersOutput(instType, req.InstID), nil
	}

	// 2. Build cancel requests (max 20 per batch)
	var cancelRequests []tradeRequests.CancelOrder
	for _, order := range resp.Orders {
		cancelRequests = append(cancelRequests, tradeRequests.CancelOrder{
			InstID: order.InstID,
			OrdID:  order.OrdID,
		})
	}

	// 3. Cancel orders in batches of 20
	totalCancelled := 0
	totalFailed := 0
	var failedOrders []*tradeModels.CancelOrder

	for i := 0; i < len(cancelRequests); i += 20 {
		end := i + 20
		if end > len(cancelRequests) {
			end = len(cancelRequests)
		}
		batch := cancelRequests[i:end]

		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return "", fmt.Errorf("rate limiter wait failed: %w", err)
		}

		// Cancel batch
		batchResp, err := c.svcCtx.OKXClient.Rest.Trade.CandleOrder(batch)
		if err != nil {
			return "", fmt.Errorf("batch cancel failed: %w", err)
		}

		// Check response code
		if batchResp.Code.Int() != 0 {
			return "", &okex.OKXError{
				Code:     batchResp.Code.Int(),
				Msg:      batchResp.Msg,
				Endpoint: "CandleOrder",
			}
		}

		// Count results
		for _, result := range batchResp.CancelOrders {
			if float64(result.SCode) == 0 {
				totalCancelled++
			} else {
				totalFailed++
				failedOrders = append(failedOrders, result)
			}
		}
	}

	return c.formatCancelAllOutput(totalCancelled, totalFailed, failedOrders, len(resp.Orders)), nil
}

// formatNoOrdersOutput returns a message when no orders are found
func (c *OkxCancelAllOrdersTool) formatNoOrdersOutput(instType, instID string) string {
	output := ""
	output += "# Cancel All Orders\n\n"
	if instID != "" {
		output += fmt.Sprintf("No pending orders found for %s (%s).\n", instID, instType)
	} else {
		output += fmt.Sprintf("No pending orders found for %s.\n", instType)
	}
	output += "---\n\n"
	return output
}

// formatCancelAllOutput formats the cancel all result
func (c *OkxCancelAllOrdersTool) formatCancelAllOutput(cancelled, failed int, failedOrders []*tradeModels.CancelOrder, total int) string {
	output := ""
	output += fmt.Sprintf("# Cancel All Orders Results\n\n")
	output += fmt.Sprintf("Total pending orders: %d\n", total)
	output += fmt.Sprintf("Successfully cancelled: %d\n", cancelled)
	output += fmt.Sprintf("Failed: %d\n\n", failed)

	if len(failedOrders) > 0 {
		output += "## Failed Orders\n\n"
		output += "```markdown\n"
		output += "| OrdId | SCode | SMsg |\n"
		output += "| :---- | :---- | :--- |\n"
		for _, f := range failedOrders {
			sCodeStr := fmt.Sprintf("%.0f", float64(f.SCode))
			output += fmt.Sprintf("| %s | %s | %s |\n",
				f.OrdID,
				sCodeStr,
				f.SMsg,
			)
		}
		output += "\n```\n\n"
	}

	output += "---\n\n"
	return output
}

// NewOkxCancelAllOrdersTool creates a new OkxCancelAllOrdersTool instance
func NewOkxCancelAllOrdersTool(svcCtx *svc.ServiceContext) *OkxCancelAllOrdersTool {
	return &OkxCancelAllOrdersTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
