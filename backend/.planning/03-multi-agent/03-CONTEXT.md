# 多 Agent 架构设计需求

本文档记录了 TradingEino 项目的 Multi-Agent 架构设计、SubAgent 需求和 Tool 规划。

---

## 一、当前架构状态

### 1.1 现有 Agent 结构

```
OKXWatcher (DeepAgent - 总协调器)
├── RiskOfficer (ChatModelAgent - 风控)
│   └── 工具：okx-get-positions-tool
└── SentimentAnalyst (ChatModelAgent - 情绪)
    └── 工具：okx-get-funding-rate-tool
```

### 1.2 现有 Tool 清单

| Tool | 位置 | 用途 |
|------|------|------|
| `okx-candlesticks-tool` | ✅ 已有 | K 线数据 +20+ 技术指标 (MACD/RSI/布林带/KDJ/ATR 等) |
| `okx-get-positions-tool` | ✅ 已有 | 查询当前持仓 + 最大买卖力量 |
| `okx-get-funding-rate-tool` | ✅ 已有 | 查询永续合约资金费率 |

### 1.3 触发机制

OKXWatcher 通过定时调度触发，接收如下格式的 prompt：

```markdown
# 交易对分析请求

**当前时间**: {{.Now}} ({{.Timezone}})
**分析标的**: {{.Symbol}}

## 任务说明

请对上述交易对进行全面的市场分析，包括但不限于:
1. 当前市场状态 - 价格、成交量、涨跌幅等基础数据
2. 技术面分析 - 关键支撑/阻力位、趋势判断、技术指标
3. 资金面分析 - 资金费率、持仓量变化、多空比
4. 交易策略建议 - 入场点位、止盈止损、仓位管理
```

---

## 二、目标架构

### 2.1 架构设计原则

1. **纯分析 Agent 与执行 Agent 分离** - 分析 Agent 只负责输出建议，执行 Agent 负责实际交易
2. **工具原子化** - 每个 Tool 只负责单一功能
3. **职责单一** - 每个 SubAgent 只负责一个专业领域
4. **层级清晰** - 只有一个 DeepAgent (OKXWatcher) 作为顶层协调器

### 2.2 目标架构图

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

**Agent 与 Tool 分配表**:

| Agent | 类型 | 分配的 Tool |
|-------|------|-------------|
| **OKXWatcher** | DeepAgent (协调器) | 无 (仅调度 SubAgent) |
| **TechnoAgent** | ChatModelAgent | `okx-candlesticks-tool` |
| **SentimentAnalyst** | ChatModelAgent | `okx-get-funding-rate-tool` |
| **PositionManager** | ChatModelAgent | `okx-get-positions-tool`, `okx-get-orders-tool`, `okx-account-balance-tool`, `okx-liquidation-price-tool` |
| **Executor** | ChatModelAgent | `okx-place-order-tool`, `okx-cancel-order-tool`, `okx-get-order-tool`, `okx-close-position-tool` |
| **FlowAnalyzer** | ChatModelAgent | `okx-orderbook-tool`, `okx-trades-history-tool` |

---

## 三、SubAgent 需求

### 3.1 纯分析型 SubAgent

这类 Agent 只负责分析，不执行交易。

#### 3.1.1 TechnoAgent (技术分析 Agent)

**职责**: 专职技术指标分析，生成技术信号

**分配工具**:
- `okx-candlesticks-tool` (K 线数据 +20+ 技术指标)

**输出**:
- 趋势判断 (多/空/震荡)
- 支撑/阻力位
- 技术指标信号 (MACD 金叉/死叉、RSI 超买/超卖等)
- 置信度评分

**DESCRIPTION.md 建议**:
```markdown
技术分析专家，专注于 K 线形态和技术指标分析。
输入：交易对符号
输出：趋势判断、支撑阻力位、指标信号、置信度
```

---

#### 3.1.2 FlowAnalyzer (订单流分析 Agent)

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

**DESCRIPTION.md 建议**:
```markdown
订单流分析专家，通过分析订单簿和成交明细识别资金流向。
输入：交易对符号
输出：大单流向、买卖盘比例、订单簿深度分析
```

