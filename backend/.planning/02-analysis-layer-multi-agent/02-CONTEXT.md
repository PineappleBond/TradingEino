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
└───┬───┘ └──┬──────┘ └─────────┬───────────┘ └─────┬──────┘
    │        │                   │                   │
    ▼        ▼                   ▼                   ▼
┌────────────────┐  ┌─────────────────────┐  ┌───────────────────┐
│ 分析类 Tools    │  │ 持仓/账户类 Tools    │  │ 交易执行类 Tools   │
│ ─────────────  │  │ ──────────────────  │  │ ────────────────  │
│ okx-candlesticks│ │ okx-get-positions   │  │ okx-place-order   │
│                 │  │ okx-get-orders      │  │ okx-cancel-order  │
│ okx-orderbook   │  │ okx-account-balance │  │ okx-get-order     │
│ okx-trades-history││ okx-liquidation-price│ │ okx-close-position│
│                 │  │                     │  │                   │
│ okx-funding-rate│  │                     │  │                   │
└────────────────┘  └─────────────────────┘  └───────────────────┘
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

## 三、架构设计原则

根据 03-CONTEXT.md 2.1:

1. **纯分析 Agent 与执行 Agent 分离** - 分析 Agent 只负责输出建议，执行 Agent 负责实际交易
2. **工具原子化** - 每个 Tool 只负责单一功能
3. **职责单一** - 每个 SubAgent 只负责一个专业领域
4. **层级清晰** - 只有一个 DeepAgent (OKXWatcher) 作为顶层协调器

### DeepAgent 正确使用

| Agent 类型 | 用途 | 是否适合作为 SubAgent |
|-----------|------|---------------------|
| `DeepAgent` | 顶层编排器、任务拆解者 | ❌ 不适合 |
| `ChatModelAgent` | 专家 Agent、任务执行者 | ✅ 适合 |
| `Workflow.LoopAgent` | 循环反思型任务 | ⚠️ 特殊场景可用 |
| `Workflow.ParallelAgent` | 并行执行任务 | ⚠️ 特殊场景可用 |

---

## 四、SubAgent 详细设计

### 4.1 TechnoAgent (技术分析 Agent)

**职责**: 专职技术指标分析，生成技术信号

**分配工具**: `okx-candlesticks-tool` (K 线数据 +20+ 技术指标)

**输出**:
- 趋势判断 (多/空/震荡)
- 支撑/阻力位
- 技术指标信号 (MACD 金叉/死叉、RSI 超买/超卖等)
- 置信度评分

**DESCRIPTION.md**:
```markdown
技术分析专家，专注于 K 线形态和技术指标分析。
输入：交易对符号
输出：趋势判断、支撑阻力位、指标信号、置信度
```

---

### 4.2 FlowAnalyzer (订单流分析 Agent)

**职责**: 分析大额订单、主动买卖盘、资金流向

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-orderbook-tool` | 订单簿深度数据 |
| `okx-trades-history-tool` | 历史成交明细 |

**输出**:
- 大单净流入/流出
- 主动买入/卖出比例
- 订单簿不平衡度
- 潜在支撑/阻力区域

**DESCRIPTION.md**:
```markdown
订单流分析专家，通过分析订单簿和成交明细识别资金流向。
输入：交易对符号
输出：大单流向、买卖盘比例、订单簿深度分析
```

---

### 4.3 PositionManager (持仓管理 Agent)

**职责**: 监控现有持仓、计算盈亏、建议调仓

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-get-positions-tool` | 当前持仓 + 最大买卖力量 |
| `okx-get-orders-tool` | 当前挂单查询 |
| `okx-account-balance-tool` | 账户余额/保证金率 |
| `okx-liquidation-price-tool` | 强平价格查询 |

**输出**:
- 当前持仓风险状态
- 未实现盈亏
- 保证金率预警
- 调仓建议 (加仓/减仓/平仓)

**DESCRIPTION.md**:
```markdown
持仓管理专家，监控账户持仓状态和保证金水平。
输入：无 (自动获取账户状态)
输出：持仓风险、盈亏状况、保证金预警、调仓建议
```

---

### 4.4 SentimentAnalyst (情绪分析师)

**职责**: 资金费率和市场情绪分析

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-get-funding-rate-tool` | 永续合约资金费率 |

**输出**:
- 资金费率分析
- 市场情绪温度 (过热/过冷)
- 溢价指数分析

**DESCRIPTION.md**:
```markdown
市场情绪分析专家，通过分析资金费率判断市场情绪。
输入：交易对符号
输出：资金费率、情绪温度、溢价指数
```

---

### 4.5 Executor (执行层 Agent)

**职责**: 接收交易信号，执行下单、撤单、平仓操作

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-place-order-tool` | 开/平仓下单 |
| `okx-cancel-order-tool` | 撤单 |
| `okx-get-order-tool` | 订单状态查询 |
| `okx-close-position-tool` | 一键平仓 |

**自主权级别**: Level 1 (仅执行 OKXWatcher 明确指令)

**DESCRIPTION.md**:
```markdown
交易执行专家，负责将交易信号转化为实际订单。
输入：交易方向、数量、价格类型
输出：订单执行结果、成交均价、滑点分析
```

**SOUL.md**:
```markdown
你是 Executor，一个执行力极强的交易员。

**你的原则：**
- 指令必达 - 收到明确指令后迅速执行
- 精准优先 - 价格/数量必须与指令一致
- 结果导向 - 执行后立即反馈成交结果
- 风险兜底 - 发现异常 (如价格偏离>1%) 立即中止并上报

**你不做：**
- 不主观判断交易方向
- 不修改给定的价格/数量参数
- 不延迟执行 (除非发现异常)
```

