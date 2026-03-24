# Domain Pitfalls: Crypto Trading Bots

**Domain:** AI-Powered Cryptocurrency Trading System
**Researched:** 2026-03-24
**Confidence:** MEDIUM (based on codebase analysis + established trading system patterns)

---

## Critical Pitfalls

Mistakes that cause financial losses, account bans, or system failures.

### 1. Silent Order Failures (Error Handling Anti-Pattern)

**What goes wrong:**
- Trading tools return errors as success strings: `return err.Error(), nil` instead of `return "", err`
- Agent receives "error message" as valid data and proceeds with execution
- Orders placed with wrong parameters, or orders fail silently while system believes success

**Why it happens:**
- Tool interface returns `(string, error)` — easy to confuse which channel carries errors
- Testing with sandbox APIs masks issues (more forgiving error responses)
- No centralized error validation layer before order submission

**Consequences:**
- **Financial loss:** Orders execute at wrong price/size
- **Missed opportunities:** Failed orders not retried or logged
- **Audit gap:** Cannot trace why trades were placed

**Warning signs:**
- Tool logs show "success" but exchange shows no order
- Agent conversation logs contain error text but continue flow
- No structured error categorization (retryable vs non-retryable)

**Prevention:**
- Enforce error-return discipline: tools MUST return `("", err)` for failures
- Add response validation layer: parse OKX `sCode`/`sMsg` fields before returning success
- Implement structured error types with `IsRetryable()` interface (already exists: `OKXWatcherRetryableError`)
- Add integration tests against OKX sandbox that verify error paths

**Phase mapping:** Foundation phase (before any live trading)

---

### 2. API Rate Limiting Violations

**What goes wrong:**
- No rate limiter on API calls → exchange bans IP/account
- Concurrent agent loops hammer endpoints (MaxIteration: 100 allows 100+ calls per task)
- Retry storms during network issues amplify the problem

**Why it happens:**
- OKX has multiple rate limit tiers (per second, per minute, per day) — easy to miss one
- Agent tool calls are dynamic — hard to predict total API calls per iteration
- Rate limiter implemented in one tool but not others (inconsistent coverage)

**Consequences:**
- **Temporary ban:** IP blocked for hours
- **Account suspension:** Repeated violations → permanent ban
- **Data starvation:** Agent runs without market data, makes stale decisions

**Warning signs:**
- OKX returns HTTP 429 or error code 6001x (rate limit errors)
- Tools have no `rate.Limiter` or inconsistent limiter config
- `context.Background()` used instead of propagated context with timeout

**Prevention:**
- Add `rate.Limiter` to ALL API tools (currently only in `OkxCandlesticksTool`)
- Use conservative limits: OKX REST API = 5-10 req/s depending on endpoint
- Implement circuit breaker: pause all trading after N consecutive rate limit errors
- Add rate limit telemetry: track calls/second per endpoint

**Current state:** Only `OkxCandlesticksTool` has rate limiting (`rate.Every(time.Second/10)`). Positions and funding rate tools are unprotected.

**Phase mapping:** Foundation phase (immediate fix required)

---

### 3. Missing Circuit Breaker for Real Trading

**What goes wrong:**
- Executor places orders indefinitely without kill-switch
- No maximum loss limit per day/session
- No validation that order size matches account balance

**Why it happens:**
- Focus on analysis agents, execution is "simple" — assumed safe
- Risk monitoring deferred to "later phase"
- No clear separation between analysis mode and execution mode

**Consequences:**
- **Catastrophic loss:** Bug or bad signal drains account
- **Cascading failures:** One bad trade triggers margin calls, forced liquidation
- **No recovery:** Cannot pause trading to investigate issues

**Warning signs:**
- No independent RiskMonitor process
- Executor agent runs at same autonomy level as analysis agents
- No daily loss limit configuration

**Prevention:**
- Implement RiskMonitor as independent goroutine (not tied to OKXWatcher schedule)
- Hard circuit breaker: pause trading if daily loss > X% or N consecutive losses
- Require explicit enable/disable for execution mode
- Pre-trade validation: check balance, position limits, daily volume

**Phase mapping:** Risk Management phase (MUST precede Executor implementation)

---

### 4. Stop-Loss/Take-Profit Implementation Errors