---

#### 3.1.3 PositionManager (持仓管理 Agent)

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

**DESCRIPTION.md 建议**:
```markdown
持仓管理专家，监控账户持仓状态和保证金水平。
输入：无 (自动获取账户状态)
输出：持仓风险、盈亏状况、保证金预警、调仓建议
```

---

#### 3.1.4 SentimentAnalyst (情绪分析师)

**职责**: 资金费率和市场情绪分析

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-get-funding-rate-tool` | 永续合约资金费率 |

**输出**:
- 资金费率分析
- 市场情绪温度 (过热/过冷)
- 溢价指数分析

**DESCRIPTION.md 建议**:
```markdown
市场情绪分析专家，通过分析资金费率判断市场情绪。
输入：交易对符号
输出：资金费率、情绪温度、溢价指数
```

---

### 3.2 执行型 SubAgent

#### 3.2.1 Executor (执行层 Agent)

**职责**: 接收交易信号，执行下单、撤单、平仓操作

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-place-order-tool` | 开/平仓下单 |
| `okx-cancel-order-tool` | 撤单 |
| `okx-get-order-tool` | 订单状态查询 |
| `okx-close-position-tool` | 一键平仓 |

**自主权级别** (可配置):
- **Level 1**: 仅执行 OKXWatcher 明确指令 (推荐起点)
- **Level 2**: 可自主判断最优下单时机/价格
- **Level 3**: 可自主止损/止盈

**DESCRIPTION.md 建议**:
```markdown
交易执行专家，负责将交易信号转化为实际订单。
输入：交易方向、数量、价格类型
输出：订单执行结果、成交均价、滑点分析
```

**SOUL.md 建议**:
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

## 四、Tool 详细设计

### 4.1 完整 Tool 清单

| 分类 | Tool 名称 | 用途 | 状态 |
|------|-----------|------|------|
| **K 线数据** | `okx-candlesticks-tool` | K 线数据 +20+ 技术指标 | ✅ 已有 |
| **持仓查询** | `okx-get-positions-tool` | 当前持仓 + 最大买卖力量 | ✅ 已有 |
| **资金费率** | `okx-get-funding-rate-tool` | 永续合约资金费率 | ✅ 已有 |
| **订单簿** | `okx-orderbook-tool` | 订单簿深度数据 | 待实现 |
| **成交明细** | `okx-trades-history-tool` | 历史成交记录 | 待实现 |
| **挂单查询** | `okx-get-orders-tool` | 当前挂单查询 | 待实现 |
| **账户余额** | `okx-account-balance-tool` | 账户余额/保证金率 | 待实现 |
| **强平价格** | `okx-liquidation-price-tool` | 强平价格查询 | 待实现 |
| **下单交易** | `okx-place-order-tool` | 开/平仓下单 | 待实现 |
| **撤单交易** | `okx-cancel-order-tool` | 撤单 | 待实现 |
| **订单查询** | `okx-get-order-tool` | 订单状态查询 | 待实现 |
| **一键平仓** | `okx-close-position-tool` | 一键平仓 | 待实现 |

---

### 4.2 数据查询类 Tool

#### 4.2.1 okx-orderbook-tool

```go
type OkxOrderBookTool struct {
    svcCtx *svc.ServiceContext
    limiter *rate.Limiter
}

// 输入参数
{
    "symbol": {
        "type": "string",
        "desc": "交易对，永续合约必须带-SWAP 后缀，如 ETH-USDT-SWAP",
        "required": true
    },
    "depth": {
        "type": "integer",
        "desc": "订单簿深度档位 (5/10/20/50/100)",
        "required": false,
        "default": 20
    }
}

// 输出格式
{
    "bids": [[price, size], ...],
    "asks": [[price, size], ...],
    "spread": 价差,
    "imbalance": 不平衡度 (买量 - 卖量)/(买量 + 卖量)
}
```

---

#### 4.2.2 okx-trades-history-tool

