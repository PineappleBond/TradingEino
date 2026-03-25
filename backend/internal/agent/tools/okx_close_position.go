package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	accountRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxClosePositionTool closes positions via OKX REST API with percentage support
type OkxClosePositionTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxClosePositionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-close-position",
		Desc:  "Close position partially or fully by percentage on OKX exchange",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP), required",
				Required: true,
			},
			"posSide": {
				Type:     schema.String,
				Desc:     "Position side: long/short/net, leave empty to query all",
				Required: false,
			},
			"percentage": {
				Type:     schema.Number,
				Desc:     "Close percentage 0-100, default 100 (full close)",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun executes the close position tool
func (c *OkxClosePositionTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstID     string  `json:"instID"`
		PosSide    string  `json:"posSide,omitempty"`
		Percentage float64 `json:"percentage,omitempty"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required fields
	if req.InstID == "" {
		return "", fmt.Errorf("instID is required")
	}

	// Validate percentage range
	percentage := req.Percentage
	// Check if percentage was explicitly provided (non-zero) and validate
	if req.Percentage != 0 && percentage < 0 {
		return "", fmt.Errorf("percentage must be between 0 and 100, got %.2f", percentage)
	}
	if percentage <= 0 {
		percentage = 100 // Default to 100% if not provided or zero
	}
	if percentage > 100 {
		return "", fmt.Errorf("percentage must be between 0 and 100, got %.2f", percentage)
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-close-position: rate limiter wait failed", err, "instID", req.InstID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v\n**交易对：** %s", err, req.InstID), nil
	}

	// For 100% close, use ClosePosition endpoint directly
	if percentage == 100 {
		return c.closeFullPosition(ctx, req.InstID, req.PosSide)
	}

	// For partial close (< 100%), we need to:
	// 1. Get current position size
	// 2. Calculate order size = positionSize * (percentage / 100)
	// 3. Place opposite market order
	return c.closePartialPosition(ctx, req.InstID, req.PosSide, percentage)
}

// closeFullPosition closes 100% of the position using ClosePosition endpoint
func (c *OkxClosePositionTool) closeFullPosition(ctx context.Context, instID, posSide string) (string, error) {
	// Build close position request
	closeReq := tradeRequests.ClosePosition{
		InstID:  instID,
		PosSide: okex.PositionSide(posSide),
		MgnMode: okex.MarginCrossMode,
	}

	// Call ClosePosition API
	resp, err := c.svcCtx.OKXClient.Rest.Trade.ClosePosition(closeReq)
	if err != nil {
		logger.Errorf(ctx, "okx-close-position: API call failed", err, "instID", instID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v\n**交易对：** %s", err, instID), nil
	}

	// Check response code (EXEC-06: sCode/sMsg validation)
	if resp.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-close-position: response code error", nil, "instID", instID, "code", resp.Code.Int(), "msg", resp.Msg)
		return fmt.Sprintf("**平仓失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** ClosePosition\n**交易对：** %s", resp.Code.Int(), resp.Msg, instID), nil
	}

	// Check for empty response
	if len(resp.ClosePositions) == 0 {
		logger.Errorf(ctx, "okx-close-position: empty response", nil, "instID", instID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** 仓位不存在或已平仓\n**交易对：** %s", instID), nil
	}

	result := resp.ClosePositions[0]

	// Format output
	output := c.formatClosePositionOutput(result.InstID, string(result.PosSide), 100)
	return output, nil
}

// closePartialPosition closes a percentage of the position by placing an opposite market order
func (c *OkxClosePositionTool) closePartialPosition(ctx context.Context, instID, posSide string, percentage float64) (string, error) {
	// 1. Get current position
	instIDs := []string{instID}
	posReq := accountRequests.GetPositions{
		InstID: instIDs,
	}

	// Wait for rate limiter for Account endpoint
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-close-position: rate limiter wait failed (GetPositions)", err, "instID", instID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	posResp, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(posReq)
	if err != nil {
		logger.Errorf(ctx, "okx-close-position: API call failed (GetPositions)", err, "instID", instID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** API 调用失败（查询仓位）\n**错误信息：** %v", err), nil
	}

	// Check response code
	if posResp.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-close-position: response code error (GetPositions)", nil, "instID", instID, "code", posResp.Code.Int(), "msg", posResp.Msg)
		return fmt.Sprintf("**平仓失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** GetPositions", posResp.Code.Int(), posResp.Msg), nil
	}

	// Find the position for the given instID and posSide
	var positionSize float64
	var actualPosSide string
	found := false

	for _, pos := range posResp.Positions {
		if pos.InstID == instID {
			// If posSide is specified, match it
			if posSide == "" || string(pos.PosSide) == posSide {
				positionSize = float64(pos.Pos)
				actualPosSide = string(pos.PosSide)
				found = true
				break
			}
		}
	}

	if !found || positionSize <= 0 {
		return "", fmt.Errorf("no open position found for %s", instID)
	}

	// 2. Calculate order size
	orderSize := positionSize * (percentage / 100)

	// 3. Determine opposite side for closing
	oppositeSide := getOppositeSide(actualPosSide)

	// 4. Place market order to close
	placeOrderReq := tradeRequests.PlaceOrder{
		InstID:  instID,
		Side:    okex.OrderSide(oppositeSide),
		PosSide: okex.PositionSide(actualPosSide),
		OrdType: okex.OrderMarket,
		Sz:      fmt.Sprintf("%.8f", orderSize),
		TdMode:  okex.TradeCrossMode,
	}

	// Wait for rate limiter for Trade endpoint
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-close-position: rate limiter wait failed (PlaceOrder)", err, "instID", instID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v\n**交易对：** %s", err, instID), nil
	}

	orderResp, err := c.svcCtx.OKXClient.Rest.Trade.PlaceOrder([]tradeRequests.PlaceOrder{placeOrderReq})
	if err != nil {
		logger.Errorf(ctx, "okx-close-position: API call failed (PlaceOrder)", err, "instID", instID)
		return fmt.Sprintf("**平仓失败**\n\n**错误类型：** API 调用失败（下单）\n**错误信息：** %v", err), nil
	}

	// Check response code
	if orderResp.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-close-position: response code error (PlaceOrder)", nil, "instID", instID, "code", orderResp.Code.Int(), "msg", orderResp.Msg)
		return fmt.Sprintf("**平仓失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** PlaceOrder", orderResp.Code.Int(), orderResp.Msg), nil
	}

	// Check for empty response
	if len(orderResp.PlaceOrders) == 0 {
		logger.Errorf(ctx, "okx-close-position: empty response (PlaceOrder)", nil, "instID", instID)
		return "**平仓失败**\n\n**错误类型：** 空响应", nil
	}

	orderResult := orderResp.PlaceOrders[0]

	// Check order-level sCode
	if int64(orderResult.SCode) != 0 {
		logger.Errorf(ctx, "okx-close-position: order-level sCode error", nil, "instID", instID, "sCode", orderResult.SCode, "sMsg", orderResult.SMsg)
		return fmt.Sprintf("**平仓失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** PlaceOrder", int64(orderResult.SCode), orderResult.SMsg), nil
	}

	// Format output
	output := c.formatClosePositionOutput(instID, actualPosSide, percentage)
	output += fmt.Sprintf("\n## Order Details\n\n")
	output += fmt.Sprintf("```markdown\n")
	output += fmt.Sprintf("| OrdId | Side | Size | State |\n")
	output += fmt.Sprintf("| :---- | :--- | :--- | :---- |\n")
	output += fmt.Sprintf("| %s | %s | %.8f | %s |\n", orderResult.OrdID, oppositeSide, orderSize, "submitted")
	output += fmt.Sprintf("```\n")

	return output, nil
}

// getOppositeSide returns the opposite side for closing a position
func getOppositeSide(posSide string) string {
	switch posSide {
	case "long":
		return "sell"
	case "short":
		return "buy"
	default: // net or empty - assume sell to close
		return "sell"
	}
}

// formatClosePositionOutput formats the close position result as a Markdown table
func (c *OkxClosePositionTool) formatClosePositionOutput(instID, posSide string, percentage float64) string {
	output := ""
	output += fmt.Sprintf("# Position Closed\n\n")
	output += "```markdown\n"
	output += "| InstId | PosSide | Closed Percentage |\n"
	output += "| :----- | :------ | :---------------- |\n"
	output += fmt.Sprintf("| %s | %s | %.2f%% |\n", instID, posSide, percentage)
	output += "\n```\n---\n\n"
	return output
}

// NewOkxClosePositionTool creates a new OkxClosePositionTool instance
func NewOkxClosePositionTool(svcCtx *svc.ServiceContext) *OkxClosePositionTool {
	return &OkxClosePositionTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade/Account endpoint
	}
}
