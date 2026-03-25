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
	traderesponses "github.com/PineappleBond/TradingEino/backend/pkg/okex/responses/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxPlaceOrderWithSlTpTool places orders with stop-loss/take-profit in a single call
type OkxPlaceOrderWithSlTpTool struct {
	svcCtx    *svc.ServiceContext
	limiter   *rate.Limiter
	mockTrade mockTrade // for testing only
}

// NewOkxPlaceOrderWithSlTpTool creates a new OkxPlaceOrderWithSlTpTool instance
func NewOkxPlaceOrderWithSlTpTool(svcCtx *svc.ServiceContext) *OkxPlaceOrderWithSlTpTool {
	return &OkxPlaceOrderWithSlTpTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade endpoint
	}
}

func (c *OkxPlaceOrderWithSlTpTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-place-order-with-sl-tp",
		Desc:  "下单同时附加止损/止盈（SL/TP）条件单",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "交易对，如 ETH-USDT-SWAP",
				Required: true,
			},
			"side": {
				Type:     schema.String,
				Desc:     "订单方向：buy 或 sell",
				Enum:     []string{"buy", "sell"},
				Required: true,
			},
			"posSide": {
				Type:     schema.String,
				Desc:     "仓位模式：long/short/net，默认 net",
				Enum:     []string{"long", "short", "net"},
				Required: false,
			},
			"ordType": {
				Type:     schema.String,
				Desc:     "订单类型：conditional (条件单)",
				Enum:     []string{"conditional"},
				Required: true,
			},
			"size": {
				Type:     schema.String,
				Desc:     "主订单数量",
				Required: true,
			},
			"slTriggerPx": {
				Type:     schema.String,
				Desc:     "止损触发价格，和 tpTriggerPx 至少填一个",
				Required: false,
			},
			"slOrderPx": {
				Type:     schema.String,
				Desc:     "止损委托价格，留空或 -1 表示市价单",
				Required: false,
			},
			"tpTriggerPx": {
				Type:     schema.String,
				Desc:     "止盈触发价格，和 slTriggerPx 至少填一个",
				Required: false,
			},
			"tpOrderPx": {
				Type:     schema.String,
				Desc:     "止盈委托价格，留空或 -1 表示市价单",
				Required: false,
			},
		}),
	}, nil
}

func (c *OkxPlaceOrderWithSlTpTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	// 1. Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-place-order-with-sl-tp: rate limiter wait failed", err)
		return fmt.Sprintf("**订单执行失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	// 2. Parse JSON arguments
	var params struct {
		InstID      string `json:"instID"`
		Side        string `json:"side"`
		PosSide     string `json:"posSide"`
		OrdType     string `json:"ordType"`
		Size        string `json:"size"`
		Price       string `json:"price"`
		SlTriggerPx string `json:"slTriggerPx"`
		SlOrderPx   string `json:"slOrderPx"`
		TpTriggerPx string `json:"tpTriggerPx"`
		TpOrderPx   string `json:"tpOrderPx"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		logger.Errorf(ctx, "okx-place-order-with-sl-tp: failed to parse arguments", err)
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// 3. Validate at least one of slTriggerPx or tpTriggerPx is provided
	slTriggerPx := 0.0
	tpTriggerPx := 0.0
	slOrderPx := 0.0
	tpOrderPx := 0.0
	hasSlOrderPx := false
	hasTpOrderPx := false

	if params.SlTriggerPx != "" {
		if _, err := fmt.Sscanf(params.SlTriggerPx, "%f", &slTriggerPx); err != nil {
			return "", fmt.Errorf("invalid slTriggerPx format: %w", err)
		}
	}
	if params.SlOrderPx != "" {
		hasSlOrderPx = true
		// -1 means market order for OKX
		if params.SlOrderPx == "-1" {
			slOrderPx = -1
		} else if _, err := fmt.Sscanf(params.SlOrderPx, "%f", &slOrderPx); err != nil {
			return "", fmt.Errorf("invalid slOrderPx format: %w", err)
		}
	}
	if params.TpTriggerPx != "" {
		if _, err := fmt.Sscanf(params.TpTriggerPx, "%f", &tpTriggerPx); err != nil {
			return "", fmt.Errorf("invalid tpTriggerPx format: %w", err)
		}
	}
	if params.TpOrderPx != "" {
		hasTpOrderPx = true
		// -1 means market order for OKX
		if params.TpOrderPx == "-1" {
			tpOrderPx = -1
		} else if _, err := fmt.Sscanf(params.TpOrderPx, "%f", &tpOrderPx); err != nil {
			return "", fmt.Errorf("invalid tpOrderPx format: %w", err)
		}
	}

	if slTriggerPx <= 0 && tpTriggerPx <= 0 {
		return "", fmt.Errorf("at least one of slTriggerPx or tpTriggerPx must be provided")
	}

	// 4. Parse side
	var side okex.OrderSide
	switch params.Side {
	case "buy":
		side = okex.OrderBuy
	case "sell":
		side = okex.OrderSell
	default:
		return "", fmt.Errorf("invalid side: %s", params.Side)
	}

	// 5. Parse posSide (default to net)
	var posSide okex.PositionSide
	switch params.PosSide {
	case "long":
		posSide = okex.PositionLongSide
	case "short":
		posSide = okex.PositionShortSide
	case "net", "":
		posSide = okex.PositionNetSide
	default:
		return "", fmt.Errorf("invalid posSide: %s", params.PosSide)
	}

	// 6. Build PlaceAlgoOrder request with conditional order type
	sizeVal := 0.0
	if params.Size != "" {
		if _, err := fmt.Sscanf(params.Size, "%f", &sizeVal); err != nil {
			return "", fmt.Errorf("invalid size format: %w", err)
		}
	}

	req := traderequests.PlaceAlgoOrder{
		InstID:    params.InstID,
		TdMode:    okex.TradeCrossMode,
		Side:      side,
		PosSide:   posSide,
		OrdType:   okex.AlgoOrderConditional,
		Sz:        int64(sizeVal),
		StopOrder: traderequests.StopOrder{},
	}

	// Only set trigger prices if provided (non-zero)
	if slTriggerPx > 0 {
		req.SlTriggerPx = &slTriggerPx
	}
	if tpTriggerPx > 0 {
		req.TpTriggerPx = &tpTriggerPx
	}

	// Only set order prices if explicitly provided
	if hasSlOrderPx {
		req.SlOrdPx = &slOrderPx
	}
	if hasTpOrderPx {
		req.TpOrdPx = &tpOrderPx
	}

	// 7. Call PlaceAlgoOrder API
	var result traderesponses.PlaceAlgoOrder
	var err error

	if c.mockTrade != nil {
		result, err = c.mockTrade.PlaceAlgoOrder(req)
	} else {
		result, err = c.svcCtx.OKXClient.Rest.Trade.PlaceAlgoOrder(req)
	}

	if err != nil {
		logger.Errorf(ctx, "okx-place-order-with-sl-tp: API call failed", err, "instID", params.InstID, "side", params.Side, "size", params.Size)
		return fmt.Sprintf("**订单执行失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			err, params.InstID, params.Side, params.Size), nil
	}

	// 8. Validate OKX response code
	if result.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-place-order-with-sl-tp: response code error", nil, "instID", params.InstID, "side", params.Side, "size", params.Size, "code", result.Code.Int(), "msg", result.Msg)
		return fmt.Sprintf("**订单执行失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** PlaceAlgoOrder\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			result.Code.Int(), result.Msg, params.InstID, params.Side, params.Size), nil
	}

	// 9. Validate algo order sCode/sMsg (EXEC-06)
	if len(result.PlaceAlgoOrders) == 0 {
		logger.Errorf(ctx, "okx-place-order-with-sl-tp: empty response", nil, "instID", params.InstID)
		return "", fmt.Errorf("no algo order result returned")
	}

	algoResult := result.PlaceAlgoOrders[0]
	if algoResult.SCode != 0 {
		logger.Errorf(ctx, "okx-place-order-with-sl-tp: order-level sCode error", nil, "instID", params.InstID, "side", params.Side, "size", params.Size, "sCode", algoResult.SCode, "sMsg", algoResult.SMsg)
		return fmt.Sprintf("**订单执行失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**交易对：** %s\n**方向：** %s\n**数量：** %s",
			int64(algoResult.SCode), algoResult.SMsg, params.InstID, params.Side, params.Size), nil
	}

	// 10. Format output as Markdown table
	output := c.formatOutput(algoResult, struct {
		InstID      string
		Side        string
		PosSide     string
		OrdType     string
		Size        string
		Price       string
		SlTriggerPx string
		SlOrderPx   string
		TpTriggerPx string
		TpOrderPx   string
	}{
		InstID:      params.InstID,
		Side:        params.Side,
		PosSide:     params.PosSide,
		OrdType:     params.OrdType,
		Size:        params.Size,
		Price:       params.Price,
		SlTriggerPx: params.SlTriggerPx,
		SlOrderPx:   params.SlOrderPx,
		TpTriggerPx: params.TpTriggerPx,
		TpOrderPx:   params.TpOrderPx,
	})
	return output, nil
}

// formatOutput formats the algo order result as a Markdown table
func (c *OkxPlaceOrderWithSlTpTool) formatOutput(algoResult *trademodels.PlaceAlgoOrder, params struct {
	InstID      string
	Side        string
	PosSide     string
	OrdType     string
	Size        string
	Price       string
	SlTriggerPx string
	SlOrderPx   string
	TpTriggerPx string
	TpOrderPx   string
}) string {
	headers := []string{"algoId", "instID", "side", "sCode", "sMsg"}
	rows := [][]string{
		{
			algoResult.AlgoID,
			params.InstID,
			params.Side,
			fmt.Sprintf("%d", int64(algoResult.SCode)),
			algoResult.SMsg,
		},
	}

	table := xmd.CreateMarkdownTable(headers, rows)

	output := "## Order with SL/TP Placed\n\n"
	output += "```markdown\n"
	output += table
	output += "\n```\n"

	if params.SlTriggerPx != "" || params.TpTriggerPx != "" {
		output += "\n### Stop-Loss/Take-Profit Details\n\n"
		output += "```markdown\n"
		if params.SlTriggerPx != "" {
			slOrderPxDisplay := coalesce(params.SlOrderPx, "-1")
			if slOrderPxDisplay == "-1" {
				slOrderPxDisplay = "Market"
			}
			output += fmt.Sprintf("- Stop-Loss: Trigger=%s, Order=%s\n", params.SlTriggerPx, slOrderPxDisplay)
		}
		if params.TpTriggerPx != "" {
			tpOrderPxDisplay := coalesce(params.TpOrderPx, "-1")
			if tpOrderPxDisplay == "-1" {
				tpOrderPxDisplay = "Market"
			}
			output += fmt.Sprintf("- Take-Profit: Trigger=%s, Order=%s\n", params.TpTriggerPx, tpOrderPxDisplay)
		}
		output += "\n```\n"
	}

	return output
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