```go
type OkxTradesHistoryTool struct {
    svcCtx *svc.ServiceContext
    limiter *rate.Limiter
}

// 输入参数
{
    "symbol": {
        "type": "string",
        "desc": "交易对，永续合约必须带-SWAP 后缀",
        "required": true
    },
    "limit": {
        "type": "integer",
        "desc": "获取最近 N 条成交 (最大 500)",
        "required": false,
        "default": 100
    }
}

// 输出格式
{
    "trades": [
        {"price": 3500, "size": 100, "side": "buy", "timestamp": "..."},
        ...
    ],
    "net_inflow": 净流入量,
    "active_buy_ratio": 主动买入比例
}
```

---

#### 4.2.3 okx-get-orders-tool

```go
type OkxGetOrdersTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{
    "symbol": {
        "type": "string",
        "desc": "交易对，留空表示所有交易对",
        "required": false
    },
    "order_type": {
        "type": "string",
        "desc": "订单类型 (pending: 待成交, all: 所有)",
        "required": false,
        "default": "pending"
    }
}

// 输出格式
{
    "orders": [
        {
            "order_id": "...",
            "symbol": "ETH-USDT-SWAP",
            "side": "buy",
            "price": 3400,
            "size": 100,
            "filled": 50
        },
        ...
    ]
}
```

---

#### 4.2.4 okx-account-balance-tool

```go
type OkxAccountBalanceTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{} // 无需参数

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

#### 4.2.5 okx-liquidation-price-tool

```go
type OkxLiquidationPriceTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{
    "symbol": {
        "type": "string",
        "desc": "交易对",
        "required": false // 留空表示所有持仓
    }
}

// 输出格式
{
    "liquidations": [
        {
            "symbol": "ETH-USDT-SWAP",
            "side": "long",
            "liquidation_price": 3000,
            "current_price": 3500,
            "distance": "-14.3%"
        },
        ...
    ]
}
```

---

### 4.3 交易执行类 Tool

#### 4.3.1 okx-place-order-tool

```go
type TradingPlaceOrderTool struct {
    svcCtx *svc.ServiceContext
    limiter *rate.Limiter
}

// 输入参数
{
    "symbol": {
        "type": "string",
        "desc": "交易对，如 ETH-USDT-SWAP",
        "required": true
    },
    "side": {
        "type": "string",
        "desc": "方向 (buy/sell)",
        "required": true
    },
    "pos_side": {
        "type": "string",
        "desc": "持仓方向 (long/short)",
        "required": true
    },
    "size": {
        "type": "number",
        "desc": "订单数量 (张数)",
        "required": true
    },
    "price": {
        "type": "number",
        "desc": "委托价格 (限价单必填)",
        "required": false
    },
    "order_type": {
        "type": "string",
        "desc": "订单类型 (limit/market)",
        "required": true
    },
    "reduce_only": {
        "type": "boolean",
        "desc": "仅减仓 (平仓单设为 true)",
        "required": false,
        "default": false
    }
}

// 输出格式
{
    "order_id": "...",
    "status": "submitted",
    "symbol": "ETH-USDT-SWAP",
    "price": 3500,
    "size": 100
}
```

---

#### 4.3.2 okx-cancel-order-tool

```go
type TradingCancelOrderTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{
    "order_id": {
        "type": "string",
        "desc": "订单 ID",
        "required": true
    },
    "symbol": {
        "type": "string",
        "desc": "交易对",
        "required": true
    }
}

// 输出格式
{
    "order_id": "...",
    "status": "cancelled",
    "symbol": "ETH-USDT-SWAP"
}
```

---

#### 4.3.3 okx-get-order-tool

```go
type TradingGetOrderTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{
    "order_id": {
        "type": "string",
        "desc": "订单 ID",
        "required": true
    },
    "symbol": {
        "type": "string",
        "desc": "交易对",
        "required": true
    }
}

// 输出格式
{
    "order_id": "...",
    "symbol": "...",
    "status": "filled", // filled/pending/cancelled
    "filled_size": 100,
    "avg_price": 3500,
    "side": "buy",
    "pos_side": "long"
}
```

---

#### 4.3.4 okx-close-position-tool

```go
type TradingClosePositionTool struct {
    svcCtx *svc.ServiceContext
}

