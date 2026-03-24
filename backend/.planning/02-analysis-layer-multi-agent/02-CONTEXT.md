# Phase 2: Analysis Layer — Multi-Agent Architecture

**Phase Number**: 2
**Created**: 2026-03-25
**Status**: Ready for Planning

---

## 一、Phase 目标

根据 ROADMAP.md 和 03-CONTEXT.md，Phase 2 必须完成：

### 需求覆盖

| Requirement | 描述 | 状态 |
|-------------|------|------|
| ANAL-02 | TechnoAgent (ChatModelAgent) — K 线数据 +20+ 技术指标分析 | Pending |
| ANAL-03 | FlowAnalyzer (ChatModelAgent) — 订单簿 + 成交明细分析 | Pending |
| ANAL-04 | PositionManager (ChatModelAgent) — 持仓管理 (原 RiskOfficer) | Pending |
| ANAL-05 | OKXWatcher orchestrates SubAgents via DeepAgent pattern | Pending |
| ANAL-06 | All SubAgents have DESCRIPTION.md and SOUL.md files | Pending |

**注**：ANAL-01 SentimentAnalyst 已存在，需确认其 ChatModelAgent 配置。

---

## 二、目标架构

根据 03-CONTEXT.md:

```
┌─────────────────────────────────────────────────────────┐
│         OKXWatcher (DeepAgent - 总协调器)                │
│  触发：定时调度 / 价格异动                               │
│  职责：市场分析、策略生成、信号路由                       │
│  工具：无 (纯协调，不直接调用工具)                         │
└─────────────┬───────────────────────────────────────────┘
              │
    ┌─────────┼──────────────────────┬────────────────────────┐
    │         │                      │                        │
┌───▼───┐ ┌──▼──────┐ ┌────────────▼────────┐ ┌─────▼──────┐
│Techno │ │Sentiment│ │   PositionManager   │ │  Executor  │
│       │ │Analyst  │ │   (原 RiskOfficer)  │ │  (执行层)  │
└───────┘ └─────────┘ └─────────────────────┘ └────────────┘
```

### Agent 与 Tool 分配

| Agent | 类型 | 分配的 Tool | 工具状态 |
|-------|------|-------------|----------|
| OKXWatcher | DeepAgent (协调器) | 无 | — |
| TechnoAgent | ChatModelAgent | `okx-candlesticks-tool` | ✅ 已有 |
| SentimentAnalyst | ChatModelAgent | `okx-get-funding-rate-tool` | ✅ 已有 |
| PositionManager | ChatModelAgent | `okx-get-positions-tool`, `okx-get-orders-tool`, `okx-account-balance-tool`, `okx-liquidation-price-tool` | ⚠️ 部分需要实现 |
| FlowAnalyzer | ChatModelAgent | `okx-orderbook-tool`, `okx-trades-history-tool` | ❌ 需要实现 |
| Executor | ChatModelAgent | `okx-place-order-tool`, `okx-cancel-order-tool`, `okx-get-order-tool`, `okx-close-position-tool` | ✅ Phase 3 完成 |

---

## 三、实现顺序

按依赖关系和工具可用性排序：

### Wave 1: TechnoAgent + SentimentAnalyst

这两个 Agent 仅需已有 tools，可立即实现：

1. **TechnoAgent**
   - 工具：`okx-candlesticks-tool` ✅ 已有
   - 输出：趋势判断、支撑/阻力位、技术指标信号、置信度

2. **SentimentAnalyst**（重构确认）
   - 工具：`okx-get-funding-rate-tool` ✅ 已有
   - 输出：资金费率分析、情绪温度、溢价指数

### Wave 2: PositionManager

需要补充 2 个 tools：

- `okx-get-positions-tool` ✅ 已有
- `okx-get-orders-tool` ✅ Phase 3 已有
- `okx-account-balance-tool` ⚠️ 待实现
- `okx-liquidation-price-tool` ⚠️ 待实现

### Wave 3: FlowAnalyzer

