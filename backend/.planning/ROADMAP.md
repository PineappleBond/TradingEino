# TradingEino Roadmap

**Project:** AI-Powered Multi-Agent Crypto Trading System
**Core Value:** Automated market analysis and execution that makes data-driven trading decisions without emotional bias
**Created:** 2026-03-24

---

## Phases

- [x] **Phase 1: Foundation & Safety** - Critical infrastructure fixes: error handling, rate limiting, singleton pattern, context propagation, graceful shutdown
- [ ] **Phase 2: Analysis Layer — Multi-Agent Architecture** - Implement 4 SubAgents (Techno, FlowAnalyzer, PositionManager, Sentiment) with DeepAgent coordination
- [x] **Phase 3: Execution Automation** - OKX trading tools and Executor Agent with Level 1 autonomy
- [ ] **Phase 4: RAG Decision Memory** - Vector storage infrastructure and decision save/search integration
- [ ] **Phase 5: Risk Management Layer** - Independent risk monitoring, auto stop-loss/take-profit, circuit breaker

---

## Phase Details

### Phase 1: Foundation & Safety

**Goal:** System operates safely with proper error handling, rate limiting, and resource management

**Depends on:** Nothing (foundation phase)

**Requirements:** FOUND-01, FOUND-02, FOUND-03, FOUND-04, FOUND-05

**Success Criteria** (what must be TRUE):
1. All API tools return errors properly (`"", err`) instead of masking as success (`err.Error(), nil`)
2. All API tools have `rate.Limiter` with conservative limits (5 req/s for trade endpoints)
3. Agents initialize via singleton pattern with `sync.Once` instead of global variables
4. Context is propagated throughout agent initialization and tool execution (cancellation works end-to-end)
5. Application shuts down gracefully on exit (resources cleaned up, no goroutine leaks)

**Plans:** 3 plans

**Plans:**
- [x] 01-foundation-safety-01-PLAN.md — OKXError 统一错误类型 + 三个 Tool 的速率限制
- [x] 01-foundation-safety-02-PLAN.md — sync.Once 单例模式 + 上下文传播
- [x] 01-foundation-safety-03-PLAN.md — 优雅关闭实现（含 checkpoint 验证）

---

### Phase 2: Analysis Layer — Multi-Agent Architecture

**Goal:** Implement target architecture with 4 specialized SubAgents coordinated by OKXWatcher

**Depends on:** Phase 1 (requires safe tool infrastructure)

**Requirements:** ANAL-02, ANAL-03, ANAL-04, ANAL-05, ANAL-06

**Success Criteria** (what must be TRUE):
1. TechnoAgent (ChatModelAgent) analyzes K-line data + 20+ technical indicators
2. FlowAnalyzer (ChatModelAgent) analyzes orderbook and trade history
3. PositionManager (ChatModelAgent) monitors positions and account balance
4. SentimentAnalyst (ChatModelAgent) analyzes funding rate sentiment
5. OKXWatcher orchestrates all 4 SubAgents via DeepAgent coordinator pattern
6. Each SubAgent has DESCRIPTION.md and SOUL.md documentation files

**Plans:** TBD

---

### Phase 3: Execution Automation

**Goal:** Autonomous trade execution with Level 1 autonomy (explicit commands only)

**Depends on:** Phase 1 (requires safe tool infrastructure)

**Requirements:** EXEC-01, EXEC-02, EXEC-03, EXEC-04, EXEC-05, EXEC-06, EXEC-07, DATA-01, DATA-02, DATA-03, DATA-04, MKT-01, MKT-02, MKT-03, MKT-04

**Success Criteria** (what must be TRUE):
1. User can place limit and market orders via `okx-place-order-tool`
2. User can cancel pending orders via `okx-cancel-order-tool`
3. User can query order status via `okx-get-order-tool`
4. User can close positions via `okx-close-position-tool`
5. Executor Agent executes trades only on explicit OKXWatcher commands (Level 1 autonomy)
6. Stop-loss/take-profit orders use OKX native `sl_tp` algo order type
7. Order responses validate OKX `sCode`/`sMsg` fields (detect silent failures)
8. Market analysis tools available: orderbook, trades history, funding rate

**Plans:** 4 plans

**Plans:**
- [x] 03-execution-automation-01-PLAN.md — P0 核心订单工具（下单、撤单、查询）
- [x] 03-execution-automation-02-PLAN.md — P0 止盈止损工具（附加 SL/TP、下单带 SL/TP）
- [x] 03-execution-automation-03-PLAN.md — Executor Agent（Level 1 自主性）
- [x] 03-execution-automation-04-PLAN.md — P1 批量操作工具（批量下单、批量撤单、历史查询、平仓）

