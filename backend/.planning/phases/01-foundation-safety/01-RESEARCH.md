# Phase 1: Foundation & Safety - Research

**Researched:** 2026-03-24
**Domain:** Go error handling, rate limiting, singleton patterns, context propagation, graceful shutdown
**Confidence:** HIGH

## Summary

Phase 1 focuses on establishing safe operational infrastructure for the TradingEino system. The phase addresses five critical requirements (FOUND-01 through FOUND-05) that ensure the system handles errors properly, respects API rate limits, manages resources correctly, and shuts down gracefully.

**Key findings:**
1. Current Tool implementations already follow correct error handling patterns (`return "", err`)
2. Only `OkxCandlesticksTool` has a rate limiter; `OkxGetPositionsTool` and `OkxGetFundingRateTool` need limiters added
3. Agents use global variables without `sync.Once` protection - needs refactoring
4. `InitAgents` uses `context.Background()` internally instead of accepting parent context
5. `main.go` lacks signal handling and graceful shutdown logic

**Primary recommendation:** Implement fixes in this order: error handling audit → rate limiting → singleton pattern → context propagation → graceful shutdown

## User Constraints (from CONTEXT.md)

### Locked Decisions
- 所有 Tool 的 `InvokableRun` 统一返回 `("", err)` 格式
- OKX API 错误使用 `OKXError` 结构体表示：
  - 文件位置：`pkg/okex/okx_error.go`
  - 字段：`Code int` (OKX 错误码), `Msg string` (错误消息), `Endpoint string` (请求端点，用于调试)
  - 方法：`Error() string`, `Unwrap() error` (支持 errors.As 解包)
- Tool 层负责检查 `result.Code != 0` 并返回 `&OKXError{...}`
- 直接返回 `&OKXError` 实例，不使用 `fmt.Errorf` 包装
- 每个 Tool 在 `NewXxxTool` 函数内初始化自己的 `limiter *rate.Limiter`
- 严格限流配置（按端点类型分类硬编码）：
  - Trade/Account 端点：5 次/秒 (burst=1)
  - Market/Public 端点：10 次/秒 (burst=2)
  - Funding 端点：1 次/秒 (burst=1)
- API 调用前必须调用 `limiter.Wait(ctx)`
- 保持全局变量 `_agents` 模式
- 添加文档说明：必须在 main 中先调用 `InitAgents()` 后才能使用 `Agents()`
- `InitAgents(ctx context.Context, svcCtx *svc.ServiceContext)` 签名修改，接收父上下文参数
- 使用 `os/signal` 包监听 SIGINT/SIGTERM 信号
- 显式顺序关闭流程：Server → Scheduler → Agents → DB → Logger

### Claude's Discretion
- OKXError 的具体实现风格（保持与项目现有代码一致）
- 限流器的具体 burst 值可根据实际 API 限制微调
- 优雅关闭的日志输出格式

### Deferred Ideas (OUT OF SCOPE)
- 自定义错误分类（业务错误/限流错误/网络错误）— 未来扩展，当前仅需 OKXError
- 限流器配置化（从 config.yaml 读取）— 当前硬编码，未来可配置
- 依赖注入模式重构 — 保持全局变量模式，不重构
- ServiceContext 统一管理限流器 — 各 Tool 自行管理

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| FOUND-01 | All tools return errors properly (`"", err`) instead of (`err.Error(), nil`) | Go error handling patterns, OKXError structure design |
| FOUND-02 | All API tools have rate.Limiter with conservative limits (5 req/s for trade endpoints) | golang.org/x/time/rate usage, OKX API rate limit categories |
| FOUND-03 | Agents use singleton pattern with sync.Once instead of global variables | sync.Once patterns for lazy initialization |
| FOUND-04 | Context propagation throughout agent initialization and tool execution | context.Context threading best practices |
| FOUND-05 | Graceful shutdown on application exit | os/signal handling, resource cleanup order |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/time` | v0.15.0 | Rate limiting with `rate.Limiter` | Official Go time utilities, built-in rate limiting |
| `context` | stdlib | Context propagation | Go standard context package |
| `os/signal` | stdlib | Signal handling | Go standard signal handling |
| `sync.Once` | stdlib | Singleton pattern | Go standard synchronization primitive |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/robfig/cron/v3` | v3.0.1 | Scheduler with Stop() support | Background job scheduling |
| `github.com/sirupsen/logrus` | v1.9.3 | Structured logging | Used by Eino framework |
| Custom logger (`internal/logger`) | - | JSON logging with stack traces | Project-specific logging wrapper |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `rate.Limiter` | `golang.org/x/sync/semaphore` | rate.Limiter provides token bucket algorithm, better for API limits |
| `sync.Once` | Manual mutex + flag | sync.Once is atomic and race-condition free |
| Global agents | Dependency injection | Global is simpler, DI is more testable (deferred to future) |

