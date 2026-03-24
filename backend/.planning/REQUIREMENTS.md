# Requirements: TradingEino

**Defined:** 2026-03-24
**Core Value:** Automated market analysis and execution that makes data-driven trading decisions without emotional bias

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Foundation & Safety

- [ ] **FOUND-01**: All tools return errors properly (`"", err`) instead of (`err.Error(), nil`)
- [ ] **FOUND-02**: All API tools have rate.Limiter with conservative limits (5 req/s for trade endpoints)
- [x] **FOUND-03**: Agents use singleton pattern with sync.Once instead of global variables
- [x] **FOUND-04**: Context propagation throughout agent initialization and tool execution
- [x] **FOUND-05**: Graceful shutdown on application exit

### Analysis Layer

- [ ] **ANAL-01**: RiskOfficer refactored from DeepAgent to ChatModelAgent
- [ ] **ANAL-02**: SentimentAnalyst refactored from DeepAgent to ChatModelAgent
- [ ] **ANAL-03**: OKXWatcher orchestrates SubAgents via DeepAgent pattern
- [ ] **ANAL-04**: SubAgents have proper DESCRIPTION.md and SOUL.md

### Execution Layer

- [ ] **EXEC-01**: okx-place-order-tool supports limit and market order types
- [ ] **EXEC-02**: okx-cancel-order-tool cancels pending orders
- [ ] **EXEC-03**: okx-get-order-tool queries order status
- [x] **EXEC-04**: Executor Agent with Level 1 autonomy (explicit commands only)
- [x] **EXEC-05**: OKX native sl_tp algo orders for stop-loss/take-profit
- [x] **EXEC-06**: Order response validation (OKX sCode/sMsg field checks)

### RAG Memory

- [ ] **RAG-01**: Redis Stack deployed for vector storage
- [ ] **RAG-02**: Ollama + m3e-base running locally for embeddings
- [ ] **RAG-03**: okx-decision-save-tool saves decision records with metadata
- [ ] **RAG-04**: okx-decision-search-tool retrieves historical decisions by similarity
- [ ] **RAG-05**: OKXWatcher integrates RAG retrieval before generating decisions

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Risk Management

- **RISK-01**: Independent RiskMonitor goroutine (30s schedule)
- **RISK-02**: Telegram alert integration for risk notifications
- **RISK-03**: Forced position reduction on margin ratio > 90%
- **RISK-04**: Daily loss limit circuit breaker (5% warning, 10% halt)
- **RISK-05**: Consecutive loss limit (stop after N losing trades)

### Enhanced Features

- **ENH-01**: Trailing stop-loss using OKX trailing order type
- **ENH-02**: Paper trading mode with simulated execution
- **ENH-03**: Multi-timeframe analysis (1m/5m/15m/1h/4h/1d)
- **ENH-04**: Position sizing calculator based on risk percentage

### Analysis Expansion

- **TECH-01**: TechnoAgent for dedicated technical analysis
- **TECH-02**: FlowAnalyzer for order book and trade flow analysis
- **TECH-03**: Additional data tools (orderbook, trades history, open interest)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Mobile app | Web-first, embedded UI sufficient per PROJECT.md |
| Social/copy trading | Regulatory risk, not core to trading value |
| Real-time chat | High complexity, not relevant to trading core value |
| Backtesting engine | Significant effort, separate product capability |
| DCA automation | Strategy layer, not core infrastructure |
| Multi-exchange support | OKX-only focus for v1 |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | Phase 1 | Complete |
| FOUND-02 | Phase 1 | Complete |
| FOUND-03 | Phase 1 | Complete |
| FOUND-04 | Phase 1 | Complete |
| FOUND-05 | Phase 1 | Complete |
| ANAL-01 | Phase 2 | Pending |
| ANAL-02 | Phase 2 | Pending |
| ANAL-03 | Phase 2 | Pending |
| ANAL-04 | Phase 2 | Pending |
| EXEC-01 | Phase 3 | Pending |
| EXEC-02 | Phase 3 | Pending |
| EXEC-03 | Phase 3 | Pending |
| EXEC-04 | Phase 3 | Complete |
| EXEC-05 | Phase 3 | Complete |
| EXEC-06 | Phase 3 | Complete |
| RAG-01 | Phase 4 | Pending |
| RAG-02 | Phase 4 | Pending |
| RAG-03 | Phase 4 | Pending |
| RAG-04 | Phase 4 | Pending |
| RAG-05 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0 ✓
- Complete: 11/20 (Phase 1 + Phase 3 Plans 01-03 complete)

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-24 - Phase 1 complete, Phase 3 Plans 01-03 complete (EXEC-04, EXEC-05, EXEC-06)*