// 输入参数
{
    "symbol": {
        "type": "string",
        "desc": "交易对",
        "required": true
    },
    "pos_side": {
        "type": "string",
        "desc": "持仓方向 (long/short)",
        "required": true
    },
    "percentage": {
        "type": "number",
        "desc": "平仓比例 (0-100, 100 表示全部平仓)",
        "required": false,
        "default": 100
    }
}

// 输出格式
{
    "symbol": "ETH-USDT-SWAP",
    "pos_side": "long",
    "closed_size": 100,
    "avg_close_price": 3550,
    "realized_pnl": 500
}
```

---

## 五、协作流程

### 5.1 分析模式流程

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

### 5.2 执行模式流程

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

## 六、实现优先级

### P0 - 基础设施

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 重构 RiskOfficer | 从 DeepAgent 改为 ChatModelAgent | P0 |
| 错误处理修复 | 修复所有 Tool 的错误返回逻辑 | P0 |
| 添加 rate.Limiter | 为所有 API 工具添加限流器 | P0 |

### P1 - 执行层

| 任务 | 描述 | 优先级 |
|------|------|--------|
| okx-place-order-tool | 下单工具 | P1 |
| okx-cancel-order-tool | 撤单工具 | P1 |
| Executor Agent | 执行层 Agent | P1 |

### P2 - 数据层

| 任务 | 描述 | 优先级 |
|------|------|--------|
| okx-orderbook-tool | 订单簿数据 | P2 |
| okx-get-orders-tool | 挂单查询 | P2 |
| okx-account-balance-tool | 账户余额 | P2 |

### P3 - 分析层

| 任务 | 描述 | 优先级 |
|------|------|--------|
| TechnoAgent | 技术分析 Agent | P3 |
| FlowAnalyzer | 订单流分析 Agent | P3 |
| PositionManager | 持仓管理 Agent | P3 |

### P4 - 完善

| 任务 | 描述 | 优先级 |
|------|------|--------|
| okx-get-order-tool | 订单查询工具 | P4 |
| okx-close-position-tool | 一键平仓工具 | P4 |
| okx-liquidation-price-tool | 强平价格查询 | P4 |
| okx-trades-history-tool | 成交明细工具 | P4 |

---

## 七、RAG 向量库设计

### 7.1 为什么需要 RAG

**问题**: LLM 本身无记忆，无法知道之前的决策历史和执行结果。

**解决方案**: 使用 RAG (Retrieval-Augmented Generation) 向量库保存历史决策，在分析时检索相关历史记录供 LLM 参考。

### 7.2 历史决策记录结构

每条历史记录包含：

```json
{
  "id": "decision-20260324-ETH-USDT-SWAP-001",
  "timestamp": "2026-03-24T10:30:00Z",
  "symbol": "ETH-USDT-SWAP",
  "type": "analysis | decision | execution",

  // 决策输入
  "input": {
    "trigger_type": "scheduled | price_alert",
    "market_context": {
      "price": 3500,
      "change_24h": 2.5
    }
  },

  // 各 Agent 分析结果
  "analysis": {
    "techno": {
      "trend": "bullish",
      "support": 3450,
      "resistance": 3550,
      "signals": ["MACD 金叉", "RSI 55"],
      "confidence": 0.75
    },
    "sentiment": {
      "funding_rate": 0.0002,
      "sentiment": "neutral",
      "premium_index": 0.003
    },
    "position": {
      "current_position": "long 100",
      "unrealized_pnl": 500,
      "margin_ratio": 0.15
    },
    "flow": {
      "net_inflow": 1500000,
      "active_buy_ratio": 0.65,
      "imbalance": 0.3
    }
  },

  // OKXWatcher 最终决策
  "decision": {
    "action": "hold | open_long | open_short | add_position | reduce_position | close",
    "reason": "技术面看涨，资金费率中性，建议持有现有仓位",
    "confidence": 0.7,
    "risk_level": "medium"
  },

  // 执行结果 (如有)
  "execution": {
    "order_id": "12345678",
    "action": "buy",
    "size": 50,
    "price": 3500,
    "status": "filled",
    "avg_price": 3498.5,
    "fee": 0.5
  },

  // 后续追踪
  "follow_up": {
    "exit_price": 3600,
    "realized_pnl": 5000,
    "exit_reason": "止盈",
    "outcome_rating": "profitable"
  }
}
```

### 7.3 向量库技术选型

| 组件 | 选型 | 用途 |
|------|------|------|
| **向量库** | Redis Stack (RediSearch) | 存储决策记录 + 向量检索 |
| **Embedding** | m3e-base (本地运行) | 将决策内容转为向量 |
| **检索方式** | 混合检索 (向量 + 关键词) | 提高检索准确率 |
| **ChatModel** | aliyuncodingplan/qwen3.5-plus | 决策生成 (使用现有配置) |

**硬件适配** (Apple M2 Pro 32GB):
- `m3e-base` (110M 参数) - 内存占用~500MB，中文效果最优
- `m3e-large` (330M 参数) - 内存占用~1GB，更高精度
- 使用 Ollama 或 llama.cpp 本地运行，无需额外 API

**本地 Embedding 配置示例**:
```go
// 使用 Ollama 运行 m3e-base
// ollama pull m3e-base
embedder, err := ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
    Model: "m3e-base",
    ServerURL: "http://localhost:11434",
})
```

### 7.4 RAG 使用场景

#### 场景 1：决策时参考历史相似情况

```
OKXWatcher 决策流程:
1. 接收当前市场状态
2. 生成查询 Query
3. 从向量库检索相似历史决策 (Top K=5)
4. 将历史记录 + 当前状态 拼接 Prompt
5. LLM 生成新决策
```

**检索条件**:
- 相同交易对 (symbol)
- 相似技术指标 (MACD/RSI 接近)
- 相似市场情绪 (资金费率接近)
- 时间范围 (最近 30 天)

#### 场景 2：决策后评估

```
决策后评估流程:
1. 检索历史上相似决策的结果
2. 统计胜率/盈亏比
3. 生成评估报告
4. 用于优化后续决策策略
```

### 7.5 RAG 工具设计

#### 7.5.1 okx-decision-save-tool

**职责**: 将决策记录保存到向量库

**输入参数**:
```json
{
  "symbol": { "type": "string", "desc": "交易对", "required": true },
  "type": { "type": "string", "desc": "记录类型 (analysis/decision/execution)", "required": true },
  "content": { "type": "object", "desc": "决策内容", "required": true },
  "tags": { "type": "array", "desc": "标签 (如 ['MACD 金叉', '突破阻力'])", "required": false }
}
```

#### 7.5.2 okx-decision-search-tool

**职责**: 从向量库检索历史决策

**输入参数**:
```json
{
  "symbol": { "type": "string", "desc": "交易对", "required": false },
  "query": { "type": "string", "desc": "检索关键词", "required": true },
  "time_range": { "type": "object", "desc": "时间范围", "required": false },
  "top_k": { "type": "integer", "desc": "返回数量，默认 5", "required": false },
  "filters": { "type": "object", "desc": "过滤条件 (如 type=decision, outcome=profitable)", "required": false }
}
```

**输出格式**:
```json
{
  "total": 10,
  "documents": [
    {
      "id": "decision-001",
      "timestamp": "...",
      "symbol": "ETH-USDT-SWAP",
      "decision": { "action": "open_long", "reason": "..." },
      "outcome": { "realized_pnl": 5000, "rating": "profitable" },
      "similarity_score": 0.92
    }
  ]
}
```

### 7.6 决策记忆流程

```
┌─────────────────────────────────────────────────────────┐
│                    决策记忆闭环                          │
│                                                         │
│   ┌─────────┐      ┌─────────┐      ┌─────────┐        │
│   │ 决策前  │ ───► │ 决策中  │ ───► │ 决策后  │        │
│   │ 检索    │      │ 执行    │      │ 记录    │        │
│   └────┬────┘      └────┬────┘      └────┬────┘        │
│        │                │                │              │
│        │                ▼                │              │
│        │         ┌──────────────┐        │              │
│        └─────────│  RAG 向量库   │◄───────┘              │
│                  └──────────────┘                        │
└─────────────────────────────────────────────────────────┘

