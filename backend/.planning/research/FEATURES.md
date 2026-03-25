# Feature Landscape

**Domain:** Crypto Trading Bot / Automated Trading System
**Researched:** 2026-03-24
**Confidence:** MEDIUM (based on established trading system patterns; recommend validation against OKX API capabilities and competitor products)

---

## Table Stakes

Features users expect. Missing = product feels incomplete or unsafe.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Order Placement (Market/Limit)** | Core trading function; cannot trade without it | Low | OKX has `place-order` API; need proper error handling |
| **Order Query/Status** | Users need to know if orders filled | Low | OKX `get-order` API; poll for status updates |
| **Order Cancellation** | Essential for changing strategy mid-trade | Low | OKX `cancel-order` API; must handle partial fills |
| **Position Tracking** | "What am I holding?" — basic question | Low | Already exists via `okx-get-positions-tool` |
| **Balance/Portfolio View** | "How much do I have?" — fundamental | Low | Need `get-balance` tool; show available vs locked |
| **Stop-Loss Orders** | Risk management baseline; users won't trade without | Medium | Can use OKX native SL orders OR custom monitor+trigger |
| **Take-Profit Orders** | Profit realization; expected alongside SL | Medium | Often paired with SL as OCO (One-Cancels-Other) |
| **Trade History/Log** | Audit trail; tax reporting; performance review | Low | Store all executions in SQLite; exportable CSV |
| **Basic P&L Tracking** | "Am I making money?" — primary metric | Low | Calculate from entry price + current price |
| **Connection Health Monitoring** | Users need to know if bot is actually running | Medium | Heartbeat checks, API rate limit tracking, reconnection logic |
| **Error Notifications** | When things break, users must know | Low | Log + alert on failed orders, API errors, disconnections |

---

## Differentiators

Features that set product apart. Not expected, but highly valued when present.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **AI Decision Memory (RAG)** | "Why did you do that?" — Explainable AI; learn from past decisions | High | Planned: Redis Stack + m3e-base; query by similarity |
| **Multi-Agent Orchestration** | Specialized agents = better decisions; audit per-agent performance | High | Already have DeepAgent; needs Executor agent added |
| **Trailing Stop-Loss** | Lock in profits as price moves favorably | Medium | Not native on most exchanges; requires custom monitoring |
| **DCA (Dollar-Cost Averaging) Automation** | Popular strategy; systematic entry/exit | Medium | Scheduled recurring orders at price intervals |
| **Position Sizing Calculator** | Risk-per-trade automation; Kelly criterion or fixed % | Low | Calculate size based on account balance + risk tolerance |
| **Multi-Timeframe Analysis** | Confirm signals across 5m/15m/1h; reduces false positives | Low-Medium | Already have indicators; extend to multiple TFs |
| **Backtesting Engine** | Test strategies on historical data before risking capital | High | Significant effort; could use existing libraries (backtrader, freqtrade) |
| **Paper Trading Mode** | Risk-free testing; essential before live trading | Medium | Simulate fills without real orders; track hypothetical P&L |
| **Risk Circuit Breaker** | Auto-halt trading after X% loss or anomalous behavior | Medium | ADR-005 mentions independent RiskMonitor layer |
| **Funding Rate Arbitrage Alerts** | Capture basis trading opportunities | Low | Already have funding rate data; add threshold alerts |
| **Strategy Templates** | Pre-built configurations (trend following, mean reversion) | Medium | Lower barrier to entry; users can modify |

---

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Martingale / Grid Trading** | High risk of ruin; attracts gamblers not traders | Focus on directional strategies with defined risk |
| **Copy Trading / Social Features** | Liability nightmare; regulatory gray area | Keep system autonomous, not a signal service |
| **Leverage Recommendations** | Users can set their own leverage; don't encourage risky behavior | Provide risk calculations, let user decide |
| **Real-Time Chat / Community** | Off-mission; cost center without trading value | Web dashboard with logs is sufficient (per PROJECT.md) |
| **Mobile App** | Explicitly out of scope per PROJECT.md | Responsive web UI; PWA if needed |
| **Prediction Markets / Gambling** | Regulatory risk; not trading | Stay focused on spot/perpetual trading |
| **Over-Optimized Backtests** | Curve-fitting gives false confidence | If building backtester, emphasize walk-forward analysis |
| **Black-Box "AI Magic"** | Users don't trust what they can't understand | Explainable decisions via RAG memory + decision logs |

---

## Feature Dependencies

```
Order Management (place/query/cancel)
  └──> Position Tracking (positions depend on order fills)
       └──> P&L Calculation (needs entry price from position)
            └──> Stop-Loss/Take-Profit (calculates levels from entry)
                 └──> Trailing Stop (depends on SL/TP infrastructure)

Decision Memory (RAG)
  └──> Requires: Trade History (stores decisions + outcomes)
  └──> Enables: Explainable AI ("why did you trade?")

Multi-Agent System
  └──> Requires: Decision History (agents need context)
  └──> Enables: Specialized analysis (RiskOfficer, SentimentAnalyst, Executor)

Risk Circuit Breaker
  └──> Requires: Position Tracking + P&L
  └──> Requires: RiskMonitor as independent layer (ADR-005)

Paper Trading Mode
  └──> Requires: Full order management stack
  └──> Requires: Simulated fill logic
```

---

## MVP Recommendation (This Milestone)

Prioritize these for production-ready trading:

1. **Order Placement Tool** — Executor agent needs `place-order` capability
2. **Order Query/Cancel Tools** — Complete order management lifecycle
3. **Stop-Loss/Take-Profit Automation** — Risk management is non-negotiable
4. **Trade History Logging** — Every decision + execution recorded
5. **Basic Error Handling + Alerts** — Know when system fails
6. **Connection Health Monitoring** — Detect API disconnections, rate limit hits

Defer to later milestones:
- **RAG Memory**: Important but can follow after execution works
- **Trailing Stop**: Enhancement on top of basic SL/TP
- **Backtesting**: Separate product capability; significant effort
- **DCA Automation**: Strategy layer; not core infrastructure
- **Paper Trading**: Should come before live trading, but requires full stack first

---

## Gaps Analysis (Current vs Target)

| Capability | Current Status | Gap |
|------------|---------------|-----|
| Market Data | Complete (K-line, indicators, funding rate) | None |
| Position Query | Complete | None |
| Order Execution | Missing | Need: place-order, cancel-order, get-order tools + Executor agent |
| Order Management | Missing | Need: track open orders, handle partial fills |
| Risk Management | Planned (ADR-005) | Need: RiskMonitor implementation, SL/TP logic |
| Decision History | Not started | Need: RAG infrastructure, decision logging |
| Audit Trail | Minimal | Need: structured trade history with reasoning |
| Alerts/Notifications | Minimal | Need: error alerts, health monitoring |

---

## Sources

- OKX API documentation (v5 trading endpoints)
- Common patterns from established bots: Freqtrade, Hummingbot, 3Commas, Cryptohopper
- Trading system design principles from algorithmic trading literature
- PROJECT.md existing requirements and ADRs

**Confidence Notes:**
- Table stakes features are well-established across trading platforms
- Differentiator categorization based on feature availability in retail trading bots
- Recommend validating specific OKX API capabilities during implementation (native SL/TP order types, OCO support)
