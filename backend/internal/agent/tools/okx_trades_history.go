package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	marketrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

type OkxTradesHistoryTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

func (c *OkxTradesHistoryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-trades-history-tool",
		Desc:  "获取历史成交明细，识别主动买入/卖出和大单动向",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"symbol": &schema.ParameterInfo{
				Type:     schema.String,
				Desc:     "交易对，比如 ETH-USDT-SWAP, BTC-USDT",
				Enum:     nil,
				Required: true,
			},
			"limit": &schema.ParameterInfo{
				Type:     schema.Number,
				Desc:     "获取最近的成交条数，默认 20",
				Required: false,
			},
		}),
	}, nil
}

func (c *OkxTradesHistoryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		Symbol string `json:"symbol"`
		Limit  *int   `json:"limit"`
	}
	var request Request
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", err
	}

	// Default limit is 20
	limit := 20
	if request.Limit != nil {
		limit = *request.Limit
	}

	// Wait for rate limiter before making API call (10 req/s for Market endpoint)
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Call OKX API
	tradesResp, err := c.svcCtx.OKXClient.Rest.Market.GetTrades(marketrequests.GetTrades{
		InstID: request.Symbol,
		Limit:  int64(limit),
	})
	if err != nil {
		return "", err
	}

	// Check OKX response code
	if tradesResp.Code.Int() != 0 {
		return "", &okex.OKXError{
			Code:     tradesResp.Code.Int(),
			Msg:      tradesResp.Msg,
			Endpoint: "GetTrades",
		}
	}

	if len(tradesResp.Trades) == 0 {
		return "无成交明细数据", nil
	}

	output := ""
	output += fmt.Sprintf("# 历史成交明细 (%s)\n\n", request.Symbol)
	output += fmt.Sprintf("**数据时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Output trade table
	output += "```markdown\n| 时间 | 价格 | 数量 | 方向 | 成交额 (USD) |\n"
	output += "| :--- | :--- | :----- | :--- | :---------- |\n"

	totalBuyVolume := 0.0
	totalSellVolume := 0.0
	largeTradeCount := 0

	for _, trade := range tradesResp.Trades {
		tradeTime := time.Time(trade.TS).Format("15:04:05")
		price := float64(trade.Px)
		size := float64(trade.Sz)
		side := string(trade.Side)
		tradeValue := price * size

		// Track buy/sell volumes
		if trade.Side == "buy" {
			totalBuyVolume += tradeValue
			side = "主动买入 🟢"
		} else if trade.Side == "sell" {
			totalSellVolume += tradeValue
			side = "主动卖出 🔴"
		}

		// Highlight large trades (>100,000 USD)
		largeTradeMarker := ""
		if tradeValue > 100000 {
			largeTradeMarker = " ⚠️**大单**"
			largeTradeCount++
		}

		output += fmt.Sprintf("| %s | %.4f | %.4f | %s | %.2f%s |\n",
			tradeTime, price, size, side, tradeValue, largeTradeMarker)
	}
	output += "\n```\n\n"

	// Analysis summary
	output += "## 成交分析\n\n"
	output += fmt.Sprintf("- **总成交笔数**: %d\n", len(tradesResp.Trades))
	output += fmt.Sprintf("- **主动买入总额**: $%.2f\n", totalBuyVolume)
	output += fmt.Sprintf("- **主动卖出总额**: $%.2f\n", totalSellVolume)
	output += fmt.Sprintf("- **大单数量 (>10 万 USD)**: %d\n", largeTradeCount)

	// Buy/Sell ratio
	if totalBuyVolume > 0 && totalSellVolume > 0 {
		ratio := totalBuyVolume / totalSellVolume
		output += fmt.Sprintf("- **买卖比 (Buy/Sell Ratio)**: %.4f\n", ratio)

		if ratio > 1.5 {
			output += "- **信号**: 主动买入占优，市场情绪偏多\n"
		} else if ratio < 0.67 {
			output += "- **信号**: 主动卖出占优，市场情绪偏空\n"
		} else {
			output += "- **信号**: 买卖力量相对均衡\n"
		}
	}

	return output, nil
}

func NewOkxTradesHistoryTool(svcCtx *svc.ServiceContext) *OkxTradesHistoryTool {
	return &OkxTradesHistoryTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2), // 10 req/s for Market endpoint (burst=2)
	}
}