流程说明:
1. 决策前：检索相似历史决策，供 LLM 参考
2. 决策中：执行交易操作
3. 决策后：保存完整决策记录到向量库
```

### 7.7 实现优先级

| 任务 | 描述 | 优先级 |
|------|------|--------|
| okx-decision-save-tool | 保存决策记录 | P1 |
| okx-decision-search-tool | 检索历史决策 | P1 |
| Redis Stack 部署 | 向量库基础设施 | P1 |
| Ollama + m3e-base 配置 | 本地 Embedding 模型 | P1 |
| OKXWatcher RAG 集成 | 决策时自动检索 | P2 |
| 决策后评估工具 | 统计胜率/盈亏比 | P3 |

---

## 十一、成熟交易所需补充模块

### 11.1 风控体系 (Risk Management)

#### 11.1.1 RiskMonitor (独立风控 Agent)

**职责**: 实时监控账户风险，独立于交易决策链路

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-get-positions-tool` | 实时持仓监控 |
| `okx-account-balance-tool` | 保证金率监控 |
| `okx-liquidation-price-tool` | 强平价监控 |

**风控规则**:
| 规则 | 阈值 | 动作 |
|------|------|------|
| 保证金率 | > 80% | 警告 |
| 保证金率 | > 90% | 强制减仓 |
| 强平价距离 | < 3% | 警告 |
| 强平价距离 | < 2% | 强制平仓 |
| 单一仓位 | > 50% 总资金 | 警告 |
| 日亏损 | > 5% 总资金 | 停止交易 24h |
| 日亏损 | > 10% 总资金 | 强制平仓所有仓位 |

