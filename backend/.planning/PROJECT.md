# TradingEino

## What This Is

TradingEino is an AI-powered multi-agent cryptocurrency trading system built on the Cloudwego Eino framework. It monitors OKX exchange markets, analyzes technical indicators and sentiment, and executes trades autonomously.

## Core Value

Automated market analysis and execution that makes data-driven trading decisions without emotional bias.

## Requirements

### Validated

- ✓ K-line data + 20+ technical indicators (MACD, RSI, Bollinger Bands, KDJ, ATR) — existing via `okx-candlesticks-tool`
- ✓ Position querying with max buy/sell power — existing via `okx-get-positions-tool`
- ✓ Funding rate data for sentiment analysis — existing via `okx-get-funding-rate-tool`
- ✓ Multi-agent orchestration with DeepAgent coordinator — existing architecture

### Active

- [ ] Fix Tool error handling (return errors properly instead of as success strings)
- [ ] Refactor SubAgents from DeepAgent to ChatModelAgent (RiskOfficer, SentimentAnalyst)
- [ ] Add rate limiter (rate.Limiter) for all API tools
- [ ] Implement Executor Agent for trade execution
- [ ] Implement trading tools: place-order, cancel-order, get-order, close-position
- [ ] Add RAG vector memory for decision history (Redis Stack + m3e-base)
- [ ] Implement RiskMonitor as independent monitoring layer
- [ ] Add auto stop-loss/take-profit mechanism
- [ ] Implement circuit breaker for risk management

### Out of Scope

- Mobile app — Web-first, embedded UI is sufficient
- Real-time chat features — Not relevant to trading core value
- Video/streaming content — Storage/bandwidth cost, defer indefinitely

## Context

**Technical Environment:**
- Go 1.26.1 backend
- Gin v1.12.0 for REST API server
- Cloudwego Eino v0.8.4 for multi-agent orchestration
- SQLite3 (pure Go, no CGO) for data persistence
- Pre-built Vue frontend embedded in binary

**Current Architecture:**
```
OKXWatcher (DeepAgent - Coordinator)
├── RiskOfficer (needs refactor to ChatModelAgent)
└── SentimentAnalyst (needs refactor to ChatModelAgent)
```

**Known Issues:**
1. DeepAgent滥用 — SubAgents are also DeepAgents (should be ChatModelAgent)
2. Global variable pollution — Agents stored in global vars
3. Tool error handling wrong — Returns `err.Error(), nil` instead of `"", err`
4. No rate limiting on API calls
5. Missing resource cleanup on init failure

## Constraints

- **Tech Stack**: Must use Cloudwego Eino framework — already invested, team familiarity
- **Exchange**: OKX only — API access and sandbox available
- **Deployment**: macOS M2 Pro 32GB for development, Linux for production
- **Latency**: Sub-second response for analysis, execution can be 1-2 seconds
- **Safety**: Must have risk controls before any real trading

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| DeepAgent only for coordinator (OKXWatcher) | Avoids层级冗余，SubAgents don't need task decomposition | ✓ Approved — ADR-002 |
| Analysis/Execution separation | Clean audit trail, independent testing | ✓ Approved — ADR-001 |
| Tool atomic化 | Each tool does one thing well | ✓ Approved — ADR-003 |
| RAG with Redis Stack + m3e-base | Local Embedding on M2, no external API dependency | ✓ Approved — ADR-004 |
| Independent RiskMonitor layer | Real-time monitoring, can't depend on OKXWatcher schedule | ✓ Approved — ADR-005 |
| Executor starts at Level 1 | Only execute explicit commands, earn autonomy over time | — Pending |

---
*Last updated: 2026-03-24 after codebase mapping and context review*
