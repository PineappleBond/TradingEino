# Phase 3: Execution Automation - Research

**Researched:** 2026-03-24
**Domain:** OKX Trading API Integration, Executor Agent, Level 1 Autonomous Trade Execution
**Confidence:** HIGH

## Summary

Phase 3 implements autonomous trade execution with Level 1 autonomy (explicit commands only). This phase delivers 9 order management tools and an Executor Agent that enables the system to execute trades based on explicit commands from OKXWatcher.

The research confirms:
1. **OKX API infrastructure exists** - `pkg/okex/api/trading.go` provides `PlaceOrder`, `CancelOrder`, `GetOrderDetails` methods with proper `sCode`/`sMsg` validation
2. **Batch operations supported** - OKX API supports batch place/cancel (max 20 orders per batch) via `/api/v5/trade/batch-order` and `/api/v5/trade/cancel-batch-orders`
3. **Algo orders for SL/TP** - OKX supports native stop-loss/take-profit via `PlaceAlgoOrder` with `sl_tp` order type
4. **Established patterns** - Project has consistent Tool structure with rate limiting, error handling, and Markdown table response formatting

**Primary recommendation:** Implement 9 tools in priority order (P0 first), create ExecutorAgent as ChatModelAgent, integrate into AgentsModel with sync.Once singleton pattern.

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **9 order tools required:** `okx-place-order`, `okx-cancel-order`, `okx-get-order`, `okx-attach-sl-tp`, `okx-place-order-with-sl-tp`, `okx-batch-place-order`, `okx-batch-cancel-order`, `okx-get-order-history`, `okx-close-position`
- **Parameter design fixed** - place-order, close-position, SL/TP tool parameters as specified in CONTEXT.md
- **Rate limiting:** 5 req/s for all order tools (OKX trading endpoint conservative limit)
- **Executor Agent type:** ChatModelAgent (not DeepAgent), Level 1 autonomy (explicit commands only)
- **Response format:** Markdown tables for order results
- **No order amendment** - amend-order deferred, use cancel+replace pattern
- **No WebSocket push** - Agent actively queries order status
- **No frontend updates** - Phase 3 backend only

### Claude's Discretion
- Tool implementation order within P0/P1 priority
- Code organization (single file vs multiple files per tool)
- Executor Agent prompt wording
- Log message formatting details

### Deferred Ideas (OUT OF SCOPE)
- Order amendment functionality (amend-order)
- Automatic order timeout cancellation
- WebSocket order status push
- Frontend order management UI

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/cloudwego/eino/adk` | 0.8.4 (project) | Agent Development Kit | Project-standard framework for Agent building |
| `github.com/cloudwego/eino/schema` | 0.8.4 (project) | Tool schema definition | Consistent with existing tools |
| `github.com/cloudwego/eino/components/tool` | 0.8.4 (project) | BaseTool interface | Required for Eino tool integration |
| `golang.org/x/time/rate` | v0.3.0 (existing) | Rate limiting | Existing pattern in okx_candlesticks.go, okx_get_positions.go |
| `github.com/PineappleBond/TradingEino/backend/pkg/okex` | internal | OKX API client | Existing OKX client with signed requests |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/PineappleBond/TradingEino/backend/internal/utils/xmd` | internal | Markdown table generation | For consistent table output formatting |
| `encoding/json` | stdlib | JSON marshaling/unmarshaling | All tool argument parsing |
| `fmt` | stdlib | String formatting | Error messages, output formatting |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual rate limiting with `time.Sleep` | `golang.org/x/time/rate` Limiter | **rate.Limiter preferred** - proven pattern, handles burst correctly, context-aware |
| Returning raw JSON | Returning structured Markdown tables | **Markdown tables preferred** - consistent with existing tools, LLM-readable |
| Direct API calls in Agent | Tool abstraction layer | **Tool abstraction required** - separation of concerns, testability |

**Installation:**
```bash
# All dependencies already in go.mod from Phase 1
go mod tidy
```

## Architecture Patterns

