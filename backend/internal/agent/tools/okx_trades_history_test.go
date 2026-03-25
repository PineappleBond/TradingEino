package tools

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/stretchr/testify/assert"
)

func TestOkxTradesHistoryTool_Info(t *testing.T) {
	ctx := context.Background()

	// Create mock service context
	svcCtx := &svc.ServiceContext{}

	// Create tool
	tool := NewOkxTradesHistoryTool(svcCtx)

	// Get tool info
	info, err := tool.Info(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "okx-trades-history-tool", info.Name)
	assert.NotEmpty(t, info.Desc)
}

func TestOkxTradesHistoryTool_RateLimiter(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	tool := NewOkxTradesHistoryTool(svcCtx)

	// Verify rate limiter is initialized (10 req/s for Market endpoint)
	assert.NotNil(t, tool.limiter)
	assert.Equal(t, 2, tool.limiter.Burst())
}
