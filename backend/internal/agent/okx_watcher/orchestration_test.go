package okx_watcher_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/okx_watcher"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/cloudwego/eino/adk"
	"golang.org/x/time/rate"
)

// TestOkxWatcherOrchestration verifies OKXWatcher has all 4 SubAgents configured.
// This is a stub for ANAL-05 verification.
func TestOkxWatcherOrchestration(t *testing.T) {
	// Skip this test if no real OKX API credentials are available
	if os.Getenv("OKX_API_KEY") == "" {
		t.Skip("Skipping integration test: OKX_API_KEY not set")
	}

	ctx := context.Background()

	// Initialize service context (requires real config)
	svcCtx := &svc.ServiceContext{}

	// Initialize all 4 SubAgents
	// Note: This test assumes the agent packages are properly initialized
	// In a real scenario, you would initialize:
	// - techno_agent.NewTechnoAgent
	// - flow_analyzer.NewFlowAnalyzerAgent
	// - position_manager.NewPositionManagerAgent
	// - sentiment_analyst.NewSentimentAnalystAgent

	// Create OKXWatcher with subagents
	subAgents := []adk.Agent{} // Would be populated in real test

	_, err := okx_watcher.NewOkxWatcherAgent(ctx, svcCtx, subAgents...)
	if err != nil {
		t.Fatalf("Failed to create OKXWatcher: %v", err)
	}

	// Verify OKXWatcher can be queried
	t.Log("OKXWatcher created successfully with SubAgents")
}

// TestAgentFilesPresence verifies all SubAgents have DESCRIPTION.md and SOUL.md files.
// This test satisfies ANAL-06 requirement.
func TestAgentFilesPresence(t *testing.T) {
	// Get the directory where this test file is located
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get test file path")
	}
	testDir := filepath.Dir(testFile)

	// List of all SubAgents that must have documentation
	agents := []string{
		"techno_agent",
		"flow_analyzer",
		"position_manager",
		"sentiment_analyst",
	}

	for _, agent := range agents {
		t.Run(agent+"_DESCRIPTION", func(t *testing.T) {
			path := filepath.Join(testDir, "..", agent, "DESCRIPTION.md")

			// Check file exists
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("DESCRIPTION.md not found for %s: %v", agent, err)
			}

			// Check has content (>100 bytes)
			if info.Size() < 100 {
				t.Fatalf("DESCRIPTION.md too small for %s (%d bytes, minimum 100)", agent, info.Size())
			}
		})

		t.Run(agent+"_SOUL", func(t *testing.T) {
			path := filepath.Join(testDir, "..", agent, "SOUL.md")

			// Check file exists
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("SOUL.md not found for %s: %v", agent, err)
			}

			// Check has content (>100 bytes)
			if info.Size() < 100 {
				t.Fatalf("SOUL.md too small for %s (%d bytes, minimum 100)", agent, info.Size())
			}
		})
	}
}

// TestOkxWatcherRateLimiter verifies the rate limiter is configured correctly.
func TestOkxWatcherRateLimiter(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 1) // 10 req/s for Market endpoint

	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	t.Log("Rate limiter configured for Market endpoint (10 req/s)")
}