**What goes wrong:**
- Stop-loss orders placed at wrong price (off by decimals, wrong side)
- Take-profit orders not updated when position averages down/up
-OCO (One-Cancels-Other) orders not atomic — both SL and TP active simultaneously
- Stop-loss triggered by wick (temporary spike) instead of close

**Why it happens:**
- Confusion between mark price vs last price for liquidation triggers
- Not accounting for funding rate impact on PnL
- Position size changes (partial close) but SL/TP orders not adjusted

**Consequences:**
- **Premature exit:** Stop-loss hit by noise, price reverses immediately
- **No exit:** Take-profit never filled, position reverses to loss
- **Double exposure:** Both SL and TP active, unintended re-entry

**Warning signs:**
- SL/TP logic uses `price` field without validating against current market
- No subscription to OKX websocket for real-time price updates
- Position queries infrequent (stale data for SL calculation)

**Prevention:**
- Use OKX native OCO order type (if available) or implement atomic order pair updates
- Trigger on mark price for liquidation-sensitive positions
- Implement trailing stop-loss for trending conditions
- Update SL/TP after every partial fill or position change
- Add slippage tolerance: don't place orders too close to current price

**Phase mapping:** Execution Automation phase

---

### 5. Context Cancellation and Resource Leaks

**What goes wrong:**
- `context.Background()` used directly instead of propagated context
- Scheduler `Stop()` cancels all tasks at once — no graceful shutdown
- Database connections, HTTP clients, websocket subscriptions not closed on failure
- Global agent state persists across restarts (no cleanup)

**Why it happens:**
- Quick iteration: "just get it working" during development
- Context propagation seems optional until production issues arise
- No testing of shutdown/failure scenarios

**Consequences:**
- **Goroutine leaks:** Orphaned workers consume memory
- **Resource exhaustion:** DB connection pool depleted
- **Inconsistent state:** Task killed mid-transaction, partial writes
- **Memory bloat:** Agent conversation history accumulates

**Warning signs:**
- `context.Background()` calls in handler/service layer
- `defer cancel()` missing after `context.WithCancel`
- Global variables holding agent instances (`var _agents *AgentsModel`)
- No `defer` cleanup for resources after init failure

**Prevention:**
- Propagate context through entire call chain (scheduler → handler → agent → tool)
- Implement graceful shutdown with timeout: allow N seconds for running tasks to complete
- Use `defer` for ALL resource cleanup immediately after acquisition
- Add shutdown integration test: start app, send SIGINT, verify clean exit
- Use `sync.Once` for singleton initialization (agents, db connections)

**Current state:** `scheduler.go:162` and `chatmodel.go:13` use `context.Background()` directly. `okx_watcher_handler.go:171-228` has multiple `_ = h.cronExecutionLogRepository.Create(...)` with ignored errors.

**Phase mapping:** Foundation phase (refactor before adding execution features)

---

### 6. Global Variable and State Pollution

**What goes wrong:**
- Agents stored in global `var _agents *AgentsModel`
- No thread-safe initialization — race condition if InitAgents called concurrently
- Agent state not recoverable after restart (lost conversation history, positions)
- Config changes require full restart

**Why it happens:**
- Go singleton pattern without proper synchronization
- No requirement for state persistence identified early
- Testing single-instance only, no concurrent access scenarios

**Consequences:**
- **Nil pointer panics:** `Agents()` returns nil if called before `InitAgents()`
- **Race conditions:** Concurrent tests or HTTP requests trigger data races
- **State loss:** Restart loses all agent context, must re-initialize
- **Non-testable:** Global state hard to mock or reset between tests

**Warning signs:**
- Global `var _something *Something` pattern
- `InitSomething()` function without `sync.Once`
- No dependency injection — components reach for globals
- Agent initialization in `main()` without error handling

**Prevention:**
- Replace global with proper singleton: `sync.Once` for init, mutex for access
- Use dependency injection: pass `*ServiceContext` to all components
- Add state persistence layer: store agent conversation history, active positions
- Implement hot-reload: watch config file, recreate components without restart

**Current state:** `internal/agent/agents.go` uses `var _agents *AgentsModel` with no synchronization.

**Phase mapping:** Foundation phase (architectural refactor)

---

### 7. SQLite Contention Under Concurrent Load

**What goes wrong:**
- SQLite file locks cause write contention during concurrent task execution
- Scheduler parallelism (default: 5) exceeds SQLite write capacity
- Execution logs queue up, blocking task completion

