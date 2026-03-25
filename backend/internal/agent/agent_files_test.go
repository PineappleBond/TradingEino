package agent_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestAgentFiles verifies ANAL-06 compliance:
// All SubAgents must have DESCRIPTION.md and SOUL.md files with minimum content.
// This test fails fast if any documentation is missing or too small.
func TestAgentFiles(t *testing.T) {
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
			path := filepath.Join(testDir, agent, "DESCRIPTION.md")

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
			path := filepath.Join(testDir, agent, "SOUL.md")

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
