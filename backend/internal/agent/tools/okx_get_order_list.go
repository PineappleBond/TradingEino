package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/internal/utils/xmd"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	trademodels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxGetOrderListTool queries all pending orders via OKX REST API
type OkxGetOrderListTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxGetOrderListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-order-list",
		Desc:  "Query all pending (incomplete) orders under the current account",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Filter by instrument ID (e.g., BTC-USDT-SWAP), optional - leave empty for all",
				Required: false,
			},
			"instType": {
				Type:     schema.String,
				Desc:     "Filter by instrument type: SWAP, FUTURES, SPOT, OPTION, MARGIN",
				Enum:     []string{"SWAP", "FUTURES", "SPOT", "OPTION", "MARGIN"},
				Required: false,
			},
			"state": {
				Type:     schema.String,
				Desc:     "Filter by order state: live, partially_filled",
				Enum:     []string{"live", "partially_filled"},
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun executes the get order list tool
func (c *OkxGetOrderListTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstID   string `json:"instID"`
		InstType string `json:"instType"`
		State    string `json:"state"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-get-order-list: rate limiter wait failed", err)
		return fmt.Sprintf("**订单列表查询失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	// Build request
	orderReq := traderequests.OrderList{}
	if req.InstID != "" {
		orderReq.InstID = req.InstID
	}
	if req.InstType != "" {
		orderReq.InstType = okex.InstrumentType(req.InstType)
	}
	if req.State != "" {
		orderReq.State = okex.OrderState(req.State)
	}

	// Get order list
	resp, err := c.svcCtx.OKXClient.Rest.Trade.GetOrderList(orderReq)
	if err != nil {
		logger.Errorf(ctx, "okx-get-order-list: API call failed", err)
		return fmt.Sprintf("**订单列表查询失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v", err), nil
	}

	// Check response code
	if resp.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-get-order-list: response code error", nil, "code", resp.Code.Int(), "msg", resp.Msg)
		return fmt.Sprintf("**订单列表查询失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** GetOrderList", resp.Code.Int(), resp.Msg), nil
	}

	// Format output as Markdown table
	output := c.formatOutput(resp.Orders)
	return output, nil
}

// formatOutput formats the order list as a Markdown table
func (c *OkxGetOrderListTool) formatOutput(orders []*trademodels.Order) string {
	if len(orders) == 0 {
		return "## 当前无未完成订单\n\n当前账户下没有任何未成交或部分成交的订单。"
	}

	output := ""
	output += fmt.Sprintf("## 当前未完成订单 (%d 个)\n\n", len(orders))

	headers := []string{"订单 ID", "交易对", "方向", "仓位方向", "类型", "状态", "数量", "价格", "已成交", "未成交", "杠杆"}
	rows := [][]string{}

	for _, order := range orders {
		sz := float64(order.Sz)
		accFillSz := float64(order.AccFillSz)
		unfillSz := sz - accFillSz

		rows = append(rows, []string{
			order.OrdID,
			order.InstID,
			string(order.Side),
			string(order.PosSide),
			string(order.OrdType),
			string(order.State),
			fmt.Sprintf("%.4f", sz),
			fmt.Sprintf("%.4f", float64(order.Px)),
			fmt.Sprintf("%.4f", accFillSz),
			fmt.Sprintf("%.4f", unfillSz),
			fmt.Sprintf("%.0f", float64(order.Lever)),
		})
	}

	table := xmd.CreateMarkdownTable(headers, rows)
	output += "```markdown\n"
	output += table
	output += "\n```\n---\n\n"

	return output
}

// NewOkxGetOrderListTool creates a new OkxGetOrderListTool instance
func NewOkxGetOrderListTool(svcCtx *svc.ServiceContext) *OkxGetOrderListTool {
	return &OkxGetOrderListTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
