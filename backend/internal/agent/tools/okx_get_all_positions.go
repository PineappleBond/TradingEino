package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/internal/utils/xmd"
	accountmodels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/account"
	accountrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxGetAllPositionsTool queries all positions under the current account
type OkxGetAllPositionsTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxGetAllPositionsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-all-positions",
		Desc:  "Query all positions under the current account (no parameters required)",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

// InvokableRun executes the get all positions tool
func (c *OkxGetAllPositionsTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Errorf(ctx, "okx-get-all-positions: rate limiter wait failed", err)
		return fmt.Sprintf("**持仓查询失败**\n\n**错误类型：** 限流等待失败\n**错误信息：** %v", err), nil
	}

	// Get all positions (empty request = all positions)
	resp, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(accountrequests.GetPositions{
		InstID:   nil,
		PosID:    nil,
		InstType: "",
	})
	if err != nil {
		logger.Errorf(ctx, "okx-get-all-positions: API call failed", err)
		return fmt.Sprintf("**持仓查询失败**\n\n**错误类型：** API 调用失败\n**错误信息：** %v", err), nil
	}

	// Check response code
	if resp.Code.Int() != 0 {
		logger.Errorf(ctx, "okx-get-all-positions: response code error", nil, "code", resp.Code.Int(), "msg", resp.Msg)
		return fmt.Sprintf("**持仓查询失败**\n\n**错误代码：** %d\n**错误信息：** %s\n**接口：** GetPositions", resp.Code.Int(), resp.Msg), nil
	}

	// Filter out empty positions
	availablePositions := make([]*accountmodels.Position, 0)
	for _, position := range resp.Positions {
		if position.Pos != 0 {
			availablePositions = append(availablePositions, position)
		}
	}

	// Format output as Markdown table
	output := c.formatOutput(availablePositions)
	return output, nil
}

// formatOutput formats the positions as a Markdown table
func (c *OkxGetAllPositionsTool) formatOutput(positions []*accountmodels.Position) string {
	if len(positions) == 0 {
		return "## 当前无持仓\n\n当前账户下没有任何持仓。"
	}

	output := ""
	output += fmt.Sprintf("## 当前持仓 (%d 个)\n\n", len(positions))

	headers := []string{"交易对", "持仓方向", "持仓数量", "开仓均价", "未实现收益", "未实现收益率", "杠杆", "强平价", "最新价", "盈亏平衡价"}
	rows := [][]string{}

	for _, position := range positions {
		rows = append(rows, []string{
			position.InstID,
			string(position.PosSide),
			fmt.Sprintf("%.4f", position.Pos),
			fmt.Sprintf("%.4f", position.AvgPx),
			fmt.Sprintf("%.4f", position.Upl),
			fmt.Sprintf("%.4f%%", position.UplRatio*100),
			fmt.Sprintf("%.0fx", position.Lever),
			fmt.Sprintf("%.4f", position.LiqPx),
			fmt.Sprintf("%.4f", position.Last),
			fmt.Sprintf("%.4f", position.BePx),
		})
	}

	table := xmd.CreateMarkdownTable(headers, rows)
	output += "```markdown\n"
	output += table
	output += "\n```\n---\n\n"

	return output
}

// NewOkxGetAllPositionsTool creates a new OkxGetAllPositionsTool instance
func NewOkxGetAllPositionsTool(svcCtx *svc.ServiceContext) *OkxGetAllPositionsTool {
	return &OkxGetAllPositionsTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Account endpoint
	}
}
