package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxCancelOrderTool cancels orders via OKX REST API
type OkxCancelOrderTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxCancelOrderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-cancel-order",
		Desc:  "Cancel pending orders on OKX exchange",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP)",
				Required: true,
			},
			"ordID": {
				Type:     schema.String,
				Desc:     "OKX order ID to cancel",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun executes the cancel order tool
func (c *OkxCancelOrderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
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
		return fmt.Sprintf("**撤单失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v\n**订单 ID：** %s", err, req.OrdID), nil
	}

	// Cancel the order
	resp, err := c.svcCtx.OKXClient.Rest.Trade.CandleOrder([]tradeRequests.CancelOrder{
		{
			InstID: req.InstID,
			OrdID:  req.OrdID,
		},
	})
	if err != nil {
		return fmt.Sprintf("**撤单失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v\n**订单 ID：** %s\n**交易对：** %s", err, req.OrdID, req.InstID), nil
	}

	// Check response code
	if resp.Code.Int() != 0 {
		return fmt.Sprintf("**撤单失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** CancelOrder\n**订单 ID：** %s\n**交易对：** %s", resp.Code.Int(), resp.Msg, req.OrdID, req.InstID), nil
	}

	// Check for empty response
	if len(resp.CancelOrders) == 0 {
		return fmt.Sprintf("**撤单失败**\n\n**错误类型：** 空响应\n**订单 ID：** %s\n**交易对：** %s", req.OrdID, req.InstID), nil
	}

	result := resp.CancelOrders[0]

	// Check for order-level errors (sCode/sMsg validation - EXEC-06 requirement)
	sCode := float64(result.SCode)

	if sCode != 0 {
		return fmt.Sprintf("**撤单失败**\n\n**错误代码：** %.0f\n**错误信息：** %s\n**订单 ID：** %s\n**交易对：** %s", sCode, result.SMsg, req.OrdID, req.InstID), nil
	}

	// Format output as Markdown table
	output := c.formatCancelOrderOutput(result)
	return output, nil
}

// formatCancelOrderOutput formats the cancel order result as a Markdown table
func (c *OkxCancelOrderTool) formatCancelOrderOutput(result *trade.CancelOrder) string {
	output := ""
	output += "# Order Cancelled\n\n"
	output += "```markdown\n"
	output += "| OrdId | ClOrdId | State | SCode | SMsg |\n"
	output += "| :---- | :------ | :---- | :---- | :--- |\n"

	// Convert SCode to string properly
	sCodeStr := fmt.Sprintf("%.0f", float64(result.SCode))

	output += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
		result.OrdID,
		result.ClOrdID,
		"cancelled",
		sCodeStr,
		result.SMsg,
	)
	output += "\n```\n---\n\n"
	return output
}

// NewOkxCancelOrderTool creates a new OkxCancelOrderTool instance
func NewOkxCancelOrderTool(svcCtx *svc.ServiceContext) *OkxCancelOrderTool {
	return &OkxCancelOrderTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
