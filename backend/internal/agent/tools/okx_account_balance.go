package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/account"
	accountrequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxAccountBalanceTool 获取账户余额和保证金率的工具
type OkxAccountBalanceTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// NewOkxAccountBalanceTool 创建账户余额工具
func NewOkxAccountBalanceTool(svcCtx *svc.ServiceContext) *OkxAccountBalanceTool {
	return &OkxAccountBalanceTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Account endpoint
	}
}

// Info 返回工具信息
func (c *OkxAccountBalanceTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-account-balance-tool",
		Desc:  "获取账户余额和保证金率，评估整体风险敞口",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

// InvokableRun 执行工具调用
func (c *OkxAccountBalanceTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Call OKX API to get balance
	balance, err := c.svcCtx.OKXClient.Rest.Account.GetBalance(accountrequests.GetBalance{})
	if err != nil {
		return "", err
	}

	// Validate OKX response code
	if balance.Code.Int() != 0 {
		return "", &okex.OKXError{
			Code:     balance.Code.Int(),
			Msg:      balance.Msg,
			Endpoint: "GetBalance",
		}
	}

	// Format output
	output := formatBalanceOutput(balance.Balances)
	return output, nil
}

// formatBalanceOutput 格式化余额输出为 markdown 表格
func formatBalanceOutput(balances []*account.Balance) string {
	if len(balances) == 0 {
		return "# 账户余额\n\n无余额\n"
	}

	output := ""
	output += "# 账户余额\n\n"
	output += "```markdown\n"
	output += "| 币种 | 总权益 | 可用 | 冻结 | 负债 |\n"
	output += "| :-- | :-- | :-- | :-- | :-- |\n"

	totalEquity := 0.0
	totalLiability := 0.0

	for _, balance := range balances {
		for _, detail := range balance.Details {
			output += formatBalanceRow(detail)
			totalEquity += float64(detail.Eq)
			totalLiability += float64(detail.Liab)
		}
	}

	output += "\n```\n\n"

	// Calculate and display margin ratio
	if totalLiability > 0 {
		marginRatio := (totalEquity - totalLiability) / totalEquity * 100
		output += formatMarginRatio(marginRatio, totalLiability)
	}

	return output
}

// formatBalanceRow 格式化单行余额数据
func formatBalanceRow(detail *account.BalanceDetails) string {
	return fmt.Sprintf("| %s | %.2f | %.2f | %.2f | %.2f |\n",
		detail.Ccy,
		detail.Eq,
		detail.AvailEq,
		detail.FrozenBal,
		detail.Liab,
	)
}

// formatMarginRatio 格式化保证金率和风险提示
func formatMarginRatio(marginRatio, totalLiability float64) string {
	output := "## 风险分析\n\n"
	output += fmt.Sprintf("保证金率：%.2f%%\n", marginRatio)
	output += fmt.Sprintf("总负债：%.2f\n\n", totalLiability)

	if marginRatio < 20 {
		output += "**警告**: 保证金率低于 20%，存在较高风险！\n"
	} else if marginRatio < 50 {
		output += "**注意**: 保证金率低于 50%，请密切关注仓位风险。\n"
	}

	return output
}

// FormatBalanceRow 格式化单行余额数据（导出用于测试）
func FormatBalanceRow(detail *account.BalanceDetails) string {
	return fmt.Sprintf("| %s | %.2f | %.2f | %.2f | %.2f |\n",
		detail.Ccy,
		detail.Eq,
		detail.AvailEq,
		detail.FrozenBal,
		detail.Liab,
	)
}

// FormatMarginRatio 格式化保证金率和风险提示（导出用于测试）
func FormatMarginRatio(marginRatio, totalLiability float64) string {
	output := "## 风险分析\n\n"
	output += fmt.Sprintf("保证金率：%.2f%%\n", marginRatio)
	output += fmt.Sprintf("总负债：%.2f\n\n", totalLiability)

	if marginRatio < 20 {
		output += "**警告**: 保证金率低于 20%，存在较高风险！\n"
	} else if marginRatio < 50 {
		output += "**注意**: 保证金率低于 50%，请密切关注仓位风险。\n"
	}

	return output
}

// FormatBalanceOutput 格式化余额输出为 markdown 表格（导出用于测试）
func FormatBalanceOutput(balances []*account.Balance) string {
	if len(balances) == 0 {
		return "# 账户余额\n\n无余额\n"
	}

	output := ""
	output += "# 账户余额\n\n"
	output += "```markdown\n"
	output += "| 币种 | 总权益 | 可用 | 冻结 | 负债 |\n"
	output += "| :-- | :-- | :-- | :-- | :-- |\n"

	totalEquity := 0.0
	totalLiability := 0.0

	for _, balance := range balances {
		for _, detail := range balance.Details {
			output += FormatBalanceRow(detail)
			totalEquity += float64(detail.Eq)
			totalLiability += float64(detail.Liab)
		}
	}

	output += "\n```\n\n"

	// Calculate and display margin ratio
	if totalLiability > 0 {
		marginRatio := (totalEquity - totalLiability) / totalEquity * 100
		output += FormatMarginRatio(marginRatio, totalLiability)
	}

	return output
}