**Why it happens:**
- SQLite chosen for simplicity ("pure Go, no CGO")
- No load testing of concurrent writes
- `SELECT FOR UPDATE SKIP LOCKED` helps but doesn't eliminate file-level locks

**Consequences:**
- **Task timeouts:** Log writes block, task execution exceeds timeout
- **Data loss:** Silent failures when log creation ignored (`_ = repo.Create(...)`)
- **Scalability wall:** Cannot increase parallelism beyond SQLite limits

**Warning signs:**
- `database is locked` errors in logs
- Task execution time spikes correlate with log volume
- Only one concurrent writer effective despite semaphore allowing more

**Prevention:**
- Add PostgreSQL/MySQL support for production (configurable via `cfg.DB.Type`)
- Batch log writes: queue in memory, flush periodically
- Use WAL mode for SQLite: `PRAGMA journal_mode=WAL`
- Consider async logging: write to channel, dedicated writer goroutine

**Current state:** `internal/config/config.go` returns error if `cfg.DB.Type != "sqlite"`. CONCERNS.md flags this as production risk.

**Phase mapping:** Scalability phase (before horizontal scaling or high-frequency trading)

---

### 8. RAG Memory Contamination

**What goes wrong:**
- Vector embeddings mix different trading sessions/strategies
- Retrieval returns stale decisions from different market conditions
- No metadata filtering: bull market patterns retrieved during bear market
- Embedding dimension too small: different scenarios map to similar vectors

**Why it happens:**
- RAG implemented as simple "store everything, retrieve top-K"
- No schema for memory categorization (by date, symbol, strategy, outcome)
- Embedding model not fine-tuned for trading domain

**Consequences:**
- **Bad analogies:** Agent recalls inappropriate past decisions
- **Confirmation bias:** Only retrieves decisions that match current bias
- **Memory bloat:** Vector DB grows unbounded, retrieval slows

**Warning signs:**
- All memories stored with same collection/schema
- No TTL or archival policy for old memories
- Retrieval returns memories from >30 days ago without recency weighting

**Prevention:**
- Add metadata filters: `symbol`, `market_regime` (bull/bear/sideways), `outcome` (win/loss)
- Implement memory decay: older memories have lower retrieval priority
- Use hybrid search: keyword + semantic + recency scoring
- Add memory curation: periodically review and archive/delete low-value memories
- Fine-tune embedding model on trading domain text (earnings calls, trading journals)

**Phase mapping:** RAG Memory phase (requires deep research on embedding strategy)

---

### 9. Credential Exposure and Secret Management

**What goes wrong:**
- API credentials committed to git in `etc/config.yaml`
- Credentials passed as plain strings, visible in memory dumps
- No credential rotation mechanism
- Logs may contain sensitive data (order IDs tied to account)

**Why it happens:**
- Development config used as template
- No distinction between dev/sandbox vs production credential handling
- Secret management seen as "deployment problem" not "code problem"

**Consequences:**
- **Account compromise:** Stolen credentials → unauthorized trades/withdrawals
- **Regulatory issues:** Trading activity traced back to compromised account
- **Loss of funds:** Attacker drains account via API

**Warning signs:**
- `config.yaml` in git history (even if removed, still in commits)
- Credentials loaded from file without encryption
- No environment variable support for sensitive fields
- Logs contain full request/response dumps

**Prevention:**
- Add `etc/config.yaml` to `.gitignore` immediately
- Keep `config.example.yaml` with placeholder values only
- Use environment variables or secret manager (Vault, AWS Secrets Manager)
- Implement credential rotation: scheduled key refresh
- Redact sensitive fields in logs: API keys, order IDs in debug dumps

**Current state:** CONCERNS.md explicitly flags `etc/config.yaml` contains actual API credentials and chat model API key.

**Phase mapping:** Security phase (immediate action required)

---

### 10. AI Agent Hallucination in Trading Decisions

**What goes wrong:**
- Agent fabricates technical indicator values or misreads tool output
- Agent invents trading rules not in instruction (e.g., "RSI > 70 means buy" instead of sell)
- Overconfidence in LLM output: assumes agent reasoning is always sound
- No validation that agent decisions match stated strategy

**Why it happens:**
- LLMs are probabilistic — even with same input, output varies
- Agent instruction (SOUL.md) may have ambiguous phrasing
- No post-decision audit: compare agent reasoning vs action