**Installation:**
Dependencies already in `go.mod`. No additional packages needed.

## Architecture Patterns

### Recommended Project Structure
```
backend/
├── cmd/server/
│   └── main.go              # Application entry + signal handling
├── pkg/okex/
│   └── okx_error.go         # NEW: OKXError type definition
├── internal/
│   ├── agent/
│   │   ├── agents.go        # Singleton pattern with sync.Once
│   │   └── tools/
│   │       ├── okx_candlesticks.go      # Has limiter (reference)
│   │       ├── okx_get_positions.go     # NEEDS limiter
│   │       └── okx_get_fundingrate.go   # NEEDS limiter
│   ├── logger/
│   │   └── logger.go        # Has Close() method to implement
│   └── svc/
│       ├── servicecontext.go # OKXClient, DB connections
│       └── database.go      # Has fmt.Printf to fix
```

### Pattern 1: Error Handling with OKXError
**What:** Custom error type for OKX API errors that preserves error code and endpoint info
**When to use:** All Tool implementations that call OKX API

**Example:**
```go
// pkg/okex/okx_error.go
type OKXError struct {
    Code     int
    Msg      string
    Endpoint string
}

func (e *OKXError) Error() string {
    return fmt.Sprintf("OKX %s error (code=%d): %s", e.Endpoint, e.Code, e.Msg)
}

func (e *OKXError) Unwrap() error {
    return nil
}

// Tool implementation
func (c *OkxGetPositionsTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    result, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(...)
    if err != nil {
        return "", err  // Return original error
    }
    if result.Code != 0 {
        return "", &OKXError{
            Code:     result.Code,
            Msg:      result.Msg,
            Endpoint: "GetPositions",
        }
    }
    return json.Marshal(result.Data)
}
```

### Pattern 2: Rate Limiter per Tool
**What:** Each Tool owns its own `rate.Limiter` initialized in constructor
**When to use:** All Tool implementations that call external APIs

**Example:**
```go
import "golang.org/x/time/rate"

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

func (c *OkxGetPositionsTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    // 1. Wait for rate limit
    if err := c.limiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limiter wait failed: %w", err)
    }

    // 2. Call API
    result, err := c.svcCtx.OKXClient.Rest.Account.GetPositions(...)
    if err != nil {
        return "", err
    }

    // 3. Check OKX response code
    if result.Code != 0 {
        return "", &OKXError{...}
    }

    return json.Marshal(result.Data)
}
```

### Pattern 3: Singleton with sync.Once
**What:** Thread-safe lazy initialization using sync.Once
**When to use:** Global resources that must be initialized once

**Example:**
```go
var (
    agentsOnce sync.Once
    _agents    *AgentsModel
)

func Agents() *AgentsModel {
    return _agents
}

func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
    var initErr error
    agentsOnce.Do(func() {
        // Initialize sub-agents using ctx
        riskOfficerAgent, err := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
        if err != nil {
            initErr = err
            return
        }
        // ... other initializations

        _agents = &AgentsModel{
            svcCtx: svcCtx,
            ctx:    ctx,
            cancel: cancel,
        }
    })
    return initErr
}

func (a *AgentsModel) Close() error {
    if a.cancel != nil {
        a.cancel()
    }
    return nil
}
```

### Pattern 4: Context Propagation
**What:** Thread context through all initialization layers
**When to use:** All agent and tool initialization

