package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	tradeModels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxBatchCancelOrderTool cancels multiple orders (max 20) via OKX REST API
type OkxBatchCancelOrderTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxBatchCancelOrderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-batch-cancel-order",
		Desc:  "Cancel multiple pending orders (max 20) in a single batch call on OKX exchange",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"orders": {
				Type: schema.Array,
				Desc: "Array of cancel request objects. Each request must contain: instID (string, required), ordID (string, required)",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun executes the batch cancel order tool
func (c *OkxBatchCancelOrderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type CancelRequest struct {
		InstID string `json:"instID"`
		OrdID  string `json:"ordID"`
	}

	type Request struct {
		Orders []CancelRequest `json:"orders"`
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
	orders := make([]tradeRequests.CancelOrder, 0, len(req.Orders))
	for i, order := range req.Orders {
		// Validate required fields
		if order.InstID == "" {
			return "", fmt.Errorf("order[%d]: instID is required", i)
		}
		if order.OrdID == "" {
			return "", fmt.Errorf("order[%d]: ordID is required", i)
		}

		// Convert to OKX request format
		cancelRequest := tradeRequests.CancelOrder{
			InstID: order.InstID,
			OrdID:  order.OrdID,
		}
		orders = append(orders, cancelRequest)
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-batch-cancel-order: rate limiter wait failed", err)
		return fmt.Sprintf("**批量撤单失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	// Cancel batch orders (CandleOrder handles batch when len > 1)
	resp, err := c.svcCtx.OKXClient.Rest.Trade.CandleOrder(orders)
	if err != nil {
		logger.Errorf(ctx, "okx-batch-cancel-order: API call failed", err)
		return fmt.Sprintf("**批量撤单失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v", err), nil
	}

	// Check response code (EXEC-06: sCode/sMsg validation)
	if resp.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-batch-cancel-order: response code error", nil, "code", resp.Code.Int(), "msg", resp.Msg)
		return fmt.Sprintf("**批量撤单失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** CandleOrder", resp.Code.Int(), resp.Msg), nil
	}

	// Check for empty response
	if len(resp.CancelOrders) == 0 {
		logger.Errorf(ctx, "okx-batch-cancel-order: empty response", nil)
		return "**批量撤单失败**\n\n**错误类型：** 空响应", nil
	}

	// Categorize results by sCode
	successes := make([]*tradeModels.CancelOrder, 0)
	failures := make([]*tradeModels.CancelOrder, 0)

	for _, result := range resp.CancelOrders {
		if float64(result.SCode) == 0 {
			successes = append(successes, result)
		} else {
			failures = append(failures, result)
		}
	}

	// Format output as Markdown tables
	output := c.formatBatchCancelOrderOutput(successes, failures, len(req.Orders))
	return output, nil
}

// formatBatchCancelOrderOutput formats the batch cancel order result as Markdown tables
func (c *OkxBatchCancelOrderTool) formatBatchCancelOrderOutput(successes, failures []*tradeModels.CancelOrder, total int) string {
	output := ""
	output += fmt.Sprintf("# Batch Order Cancellation (%d orders)\n\n", total)

	// Success section
	if len(successes) > 0 {
		output += "## 成功取消\n\n"
		output += "```markdown\n"
		output += "| OrdId | ClOrdId | SCode | SMsg |\n"
		output += "| :---- | :------ | :---- | :--- |\n"
		for _, s := range successes {
			sCodeStr := fmt.Sprintf("%.0f", float64(s.SCode))
			output += fmt.Sprintf("| %s | %s | %s | %s |\n",
				s.OrdID,
				s.ClOrdID,
				sCodeStr,
				s.SMsg,
			)
		}
		output += "\n```\n\n"
	}

	// Failure section
	if len(failures) > 0 {
		output += "## 失败取消\n\n"
		output += "```markdown\n"
		output += "| 请求索引 | OrdId | SCode | SMsg |\n"
		output += "| :------- | :---- | :---- | :--- |\n"
		for i, f := range failures {
			sCodeStr := fmt.Sprintf("%.0f", float64(f.SCode))
			output += fmt.Sprintf("| %d | %s | %s | %s |\n",
				i+1,
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

// NewOkxBatchCancelOrderTool creates a new OkxBatchCancelOrderTool instance
func NewOkxBatchCancelOrderTool(svcCtx *svc.ServiceContext) *OkxBatchCancelOrderTool {
	return &OkxBatchCancelOrderTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}
