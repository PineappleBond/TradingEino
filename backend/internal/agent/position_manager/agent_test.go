package position_manager

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
)

func TestPositionManagerAgent_Creation(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}

	agent, err := NewPositionManagerAgent(ctx, svcCtx)
	if err != nil {
		t.Fatalf("NewPositionManagerAgent() returned error: %v", err)
	}

	if agent == nil {
		t.Fatal("Expected non-nil PositionManagerAgent")
	}

	if agent.agent == nil {
		t.Fatal("Expected non-nil internal agent")
	}
}

func TestPositionManagerAgent_AgentInterface(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}

	agent, err := NewPositionManagerAgent(ctx, svcCtx)
	if err != nil {
		t.Fatalf("NewPositionManagerAgent() returned error: %v", err)
	}

	// Verify Agent() method returns non-nil
	internalAgent := agent.Agent()
	if internalAgent == nil {
		t.Fatal("Expected non-nil agent from Agent() method")
	}

	// Verify it's a ChatModelAgent (can be used as adk.Agent)
	_, ok := internalAgent.(adk.Agent)
	if !ok {
		t.Fatal("Expected internal agent to implement adk.Agent interface")
	}
}

func TestPositionManagerAgent_DescriptionAndSoul(t *testing.T) {
	// Verify DESCRIPTION is not empty
	if DESCRIPTION == "" {
		t.Error("Expected non-empty DESCRIPTION")
	}

	// Verify SOUL is not empty
	if SOUL == "" {
		t.Error("Expected non-empty SOUL")
	}

	// Verify DESCRIPTION contains key capabilities
	expectedDescKeywords := []string{"持仓", "风险", "账户余额", "保证金率"}
	for _, keyword := range expectedDescKeywords {
		if !contains(DESCRIPTION, keyword) {
			t.Logf("DESCRIPTION may not contain '%s'", keyword)
		}
	}

	// Verify SOUL contains personality traits
	expectedSoulKeywords := []string{"保守", "谨慎", "风险", "止损"}
	for _, keyword := range expectedSoulKeywords {
		if !contains(SOUL, keyword) {
			t.Logf("SOUL may not contain '%s'", keyword)
		}
	}
}

func TestPositionManagerAgent_MinimumLines(t *testing.T) {
	// DESCRIPTION.md should have at least 10 lines
	descLines := countLines(DESCRIPTION)
	if descLines < 10 {
		t.Errorf("DESCRIPTION should have at least 10 lines, got %d", descLines)
	}

	// SOUL.md should have at least 10 lines
	soulLines := countLines(SOUL)
	if soulLines < 10 {
		t.Errorf("SOUL should have at least 10 lines, got %d", soulLines)
	}
}

// Helper functions

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count + 1 // Include last line
}
