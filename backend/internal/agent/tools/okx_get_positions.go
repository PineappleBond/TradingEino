package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/account"
	accountrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type OkxGetPositionsTool struct {
	svcCtx *svc.ServiceContext
}

func (c *OkxGetPositionsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-positions-tool",
		Desc:  "调用OKX接口获取当前仓位和最大购买力的工具",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"symbol": &schema.ParameterInfo{
				Type:     schema.String,
				Desc:     "交易对,比如ETH-USDT-SWAP,BTC-USDT",
				Enum:     nil,
				Required: true,
			},
			"leverage": &schema.ParameterInfo{
				Type:     schema.Number,
				Desc:     "杠杆倍数,整形,默认1倍",
				Required: false,
			},
		}),
	}, nil
}

func (c *OkxGetPositionsTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		Symbol   string `json:"symbol"`
		Leverage *int   `json:"leverage"`
	}
	var request Request
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", err
	}
	leverage := 1
	if request.Leverage != nil {
		leverage = *request.Leverage
	}
	output := ""
	output += "# 当前仓位\n\n"
	{
		instIDs := make([]string, 0)
		instIDs = append(instIDs, request.Symbol)
		getPositions, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(accountrequests.GetPositions{
			InstID:   instIDs,
			PosID:    nil,
			InstType: "",
		})
		if err != nil {
			return "", err
		}
		if getPositions.Code != 0 {
			return "", fmt.Errorf("OKX API error: %s", getPositions.Msg)
		}
		availablePositions := make([]*account.Position, 0)
		for _, position := range getPositions.Positions {
			if position.Pos == 0 {
				continue
			}
			availablePositions = append(availablePositions, position)
		}
		if len(availablePositions) == 0 {
			output += "无仓位\n---\n\n"
		} else {
			output += "```markdown\n| 保证金模式 | 持仓ID | 持仓方向 | 持仓数量 | 开仓均价 | 未实现收益 | 未实现收益率 | 杠杆倍数 | 预估强平价 | 最新成交价 | 盈亏平衡价 |\n"
			output += "| :------- | :----- | :------ | :------ | :----- | :------- | :--------- | :------ | :------- | :------- | :------- |\n"
			for _, position := range availablePositions {
				output += fmt.Sprintf(
					"| %s | %s | %s | %.4f | %.4f | %.4f | %.4f | %d | %.4f | %.4f | %.4f |\n",
					position.MgnMode,
					position.PosID,
					position.PosSide,
					position.Pos,
					position.AvgPx,
					position.Upl,
					position.UplRatio,
					int(position.Lever),
					position.LiqPx,
					position.Last,
					position.BePx,
				)
			}
			output += "\n```\n---\n\n"
		}
	}
	output += "# 最大购买力(张合约)\n\n"
	{
		getMaxAvailSizeCCY := strings.TrimSuffix(request.Symbol, "-SWAP")
		getMaxAvailSizeCCY = strings.TrimSuffix(getMaxAvailSizeCCY, "-USDT")
		getMaxAvailSizeCCY = strings.TrimSuffix(getMaxAvailSizeCCY, "-USDC")
		getMaxAvailSizeCCY = strings.TrimSuffix(getMaxAvailSizeCCY, "-USD")
		getMaxTradeAmount, err := c.svcCtx.OKXClient.Rest.Account.GetMaxBuySellAmount(accountrequests.GetMaxBuySellAmount{
			Ccy:      getMaxAvailSizeCCY,
			InstID:   []string{request.Symbol},
			Px:       0,
			Leverage: leverage,
			TdMode:   "cross",
		})
		if err != nil {
			return "", err
		}
		if getMaxTradeAmount.Code != 0 {
			return "", fmt.Errorf("OKX API error: %s", getMaxTradeAmount.Msg)
		}
		if len(getMaxTradeAmount.MaxBuySellAmounts) != 1 {
			return "", fmt.Errorf("get max available trade amount amounts: %d", len(getMaxTradeAmount.MaxBuySellAmounts))
		}
		output += "```markdown\n| 最大可买 | 最大可卖 |\n"
		output += "| :----- | :------ |\n"
		for _, maxBuySellAmount := range getMaxTradeAmount.MaxBuySellAmounts {
			output += fmt.Sprintf("| %.3f | %.3f |\n", maxBuySellAmount.MaxBuy, maxBuySellAmount.MaxSell)
		}
		output += "\n```\n---\n\n"
	}
	return output, nil
}

func NewOkxGetPositionsTool(svcCtx *svc.ServiceContext) *OkxGetPositionsTool {
	return &OkxGetPositionsTool{svcCtx: svcCtx}
}