需要实现 2 个新 tools：

- `okx-orderbook-tool` ❌ 待实现
- `okx-trades-history-tool` ❌ 待实现

### Wave 4: OKXWatcher 协调器集成

- 将 4 个 SubAgents 整合到 OKXWatcher DeepAgent
- 实现协作流程（分析模式）
- 更新触发调度逻辑

---

## 四、Codebase 上下文

### 现有 Agent 位置

```
internal/agent/
├── agents.go                  # Agent 单例初始化
├── okx_watcher/
│   ├── okx_watcher.go         # DeepAgent 协调器
│   ├── DESCRIPTION.md
│   └── SOUL.md
├── risk_officer/              # 需重构为 PositionManager
│   ├── risk_officer.go
│   ├── DESCRIPTION.md
│   └── SOUL.md
└── sentiment_analyst/
    ├── sentiment_analyst.go
    ├── DESCRIPTION.md
    └── SOUL.md
```

### 现有 Tools 位置

```
internal/agent/tools/
├── okx_candlesticks.go        # TechnoAgent 用
├── okx_get_fundingrate.go     # SentimentAnalyst 用
├── okx_get_positions.go       # PositionManager 用
├── okx_get_order.go           # Phase 3 已有
├── okx_get_order_history.go   # Phase 3 已有
├── okx_place_order.go         # Phase 3 已有
├── okx_cancel_order.go        # Phase 3 已有
├── okx_close_position.go      # Phase 3 已有
└── ...
```

### 需要新建的目录结构

```
internal/agent/
├── techno/                    # 新建
│   ├── techno.go
│   ├── DESCRIPTION.md
│   └── SOUL.md
├── flow_analyzer/             # 新建
│   ├── flow_analyzer.go
│   ├── DESCRIPTION.md
│   └── SOUL.md
└── position_manager/          # 重建（原 risk_officer）
    ├── position_manager.go
    ├── DESCRIPTION.md
    └── SOUL.md
```

---

## 五、关键设计决策

### 决策 1: SubAgent 类型选择

**决策**: 所有 SubAgents 使用 ChatModelAgent，不使用 DeepAgent

**理由**:
- 03-CONTEXT.md 2.1 架构原则：层级清晰，只有一个 DeepAgent
- ChatModelAgent 更高效，适用于单一职责的分析任务
- 避免 DeepAgent 滥用导致的层级冗余

### 决策 2: OKXWatcher 协调模式

**决策**: OKXWatcher 使用 DeepAgent，通过 SubAgents 列表整合所有分析 Agent

**配置示例**:
```go
deepAgent, err := deep.New(ctx, &deep.Config{
    Name:        "OKXWatcher",
    Description: "OKX 盯盘代理，协调技术分析、情绪分析、持仓管理",
    ChatModel:   svcCtx.ChatModel,
    SubAgents:   []adk.Agent{technoAgent, sentimentAgent, positionManager, flowAnalyzer},
    MaxIteration: 10,
})
```

### 决策 3: SubAgent 指令（Instruction）

每个 SubAgent 需要明确的 Instruction：

**TechnoAgent**:
```
你是技术分析专家，专注于 K 线形态和技术指标分析。
输入：交易对符号（如 BTC-USDT-SWAP）
输出：趋势判断（多/空/震荡）、支撑/阻力位、技术指标信号（MACD/RSI 等）、置信度评分
```

**SentimentAnalyst**:
```
你是市场情绪分析专家，通过分析资金费率判断市场情绪。
输入：交易对符号
输出：资金费率分析、情绪温度（过热/正常/过冷）、溢价指数
```

**PositionManager**:
```
你是持仓管理专家，监控账户持仓状态和保证金水平。
输入：无（自动获取账户状态）
输出：当前持仓风险、未实现盈亏、保证金率预警、调仓建议
```

**FlowAnalyzer**:
```
你是订单流分析专家，通过分析订单簿和成交明细识别资金流向。
输入：交易对符号
输出：大单净流入/流出、主动买入/卖出比例、订单簿不平衡度
```

