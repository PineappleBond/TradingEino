# Project Research Summary

**Project:** TradingEino - Multi-Agent Crypto Trading Bot
**Domain:** AI-Powered Cryptocurrency Trading System
**Researched:** 2026-03-24
**Confidence:** HIGH

## Executive Summary

TradingEino is a multi-agent cryptocurrency trading system built on Cloudwego Eino framework that orchestrates specialized AI agents for market analysis, risk monitoring, and trade execution. The research confirms a well-established architecture pattern: a DeepAgent coordinator (OKXWatcher) delegates to lightweight ChatModelAgent sub-agents (technical analysis, sentiment, flow analysis), with an independent RiskMonitor layer running parallel for real-time protection.

The recommended approach prioritizes foundational safety before execution: fix error handling and rate limiting first, implement independent risk monitoring with circuit breakers, then add execution tools with configurable autonomy levels. RAG-based decision memory using Redis Stack + m3e-base embeddings enables explainable AI by storing and retrieving historical trading decisions.

Key risks center on silent order failures (tools returning errors as success strings), API rate limit violations causing account bans, and missing circuit breakers for catastrophic loss prevention. All are mitigated through enforced error-return discipline, per-tool rate limiters with circuit breakers, and independent RiskMonitor goroutine with hard loss limits that can override trading decisions.

## Key Findings

### Recommended Stack

**Existing stack (confirmed adequate):** Go 1.26.1, Cloudwego Eino 0.8.4, Gin 1.12.0, SQLite3/GORM, OpenAI-compatible LLM client, go-talib for indicators, shopspring/decimal for precision math.

**New dependencies required:**
- **Redis Stack**: Vector store for RAG memory + checkpoint storage — Eino has official Redis retriever/indexer components
- **Ollama + m3e-base**: Local embedding generation — runs on M2 Pro 32GB, no external API dependency
- **Telegram Bot API**: Real-time risk alerts — crypto industry standard for push notifications

**Critical version notes:** Eino 0.8.4+ required for RAG components; redis/go-redis/v9 has breaking changes from v8.

### Expected Features

**Must have (table stakes):**
- Order Placement/Query/Cancellation — core trading function via OKX API
- Stop-Loss/Take-Profit Orders — risk management baseline, non-negotiable
- Position Tracking + P&L Calculation — "what am I holding?" fundamental question
- Connection Health Monitoring — users need to know if bot is actually running
- Trade History/Logging — audit trail for every decision and execution

**Should have (competitive):**
- AI Decision Memory (RAG) — explainable AI, learn from past decisions
- Multi-Agent Orchestration — specialized agents for better decisions
- Trailing Stop-Loss — lock in profits as price moves favorably
- Risk Circuit Breaker — auto-halt after X% loss or anomalous behavior
- Paper Trading Mode — risk-free testing before live capital

**Defer (v2+):**
- Backtesting Engine — significant effort, separate product capability
- DCA Automation — strategy layer, not core infrastructure
- Copy Trading/Social Features — regulatory risk, out of scope
- Mobile App — explicitly out of scope per PROJECT.md

### Architecture Approach

Three-layer architecture with clear separation: **Coordination Layer** (OKXWatcher DeepAgent), **Analysis/Execution/Risk Layers** (ChatModelAgent sub-agents), and **Infrastructure Layer** (RAG, OKX API, persistence). RiskMonitor runs as independent parallel process, not tied to OKXWatcher schedule.

**Major components:**
1. **OKXWatcher (DeepAgent)** — Top-level coordinator, orchestrates sub-agents, generates trade signals
2. **Executor Agent (ChatModelAgent)** — Trade execution with configurable autonomy (Level 1-3)
3. **RiskMonitor (Independent)** — Parallel goroutine with own scheduler, can override trading decisions
4. **RAG Store** — Vector database for decision memory using Redis Stack + m3e-base embeddings
5. **CircuitBreaker** — Interrupt/Resume pattern for trade approval workflow

### Critical Pitfalls

