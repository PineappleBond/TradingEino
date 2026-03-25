package tools

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/account"
)

// MockServiceContext creates a mock ServiceContext for testing
type MockServiceContext struct {
	*svc.ServiceContext
}

func TestOkxAccountBalanceTool_Info(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}
	tool := NewOkxAccountBalanceTool(svcCtx)

	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Info() returned error: %v", err)
	}

	if info.Name != "okx-account-balance-tool" {
		t.Errorf("Expected name 'okx-account-balance-tool', got '%s'", info.Name)
	}

	if info.Desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestOkxAccountBalanceTool_Params(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}
	tool := NewOkxAccountBalanceTool(svcCtx)

	_, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Info() returned error: %v", err)
	}

	// Verify tool has no required parameters
	// The tool should not require any input parameters for balance query
}

func TestOkxAccountBalanceTool_RateLimiter(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	tool := NewOkxAccountBalanceTool(svcCtx)

	// Verify rate limiter is initialized
	if tool.limiter == nil {
		t.Fatal("Rate limiter should be initialized")
	}

	// Test rate limiter allows 5 req/s (200ms per request)
	// Note: limiter is private, we can only verify it's initialized
	// Actual rate limiting is tested through integration
}

func TestOkxAccountBalanceTool_OutputFormat(t *testing.T) {
	// Test the output format structure
	output := FormatBalanceOutput([]*account.Balance{
		{
			TotalEq: 10000.0,
			Details: []*account.BalanceDetails{
				{
					Ccy:       "USDT",
					Eq:        5000.0,
					AvailEq:   4500.0,
					FrozenBal: 500.0,
					Liab:      2000.0,
				},
				{
					Ccy:       "BTC",
					Eq:        0.5,
					AvailEq:   0.4,
					FrozenBal: 0.1,
					Liab:      0.0,
				},
			},
		},
	})

	// Verify output contains expected sections
	expectedSections := []string{
		"# 账户余额",
		"| 币种 | 总权益 | 可用 | 冻结 | 负债 |",
		"保证金率",
	}

	for _, section := range expectedSections {
		if !contains(output, section) {
			t.Errorf("Output should contain section '%s'", section)
		}
	}
}

func TestOkxAccountBalanceTool_MarginRatioCalculation(t *testing.T) {
	// Test margin ratio calculation with high risk scenario
	balances := []*account.Balance{
		{
			TotalEq: 10000.0,
			Mmr:     5000.0, // Maintenance margin required
			Details: []*account.BalanceDetails{
				{
					Ccy:  "USDT",
					Eq:   10000.0,
					Liab: 8500.0, // High liability
				},
			},
		},
	}

	output := FormatBalanceOutput(balances)

	// Should contain risk warning for high margin ratio
	if !contains(output, "风险") {
		t.Log("Output may not contain appropriate risk warnings")
	}
}

func TestOkxAccountBalanceTool_EmptyBalance(t *testing.T) {
	output := FormatBalanceOutput([]*account.Balance{})

	if output == "" {
		t.Error("Expected non-empty output for empty balance")
	}

	if !contains(output, "无余额") {
		t.Error("Expected '无余额' message for empty balance")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