**Example:**
```go
// main.go
ctx := context.Background()
logger.Info(ctx, "application starting")

svcCtx := svc.NewServiceContext(*cfg)

// Pass ctx to InitAgents
err = agent.InitAgents(ctx, svcCtx)
if err != nil {
    // handle error
}

// In agents.go
func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
    // Use the passed ctx, NOT context.Background()
    riskOfficerAgent, err := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
    // ...
}

// In risk_officer/agent.go
func NewRiskOfficerAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*RiskOfficerAgent, error) {
    // Use ctx for all operations
    agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        // ...
    })
    // ...
}
```

### Pattern 5: Graceful Shutdown
**What:** Signal handling with ordered resource cleanup
**When to use:** Application main function

**Example:**
```go
// main.go
func main() {
    // ... setup ...

    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Wait for shutdown signal
    <-sigChan
    logger.Info(ctx, "shutdown signal received")

    // Ordered shutdown
    // 1. Stop HTTP server
    if err := server.Shutdown(ctx); err != nil {
        logger.Error(ctx, "failed to shutdown server", err)
    }

    // 2. Stop scheduler
    if err := scheduler.Stop(); err != nil {
        logger.Error(ctx, "failed to stop scheduler", err)
    }

    // 3. Close agents
    if err := agents.Close(); err != nil {
        logger.Error(ctx, "failed to close agents", err)
    }

    // 4. Close database
    db, _ := svcCtx.DB.DB()
    if err := db.Close(); err != nil {
        logger.Error(ctx, "failed to close database", err)
    }

    // 5. Close logger
    if err := logger.Close(); err != nil {
        logger.Error(ctx, "failed to close logger", err)
    }

    logger.Info(ctx, "graceful shutdown completed")
}
```

### Anti-Patterns to Avoid
- **Using `context.Background()` inside `InitAgents`** — breaks cancellation propagation
- **Returning `err.Error()` as success** — masks errors, breaks retry logic
- **Sharing rate limiters across tools** — causes unexpected throttling
- **Global variables without sync.Once** — race conditions during initialization
- **Closing resources in wrong order** — can cause panics or resource leaks

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Rate limiting | Custom token bucket | `golang.org/x/time/rate` | Battle-tested, handles edge cases, well-documented |
| Singleton pattern | Manual mutex + flag | `sync.Once` | Atomic, no race conditions, standard library |
| Error wrapping | String concatenation | `fmt.Errorf("%w")` or custom types | Supports `errors.As`, `errors.Is` |
| Signal handling | Polling os.Interrupt | `os/signal.Notify` | OS-level integration, efficient |
| Context cancellation | Custom done channels | `context.Context` | Standard pattern, composable |

**Key insight:** Go's standard library provides robust, well-tested primitives for all Phase 1 requirements. Custom implementations introduce bugs and maintenance burden.

## Common Pitfalls

### Pitfall 1: Error Masking
**What goes wrong:** Returning `err.Error()` as success string instead of `("", err)`
**Why it happens:** Confusion about error handling contract in Tool interface
**How to avoid:** Always return `("", err)` on error, `(result, nil)` on success
**Warning signs:** Code like `return err.Error(), nil` in error branches

### Pitfall 2: Missing Rate Limiters
**What goes wrong:** API calls hit rate limits, causing temporary bans
**Why it happens:** Forgetting to add limiter to new Tool implementations
**How to avoid:** Add limiter as struct field, initialize in constructor, call `Wait(ctx)` before API calls
**Warning signs:** Tool struct without `limiter *rate.Limiter` field

### Pitfall 3: Context.Background() Leakage
**What goes wrong:** Cancellation signals don't propagate to child operations
**Why it happens:** Using `context.Background()` instead of passed context
**How to avoid:** Always accept `ctx context.Context` parameter, never create new root context
**Warning signs:** `context.Background()` calls inside initialization functions

### Pitfall 4: Race Condition in Singleton
**What goes wrong:** Multiple goroutines initialize resource multiple times
**Why it happens:** Manual check-then-set without proper synchronization
**How to avoid:** Use `sync.Once` for all singleton initializations
**Warning signs:** `if _agents == nil { _agents = ... }` without mutex

### Pitfall 5: Resource Leaks on Shutdown
**What goes wrong:** Goroutines continue running after main exits
**Why it happens:** No signal handling or improper cleanup order
**How to avoid:** Setup signal handler early, implement Close() methods, wait for goroutines
**Warning signs:** No `signal.Notify` in main.go, no `sync.WaitGroup` usage

