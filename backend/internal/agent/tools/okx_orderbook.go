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

type OkxOrderbookTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

func (c *OkxOrderbookTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-orderbook-tool",
		Desc:  "获取订单簿深度数据，分析买卖盘不平衡度",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"symbol": &schema.ParameterInfo{
				Type:     schema.String,
				Desc:     "交易对，比如 ETH-USDT-SWAP, BTC-USDT",
				Enum:     nil,
				Required: true,
			},
			"depth": &schema.ParameterInfo{
				Type:     schema.Number,
				Desc:     "深度档位，可选 1/5/10/20/40/50/100/200/400，默认 5",
				Enum:     []string{"1", "5", "10", "20", "40", "50", "100", "200", "400"},
				Required: false,
			},
		}),
	}, nil
}

func (c *OkxOrderbookTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		Symbol string `json:"symbol"`
		Depth  *int   `json:"depth"`
	}
	var request Request
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", err
	}

	// Default depth is 5
	depth := 5
	if request.Depth != nil {
		depth = *request.Depth
	}

	// Wait for rate limiter before making API call (10 req/s for Market endpoint)
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Call OKX API
	orderBookResp, err := c.svcCtx.OKXClient.Rest.Market.GetOrderBook(marketrequests.GetOrderBook{
		InstID: request.Symbol,
		Sz:     depth,
	})
	if err != nil {
		return "", err
	}

	// Check OKX response code
	if orderBookResp.Code.Int() != 0 {
		return "", &okex.OKXError{
			Code:     orderBookResp.Code.Int(),
			Msg:      orderBookResp.Msg,
			Endpoint: "GetOrderBook",
		}
	}

	if len(orderBookResp.OrderBooks) == 0 {
		return "无订单簿数据", nil
	}

	orderBook := orderBookResp.OrderBooks[0]
	output := ""
	output += fmt.Sprintf("# 订单簿深度数据 (%s)\n\n", request.Symbol)
	output += fmt.Sprintf("**数据时间**: %s\n\n", time.Time(orderBook.TS).Format("2006-01-02 15:04:05"))

	// Calculate bid/ask totals for imbalance analysis
	totalBidSize := 0.0
	totalAskSize := 0.0

	// Output bid table
	output += "## 买单 (Bids)\n\n"
	output += "```markdown\n| 价格 | 数量 | 累计数量 |\n"
	output += "| :--- | :--- | :------- |\n"
	cumulativeBid := 0.0
	for _, bid := range orderBook.Bids {
		cumulativeBid += bid.Size
		totalBidSize += bid.Size
		output += fmt.Sprintf("| %.4f | %.4f | %.4f |\n", bid.DepthPrice, bid.Size, cumulativeBid)
	}
	output += "\n```\n\n"

	// Output ask table
	output += "## 卖单 (Asks)\n\n"
	output += "```markdown\n| 价格 | 数量 | 累计数量 |\n"
	output += "| :--- | :--- | :------- |\n"
	cumulativeAsk := 0.0
	for _, ask := range orderBook.Asks {
		cumulativeAsk += ask.Size
		totalAskSize += ask.Size
		output += fmt.Sprintf("| %.4f | %.4f | %.4f |\n", ask.DepthPrice, ask.Size, cumulativeAsk)
	}
	output += "\n```\n\n"

	// Calculate and display imbalance ratio
	output += "## 买卖盘不平衡度分析\n\n"

	if totalBidSize > 0 && totalAskSize > 0 {
		ratio := totalBidSize / totalAskSize
		output += fmt.Sprintf("- **买单总量**: %.4f\n", totalBidSize)
		output += fmt.Sprintf("- **卖单总量**: %.4f\n", totalAskSize)
		output += fmt.Sprintf("- **买卖比 (Bid/Ask Ratio)**: %.4f\n", ratio)

		if ratio > 2.0 {
			output += "- **信号**: 买盘明显强于卖盘，可能存在上涨压力\n"
		} else if ratio < 0.5 {
			output += "- **信号**: 卖盘明显强于买盘，可能存在下跌压力\n"
		} else {
			output += "- **信号**: 买卖盘相对平衡\n"
		}
	} else if totalBidSize > 0 {
		output += "- **买单总量**: %.4f\n"
		output += "- **卖单总量**: 0\n"
		output += "- **信号**: 只有买单，极端行情\n"
	} else if totalAskSize > 0 {
		output += "- **买单总量**: 0\n"
		output += "- **卖单总量**: %.4f\n"
		output += "- **信号**: 只有卖单，极端行情\n"
	}

	return output, nil
}

func NewOkxOrderbookTool(svcCtx *svc.ServiceContext) *OkxOrderbookTool {
	return &OkxOrderbookTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2), // 10 req/s for Market endpoint (burst=2)
	}
}
