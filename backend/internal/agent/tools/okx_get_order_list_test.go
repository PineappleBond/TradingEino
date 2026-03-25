package tools_test

import (
	"context"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/stretchr/testify/assert"
)

func TestOkxGetOrderListTool_Info(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	tool := tools.NewOkxGetOrderListTool(svcCtx)

	ctx := context.Background()
	info, err := tool.Info(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "okx-get-order-list", info.Name)
	assert.NotEmpty(t, info.Desc)
}