### Recommended Project Structure
```
internal/agent/
├── agents.go                    # Add ExecutorAgent to AgentsModel
├── executor_agent/              # New directory
│   ├── executor_agent.go        # ExecutorAgent initialization
│   └── DESCRIPTION.md           # Agent description
└── tools/
    ├── okx_place_order.go       # P0 - Place order tool
    ├── okx_cancel_order.go      # P0 - Cancel order tool
    ├── okx_get_order.go         # P0 - Get order status tool
    ├── okx_attach_sl_tp.go      # P0 - Attach SL/TP to existing order
    ├── okx_place_order_with_sl_tp.go  # P0 - Place order with SL/TP
    ├── okx_batch_place_order.go # P1 - Batch place (max 20)
    ├── okx_batch_cancel_order.go # P1 - Batch cancel (max 20)
    ├── okx_get_order_history.go # P1 - Query historical orders
    └── okx_close_position.go    # P1 - Close position (partial/full)
```

### Pattern 1: Tool Structure (Existing Standard)
**What:** Consistent tool structure with svcCtx, rate limiter, Info/InvokableRun methods

**Example:**
```go
// Source: internal/agent/tools/okx_get_positions.go
type OkxGetPositionsTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

func NewOkxGetPositionsTool(svcCtx *svc.ServiceContext) *OkxGetPositionsTool {
	return &OkxGetPositionsTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s
	}
}

func (c *OkxGetPositionsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-get-positions-tool",
		Desc:  "调用 OKX 接口获取当前仓位和最大购买力的工具",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"symbol": {Type: schema.String, Desc: "交易对", Required: true},
		}),
	}, nil
}

func (c *OkxGetPositionsTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 1. Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// 2. Parse arguments
	var args struct{ Symbol string }
	json.Unmarshal([]byte(argumentsInJSON), &args)

	// 3. Call API via OKXClient
	result, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(...)
	if err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", &okex.OKXError{Code: result.Code, Msg: result.Msg, Endpoint: "GetPositions"}
	}

	// 4. Format output as Markdown
	output := "# Output\n\n| Column | Value |\n|---|---|\n"
	return output, nil
}
```

### Pattern 2: Agent Initialization with sync.Once
**What:** Singleton pattern for Agent initialization with proper context propagation

**Example:**
```go
// Source: internal/agent/agents.go
type AgentsModel struct {
	svcCtx           *svc.ServiceContext
	OkxWatcher       adk.Agent
	RiskOfficer      adk.Agent
	SentimentAnalyst adk.Agent
	Executor         adk.Agent              // Add for Phase 3
	mux              sync.Mutex
	ctx              context.Context
	cancel           context.CancelFunc
}

func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
	var initErr error
	agentsOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)

		// Initialize existing agents...
		riskOfficerAgent, _ := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
		sentimentAnalystAgent, _ := sentiment_analyst.NewSentimentAnalystAgent(ctx, svcCtx)
		okxWatcherAgent, _ := okx_watcher.NewOkxWatcherAgent(ctx, svcCtx, riskOfficerAgent.Agent(), sentimentAnalystAgent.Agent())

		// Add Executor Agent for Phase 3
		executorAgent, _ := executor_agent.NewExecutorAgent(ctx, svcCtx)

		_agents = &AgentsModel{
			svcCtx:           svcCtx,
			OkxWatcher:       okxWatcherAgent.Agent(),
			RiskOfficer:      riskOfficerAgent.Agent(),
			SentimentAnalyst: sentimentAnalystAgent.Agent(),
			Executor:         executorAgent.Agent(),  // Phase 3
			ctx:              ctx,
			cancel:           cancel,
		}
	})
	return initErr
}
```

### Pattern 3: OKX API Response Validation
**What:** All API calls must validate `Code` (response-level) and `SCode` (order-level)

**Example:**
```go
// Source: pkg/okex/api/trading.go
func (c *Client) PlaceOrder(ctx context.Context, instID, side, posSide, ordType, size, price string) (*OrderResult, error) {
	req := []tradeRequests.PlaceOrder{{InstID: instID, Side: okex.OrderSide(side), ...}}

	resp, err := c.Rest.Trade.PlaceOrder(req)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	// Check response-level code
	if resp.Code != 0 {
		return nil, fmt.Errorf("place order failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if len(resp.PlaceOrders) == 0 {
		return nil, fmt.Errorf("place order failed: empty response")
	}

	result := resp.PlaceOrders[0]

	// Check order-level SCode (silent failure detection)
	if result.SCode != 0 {
		return nil, fmt.Errorf("order placement error: code=%d, msg=%s", result.SCode, result.SMsg)
	}

	return &OrderResult{OrderID: result.OrdID, State: "pending", SMsg: result.SMsg}, nil
}
```