---

#### 11.1.2 自动止损/止盈

**方式**: 使用 OKX 条件单 (Algo Order)

**分配工具**:
| Tool | 用途 |
|------|------|
| `okx-place-algo-order-tool` | 设置条件止损/止盈单 |
| `okx-cancel-algo-order-tool` | 取消条件单 |

**止损策略**:
- 固定百分比止损 (如 -3%)
- 技术位止损 (如跌破支撑位)
- 移动止损 (Trailing Stop)

**止盈策略**:
- 固定百分比止盈 (如 +5%)
- 分批止盈 (50% 仓位在 +5%, 50% 在 +10%)
- 移动止盈 (从最高点回撤 2% 止盈)

---

### 11.2 回测与策略验证

#### 11.2.1 BacktestAgent (回测 Agent)

**职责**: 在历史数据上验证交易策略

**数据需求**:
- 历史 K 线数据 (已有 `okx-candlesticks-tool`)
- 历史资金费率数据
- 历史订单簿快照 (可选)

**输出指标**:
| 指标 | 说明 |
|------|------|
| 总收益率 | 回测期间总收益 |
| 年化收益率 | 折算年化 |
| 夏普比率 | 收益/波动比 |
| 最大回撤 | 最大亏损幅度 |
| 胜率 | 盈利交易占比 |
| 盈亏比 | 平均盈利/平均亏损 |
| 交易次数 | 总交易数 |

---

#### 11.2.2 PaperTradingAgent (纸面交易 Agent)

**职责**: 模拟真实交易，验证策略但不实际下单

**工作流程**:
1. 接收真实市场信号
2. 生成模拟交易决策
3. 记录模拟成交结果
4. 追踪模拟持仓盈亏

**用途**:
- 新策略上线前验证
- 实盘前的"热身"运行
- 对比实盘与模拟盘差异

---

### 11.3 市场数据增强

#### 11.3.1 新增数据 Tool

| Tool | 用途 | 优先级 |
|------|------|--------|
| `okx-open-interest-tool` | 持仓量 (OI) 数据 | P2 |
| `okx-long-short-ratio-tool` | 多空持仓人数比 | P2 |
| `okx-taker-volume-tool` | 主动买入/卖出成交量 | P2 |
| `okx-premium-index-tool` | 溢价指数 | P2 |
| `okx-spot-futures-basis-tool` | 现货 - 期货基差 | P3 |