---

### Phase 4: RAG Decision Memory

**Goal:** Decision history storage and retrieval for explainable AI

**Depends on:** Phase 3 (requires trade execution infrastructure for decision records)

**Requirements:** RAG-01, RAG-02, RAG-03, RAG-04, RAG-05

**Success Criteria** (what must be TRUE):
1. Redis Stack deployed and running for vector storage
2. Ollama + m3e-base running locally for embedding generation
3. Decision records saved with metadata via `okx-decision-save-tool`
4. Historical decisions retrieved by similarity via `okx-decision-search-tool`
5. OKXWatcher integrates RAG retrieval before generating decisions

**Plans:** TBD

---

### Phase 5: Risk Management Layer

**Goal:** Independent risk monitoring and protection mechanisms

**Depends on:** Phase 3 (requires execution infrastructure for forced reduction)

**Requirements:** RISK-01, RISK-02, RISK-03

**Success Criteria** (what must be TRUE):
1. RiskMonitor Agent runs independently on 30s schedule
2. Auto stop-loss/take-profit triggers on price conditions
3. Circuit breaker halts trading on daily loss limit (5% warning, 10% stop)
4. Margin ratio monitoring with forced reduction > 90%

**Plans:** TBD

---

## Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Safety | 3/3 | Complete | 01-foundation-safety-01, 01-foundation-safety-02, 01-foundation-safety-03 |
| 2. Analysis Layer — Multi-Agent | 0/6 | Not started | - |
| 3. Execution Automation | 4/4 | Complete | 03-01, 03-02, 03-03, 03-04 |
| 4. RAG Decision Memory | 0/5 | Not started | - |
| 5. Risk Management Layer | 0/3 | Not started | - |

---

## Requirement Coverage

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | Phase 1 | Complete |
| FOUND-02 | Phase 1 | Complete |
| FOUND-03 | Phase 1 | Complete |
| FOUND-04 | Phase 1 | Complete |
| FOUND-05 | Phase 1 | Complete |
| ANAL-01 | Phase 2 | Complete |
| ANAL-02 | Phase 2 | Pending |
| ANAL-03 | Phase 2 | Pending |
| ANAL-04 | Phase 2 | Pending |
| ANAL-05 | Phase 2 | Pending |
| ANAL-06 | Phase 2 | Pending |
| EXEC-01 | Phase 3 | Complete |
| EXEC-02 | Phase 3 | Complete |
| EXEC-03 | Phase 3 | Complete |
| EXEC-04 | Phase 3 | Complete |
| EXEC-05 | Phase 3 | Complete |
| EXEC-06 | Phase 3 | Complete |
| EXEC-07 | Phase 3 | Complete |
| DATA-01 | Phase 3 | Complete |
| DATA-02 | Phase 3 | Pending |
| DATA-03 | Phase 3 | Pending |
| DATA-04 | Phase 3 | Pending |
| MKT-01 | Phase 3 | Complete |
| MKT-02 | Phase 3 | Pending |
| MKT-03 | Phase 3 | Pending |
| MKT-04 | Phase 3 | Complete |
| RAG-01 | Phase 4 | Pending |
| RAG-02 | Phase 4 | Pending |
| RAG-03 | Phase 4 | Pending |
| RAG-04 | Phase 4 | Pending |
| RAG-05 | Phase 4 | Pending |
| RISK-01 | Phase 5 | Pending |
| RISK-02 | Phase 5 | Pending |
| RISK-03 | Phase 5 | Pending |

**Coverage:** 33/33 v1 requirements mapped ✓

---

## Deferred to v2

| Requirement | Category | Reason |
|-------------|----------|--------|
| ENH-01 ~ ENH-05 | Enhanced Features | Deferred to v2+ |
| DATA-05 ~ DATA-09 | Data Tools Expansion | Deferred to v2+ |
| TECH-01 ~ TECH-02 | Analysis Expansion | Deferred to v2+ |
| MON-01 ~ MON-03 | System Monitoring | Deferred to v2+ |

---

*Roadmap created: 2026-03-24*
*Roadmap updated: 2026-03-24 - Phase 1 plans created (3 plans in 2 waves)*
*Roadmap updated: 2026-03-24 - Phase 1 complete (3/3 plans)*
*Roadmap updated: 2026-03-24 - Phase 3 plans created (4 plans in 3 waves)*
*Roadmap updated: 2026-03-24 - Phase 3 complete (4/4 plans)*
*Roadmap updated: 2026-03-24 — 根据 03-CONTEXT.md 多 Agent 架构设计调整，新增 Phase 5 风控层*