### 决策 4: Tool 复用策略

| Tool | 来源 |
|------|------|
| okx-candlesticks-tool | 已有，复用 |
| okx-get-funding-rate-tool | 已有，复用 |
| okx-get-positions-tool | 已有，复用 |
| okx-get-orders-tool | Phase 3 已有，复用 |
| okx-account-balance-tool | 需要实现 |
| okx-liquidation-price-tool | 需要实现 |
| okx-orderbook-tool | 需要实现 |
| okx-trades-history-tool | 需要实现 |

### 决策 5: 文件名规范

- Agent 文件：`{agent_name}.go`（小写，下划线分隔）
- 文档文件：`DESCRIPTION.md`, `SOUL.md`（大写）
- 目录名：`{agent_name}/`（小写，下划线分隔）

---

## 六、实现清单

### Wave 1: TechnoAgent + SentimentAnalyst

- [ ] 创建 `internal/agent/techno/techno.go`
- [ ] 创建 `internal/agent/techno/DESCRIPTION.md`
- [ ] 创建 `internal/agent/techno/SOUL.md`
- [ ] 在 `agents.go` 中集成 TechnoAgent
- [ ] 确认 SentimentAnalyst 配置正确
- [ ] 更新 `agents.go` 中的 SentimentAnalyst（如需要）

### Wave 2: PositionManager

- [ ] 实现 `internal/agent/tools/okx_account_balance.go`
- [ ] 实现 `internal/agent/tools/okx_liquidation_price.go`
- [ ] 创建 `internal/agent/position_manager/position_manager.go`
- [ ] 创建 `internal/agent/position_manager/DESCRIPTION.md`
- [ ] 创建 `internal/agent/position_manager/SOUL.md`
- [ ] 在 `agents.go` 中集成 PositionManager
- [ ] 迁移/删除原 `internal/agent/risk_officer/`

### Wave 3: FlowAnalyzer

- [ ] 实现 `internal/agent/tools/okx_orderbook.go`
- [ ] 实现 `internal/agent/tools/okx_trades_history.go`
- [ ] 创建 `internal/agent/flow_analyzer/flow_analyzer.go`
- [ ] 创建 `internal/agent/flow_analyzer/DESCRIPTION.md`
- [ ] 创建 `internal/agent/flow_analyzer/SOUL.md`
- [ ] 在 `agents.go` 中集成 FlowAnalyzer

### Wave 4: OKXWatcher 协调器

- [ ] 更新 `internal/agent/okx_watcher/okx_watcher.go` 整合 4 个 SubAgents
- [ ] 更新 `internal/agent/okx_watcher/DESCRIPTION.md`
- [ ] 更新 `internal/agent/okx_watcher/SOUL.md`
- [ ] 更新 `internal/agent/agents.go` 中的 OKXWatcher 初始化

---

## 七、Success Criteria（来自 ROADMAP.md）

1. ✓ TechnoAgent (ChatModelAgent) analyzes K-line data + 20+ technical indicators
2. ✓ FlowAnalyzer (ChatModelAgent) analyzes orderbook and trade history
3. ✓ PositionManager (ChatModelAgent) monitors positions and account balance
4. ✓ SentimentAnalyst (ChatModelAgent) analyzes funding rate sentiment
5. ✓ OKXWatcher orchestrates all 4 SubAgents via DeepAgent coordinator pattern
6. ✓ Each SubAgent has DESCRIPTION.md and SOUL.md documentation files

---

## 八、Deferred Ideas（范围外）

以下想法记录但不在此 Phase 实现：

- TechnoAgent 扩展更多技术指标
- FlowAnalyzer 实时推送订阅
- PositionManager 自动调仓建议执行
- 额外的 SubAgents（BacktestAgent, PaperTradingAgent 等）

---

*CONTEXT.md created: 2026-03-25*
*Next step: Run `/gsd:plan-phase 2` to create implementation plan*