---

### 11.4 运营监控

#### 11.4.1 MonitorAgent (系统监控 Agent)

**职责**: 监控系统健康状态

**监控项**:
| 监控项 | 说明 | 告警方式 |
|--------|------|----------|
| OKX API 延迟 | > 3 秒警告 | 通知 |
| OKX API 错误率 | > 5% 警告 | 通知 |
| 账户余额变化 | 异常变动 | 通知 |
| 持仓异常 | 未授权开仓 | 紧急通知 |
| 决策频率异常 | 过于频繁 | 警告 |

**通知渠道**:
- Telegram Bot
- 钉钉机器人
- 邮件

---

#### 11.4.2 日志审计

**日志类型**:
| 日志类型 | 内容 | 保存期限 |
|----------|------|----------|
| 决策日志 | 各 Agent 分析结果 + 最终决策 | 1 年 |
| 交易日志 | 订单提交/成交/撤单详情 | 1 年 |
| 风控日志 | 风控触发记录 | 2 年 |
| 系统日志 | API 调用/错误/性能 | 30 天 |

---

### 11.5 配置管理

#### 11.5.1 策略配置

```json
{
  "strategy": {
    "enabled": true,
    "symbols": ["BTC-USDT-SWAP", "ETH-USDT-SWAP"],
    "max_position_size": 1000,
    "max_daily_loss": 0.05,
    "trading_hours": {
      "start": "00:00",
      "end": "23:59"
    },
    "risk_limits": {
      "margin_ratio_max": 0.8,
      "single_position_max": 0.5,
      "stop_loss_pct": 0.03,
      "take_profit_pct": 0.05
    }
  }
}
```

---

### 11.6 完整架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                      OKXWatcher (DeepAgent)                      │
│                   总协调器 - 市场分析 + 策略生成                  │
└─────────────┬───────────────────────────────────────────────────┘
              │
    ┌─────────┼────────────┬───────────────┬──────────┬───────────┐
    │         │            │               │          │           │
┌───▼───┐ ┌──▼────┐ ┌─────▼────┐ ┌───────▼──┐ ┌────▼────┐ ┌────▼────┐
│Techno │ │Flow   │ │Sentiment │ │Position  │ │ Executor│ │Backtest │
│       │ │Analyzer│ │Analyst  │ │Manager   │ │         │ │(模拟)  │
└───────┘ └───────┘ └──────────┘ └──────────┘ └─────────┘ └─────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      独立风控层 (并行监控)                        │
│  ┌──────────────┐  ┌───────────────┐  ┌─────────────────────┐   │
│  │ RiskMonitor  │  │ AutoStopLoss  │  │ CircuitBreaker      │   │
│  │ 实时风险监控  │  │ 自动止损止盈   │  │ 熔断机制            │   │
│  └──────────────┘  └───────────────┘  └─────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      支撑服务层                                   │
│  ┌──────────────┐  ┌───────────────┐  ┌─────────────────────┐   │
│  │ RAG 向量库    │  │ 日志审计       │  │ 配置管理             │   │
│  │ (决策记忆)    │  │ + 通知告警     │  │ + 策略配置          │   │
│  └──────────────┘  └───────────────┘  └─────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

### 11.7 补充优先级

| 模块 | 任务 | 优先级 |
|------|------|--------|
| **风控** | RiskMonitor Agent | P0 |
| **风控** | 自动止损/止盈 Tool | P0 |
| **风控** | 熔断机制 | P0 |
| **数据** | okx-open-interest-tool | P2 |
| **数据** | okx-long-short-ratio-tool | P2 |
| **回测** | BacktestAgent | P3 |
| **回测** | PaperTradingAgent | P2 |
| **监控** | MonitorAgent + 通知 | P1 |
| **配置** | 策略配置管理 | P1 |

---

## 十二、待决策事项

### 12.1 Executor 自主权级别

### 12.1 Executor 自主权级别

Executor 的自主权级别需要明确：

- **Level 1**: 仅执行 OKXWatcher 明确指令 (推荐起点)
- **Level 2**: 可自主判断最优下单时机/价格
- **Level 3**: 可自主止损/止盈