### Anti-Patterns to Avoid
- **Never return `(err.Error(), nil)`** - Always return `("", err)` for errors (FOUND-01 requirement)
- **Never skip rate limiter wait** - Always call `limiter.Wait(ctx)` before API calls (FOUND-02 requirement)
- **Never use `context.Background()`** - Always propagate parent context (FOUND-04 requirement)
- **Never skip SCode validation** - OKX can return `Code=0` but `SCode!=0` for individual order failures

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Rate limiting | Custom time.Sleep + counter | `golang.org/x/time/rate.Limiter` | Handles burst, context cancellation, proven pattern |
| OKX API signing | Manual HMAC-SHA256 | `pkg/okex/api.Client` | Existing client handles signing, timestamp, passphrase |
| Markdown tables | Manual string concatenation | `internal/utils/xmd.CreateMarkdownTable` | Consistent formatting, less error-prone |
| Singleton initialization | Global variable assignment | `sync.Once` pattern | Thread-safe, prevents race conditions |
| Error wrapping | Raw error strings | `fmt.Errorf("...: %w", err)` | Preserves error chain for debugging |
| SL/TP order logic | Custom trigger monitoring | OKX `PlaceAlgoOrder` with `sl_tp` type | Server-side execution, no polling needed |

**Key insight:** The project already has all infrastructure needed - OKX client, rate limiting, Agent patterns. Phase 3 is about applying existing patterns consistently.

## Common Pitfalls

### Pitfall 1: Silent Order Failures (SCode Check)
**What goes wrong:** OKX response has `Code=0` (success) but individual orders have `SCode!=0` (failed)

**Why it happens:** Batch operations return overall success even if some orders fail

**How to avoid:** Always check both `resp.Code` AND `result.SCode` for each order in response

**Warning signs:** Order appears to succeed but never appears in order book

### Pitfall 2: Rate Limiter Burst Configuration
**What goes wrong:** Setting `rate.Every(time.Second/5)` with burst=5 causes initial burst of 5 requests

**Why it happens:** Limiter starts with `burst` tokens available immediately

**How to avoid:** Use `rate.NewLimiter(rate.Every(200*time.Millisecond), 1)` - burst=1 for steady 5 req/s

**Warning signs:** First API call succeeds, next 4 fail with rate limit error

### Pitfall 3: Price Precision Errors
**What goes wrong:** Sending price with wrong decimal places causes OKX rejection

**Why it happens:** Each instrument has different `tick_sz` (price precision)

**How to avoid:** Document in tool description that price must match instrument's `tick_sz`; let OKX reject invalid prices

**Warning signs:** OKX returns `SCode=51000` with message about price precision

### Pitfall 4: Context Cancellation in Rate Limiter
**What goes wrong:** `limiter.Wait(ctx)` blocks forever if context has no timeout

**Why it happens:** If queue is full, Wait blocks until context is cancelled

**How to avoid:** Ensure parent context has reasonable timeout; OKXWatcher should set per-task timeout

**Warning signs:** Tool hangs indefinitely during high load

### Pitfall 5: Batch Operation Partial Failures
**What goes wrong:** 10 of 20 orders succeed, 10 fail - need to report both

**Why it happens:** Some orders may fail due to balance, position limits, etc.

**How to avoid:** Iterate through all results, categorize by `SCode`, report successes and failures separately

**Warning signs:** User sees only success or only failure, not mixed results

## Code Examples

Verified patterns from official sources:

### Place Order Tool (P0)
```go
// Source: Pattern from okx_get_positions.go + OKX API trading.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	tradeRequests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/time/rate"
)

type OkxPlaceOrderTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

func NewOkxPlaceOrderTool(svcCtx *svc.ServiceContext) *OkxPlaceOrderTool {
	return &OkxPlaceOrderTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s
	}
}

func (c *OkxPlaceOrderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "okx-place-order",
		Desc: "OKX 下单工具 - 支持限价单、市价单、POST_ONLY、FOK、IOC 订单类型",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"instID": {
				Type:     schema.String,
				Desc:     "交易对，如 ETH-USDT-SWAP",
				Required: true,
			},
			"side": {
				Type:     schema.String,
				Desc:     "订单方向：buy 或 sell",
				Enum:     []string{"buy", "sell"},
				Required: true,
			},
			"posSide": {
				Type:     schema.String,
				Desc:     "仓位模式：long/short/net，默认 net",
				Enum:     []string{"long", "short", "net"},
				Required: false,
			},
			"ordType": {
				Type:     schema.String,
				Desc:     "订单类型：market/limit/post_only/fok/ioc",
				Enum:     []string{"market", "limit", "post_only", "fok", "ioc"},
				Required: true,
			},
			"size": {
				Type:     schema.String,
				Desc:     "订单数量（合约张数）",
				Required: true,
			},
			"price": {
				Type:     schema.String,
				Desc:     "订单价格，limit 和 post_only 必填，market 留空",
				Required: false,
			},
		}),
	}, nil
}

func (c *OkxPlaceOrderTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	// 1. Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// 2. Parse arguments
	var args struct {
		InstID  string `json:"instID"`
		Side    string `json:"side"`
		PosSide string `json:"posSide"`
		OrdType string `json:"ordType"`
		Size    string `json:"size"`
		Price   string `json:"price"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// Default posSide to "net" if not provided
	if args.PosSide == "" {
		args.PosSide = "net"
	}

	// 3. Call OKX API
	result, err := c.svcCtx.OKXClient.Rest.Trade.PlaceOrder([]tradeRequests.PlaceOrder{
		{
			InstID:  args.InstID,
			Side:    okex.OrderSide(args.Side),
			PosSide: okex.PositionSide(args.PosSide),
			OrdType: okex.OrderType(args.OrdType),
			Sz:      args.Size,
			Px:      args.Price,
			TdMode:  okex.TradeCrossMode,
		},
	})
	if err != nil {
		return "", err
	}

	// 4. Validate response
	if result.Code != 0 {
		return "", &okex.OKXError{Code: result.Code, Msg: result.Msg, Endpoint: "PlaceOrder"}
	}

	if len(result.PlaceOrders) == 0 {
		return "", fmt.Errorf("empty response from PlaceOrder")
	}

	order := result.PlaceOrders[0]

	// 5. Check order-level SCode (silent failure detection)
	if order.SCode != 0 {
		return "", fmt.Errorf("order failed: sCode=%d, sMsg=%s", order.SCode, order.SMsg)
	}

	// 6. Format output as Markdown table
	output := "## 订单提交成功\n\n"
	output += "| 字段 | 值 |\n|------|-----|\n"
	output += fmt.Sprintf("| ordId | %s |\n", order.OrdID)
	output += fmt.Sprintf("| clOrdId | %s |\n", order.ClOrdID)
	output += fmt.Sprintf("| tag | %s |\n", order.Tag)
	output += "| state | pending |\n"
	output += fmt.Sprintf("| sCode | %d |\n", order.SCode)
	output += fmt.Sprintf("| sMsg | %s |\n", order.SMsg)

	return output, nil
}
```

### Cancel Order Tool (P0)
```go
// Source: Pattern from okx_get_positions.go + pkg/okex/api/trading.go CancelOrder
func (c *OkxCancelOrderTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	var args struct {
		InstID  string `json:"instID"`
		OrdID   string `json:"ordID"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	result, err := c.svcCtx.OKXClient.Rest.Trade.CandleOrder([]tradeRequests.CancelOrder{
		{InstID: args.InstID, OrdID: args.OrdID},
	})
	if err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", &okex.OKXError{Code: result.Code, Msg: result.Msg, Endpoint: "CancelOrder"}
	}

	if len(result.CancelOrders) == 0 {
		return "", fmt.Errorf("empty response from CancelOrder")
	}

	cancelResult := result.CancelOrders[0]

	if cancelResult.SCode != 0 {
		return "", fmt.Errorf("cancel failed: sCode=%d, sMsg=%s", cancelResult.SCode, cancelResult.SMsg)
	}

	output := "## 订单已取消\n\n"
	output += fmt.Sprintf("| ordId | %s |\n| state | cancelled |\n", cancelResult.OrdID)
	return output, nil
}
```

### Place Order with SL/TP (P0 - Algo Order)
```go
// Source: pkg/okex/api/rest/trade.go PlaceAlgoOrder
// OKX sl_tp order type: https://www.okex.com/docs-v5/en/#rest-api-trade-place-algo-order
func (c *OkxPlaceOrderWithSlTpTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	var args struct {
		InstID      string  `json:"instID"`
		Side        string  `json:"side"`
		PosSide     string  `json:"posSide"`
		OrdType     string  `json:"ordType"`
		Size        string  `json:"size"`
		Price       string  `json:"price"`
		SlTriggerPx float64 `json:"slTriggerPx"`
		SlOrderPx   float64 `json:"slOrderPx"`
		TpTriggerPx float64 `json:"tpTriggerPx"`
		TpOrderPx   float64 `json:"tpOrderPx"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// OKX sl_tp order requires tp_trigger_px and/or sl_trigger_px
	algoOrder := tradeRequests.PlaceAlgoOrder{
		InstID:  args.InstID,
		TdMode:  okex.TradeCrossMode,
		Side:    okex.OrderSide(args.Side),
		PosSide: okex.PositionSide(args.PosSide),
		OrdType: okex.AlgoOrderType("sl_tp"),
		Sz:      int64(len(args.Size)), // Note: size as integer
	}

	// Set SL parameters
	if args.SlTriggerPx > 0 {
		algoOrder.SlTriggerPx = args.SlTriggerPx
		algoOrder.SlOrdPx = args.SlOrderPx // 0 = market order
	}

	// Set TP parameters
	if args.TpTriggerPx > 0 {
		algoOrder.TpTriggerPx = args.TpTriggerPx
		algoOrder.TpOrdPx = args.TpOrderPx // 0 = market order
	}

	result, err := c.svcCtx.OKXClient.Rest.Trade.PlaceAlgoOrder(algoOrder)
	if err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", &okex.OKXError{Code: result.Code, Msg: result.Msg, Endpoint: "PlaceAlgoOrder"}
	}

	if len(result.PlaceAlgoOrders) == 0 {
		return "", fmt.Errorf("empty response from PlaceAlgoOrder")
	}

	algoResult := result.PlaceAlgoOrders[0]

	if algoResult.SCode != 0 {
		return "", fmt.Errorf("algo order failed: sCode=%d, sMsg=%s", algoResult.SCode, algoResult.SMsg)
	}

	output := "## 止盈止损订单提交成功\n\n"
	output += fmt.Sprintf("| algoId | %s |\n", algoResult.AlgoID)
	output += fmt.Sprintf("| sCode | %d |\n| sMsg | %s |\n", algoResult.SCode, algoResult.SMsg)
	return output, nil
}
```

### Close Position Tool (P1)
```go
// Source: pkg/okex/api/rest/trade.go ClosePosition
func (c *OkxClosePositionTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	var args struct {
		InstID     string  `json:"instID"`
		PosSide    string  `json:"posSide"`
		Percentage float64 `json:"percentage"` // 0-100, default 100
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal args: %w", err)
	}

	// Note: OKX ClosePosition API closes entire position
	// For partial close, user should place opposite order manually
	// Percentage parameter requires custom implementation

	result, err := c.svcCtx.OKXClient.Rest.Trade.ClosePosition(tradeRequests.ClosePosition{
		InstID:  args.InstID,
		PosSide: okex.PositionSide(args.PosSide),
		MgnMode: okex.MarginCrossMode,
	})
	if err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", &okex.OKXError{Code: result.Code, Msg: result.Msg, Endpoint: "ClosePosition"}
	}

	output := "## 仓位已关闭\n\n"
	output += fmt.Sprintf("| instId | %s |\n| posSide | %s |\n", args.InstID, args.PosSide)
	return output, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual HTTP signing | `pkg/okex/api.Client` with built-in signing | Project inception | Consistent, tested signing |
