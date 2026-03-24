package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/internal/utils/xmd"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	trademodels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/trade"
	traderequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	traderesponses "github.com/PineappleBond/TradingEino/backend/pkg/okex/responses/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxAttachSlTpTool attaches stop-loss/take-profit to existing orders
type OkxAttachSlTpTool struct {
	svcCtx    *svc.ServiceContext
	limiter   *rate.Limiter
	mockTrade mockTrade // for testing only
}

// mockTrade interface for testing
type mockTrade interface {
	PlaceAlgoOrder(req traderequests.PlaceAlgoOrder) (traderesponses.PlaceAlgoOrder, error)
}

// NewOkxAttachSlTpTool creates a new OkxAttachSlTpTool instance
func NewOkxAttachSlTpTool(svcCtx *svc.ServiceContext) *OkxAttachSlTpTool {
	return &OkxAttachSlTpTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}

func (c *OkxAttachSlTpTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-attach-sl-tp",
		Desc:  "为已有订单附加止损/止盈（SL/TP）条件单",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "交易对，如 ETH-USDT-SWAP",
				Required: true,
			},
			"ordId": {
				Type:     schema.String,
				Desc:     "已有订单 ID",
				Required: true,
			},
			"slTriggerPx": {
				Type:     schema.String,
				Desc:     "止损触发价格，和 tpTriggerPx 至少填一个",
				Required: false,
			},
			"slOrderPx": {
				Type:     schema.String,
				Desc:     "止损委托价格，留空表示市价单",
				Required: false,
			},
			"tpTriggerPx": {
				Type:     schema.String,
				Desc:     "止盈触发价格，和 slTriggerPx 至少填一个",
				Required: false,
			},
			"tpOrderPx": {
				Type:     schema.String,
				Desc:     "止盈委托价格，留空表示市价单",
				Required: false,
			},
		}),
	}, nil
}

func (c *OkxAttachSlTpTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	// 1. Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// 2. Parse JSON arguments
	var params struct {
		InstID      string `json:"instID"`
		OrdID       string `json:"ordId"`
		SlTriggerPx string `json:"slTriggerPx"`
		SlOrderPx   string `json:"slOrderPx"`
		TpTriggerPx string `json:"tpTriggerPx"`
		TpOrderPx   string `json:"tpOrderPx"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// 3. Validate at least one of slTriggerPx or tpTriggerPx is provided
	slTriggerPx := 0.0
	tpTriggerPx := 0.0
	slOrderPx := 0.0
	tpOrderPx := 0.0

	if params.SlTriggerPx != "" {
		if _, err := fmt.Sscanf(params.SlTriggerPx, "%f", &slTriggerPx); err != nil {
			return "", fmt.Errorf("invalid slTriggerPx format: %w", err)
		}
	}
	if params.SlOrderPx != "" {
		if _, err := fmt.Sscanf(params.SlOrderPx, "%f", &slOrderPx); err != nil {
			return "", fmt.Errorf("invalid slOrderPx format: %w", err)
		}
	}
	if params.TpTriggerPx != "" {
		if _, err := fmt.Sscanf(params.TpTriggerPx, "%f", &tpTriggerPx); err != nil {
			return "", fmt.Errorf("invalid tpTriggerPx format: %w", err)
		}
	}
	if params.TpOrderPx != "" {
		if _, err := fmt.Sscanf(params.TpOrderPx, "%f", &tpOrderPx); err != nil {
			return "", fmt.Errorf("invalid tpOrderPx format: %w", err)
		}
	}

	if slTriggerPx <= 0 && tpTriggerPx <= 0 {
		return "", fmt.Errorf("at least one of slTriggerPx or tpTriggerPx must be provided")
	}

	// 4. Build PlaceAlgoOrder request with sl_tp order type
	req := traderequests.PlaceAlgoOrder{
		InstID:  params.InstID,
		TdMode:  okex.TradeCrossMode, // default to cross mode
		Side:    okex.OrderBuy,       // will be determined by the existing order
		OrdType: okex.AlgoOrderConditional,
		Sz:      0, // Not needed for attach SL/TP
		StopOrder: traderequests.StopOrder{
			SlTriggerPx: slTriggerPx,
			SlOrdPx:     slOrderPx,
			TpTriggerPx: tpTriggerPx,
			TpOrdPx:     tpOrderPx,
		},
	}

	// 5. Call PlaceAlgoOrder API
	var result traderesponses.PlaceAlgoOrder
	var err error

	if c.mockTrade != nil {
		// Testing mode
		result, err = c.mockTrade.PlaceAlgoOrder(req)
	} else {
		// Production mode
		result, err = c.svcCtx.OKXClient.Rest.Trade.PlaceAlgoOrder(req)
	}

	if err != nil {
		return "", fmt.Errorf("failed to place algo order: %w", err)
	}

	// 6. Validate OKX response code
	if result.Code != 0 {
		return "", &okex.OKXError{
			Code:     result.Code,
			Msg:      result.Msg,
			Endpoint: "PlaceAlgoOrder",
		}
	}

	// 7. Validate algo order sCode/sMsg (EXEC-06)
	if len(result.PlaceAlgoOrders) == 0 {
		return "", fmt.Errorf("no algo order result returned")
	}

	algoResult := result.PlaceAlgoOrders[0]
	if algoResult.SCode != 0 {
		return "", fmt.Errorf("algo order failed: sCode=%d, sMsg=%s", int64(algoResult.SCode), algoResult.SMsg)
	}

	// 8. Format output as Markdown table
	output := c.formatOutput(algoResult)
	return output, nil
}

// formatOutput formats the algo order result as a Markdown table
func (c *OkxAttachSlTpTool) formatOutput(algoResult *trademodels.PlaceAlgoOrder) string {
	headers := []string{"algoId", "sCode", "sMsg"}
	rows := [][]string{
		{
			algoResult.AlgoID,
			fmt.Sprintf("%d", int64(algoResult.SCode)),
			algoResult.SMsg,
		},
	}

	table := xmd.CreateMarkdownTable(headers, rows)

	output := "## SL/TP Order Attached\n\n"
	output += "```markdown\n"
	output += table
	output += "\n```\n"

	return output
}