**建议**: 从 Level 1 开始，随着系统稳定性逐步提升

### 12.2 风险监控独立 Agent

是否需要一个独立的 Monitor Agent，负责：
- 持续监控已执行订单的状态
- 触发条件止损/止盈
- 异常情况告警

### 12.3 日志/审计 Agent

是否需要 Audit Agent 负责：
- 记录所有 Agent 的决策过程
- 记录所有交易的执行详情
- 生成日报/周报

---

## 十三、相关文件

- [01-OVERVIEW.md](./01-OVERVIEW.md) - 多 Agent 架构概述
- [02-DEEP-AGENT.md](./02-DEEP-AGENT.md) - DeepAgent 使用规范
- [03-CONTEXT.md](./03-CONTEXT.md) - 本需求文档

---

## 十四、架构决策记录

### ADR-001: 分析与执行分离

**决策**: 将分析 Agent 与执行 Agent 分离

**理由**:
1. 职责清晰，便于审计
2. 执行 Agent 可以独立测试
3. 可以灵活切换执行策略

**状态**: 已批准

### ADR-002: 单一 DeepAgent 架构

**决策**: 整个系统只保留一个 DeepAgent (OKXWatcher)

**理由**:
1. 避免 DeepAgent 滥用导致的层级冗余
2. 子 Agent 使用 ChatModelAgent 更高效
3. 协作链条更清晰

**状态**: 已批准

### ADR-003: Tool 原子化

**决策**: 每个 Tool 只负责单一功能

**理由**:
1. 便于测试和维护
2. LLM 更容易理解单一功能工具
3. 可以灵活组合

**状态**: 已批准

### ADR-004: RAG 向量库设计

**决策**: 使用 Redis Stack + m3e-base 构建决策记忆系统

**理由**:
1. LLM 本身无记忆，无法追溯历史决策
2. m3e-base 可在 M2 Pro 32GB 本地运行，无需外部 API
3. Redis Stack 提供成熟的向量检索 + 结构化存储能力

**状态**: 已批准

### ADR-005: 独立风控层

**决策**: 风控模块独立于交易决策链路，并行运行

**理由**:
1. 风控需要实时监控，不能依赖 OKXWatcher 调度
2. 紧急情况下风控优先级高于交易决策
3. 可独立触发强制减仓/平仓操作

**状态**: 已批准

---

## 十五、待办事项清单

### 已完成
- [x] 多 Agent 架构设计
- [x] SubAgent 职责定义
- [x] Tool 详细设计 (12 个)
- [x] RAG 向量库设计
- [x] 成熟交易所需模块分析
- [x] 章节编号修正

### 待实现 (按优先级)

**P0 - 基础设施**
- [ ] 重构 RiskOfficer 从 DeepAgent 改为 ChatModelAgent
- [ ] 修复所有 Tool 的错误返回逻辑 (return "", err)
- [ ] 为所有 API 工具添加 rate.Limiter 限流器
- [ ] RiskMonitor Agent 实现
- [ ] 自动止损/止盈 Tool 实现
- [ ] 熔断机制实现

**P1 - 执行层**
- [ ] okx-place-order-tool
- [ ] okx-cancel-order-tool
- [ ] Executor Agent
- [ ] okx-decision-save-tool
- [ ] okx-decision-search-tool
- [ ] Redis Stack 部署
- [ ] Ollama + m3e-base 配置

**P2 - 数据层**
- [ ] okx-orderbook-tool
- [ ] okx-get-orders-tool
- [ ] okx-account-balance-tool
- [ ] okx-open-interest-tool
- [ ] okx-long-short-ratio-tool
- [ ] PaperTradingAgent

**P3 - 分析层**
- [ ] TechnoAgent
- [ ] FlowAnalyzer
- [ ] PositionManager
- [ ] BacktestAgent

**P4 - 完善**
- [ ] okx-get-order-tool
- [ ] okx-close-position-tool
- [ ] okx-liquidation-price-tool
- [ ] okx-trades-history-tool
- [ ] MonitorAgent + 通知渠道
- [ ] 策略配置管理