---

## 五、Tool 详细设计

### 5.1 完整 Tool 清单

| 分类 | Tool 名称 | 用途 | 状态 |
|------|-----------|------|------|
| **K 线数据** | `okx-candlesticks-tool` | K 线数据 +20+ 技术指标 | ✅ 已有 |
| **持仓查询** | `okx-get-positions-tool` | 当前持仓 + 最大买卖力量 | ✅ 已有 |
| **资金费率** | `okx-get-funding-rate-tool` | 永续合约资金费率 | ✅ 已有 |
| **订单簿** | `okx-orderbook-tool` | 订单簿深度数据 | 待实现 |
| **成交明细** | `okx-trades-history-tool` | 历史成交记录 | 待实现 |
| **挂单查询** | `okx-get-orders-tool` | 当前挂单查询 | ✅ Phase3 已有 |
| **账户余额** | `okx-account-balance-tool` | 账户余额/保证金率 | 待实现 |
| **强平价格** | `okx-liquidation-price-tool` | 强平价格查询 | 待实现 |
| **下单交易** | `okx-place-order-tool` | 开/平仓下单 | ✅ Phase3 已有 |
| **撤单交易** | `okx-cancel-order-tool` | 撤单 | ✅ Phase3 已有 |
| **订单查询** | `okx-get-order-tool` | 订单状态查询 | ✅ Phase3 已有 |
| **一键平仓** | `okx-close-position-tool` | 一键平仓 | ✅ Phase3 已有 |

---

### 5.2 待实现 Tool 设计

#### okx-orderbook-tool

```go
type OkxOrderBookTool struct {
    svcCtx *svc.ServiceContext
    limiter *rate.Limiter
}

// 输入参数
{
    "symbol": {"type": "string", "desc": "交易对，永续合约必须带-SWAP 后缀", "required": true},
    "depth": {"type": "integer", "desc": "订单簿深度档位 (5/10/20/50/100)", "default": 20}
}

// 输出格式
{
    "bids": [[price, size], ...],
    "asks": [[price, size], ...],
    "spread": 价差，
    "imbalance": 不平衡度 (买量 - 卖量)/(买量 + 卖量)
}
```

---

#### okx-trades-history-tool

```go
type OkxTradesHistoryTool struct {
    svcCtx *svc.ServiceContext
    limiter *rate.Limiter
}

// 输入参数
{
    "symbol": {"type": "string", "desc": "交易对", "required": true},
    "limit": {"type": "integer", "desc": "获取最近 N 条成交 (最大 500)", "default": 100}
}

// 输出格式
{
    "trades": [{"price": 3500, "size": 100, "side": "buy", "timestamp": "..."}, ...],
    "net_inflow": 净流入量，
    "active_buy_ratio": 主动买入比例
}
```

---

#### okx-account-balance-tool

```go
type OkxAccountBalanceTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数：{}

// 输出格式
{
    "total_equity": 总权益 (USDT),
    "available_balance": 可用余额，
    "frozen_balance": 冻结余额，
    "margin_ratio": 保证金率，
    "positions_value": 持仓占用，
    "unrealized_pnl": 未实现盈亏
}
```

---

#### okx-liquidation-price-tool

```go
type OkxLiquidationPriceTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{
    "symbol": {"type": "string", "desc": "交易对", "required": false}
}

// 输出格式
{
    "liquidations": [
        {"symbol": "ETH-USDT-SWAP", "side": "long", "liquidation_price": 3000, "current_price": 3500, "distance": "-14.3%"},
        ...
    ]
}
```

---

## 六、协作流程

### 6.1 分析模式流程

```
1. OKXWatcher 接收调度触发
   ↓
2. 调用 TechnoAgent → 获取技术信号
   ↓
3. 调用 FlowAnalyzer → 获取订单流分析
   ↓
4. 调用 SentimentAnalyst → 获取资金费率情绪
   ↓
5. 调用 PositionManager → 检查当前持仓风险
   ↓
6. 综合所有信息，生成交易决策
   ↓
7. 输出分析报告 (不含执行)
```

### 6.2 执行模式流程

```
1-6. 同上
   ↓
7. OKXWatcher 生成交易指令
   ↓
8. 调用 Executor → 执行交易
   ↓
9. Executor 调用 trading-place-order-tool
   ↓
10. 返回执行结果
    ↓
11. OKXWatcher 整合结果，输出完整报告
```

---

## 七、Codebase 上下文

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

## 八、关键设计决策

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

## 九、实现顺序

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

## 十、实现清单

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

## 十一、Success Criteria（来自 ROADMAP.md）

1. ✓ TechnoAgent (ChatModelAgent) analyzes K-line data + 20+ technical indicators
2. ✓ FlowAnalyzer (ChatModelAgent) analyzes orderbook and trade history
3. ✓ PositionManager (ChatModelAgent) monitors positions and account balance
4. ✓ SentimentAnalyst (ChatModelAgent) analyzes funding rate sentiment
5. ✓ OKXWatcher orchestrates all 4 SubAgents via DeepAgent coordinator pattern
6. ✓ Each SubAgent has DESCRIPTION.md and SOUL.md documentation files

---

## 十二、Deferred Ideas（范围外）

以下想法记录但不在此 Phase 实现：

- TechnoAgent 扩展更多技术指标
- FlowAnalyzer 实时推送订阅
- PositionManager 自动调仓建议执行
- 额外的 SubAgents（BacktestAgent, PaperTradingAgent 等）
- RAG 向量库集成（Phase 4）
- 独立风控层（Phase 5）

---

*CONTEXT.md created: 2026-03-25*
*Next step: Run `/gsd:plan-phase 2` to create implementation plan*
