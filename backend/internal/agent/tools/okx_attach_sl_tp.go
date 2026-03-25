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
				Desc:     "已有订单 ID（可选，用于参考）",
				Required: false,
			},
			"side": {
				Type:     schema.String,
				Desc:     "订单方向：buy 或 sell",
				Enum:     []string{"buy", "sell"},
				Required: true,
			},
			"posSide": {
				Type:     schema.String,
				Desc:     "持仓方向：long, short, net",
				Enum:     []string{"long", "short", "net"},
				Required: true,
			},
			"slTriggerPx": {
				Type:     schema.String,
				Desc:     "止损触发价格，和 tpTriggerPx 至少填一个",
				Required: false,
			},
			"slOrderPx": {
				Type:     schema.String,
				Desc:     "止损委托价格，-1 表示市价单",
				Required: false,
			},
			"tpTriggerPx": {
				Type:     schema.String,
				Desc:     "止盈触发价格，和 slTriggerPx 至少填一个",
				Required: false,
			},
			"tpOrderPx": {
				Type:     schema.String,
				Desc:     "止盈委托价格，-1 表示市价单",
				Required: false,
			},
			"sz": {
				Type:     schema.String,
				Desc:     "委托数量（需要满足最小订单金额要求）",
				Required: true,
			},
		}),
	}, nil
}