## OKX API Reference

### Rate Limits by Endpoint Category
Based on OKX API V5 documentation:

| Category | Endpoints | Rate Limit | Recommended Limiter |
|----------|-----------|------------|---------------------|
| **Trade** | Place order, Cancel order, Amend order | 60 requests/2s per API key | `rate.NewLimiter(rate.Every(200*time.Millisecond), 1)` = 5 req/s |
| **Account** | Get balance, Get positions, Get max buy/sell | 60 requests/2s per API key | `rate.NewLimiter(rate.Every(200*time.Millisecond), 1)` = 5 req/s |
| **Market** | Get tickers, Get candles, Get depth | 20 requests/2s | `rate.NewLimiter(rate.Every(100*time.Millisecond), 2)` = 10 req/s |
| **Public** | Get funding rate, Get interest rate | 20 requests/2s | `rate.NewLimiter(rate.Every(100*time.Millisecond), 2)` = 10 req/s |
| **Funding** | Get deposit address, Withdraw, Transfer | 1 request/second | `rate.NewLimiter(rate.Every(time.Second), 1)` = 1 req/s |

### Error Code Patterns
OKX API returns responses with `code` and `msg` fields:

```go
type BasicResponse struct {
    Code int    `json:"code,string"`
    Msg  string `json:"msg,omitempty"`
    Data T      `json:"data"`
}
```

**Success:** `code = "0"`
**Error:** `code != "0"` with descriptive `msg`

Common error codes:
- `0` - Success
- `50001` - Invalid API key
- `50002` - Invalid signature
- `50003` - Request timestamp expired
- `50004` - Request method invalid
- `50005` - Invalid Content-Type
- `50006` - Invalid OK-ACCESS-PASSPHRASE
- `50007` - Rate limit exceeded
- `50008` - Endpoint does not exist
- `50009` - Parameter verification error

### OKXClient Usage Pattern
```go
import (
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/api"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
)

// ServiceContext provides initialized client
type ServiceContext struct {
    OKXClient *api.Client
}

// Call example
positions, err := svcCtx.OKXClient.Rest.Account.GetPositions(accountrequests.GetPositions{
    InstID: []string{"ETH-USDT-SWAP"},
})
if err != nil {
    return "", err
}
if positions.Code != 0 {
    return "", &OKXError{
        Code:     positions.Code,
        Msg:      positions.Msg,
        Endpoint: "GetPositions",
    }
}
```

## Validation Architecture

> Skip this section entirely if workflow.nyquist_validation is explicitly set to false in .planning/config.json. If the key is absent, treat as enabled.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package (stdlib) + testify v1.11.1 |
| Config file | None — standard Go test files |
| Quick run command | `go test ./internal/agent/tools/... -run TestOkxGetPositionsTool -v` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FOUND-01 | Tools return `("", err)` on error | unit | `go test ./internal/agent/tools/... -v` | ❌ Wave 0 |
| FOUND-02 | Tools have rate.Limiter with correct limits | unit | `go test ./internal/agent/tools/... -v` | ❌ Wave 0 |
| FOUND-03 | Agents initialize via sync.Once | unit | `go test ./internal/agent/... -v` | ❌ Wave 0 |
| FOUND-04 | Context propagation works end-to-end | integration | `go test ./internal/agent/... -v` | ❌ Wave 0 |
| FOUND-05 | Graceful shutdown cleans up resources | integration | Manual verification | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/agent/tools/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/agent/tools/okx_get_positions_test.go` — covers FOUND-01, FOUND-02
- [ ] `internal/agent/tools/okx_get_fundingrate_test.go` — covers FOUND-01, FOUND-02
- [ ] `internal/agent/tools/okx_candlesticks_test.go` — covers FOUND-01, FOUND-02
- [ ] `internal/agent/agents_test.go` — covers FOUND-03, FOUND-04
- [ ] `pkg/okex/okx_error_test.go` — covers FOUND-01
- [ ] Framework install: `go mod download` — dependencies already in go.mod

## Code Examples

Verified patterns from project codebase:

### Error Handling Pattern (OkxCandlesticksTool)
```go
// Source: internal/agent/tools/okx_candlesticks.go
func (c *OkxCandlesticksTool) GetCandlesticks(ctx context.Context, symbol string, bar okex.BarSize, afterDatetime *time.Time, limit int) ([]*market.Candle, error) {
    candles := make([]*market.Candle, 0)
    for {
        // Wait for rate limit
        if err := c.limiter.Wait(ctx); err != nil {
            return nil, err
        }
        getCandlesticksHistory, err := c.svcCtx.OKXClient.Rest.Market.GetCandlesticksHistory(...)
        if err != nil {
            return nil, err
        }
        // Process response...
    }
    return candles, nil
}
```

### Rate Limiter Initialization
```go
// Source: internal/agent/tools/okx_candlesticks.go
func NewOkxCandlesticksTool(svcCtx *svc.ServiceContext) *OkxCandlesticksTool {
    return &OkxCandlesticksTool{
        svcCtx:  svcCtx,
        limiter: rate.NewLimiter(rate.Every(time.Second/10), 1), // 10 req/s
    }
}
```

### Context Usage Pattern (Scheduler)
```go
// Source: internal/service/scheduler/scheduler.go
func NewScheduler(svcCtx *svc.ServiceContext) *Scheduler {
    ctx, cancel := context.WithCancel(context.Background())
    return &Scheduler{
        ctx:    ctx,
        cancel: cancel,
        // ...
    }
}

