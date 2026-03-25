# Requirements: TradingEino

**Defined:** 2026-03-24
**Updated:** 2026-03-24 — 根据 03-CONTEXT.md 多 Agent 架构设计调整
**Core Value:** Automated market analysis and execution that makes data-driven trading decisions without emotional bias

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Foundation & Safety

- [x] **FOUND-01**: All tools return errors properly (`"", err`) instead of (`err.Error(), nil`)
- [x] **FOUND-02**: All API tools have rate.Limiter with conservative limits (5 req/s for trade endpoints)
- [x] **FOUND-03**: Agents use singleton pattern with sync.Once instead of global variables
- [x] **FOUND-04**: Context propagation throughout agent initialization and tool execution
- [x] **FOUND-05**: Graceful shutdown on application exit

### Analysis Layer — SubAgents

目标架构：OKXWatcher (DeepAgent) 协调 4 个分析型 ChatModelAgent

- [x] **ANAL-01**: SentimentAnalyst (ChatModelAgent) — 资金费率情绪分析
- [x] **ANAL-02**: TechnoAgent (ChatModelAgent) — K 线数据 +20+ 技术指标分析
- [ ] **ANAL-03**: FlowAnalyzer (ChatModelAgent) — 订单簿 + 成交明细分析
- [ ] **ANAL-04**: PositionManager (ChatModelAgent) — 持仓管理 (原 RiskOfficer)
- [ ] **ANAL-05**: OKXWatcher orchestrates SubAgents via DeepAgent pattern
- [x] **ANAL-06**: All SubAgents have DESCRIPTION.md and SOUL.md files

### Execution Layer

- [x] **EXEC-01**: okx-place-order-tool — 支持限价/市价单
- [x] **EXEC-02**: okx-cancel-order-tool — 撤销待成交订单
- [x] **EXEC-03**: okx-get-order-tool — 查询订单状态
- [x] **EXEC-04**: okx-close-position-tool — 一键平仓
- [x] **EXEC-05**: Executor Agent (ChatModelAgent) — Level 1 自主权（仅执行明确指令）
- [x] **EXEC-06**: OKX native sl_tp algo orders — 止盈止损条件单
- [x] **EXEC-07**: Order response validation — OKX sCode/sMsg 字段验证

### Data Tools — Account & Position

- [x] **DATA-01**: okx-get-positions-tool — 查询持仓 + 最大买卖力量
- [ ] **DATA-02**: okx-get-orders-tool — 查询当前挂单
- [ ] **DATA-03**: okx-account-balance-tool — 账户余额/保证金率
- [ ] **DATA-04**: okx-liquidation-price-tool — 强平价格查询

### Data Tools — Market Analysis

- [x] **MKT-01**: okx-candlesticks-tool — K 线数据 +20+ 技术指标
- [ ] **MKT-02**: okx-orderbook-tool — 订单簿深度数据
- [ ] **MKT-03**: okx-trades-history-tool — 历史成交明细
- [x] **MKT-04**: okx-get-funding-rate-tool — 资金费率查询

### RAG 记忆

- [ ] **RAG-01**: Redis Stack deployed for vector storage
- [ ] **RAG-02**: Ollama + m3e-base running locally for embeddings
- [ ] **RAG-03**: okx-decision-save-tool saves decision records with metadata
- [ ] **RAG-04**: okx-decision-search-tool retrieves historical decisions by similarity
- [ ] **RAG-05**: OKXWatcher integrates RAG retrieval before generating decisions

### Risk Management (Independent Layer)

独立风控层，并行于交易决策链路运行

- [ ] **RISK-01**: RiskMonitor Agent — 实时风险监控（30s 周期）
- [ ] **RISK-02**: Auto stop-loss/take-profit — 自动止损止盈条件单
- [ ] **RISK-03**: Circuit breaker — 熔断机制（日亏损 5% 警告，10% 停止）

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Enhanced Features

- **ENH-01**: Trailing stop-loss using OKX trailing order type
- **ENH-02**: Paper trading mode with simulated execution
- **ENH-03**: Multi-timeframe analysis (1m/5m/15m/1h/4h/1d)
- **ENH-04**: Position sizing calculator based on risk percentage
- **ENH-05**: Consecutive loss limit (stop after N losing trades)

### Data Tools Expansion

- **DATA-05**: okx-open-interest-tool — 持仓量 (OI) 数据
- **DATA-06**: okx-long-short-ratio-tool — 多空持仓人数比
- **DATA-07**: okx-taker-volume-tool — 主动买入/卖出成交量
- **DATA-08**: okx-premium-index-tool — 溢价指数
- **DATA-09**: okx-spot-futures-basis-tool — 现货 - 期货基差

### Analysis Expansion

- **TECH-01**: BacktestAgent — 历史数据回测验证策略
- **TECH-02**: PaperTradingAgent — 模拟交易验证

### System Monitoring

- **MON-01**: MonitorAgent — 系统健康状态监控
- **MON-02**: Telegram/钉钉 notification integration
- **MON-03**: Audit logging (决策日志/交易日志/风控日志)

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

**Coverage:**
- v1 requirements: 33 total
- Mapped to phases: 33
- Unmapped: 0 ✓
- Complete: 18/33 (Phase 1 + Phase 3 complete)

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-24 — 根据 03-CONTEXT.md 多 Agent 架构设计调整*
