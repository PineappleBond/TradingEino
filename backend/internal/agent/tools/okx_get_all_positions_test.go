package tools_test

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/stretchr/testify/assert"
)

func TestOkxGetAllPositionsTool_Info(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	tool := tools.NewOkxGetAllPositionsTool(svcCtx)

	ctx := context.Background()
	info, err := tool.Info(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "okx-get-all-positions", info.Name)
	assert.NotEmpty(t, info.Desc)
	// Should have no required parameters (empty params)
}