**Consequences:**
- **Random trades:** Agent acts on hallucinated signals
- **Strategy drift:** Agent gradually deviates from intended behavior
- **Unexplainable losses:** Cannot trace why bad decision was made

**Warning signs:**
- Agent tool calls don't match conversation reasoning
- Same input produces different decisions on retry
- No deterministic validation layer before execution

**Prevention:**
- Implement decision validation: check agent output against explicit rules before execution
- Add "pre-trade checklist" tool: agent must explicitly confirm conditions met
- Use lower temperature (0.1-0.3) for trading decisions
- Implement "second opinion" pattern: require two independent agent runs to agree
- Add execution audit trail: log agent reasoning, tool calls, and final decision side-by-side

**Phase mapping:** Execution phase (critical before autonomous trading)

---

## Moderate Pitfalls

### 11. Timezone and Timestamp Confusion

**What goes wrong:**
- K-line timestamps interpreted in wrong timezone
- Funding rate payment times miscalculated
- Task scheduling uses server time, user expects local time

**Prevention:**
- Standardize on UTC internally, convert only for display
- Explicitly parse timestamps with timezone: `time.Parse(time.RFC3339, ...)`
- Store timezone preference in config, use consistently

**Phase mapping:** Foundation phase

---

### 12. Precision Loss in Financial Calculations

**What goes wrong:**
- float64 rounding errors accumulate over many trades
- OKX returns prices with 8+ decimals, float64 loses precision
- PnL calculations off by small amounts, compound over time

**Prevention:**
- Use `decimal.Decimal` for all monetary values (already imported in `okx_candlesticks.go`)
- Never use `float64` for order sizes, prices, or balances
- Round only at display time, not intermediate calculations

**Phase mapping:** Foundation phase

---

### 13. Overfitting to Historical Data

**What goes wrong:**
- Backtest shows 90% win rate, live trading loses money
- Strategy parameters tuned to specific market conditions
- No out-of-sample testing

**Prevention:**
- Split data: 70% training, 30% out-of-sample validation
- Walk-forward analysis: test on rolling windows
- Add market regime detection: avoid applying bull market strategies in bear markets

**Phase mapping:** Strategy Optimization phase (requires dedicated research)

---

## Minor Pitfalls

### 14. Debug Code Left in Production

**Current state:** `okx_watcher_handler.go:190-193` has dead `if false { fmt.Printf(...) }` block.

**Prevention:**
- Use proper logger with log levels
- Remove debug code before merge, don't just comment out

**Phase mapping:** Code Quality phase

---

### 15. Ignored Database Errors

**Current state:** `okx_watcher_handler.go:171, 177, 209, 222` has `_ = h.cronExecutionLogRepository.Create(...)`.

**What goes wrong:**
- Log write failures go unnoticed
- Cannot debug issues without execution history
- Audit trail incomplete

**Prevention:**
- Log errors at minimum: `if err != nil { logger.Error(ctx, "failed to create log", err) }`
- Consider retry queue for critical logs
- Alert on persistent write failures

**Phase mapping:** Foundation phase

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Real Trading Execution | Silent order failures, No circuit breaker | Fix error handling first, add kill-switch before any live orders |
| RAG Memory Implementation | Memory contamination, stale retrieval | Design metadata schema before storing any embeddings |
| Risk Monitoring System | Independent monitoring not implemented | Run RiskMonitor as separate goroutine with own context |
| Stop-Loss/Take-Profit | Wrong price, non-atomic OCO | Use OKX native order types, validate against mark price |
| Rate Limiting Rollout | Inconsistent coverage | Audit ALL tools, add rate limiter to each |
| Agent Refactor (DeepAgent→ChatModel) | Loss of orchestration logic | Document current behavior before refactor, add integration tests |

---

## Sources

- **Codebase analysis:** Internal files reviewed 2026-03-24
- **CONCERNS.md:** Existing codebase audit (security, error handling, technical debt)
- **PROJECT.md:** Known issues list (DeepAgent misuse, global variables, error handling)
- **OKX API documentation:** Rate limits, order types, error codes (inferred from code usage)
- **Trading system best practices:** Circuit breakers, OCO orders, position management (established patterns)

**Confidence Assessment:**
- Pitfalls 1-7: HIGH confidence (directly observed in codebase)
- Pitfalls 8-10: MEDIUM confidence (inferred from architecture + domain knowledge)
- Pitfalls 11-15: HIGH confidence (specific code locations identified)