func (c *OkxAttachSlTpTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	// 1. Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-attach-sl-tp: rate limiter wait failed", err)
		return fmt.Sprintf("**附加 SL/TP 失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	// 2. Parse JSON arguments
	var params struct {
		InstID      string `json:"instID"`
		OrdID       string `json:"ordId"`
		Side        string `json:"side"`
		PosSide     string `json:"posSide"`
		SlTriggerPx string `json:"slTriggerPx"`
		SlOrderPx   string `json:"slOrderPx"`
		TpTriggerPx string `json:"tpTriggerPx"`
		TpOrderPx   string `json:"tpOrderPx"`
		Sz          string `json:"sz"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		logger.Errorf(ctx, "okx-attach-sl-tp: failed to parse arguments", err)
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// 3. Validate at least one of slTriggerPx or tpTriggerPx is provided
	slTriggerPx := 0.0
	tpTriggerPx := 0.0
	slOrderPx := 0.0
	tpOrderPx := 0.0
	sz := 0.0

	if params.SlTriggerPx != "" {
		if _, err := fmt.Sscanf(params.SlTriggerPx, "%f", &slTriggerPx); err != nil {
			return "", fmt.Errorf("invalid slTriggerPx format: %w", err)
		}
	}
	if params.SlOrderPx != "" {
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
		// -1 means market order for OKX
		if params.TpOrderPx == "-1" {
			tpOrderPx = -1
		} else if _, err := fmt.Sscanf(params.TpOrderPx, "%f", &tpOrderPx); err != nil {
			return "", fmt.Errorf("invalid tpOrderPx format: %w", err)
		}
	}
	if params.Sz != "" {
		if _, err := fmt.Sscanf(params.Sz, "%f", &sz); err != nil {
			return "", fmt.Errorf("invalid sz format: %w", err)
		}
	}

	if slTriggerPx <= 0 && tpTriggerPx <= 0 {
		return "", fmt.Errorf("at least one of slTriggerPx or tpTriggerPx must be provided")
	}

	// Parse side
	var side okex.OrderSide
	switch params.Side {
	case "buy":
		side = okex.OrderBuy
	case "sell":
		side = okex.OrderSell
	default:
		return "", fmt.Errorf("invalid side: %s", params.Side)
	}

	// Helper function to create float64 pointer
	floatPtr := func(v float64) *float64 { return &v }

	// 4. Place SL and TP as separate orders (OKX doesn't support combined SL+TP in one order)
	var results []traderesponses.PlaceAlgoOrder

	// Place SL order if provided
	if slTriggerPx > 0 {
		slReq := traderequests.PlaceAlgoOrder{
			InstID:     params.InstID,
			TdMode:     okex.TradeCrossMode,
			Side:       side,
			PosSide:    okex.PositionSide(params.PosSide),
			OrdType:    okex.AlgoOrderConditional,
			Sz:         int64(sz),
			ReduceOnly: true, // Must be true for closing position
			StopOrder: traderequests.StopOrder{
				SlTriggerPx: floatPtr(slTriggerPx),
				SlOrdPx:     floatPtr(slOrderPx),
				TpTriggerPx: nil, // No TP in SL order
				TpOrdPx:     nil,
			},
		}

		var slResult traderesponses.PlaceAlgoOrder
		var err error

		if c.mockTrade != nil {
			slResult, err = c.mockTrade.PlaceAlgoOrder(slReq)
		} else {
			slResult, err = c.svcCtx.OKXClient.Rest.Trade.PlaceAlgoOrder(slReq)
		}

		if err != nil {
			logger.Errorf(ctx, "okx-attach-sl-tp: failed to place SL order", err, "instID", params.InstID, "side", params.Side, "posSide", params.PosSide)
			return fmt.Sprintf("**附加止损失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v", err), nil
		}

		if slResult.Code.Int() != 0 {
			logger.Errorf(ctx, "okx-attach-sl-tp: SL order response code error", nil, "instID", params.InstID, "code", slResult.Code.Int(), "msg", slResult.Msg)
			return fmt.Sprintf("**附加止损失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** PlaceAlgoOrder", slResult.Code.Int(), slResult.Msg), nil
		}

		if len(slResult.PlaceAlgoOrders) == 0 {
			logger.Errorf(ctx, "okx-attach-sl-tp: SL order empty response", nil, "instID", params.InstID)
			return "**附加止损失败**\n\n**错误类型：** 空响应", nil
		}

		slAlgoResult := slResult.PlaceAlgoOrders[0]
		if slAlgoResult.SCode != 0 {
			logger.Errorf(ctx, "okx-attach-sl-tp: SL order sCode error", nil, "instID", params.InstID, "sCode", slAlgoResult.SCode, "sMsg", slAlgoResult.SMsg)
			return fmt.Sprintf("**附加止损失败**\n\n**错误代码：** %d\n**错误信息：** %s", int64(slAlgoResult.SCode), slAlgoResult.SMsg), nil
		}

		results = append(results, slResult)
	}

	// Place TP order if provided
	if tpTriggerPx > 0 {
		tpReq := traderequests.PlaceAlgoOrder{
			InstID:     params.InstID,
			TdMode:     okex.TradeCrossMode,
			Side:       side,
			PosSide:    okex.PositionSide(params.PosSide),
			OrdType:    okex.AlgoOrderConditional,
			Sz:         int64(sz),
			ReduceOnly: true, // Must be true for closing position
			StopOrder: traderequests.StopOrder{
				SlTriggerPx: nil, // No SL in TP order
				SlOrdPx:     nil,
				TpTriggerPx: floatPtr(tpTriggerPx),
				TpOrdPx:     floatPtr(tpOrderPx),
			},
		}

		var tpResult traderesponses.PlaceAlgoOrder
		var err error

		if c.mockTrade != nil {
			tpResult, err = c.mockTrade.PlaceAlgoOrder(tpReq)
		} else {
			tpResult, err = c.svcCtx.OKXClient.Rest.Trade.PlaceAlgoOrder(tpReq)
		}

		if err != nil {
			logger.Errorf(ctx, "okx-attach-sl-tp: failed to place TP order", err, "instID", params.InstID, "side", params.Side, "posSide", params.PosSide)
			return fmt.Sprintf("**附加止盈失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v", err), nil
		}

		if tpResult.Code.Int() != 0 {
			logger.Errorf(ctx, "okx-attach-sl-tp: TP order response code error", nil, "instID", params.InstID, "code", tpResult.Code.Int(), "msg", tpResult.Msg)
			return fmt.Sprintf("**附加止盈失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** PlaceAlgoOrder", tpResult.Code.Int(), tpResult.Msg), nil
		}

		if len(tpResult.PlaceAlgoOrders) == 0 {
			logger.Errorf(ctx, "okx-attach-sl-tp: TP order empty response", nil, "instID", params.InstID)
			return "**附加止盈失败**\n\n**错误类型：** 空响应", nil
		}

		tpAlgoResult := tpResult.PlaceAlgoOrders[0]
		if tpAlgoResult.SCode != 0 {
			logger.Errorf(ctx, "okx-attach-sl-tp: TP order sCode error", nil, "instID", params.InstID, "sCode", tpAlgoResult.SCode, "sMsg", tpAlgoResult.SMsg)
			return fmt.Sprintf("**附加止盈失败**\n\n**错误代码：** %d\n**错误信息：** %s", int64(tpAlgoResult.SCode), tpAlgoResult.SMsg), nil
		}

		results = append(results, tpResult)
	}

	// 5. Format output as Markdown table with all results
	output := c.formatOutputMultiple(results)
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

// formatOutputMultiple formats multiple algo order results as a Markdown table
func (c *OkxAttachSlTpTool) formatOutputMultiple(results []traderesponses.PlaceAlgoOrder) string {
	headers := []string{"Type", "algoId", "sCode", "sMsg"}
	rows := [][]string{}

	for i, result := range results {
		if len(result.PlaceAlgoOrders) > 0 {
			algoResult := result.PlaceAlgoOrders[0]
			orderType := "SL"
			if i == 1 {
				orderType = "TP"
			}
			rows = append(rows, []string{
				orderType,
				algoResult.AlgoID,
				fmt.Sprintf("%d", int64(algoResult.SCode)),
				algoResult.SMsg,
			})
		}
	}

	table := xmd.CreateMarkdownTable(headers, rows)

	output := "## SL/TP Orders Attached\n\n"
	output += "```markdown\n"
	output += table
	output += "\n```\n"

	return output
}
