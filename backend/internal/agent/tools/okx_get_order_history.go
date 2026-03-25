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

// OkxGetOrderHistoryTool queries historical orders via OKX REST API
type OkxGetOrderHistoryTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxGetOrderHistoryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-order-history",
		Desc:  "Query historical orders (last 7 days) with optional time range and instrument filters",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instType": {
				Type:     schema.String,
				Desc:     "Instrument type (SPOT/SWAP/FUTURES/OPTION/MARGIN), optional but recommended for filtering",
				Required: false,
			},
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP), leave empty for all instruments",
				Required: false,
			},
			"startTime": {
				Type:     schema.String,
				Desc:     "Start time in Unix milliseconds timestamp, optional",
				Required: false,
			},
			"endTime": {
				Type:     schema.String,
				Desc:     "End time in Unix milliseconds timestamp, optional",
				Required: false,
			},
			"limit": {
				Type:     schema.Number,
				Desc:     "Maximum number of orders to return, default 100",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun executes the get order history tool
func (c *OkxGetOrderHistoryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstType  string `json:"instType,omitempty"`
		InstID    string `json:"instID,omitempty"`
		StartTime string `json:"startTime,omitempty"`
		EndTime   string `json:"endTime,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Set default limit
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	// Build OKX request
	okxReq := tradeRequests.OrderList{
		InstID: req.InstID,
		Limit:  float64(limit),
	}

	// Add instrument type if provided
	if req.InstType != "" {
		okxReq.InstType = okex.InstrumentType(req.InstType)
	}

	// Add time range filters if provided
	if req.StartTime != "" {
		// Parse start time as Unix milliseconds
		var startMs float64
		fmt.Sscanf(req.StartTime, "%f", &startMs)
		if startMs > 0 {
			okxReq.Before = startMs // Note: OKX uses 'before' for end time in history endpoint
		}
	}
	if req.EndTime != "" {
		// Parse end time as Unix milliseconds
		var endMs float64
		fmt.Sscanf(req.EndTime, "%f", &endMs)
		if endMs > 0 {
			okxReq.After = endMs // Note: OKX uses 'after' for start time in history endpoint
		}
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Sprintf("**订单历史查询失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	// Get order history (arch=false for last 7 days)
	resp, err := c.svcCtx.OKXClient.Rest.Trade.GetOrderHistory(okxReq, false)
	if err != nil {
		return fmt.Sprintf("**订单历史查询失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v", err), nil
	}

	// Check response code (EXEC-06: sCode/sMsg validation)
	if resp.Code.Int() != 0 {
		return fmt.Sprintf("**订单历史查询失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** GetOrderHistory", resp.Code.Int(), resp.Msg), nil
	}

	// Empty result set is valid - return empty table
	if len(resp.Orders) == 0 {
		return c.formatEmptyOrderHistory(), nil
	}

	// Format output as Markdown table
	output := c.formatOrderHistoryOutput(resp.Orders)
	return output, nil
}

// formatOrderHistoryOutput formats the order history result as a Markdown table
func (c *OkxGetOrderHistoryTool) formatOrderHistoryOutput(orders []*tradeModels.Order) string {
	output := ""
	output += fmt.Sprintf("# Order History (%d orders)\n\n", len(orders))
	output += "```markdown\n"
	output += "| ordId | instId | instType | side | posSide | ordType | size | avgPx | fillSize | state | cTime |\n"
	output += "| :---- | :----- | :------- | :--- | :------ | :------ | :--- | :---- | :------- | :---- | :---- |\n"

	for _, order := range orders {
		avgPxStr := fmt.Sprintf("%.2f", float64(order.AvgPx))
		if order.AvgPx == 0 {
			avgPxStr = "-"
		}
		fillSizeStr := fmt.Sprintf("%.2f", float64(order.FillSz))
		if order.FillSz == 0 {
			fillSizeStr = "-"
		}
		cTimeStr := time.Time(order.CTime).Format("2006-01-02 15:04:05")

		output += fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %.2f | %s | %s | %s | %s |\n",
			order.OrdID,
			order.InstID,
			order.InstType,
			order.Side,
			order.PosSide,
			order.OrdType,
			float64(order.Sz),
			avgPxStr,
			fillSizeStr,
			order.State,
			cTimeStr,
		)
	}
	output += "\n```\n---\n\n"
	return output
}

// formatEmptyOrderHistory returns a message for empty order history
func (c *OkxGetOrderHistoryTool) formatEmptyOrderHistory() string {
	output := ""
	output += "# Order History\n\n"
	output += "No orders found in the specified time range.\n"
	output += "---\n\n"
	return output
}

// NewOkxGetOrderHistoryTool creates a new OkxGetOrderHistoryTool instance
func NewOkxGetOrderHistoryTool(svcCtx *svc.ServiceContext) *OkxGetOrderHistoryTool {
	return &OkxGetOrderHistoryTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