func (s *Scheduler) Stop() error {
    s.cancel()  // Cancel all child operations
    s.wg.Wait() // Wait for goroutines to complete
    s.cron.Stop()
    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual rate limiting with time.Sleep | rate.Limiter token bucket | Ongoing | Precise control, no goroutine leaks |
| fmt.Printf for debugging | Structured logger with JSON output | Project standard | Audit trail, log aggregation |
| context.Background() everywhere | Context propagation | Project standard | Cancellation works end-to-end |
| Global variables without sync | sync.Once for singletons | Phase 1 | Race-condition free |
| No shutdown handling | Signal-based graceful shutdown | Phase 1 | Clean resource cleanup |

**Deprecated/outdated:**
- `fmt.Fprintf(os.Stderr, ...)`: Replace with `logger.Error(ctx, ...)` — affects `internal/svc/database.go`, `internal/service/scheduler/handlers/okx_watcher_handler.go`
- `if false { ... }` dead code blocks: Remove entirely — found in `internal/service/scheduler/handlers/okx_watcher_handler.go:190-193`

## Open Questions

1. **Logger.Close() implementation details**
   - What we know: Logger wraps slog.Logger, may have file handles
   - What's unclear: Whether current implementation requires explicit Close()
   - Recommendation: Add Close() method that flushes buffers and closes file handles

2. **OKXError retry classification**
   - What we know: Some OKX errors are retryable (rate limit), some are not (invalid API key)
   - What's unclear: Whether to implement RetryableError interface for OKXError
   - Recommendation: Defer to Phase 3 when retry logic is needed

3. **Main.go shutdown timeout**
   - What we know: Shutdown should have timeout to prevent hanging
   - What's unclear: Appropriate timeout duration
   - Recommendation: Use 30-second timeout for graceful shutdown phase

## Sources

### Primary (HIGH confidence)
- Project codebase: `internal/agent/tools/okx_candlesticks.go` — rate limiter pattern
- Project codebase: `internal/agent/agents.go` — current agents structure
- Project codebase: `internal/service/scheduler/scheduler.go` — Stop() pattern, context usage
- Project codebase: `cmd/server/main.go` — current main without signal handling
- Project codebase: `CLAUDE.md` — project conventions and OKX API patterns
- `.planning/phases/01-foundation-safety/01-CONTEXT.md` — user decisions

### Secondary (MEDIUM confidence)
- Go standard library documentation — context, sync.Once, os/signal, rate.Limiter
- OKX API V5 documentation — rate limits, error codes (via project skill docs)

### Tertiary (LOW confidence)
- OKX API rate limits — may have changed, verify with official docs during implementation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — All libraries are Go stdlib or already in go.mod
- Architecture: HIGH — Patterns verified against existing project code
- Pitfalls: HIGH — Based on code analysis and common Go anti-patterns
- OKX API limits: MEDIUM — Based on documentation, may need verification

**Research date:** 2026-03-24
**Valid until:** 180 days (Go stdlib and project patterns are stable)
