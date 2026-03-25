package flow_analyzer

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/stretchr/testify/assert"
)

func TestFlowAnalyzerAgent_Creation(t *testing.T) {
	ctx := context.Background()

	// Create mock service context with required fields
	svcCtx := &svc.ServiceContext{}

	// Create agent
	agent, err := NewFlowAnalyzerAgent(ctx, svcCtx)

	// Verify agent is created successfully
	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.NotNil(t, agent.agent)
}

func TestFlowAnalyzerAgent_AgentAccessor(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}

	agent, err := NewFlowAnalyzerAgent(ctx, svcCtx)
	assert.NoError(t, err)

	// Verify agent accessor returns non-nil
	assert.NotNil(t, agent.Agent())
}

func TestFlowAnalyzerAgent_DescriptionAndSoul(t *testing.T) {
	// Verify DESCRIPTION and SOUL are embedded
	assert.NotEmpty(t, DESCRIPTION)
	assert.NotEmpty(t, SOUL)
}
