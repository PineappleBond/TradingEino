package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	accountRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	accountModels "github.com/PineappleBond/TradingEino/backend/pkg/okex/models/account"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

// OkxCloseAllPositionsTool closes all open positions via OKX REST API
type OkxCloseAllPositionsTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

// Info returns the tool information
func (c *OkxCloseAllPositionsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-close-all-positions",
		Desc:  "Close all open positions for a specific instrument or all instruments. First queries all positions, then closes them one by one.",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "Instrument ID (e.g., ETH-USDT-SWAP), leave empty to close all positions",
				Required: false,
			},
			"instType": {
				Type:     schema.String,
				Desc:     "Instrument type: SPOT/SWAP/FUTURES/OPTION/MARGIN, default SWAP",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun executes the close all positions tool
func (c *OkxCloseAllPositionsTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		InstID   string `json:"instID,omitempty"`
		InstType string `json:"instType,omitempty"`
	}

	var req Request
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 1. Query all positions
	instType := req.InstType
	if instType == "" {
		instType = "SWAP"
	}

	// Wait for rate limiter before making API call
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	posReq := accountRequests.GetPositions{
		InstType: okex.InstrumentType(instType),
	}

	posResp, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(posReq)
	if err != nil {
		return "", err
	}

	// Check response code
	if posResp.Code.Int() != 0 {
		return "", &okex.OKXError{
			Code:     posResp.Code.Int(),
			Msg:      posResp.Msg,
			Endpoint: "GetPositions",
		}
	}

	// Filter positions by instID if specified
	var positions []*accountModels.Position
	for _, pos := range posResp.Positions {
		// Only include positions with positive size
		if pos.Pos <= 0 {
			continue
		}
		// Filter by instID if specified
		if req.InstID != "" && pos.InstID != req.InstID {
			continue
		}
		positions = append(positions, pos)
	}

	// Check if there are any open positions
	if len(positions) == 0 {
		return c.formatNoPositionsOutput(instType, req.InstID), nil
	}

	// 2. Close each position
	totalClosed := 0
	totalFailed := 0
	var failedPositions []*failedPosition

	for _, pos := range positions {
		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return "", fmt.Errorf("rate limiter wait failed: %w", err)
		}

		// Close position using ClosePosition endpoint (100% close)
		closeReq := tradeRequests.ClosePosition{
			InstID:  pos.InstID,
			PosSide: pos.PosSide,
			MgnMode: okex.MarginCrossMode,
		}

		closeResp, err := c.svcCtx.OKXClient.Rest.Trade.ClosePosition(closeReq)
		if err != nil {
			totalFailed++
			failedPositions = append(failedPositions, &failedPosition{
				InstID:  pos.InstID,
				PosSide: string(pos.PosSide),
				ErrMsg:  err.Error(),
			})
			continue
		}

		// Check response code
		if closeResp.Code.Int() != 0 {
			totalFailed++
			failedPositions = append(failedPositions, &failedPosition{
				InstID:  pos.InstID,
				PosSide: string(pos.PosSide),
				ErrMsg:  closeResp.Msg,
			})
			continue
		}

		totalClosed++
	}

	return c.formatCloseAllOutput(totalClosed, totalFailed, failedPositions, len(positions)), nil
}

// failedPosition holds information about a failed position close
type failedPosition struct {
	InstID  string
	PosSide string
	ErrMsg  string
}

// formatNoPositionsOutput returns a message when no positions are found
func (c *OkxCloseAllPositionsTool) formatNoPositionsOutput(instType, instID string) string {
	output := ""
	output += "# Close All Positions\n\n"
	if instID != "" {
		output += fmt.Sprintf("No open positions found for %s (%s).\n", instID, instType)
	} else {
		output += fmt.Sprintf("No open positions found for %s.\n", instType)
	}
	output += "---\n\n"
	return output
}

// formatCloseAllOutput formats the close all result
func (c *OkxCloseAllPositionsTool) formatCloseAllOutput(closed, failed int, failedPositions []*failedPosition, total int) string {
	output := ""
	output += fmt.Sprintf("# Close All Positions Results\n\n")
	output += fmt.Sprintf("Total open positions: %d\n", total)
	output += fmt.Sprintf("Successfully closed: %d\n", closed)
	output += fmt.Sprintf("Failed: %d\n\n", failed)

	if len(failedPositions) > 0 {
		output += "## Failed Positions\n\n"
		output += "```markdown\n"
		output += "| InstId | PosSide | Error |\n"
		output += "| :----- | :------ | :---- |\n"
		for _, f := range failedPositions {
			output += fmt.Sprintf("| %s | %s | %s |\n",
				f.InstID,
				f.PosSide,
				f.ErrMsg,
			)
		}
		output += "\n```\n\n"
	}

	output += "---\n\n"
	return output
}

// NewOkxCloseAllPositionsTool creates a new OkxCloseAllPositionsTool instance
func NewOkxCloseAllPositionsTool(svcCtx *svc.ServiceContext) *OkxCloseAllPositionsTool {
	return &OkxCloseAllPositionsTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s for Trade/Account endpoint
	}
}
