# TradingEino Roadmap

**Project:** AI-Powered Multi-Agent Crypto Trading System
**Core Value:** Automated market analysis and execution that makes data-driven trading decisions without emotional bias
**Created:** 2026-03-24

---

## Phases

- [ ] **Phase 1: Foundation & Safety** - Critical infrastructure fixes: error handling, rate limiting, singleton pattern, context propagation, graceful shutdown
- [ ] **Phase 2: Analysis Layer Completion** - Refactor SubAgents to ChatModelAgent pattern with proper documentation
- [ ] **Phase 3: Execution Automation** - OKX trading tools and Executor Agent with Level 1 autonomy
- [ ] **Phase 4: RAG Decision Memory** - Vector storage infrastructure and decision save/search integration

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
- [ ] 01-foundation-safety-01-PLAN.md — OKXError 统一错误类型 + 三个 Tool 的速率限制
- [ ] 01-foundation-safety-02-PLAN.md — sync.Once 单例模式 + 上下文传播
- [ ] 01-foundation-safety-03-PLAN.md — 优雅关闭实现（含 checkpoint 验证）

---

### Phase 2: Analysis Layer Completion

**Goal:** Multi-agent analysis chain with properly structured SubAgents

**Depends on:** Phase 1 (requires safe tool infrastructure)

**Requirements:** ANAL-01, ANAL-02, ANAL-03, ANAL-04

**Success Criteria** (what must be TRUE):
1. RiskOfficer runs as ChatModelAgent instead of DeepAgent
2. SentimentAnalyst runs as ChatModelAgent instead of DeepAgent
3. OKXWatcher orchestrates SubAgents via DeepAgent coordinator pattern
4. Each SubAgent has DESCRIPTION.md and SOUL.md documentation files

**Plans:** TBD

---

### Phase 3: Execution Automation

**Goal:** Autonomous trade execution with Level 1 autonomy (explicit commands only)

**Depends on:** Phase 1 (requires safe tool infrastructure)

**Requirements:** EXEC-01, EXEC-02, EXEC-03, EXEC-04, EXEC-05, EXEC-06

**Success Criteria** (what must be TRUE):
1. User can place limit and market orders via `okx-place-order-tool`
2. User can cancel pending orders via `okx-cancel-order-tool`
3. User can query order status via `okx-get-order-tool`
4. Executor Agent executes trades only on explicit OKXWatcher commands (Level 1 autonomy)
5. Stop-loss/take-profit orders use OKX native `sl_tp` algo order type
6. Order responses validate OKX `sCode`/`sMsg` fields (detect silent failures)

**Plans:** TBD

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

## Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Safety | 0/3 | Planned | - |
| 2. Analysis Layer Completion | 0/4 | Not started | - |
| 3. Execution Automation | 0/6 | Not started | - |
| 4. RAG Decision Memory | 0/5 | Not started | - |

---

## Requirement Coverage

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | Phase 1 | Pending |
| FOUND-02 | Phase 1 | Pending |
| FOUND-03 | Phase 1 | Pending |
| FOUND-04 | Phase 1 | Pending |
| FOUND-05 | Phase 1 | Pending |
| ANAL-01 | Phase 2 | Pending |
| ANAL-02 | Phase 2 | Pending |
| ANAL-03 | Phase 2 | Pending |
| ANAL-04 | Phase 2 | Pending |
| EXEC-01 | Phase 3 | Pending |
| EXEC-02 | Phase 3 | Pending |
| EXEC-03 | Phase 3 | Pending |
| EXEC-04 | Phase 3 | Pending |
| EXEC-05 | Phase 3 | Pending |
| EXEC-06 | Phase 3 | Pending |
| RAG-01 | Phase 4 | Pending |
| RAG-02 | Phase 4 | Pending |
| RAG-03 | Phase 4 | Pending |
| RAG-04 | Phase 4 | Pending |
| RAG-05 | Phase 4 | Pending |

**Coverage:** 20/20 v1 requirements mapped ✓

---

## Deferred to v2

| Requirement | Category | Reason |
|-------------|----------|--------|
| RISK-01 | Risk Management | Independent RiskMonitor deferred per user decision |
| RISK-02 | Risk Management | Telegram alerts deferred per user decision |
| RISK-03 | Risk Management | Forced position reduction deferred per user decision |
| RISK-04 | Risk Management | Circuit breaker deferred per user decision |

---

*Roadmap created: 2026-03-24*
*Roadmap updated: 2026-03-24 - Phase 1 plans created (3 plans in 2 waves)*