1. **Silent Order Failures** — Tools returning `err.Error(), nil` instead of `"", err` causes agent to treat errors as success. **Prevention:** Enforce error-return discipline, add response validation for OKX `sCode`/`sMsg` fields, implement structured error types with `IsRetryable()`.

2. **API Rate Limiting Violations** — No rate limiter on API tools causes IP/account bans. **Prevention:** Add `rate.Limiter` to ALL API tools (currently only in candlesticks tool), use conservative limits (5 req/s for trade endpoints), implement circuit breaker after N consecutive rate limit errors.

3. **Missing Circuit Breaker** — Executor places orders indefinitely without kill-switch or daily loss limits. **Prevention:** Implement RiskMonitor as independent goroutine, hard circuit breaker at daily loss limits (5% warning, 10% halt), require explicit enable/disable for execution mode.

4. **Stop-Loss/Take-Profit Implementation Errors** — Wrong price calculations, non-atomic OCO orders, trigger on wicks instead of close. **Prevention:** Use OKX native `sl_tp` algo order type, trigger on mark price for liquidation-sensitive positions, update SL/TP after every partial fill.

5. **Global Variable State Pollution** — Agents stored in global `var _agents *AgentsModel` causes race conditions and state loss on restart. **Prevention:** Replace with `sync.Once` singleton, use dependency injection via ServiceContext, add state persistence for agent conversation history.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Foundation & Safety
**Rationale:** Critical pitfalls 1, 2, 5 must be fixed before any execution capability — silent failures and rate limit violations could cause immediate financial loss or account bans.
**Delivers:** Safe infrastructure for trading: proper error handling, rate limiting on all tools, context propagation, singleton pattern for agents
**Addresses:** Table stakes: Connection Health Monitoring, Error Notifications
**Avoids:** Silent order failures, API rate limiting violations, global state pollution, context cancellation leaks

### Phase 2: Analysis Layer Completion
**Rationale:** OKXWatcher needs functional sub-agents before execution can be delegated; refactoring existing agents to ChatModelAgent pattern establishes clean architecture.
**Delivers:** Multi-agent analysis: TechnoAgent, SentimentAnalyst, FlowAnalyzer all using ChatModelAgent pattern
**Uses:** Eino ChatModelAgent pattern, existing OKX data tools
**Implements:** DeepAgent coordinator with ChatModelAgent sub-agents architecture

### Phase 3: Risk Management Layer
**Rationale:** Critical pitfall 3 — RiskMonitor MUST precede Executor implementation. Independent risk monitoring with circuit breakers is non-negotiable before any live trading.
**Delivers:** Independent RiskMonitor goroutine, risk threshold configuration, Telegram alert integration, forced position reduction logic
**Addresses:** Table stakes: Stop-Loss Orders, Risk Circuit Breaker
**Avoids:** Catastrophic loss from unchecked trading, margin calls, forced liquidation

### Phase 4: Execution Automation
**Rationale:** With safety infrastructure in place (Phases 1-3), execution tools can be added with configurable autonomy levels for gradual production rollout.
**Delivers:** Executor Agent (Level 1-3 autonomy), OKX trading tools (place/cancel/get-order, close-position), circuit breaker with Interrupt/Resume
**Addresses:** Table stakes: Order Placement, Order Query/Cancel, Take-Profit Orders
**Uses:** OKX native algo order types (`sl_tp`, `trigger`) for server-side SL/TP
**Avoids:** AI agent hallucination through decision validation, pre-trade checklists

### Phase 5: RAG Decision Memory
**Rationale:** Requires Redis Stack infrastructure and embedding model setup; enables explainable AI and decision consistency but depends on trade history from Phase 4.
**Delivers:** Redis Stack deployment, Ollama + m3e-base setup, decision save/search tools, RAG integration in OKXWatcher
**Addresses:** Differentiator: AI Decision Memory (RAG)
**Avoids:** RAG memory contamination through metadata schema (symbol, market_regime, outcome filtering)

