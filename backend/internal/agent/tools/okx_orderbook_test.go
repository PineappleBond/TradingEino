package tools

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/stretchr/testify/assert"
)

func TestOkxOrderbookTool_Info(t *testing.T) {
	ctx := context.Background()

	// Create mock service context
	svcCtx := &svc.ServiceContext{}

	// Create tool
	tool := NewOkxOrderbookTool(svcCtx)

	// Get tool info
	info, err := tool.Info(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "okx-orderbook-tool", info.Name)
	assert.NotEmpty(t, info.Desc)
}

func TestOkxOrderbookTool_RateLimiter(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	tool := NewOkxOrderbookTool(svcCtx)

	// Verify rate limiter is initialized (10 req/s for Market endpoint)
	assert.NotNil(t, tool.limiter)
	assert.Equal(t, 2, tool.limiter.Burst())
}
