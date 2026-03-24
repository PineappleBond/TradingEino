package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	publicrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

type OkxGetFundingRateTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

func (c *OkxGetFundingRateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-funding-rate-tool",
		Desc:  "调用OKX接口获取永续合约的资金费率",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"symbol": &schema.ParameterInfo{
				Type:     schema.String,
				Desc:     "交易对,必选是-SWAP结尾.比如ETH-USDT-SWAP,BTC-USDT-SWAP",
				Enum:     nil,
				Required: true,
			},
		}),
	}, nil
}

func (c *OkxGetFundingRateTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		Symbol string `json:"symbol"`
	}
	var request Request
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", err
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	output := ""
	output += "# 资金费率\n\n"
	{
		getFundingRate, err := c.svcCtx.OKXClient.Rest.PublicData.GetFundingRate(publicrequests.GetFundingRate{
			InstID: request.Symbol,
		})
		if err != nil {
			return "", err
		}
		if getFundingRate.Code != 0 {
			return "", &okex.OKXError{
				Code:     getFundingRate.Code,
				Msg:      getFundingRate.Msg,
				Endpoint: "GetFundingRate",
			}
		}

		if len(getFundingRate.FundingRates) == 0 {
			output += "无仓位\n---\n\n"
		} else {
			output += "```markdown\n"
			output += "| 交易对 | 资金费率 | 资金费时间 | 下一期资金费时间 | 溢价指数 |\n"
			output += "| :---- | :----- | :------- | :------------- | :----- |\n"
			for _, fr := range getFundingRate.FundingRates {
				output += fmt.Sprintf(
					"| %s | %.10f | %s | %s | %s |\n",
					fr.InstID,
					fr.FundingRate,
					time.Time(fr.FundingTime).Format(time.RFC3339),
					time.Time(fr.NextFundingTime).Format(time.RFC3339),
					fr.Premium,
				)
			}
			output += "\n```\n---\n\n"
		}
	}
	return output, nil
}

func NewOkxGetFundingRateTool(svcCtx *svc.ServiceContext) *OkxGetFundingRateTool {
	return &OkxGetFundingRateTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2), // 10 req/s for Public endpoint
	}
}