### Phase 6: Enhanced Features
**Rationale:** Core trading stack complete; enhancements like paper trading, trailing stops, and multi-symbol support add polish without blocking MVP.
**Delivers:** Paper trading mode, trailing stop-loss, multi-timeframe analysis, position sizing calculator
**Addresses:** Differentiators: Trailing Stop-Loss, Paper Trading Mode, Multi-Timeframe Analysis

### Phase Ordering Rationale

- **Safety before execution:** Phases 1-3 establish error handling, rate limiting, and risk monitoring before any order placement capability — this ordering directly addresses critical pitfalls 1-3
- **Analysis before coordination:** Sub-agents must be functional before OKXWatcher can orchestrate them effectively
- **Risk independent from execution:** RiskMonitor runs parallel to OKXWatcher chain (every 30s vs 5min) so risk checks don't depend on analysis cycle
- **RAG after trade history:** Decision memory requires trade execution infrastructure to exist first

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3 (Risk Management):** Risk threshold calibration (margin ratio limits, liquidation distance) needs validation against actual OKX account behavior and market conditions
- **Phase 5 (RAG Memory):** Embedding strategy (metadata schema, memory decay policy, hybrid search) needs research on trading domain specificity — m3e-base may need fine-tuning

Phases with standard patterns (skip research-phase):
- **Phase 1 (Foundation):** Well-documented Go patterns for error handling, rate limiting, singleton initialization
- **Phase 2 (Analysis Layer):** Established Eino ChatModelAgent pattern from existing codebase
- **Phase 4 (Execution):** OKX algo order API is well-documented, existing client support in `pkg/okex`
- **Phase 6 (Enhancements):** Standard trading features with established implementations

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Verified against existing codebase, Eino official docs, and eino-ext component availability |
| Features | HIGH | Table stakes established across Freqtrade, Hummingbot, 3Commas; differentiators validated against competitor products |
| Architecture | HIGH | Based on existing ARCHITECTURE.md + Eino DeepAgent/ChatModelAgent patterns from official documentation |
| Pitfalls | HIGH | Directly observed in codebase audit (CONCERNS.md, PROJECT.md) with specific file locations identified |

**Overall confidence:** HIGH

### Gaps to Address

- **OKX Algo Order API validation:** Confirm `sl_tp` order type supports atomic OCO execution during Phase 4 implementation — if not, fallback to client-side monitoring required
- **m3e-base embedding quality:** Test embedding model performance on trading domain text (decision records, technical analysis) — may need fine-tuning if retrieval quality is poor
- **Risk threshold calibration:** Default margin ratio limits (80%/90%) and liquidation distances (3%/2%) need validation against actual market volatility and OKX margin requirements
- **Redis Stack deployment:** Confirm Docker image runs on target deployment platform (M2 Pro for local, cloud VM for production)

## Sources

### Primary (HIGH confidence)
- **Codebase analysis:** Internal files reviewed 2026-03-24 (`internal/agent/`, `pkg/okex/`, `internal/service/`)
- **CONCERNS.md:** Existing codebase audit documenting security, error handling, technical debt
- **PROJECT.md:** Known issues list and requirements
- **Eino documentation:** `.claude/skills/eino-skill/SKILL.md`, deep agent/interrupt/RAG reference docs
- **OKX API client:** `pkg/okex/api/trading.go`, `trade_requests.go` for algo order support

### Secondary (MEDIUM confidence)
- **OKX API documentation:** Rate limits, order types, error codes (inferred from code usage)
- **Trading system patterns:** Freqtrade, Hummingbot, 3Commas feature comparison
- **Redis Stack for vectors:** Common pattern for RAG infrastructure
- **Telegram for crypto alerts:** Industry standard for trading notifications

### Tertiary (LOW confidence, needs validation)
- **m3e-base model performance:** Assumed suitable for trading domain based on Chinese optimization — needs testing
- **OKX native SL/TP atomic execution:** Requires validation during Phase 4 implementation

---
*Research completed: 2026-03-24*
*Ready for roadmap: yes*
