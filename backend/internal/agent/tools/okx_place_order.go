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

// OkxPlaceOrderTool places orders via OKX REST API
type OkxPlaceOrderTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxPlaceOrderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-place-order",
		Desc:  "Place limit or market orders on OKX exchange",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP)",
				Required: true,
			},
			"side": {
				Type:     schema.String,
				Desc:     "Order side: buy or sell",
				Enum:     []string{"buy", "sell"},
				Required: true,
			},
			"posSide": {
				Type:     schema.String,
				Desc:     "Position side: long, short, or net (default: net)",
				Enum:     []string{"long", "short", "net"},
				Required: false,
			},
			"ordType": {
				Type:     schema.String,
				Desc:     "Order type: market, limit, post_only, fok, ioc",
				Enum:     []string{"market", "limit", "post_only", "fok", "ioc"},
				Required: true,
			},
			"size": {
				Type:     schema.String,
				Desc:     "Order size (number of contracts)",
				Required: true,
			},
			"price": {
				Type:     schema.String,
				Desc:     "Order price (required for limit/post_only, empty for market)",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun executes the place order tool
func (c *OkxPlaceOrderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstID  string `json:"instID"`
		Side    string `json:"side"`
		PosSide string `json:"posSide"`
		OrdType string `json:"ordType"`
		Size    string `json:"size"`
		Price   string `json:"price"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if req.InstID == "" {
		return "", fmt.Errorf("instID is required")
	}
	if req.Side == "" {
		return "", fmt.Errorf("side is required (buy or sell)")
	}
	if req.OrdType == "" {
		return "", fmt.Errorf("ordType is required (market, limit, post_only, fok, ioc)")
	}
	if req.Size == "" {
		return "", fmt.Errorf("size is required")
	}

	// Validate price for limit orders
	if (req.OrdType == "limit" || req.OrdType == "post_only") && req.Price == "" {
		return "", fmt.Errorf("price is required for %s order type", req.OrdType)
	}

	// Set default posSide to "net" if not provided
	if req.PosSide == "" {
		req.PosSide = "net"
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Place the order
	resp, err := c.svcCtx.OKXClient.Rest.Trade.PlaceOrder([]tradeRequests.PlaceOrder{
		{
			InstID:  req.InstID,
			Side:    okex.OrderSide(req.Side),
			PosSide: okex.PositionSide(req.PosSide),
			OrdType: okex.OrderType(req.OrdType),
			Sz:      req.Size,
			Px:      req.Price,
			TdMode:  okex.TradeCrossMode,
		},
	})
	if err != nil {
		// Return formatted error message to Agent (not error)
		return fmt.Sprintf("**订单执行失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			err, req.InstID, req.Side, req.Size), nil
	}

	// Check response code
	if resp.Code.Int() != 0 {
		// Return formatted error message to Agent (not error)
		return fmt.Sprintf("**订单执行失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** PlaceOrder\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			resp.Code.Int(), resp.Msg, req.InstID, req.Side, req.Size), nil
	}

	// Check for empty response
	if len(resp.PlaceOrders) == 0 {
		return fmt.Sprintf("**订单执行失败**\n\n**错误类型：** 空响应\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			req.InstID, req.Side, req.Size), nil
	}

	result := resp.PlaceOrders[0]

	// Check for order-level errors (sCode/sMsg validation - EXEC-06 requirement)
	var sCode int64
	if err := result.SCode.UnmarshalJSON([]byte(fmt.Sprintf("%d", result.SCode))); err != nil {
		sCode = 0
	} else {
		sCode = int64(result.SCode)
	}

	if sCode != 0 {
		// Return formatted error message to Agent (not error)
		return fmt.Sprintf("**订单执行失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			int(sCode), result.SMsg, req.InstID, req.Side, req.Size), nil
	}

	// Format output as Markdown table
	output := c.formatPlaceOrderOutput(result)
	return output, nil
}

// formatPlaceOrderOutput formats the place order result as a Markdown table
func (c *OkxPlaceOrderTool) formatPlaceOrderOutput(result *trade.PlaceOrder) string {
	output := ""
	output += "# Order Placed\n\n"
	output += "```markdown\n"
	output += "| OrdId | ClOrdId | Tag | State | SCode | SMsg |\n"
	output += "| :---- | :------ | :-- | :---- | :---- | :--- |\n"
	output += fmt.Sprintf("| %s | %s | %s | %s | %d | %s |\n",
		result.OrdID,
		result.ClOrdID,
		result.Tag,
		"live",
		int64(result.SCode),
		result.SMsg,
	)
	output += "\n```\n---\n\n"
	return output
}

// NewOkxPlaceOrderTool creates a new OkxPlaceOrderTool instance
func NewOkxPlaceOrderTool(svcCtx *svc.ServiceContext) *OkxPlaceOrderTool {
	return &OkxPlaceOrderTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