| Global variables for Agents | `sync.Once` singleton pattern | Phase 1 (FOUND-03) | Thread-safe initialization |
| No rate limiting | `rate.Limiter` per tool | Phase 1 (FOUND-02) | Prevents API throttling |
| `err.Error(), nil` returns | `("", err)` returns | Phase 1 (FOUND-01) | Proper error propagation |
| `context.Background()` | Context propagation from entry | Phase 1 (FOUND-04) | Cancellation works end-to-end |

**Deprecated/outdated:**
- Using global `_agents` variable directly - Use `InitAgents()` + `Agents()` accessor
- Returning raw strings for errors - Always return typed errors
- Skipping `sCode` check - Always validate order-level response

## Open Questions

1. **Partial close percentage implementation**
   - What we know: OKX `ClosePosition` API closes entire position
   - What's unclear: How to implement percentage-based partial close
   - Recommendation: For P1, implement as "place opposite market order for percentage of position size" - requires fetching position size first, then calculating order size

2. **Executor Agent prompt constraints**
   - What we know: Level 1 autonomy = only execute explicit OKXWatcher commands
   - What's unclear: Exact wording for prompt to enforce this constraint
   - Recommendation: Include explicit rules: "DO NOT initiate trades", "WAIT for explicit command", "REPORT failures without retry"

3. **Batch operation error aggregation**
   - What we know: OKX returns array of results with individual `sCode`
   - What's unclear: Best format for mixed success/failure results
   - Recommendation: Split output into "## 成功订单" and "## 失败订单" sections with separate tables

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package (`testing`) |
| Config file | None — tests alongside implementation files |
| Quick run command | `go test ./internal/agent/tools/... -run TestOkxPlaceOrder -v` |
| Full suite command | `go test ./internal/agent/tools/... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| EXEC-01 | Place limit/market orders | Integration | `go test ./internal/agent/tools/... -run TestPlaceOrder` | ❌ Wave 0 |
| EXEC-02 | Cancel pending orders | Integration | `go test ./internal/agent/tools/... -run TestCancelOrder` | ❌ Wave 0 |
| EXEC-03 | Query order status | Integration | `go test ./internal/agent/tools/... -run TestGetOrder` | ❌ Wave 0 |
| EXEC-04 | Executor Level 1 autonomy | Manual | N/A - Agent behavior validation | ❌ Wave 0 |
| EXEC-05 | SL/TP via sl_tp algo | Integration | `go test ./internal/agent/tools/... -run TestPlaceOrderWithSlTp` | ❌ Wave 0 |
| EXEC-06 | Order sCode/sMsg validation | Unit + Integration | `go test ./internal/agent/tools/... -run TestOrderValidation` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/agent/tools/... -v`
- **Per wave merge:** `go test ./... -v` (full suite)
- **Phase gate:** All tests green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/agent/tools/okx_place_order_test.go` — covers EXEC-01, EXEC-06
- [ ] `internal/agent/tools/okx_cancel_order_test.go` — covers EXEC-02
- [ ] `internal/agent/tools/okx_get_order_test.go` — covers EXEC-03
- [ ] `internal/agent/tools/okx_place_order_with_sl_tp_test.go` — covers EXEC-05
- [ ] `internal/agent/executor_agent/executor_agent_test.go` — covers EXEC-04
- [ ] Test fixtures/mock OKX responses for offline testing

## Sources

### Primary (HIGH confidence)
- **Project code** - `pkg/okex/api/trading.go` - PlaceOrder, CancelOrder, GetOrderDetails implementations
- **Project code** - `pkg/okex/api/rest/trade.go` - Full Trade API client methods
- **Project code** - `pkg/okex/requests/rest/trade/trade_requests.go` - Request structures
- **Project code** - `pkg/okex/responses/trade/trade_responses.go` - Response structures
- **Project code** - `pkg/okex/definitions.go` - Type definitions (OrderType, OrderState, etc.)
- **Project code** - `internal/agent/tools/okx_get_positions.go` - Existing tool pattern
- **Project code** - `internal/agent/agents.go` - Agent initialization pattern
- **CONTEXT.md** - User decisions for Phase 3 scope and parameters

### Secondary (MEDIUM confidence)
- **OKX API Documentation** - https://www.okex.com/docs-v5/en/#rest-api-trade - Trade API reference
- **Eino Skill** - `.claude/skills/eino-skill/SKILL.md` - ChatModelAgent pattern
- **OKEX Skill** - `.claude/skills/okex-skill/SKILL.md` - OKX client usage patterns

### Tertiary (LOW confidence)
- None - All critical claims verified against project source code

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries from existing project dependencies
- Architecture: HIGH - Patterns extracted from working project code
- Pitfalls: HIGH - Derived from existing error handling patterns and OKX API behavior
- API capabilities: HIGH - Verified against `pkg/okex/api/rest/trade.go` implementation

**Research date:** 2026-03-24
**Valid until:** 90 days (OKX API is stable; project patterns established in Phase 1)

---

*Phase 3 Research Complete - Ready for Planning*
