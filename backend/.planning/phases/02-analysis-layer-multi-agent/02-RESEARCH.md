# Phase 2: Analysis Layer - Multi-Agent Architecture - Research

**Researched:** 2026-03-25
**Domain:** Cloudwego Eino Multi-Agent Architecture, OKX Trading Data Analysis
**Confidence:** HIGH

## Summary

Phase 2 implements the target multi-agent analysis architecture with 4 specialized SubAgents coordinated by OKXWatcher (DeepAgent). The phase transforms the current single-agent design into a hierarchical collaboration system where:

1. **OKXWatcher** serves as the DeepAgent orchestrator with its own tools and coordination logic
2. **4 ChatModelAgent SubAgents** each specialize in a distinct analysis domain
3. **Agent-to-Agent collaboration** follows the established DeepAgent pattern from Eino ADK

The architecture is partially implemented: SentimentAnalyst (ANAL-01) is complete, and the existing RiskOfficer pattern provides a reference template for new SubAgents.

**Primary recommendation:** Follow the established ChatModelAgent pattern (seen in `sentiment_analyst/agent.go` and `risk_officer/agent.go`) to implement TechnoAgent, FlowAnalyzer, and PositionManager with their domain-specific tools and SOUL/DESCRIPTION files.

## User Constraints (from CONTEXT.md)

### Locked Decisions

*No CONTEXT.md exists for this phase - research covers full domain scope*

### Claude's Discretion

*No CONTEXT.md exists - all architecture decisions within this research are recommendations*

### Deferred Ideas (OUT OF SCOPE)

*No CONTEXT.md exists*

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Cloudwego Eino ADK | v0.8.4 (per CLAUDE.md) | Multi-Agent framework | Project standard, already integrated |
| Cloudwego Eino Components | v0.8.4 | ChatModel, Tool, Graph components | Native Eino ecosystem |
| go-talib | Latest (via `github.com/markcheno/go-talib`) | 20+ technical indicators | Already used in `okx_candlesticks.go` |
| golang.org/x/time/rate | Latest | Rate limiting for API tools | Project standard pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/shopspring/decimal` | Latest | Precise decimal calculations | Financial data handling |
| `encoding/json` | Standard | JSON parsing for tool args | All tool parameter parsing |
| `time.RFC3339` | Standard | Time formatting | Timestamp outputs |

### Installation
```bash
# Core dependencies already in go.mod per CLAUDE.md
go get github.com/cloudwego/eino@v0.8.4
go get github.com/cloudwego/eino-ext/components/model/openai@latest
go get golang.org/x/time/rate
go get github.com/markcheno/go-talib
go get github.com/shopspring/decimal
```

## Architecture Patterns

### Target Multi-Agent Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    OKXWatcher (DeepAgent)                    │
│  - ChatModel: svcCtx.ChatModel                               │
│  - Tools: okx-candlesticks-tool                              │
│  - SubAgents: TechnoAgent, FlowAnalyzer, PositionManager,    │
│              SentimentAnalyst                                │
│  - MaxIteration: 100                                         │
└─────────────┬─────────────────────────────────────────────────┘
              │ Coordinates via ReAct pattern
    ┌─────────┼──────────────┬──────────────────┬──────────────┐
    │         │              │                  │              │
┌───▼────┐ ┌──▼────────┐ ┌──▼──────────┐ ┌────▼────────┐     │
│Techno  │ │Flow       │ │Position     │ │Sentiment    │     │
│Agent   │ │Analyzer   │ │Manager      │ │Analyst      │     │
│(CMA)   │ │(CMA)      │ │(CMA)        │ │(CMA)        │     │
└────────┘ └───────────┘ └─────────────┘ └─────────────┘     │
```

**CMA = ChatModelAgent**

### Reference: Current OKXWatcher Implementation

Source: `internal/agent/okx_watcher/agent.go`

```go
func NewOkxWatcherAgent(ctx context.Context, svcCtx *svc.ServiceContext, subAgents ...adk.Agent) (*OkxWatcherAgent, error) {
    baseTools := []tool.BaseTool{
        tools.NewOkxCandlesticksTool(svcCtx),
    }

    agent, err := deep.New(ctx, &deep.Config{
        Name:        "OKXWatcher",
        Description: DESCRIPTION,
        ChatModel:   svcCtx.ChatModel,
        Instruction: SOUL,
        SubAgents:   subAgents,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: baseTools,
            },
            EmitInternalEvents: true,
        },
        MaxIteration: 100,
    })
    // ...
}
```

### Reference: SubAgent Pattern (ChatModelAgent)

Source: `internal/agent/sentiment_analyst/agent.go` and `internal/agent/risk_officer/agent.go`

```go
func NewSentimentAnalystAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*SentimentAnalystAgent, error) {
    baseTools := []tool.BaseTool{
        tools.NewOkxGetFundingRateTool(svcCtx),
    }

    agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "SentimentAnalyst",
        Description: DESCRIPTION,
        Model:       svcCtx.ChatModel,
        Instruction: SOUL,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: baseTools,
            },
            EmitInternalEvents: true,
        },
        MaxIterations: 100,
    })
    // ...
}
```

### Required Project Structure for Phase 2

```
internal/agent/
├── agents.go                          # Update: add new SubAgents to AgentsModel
├── okx_watcher/
│   ├── agent.go                       # Exists - update SubAgents list
│   ├── DESCRIPTION.md                 # Exists
│   └── SOUL.md                        # Exists
├── sentiment_analyst/                 # COMPLETE (ANAL-01)
│   ├── agent.go                       # Exists
│   ├── DESCRIPTION.md                 # Exists
│   └── SOUL.md                        # Exists
├── techno_agent/                      # NEW (ANAL-02)
│   ├── agent.go                       # Create
│   ├── DESCRIPTION.md                 # Create
│   └── SOUL.md                        # Create
├── flow_analyzer/                     # NEW (ANAL-03)
│   ├── agent.go                       # Create
│   ├── DESCRIPTION.md                 # Create
│   └── SOUL.md                        # Create
└── position_manager/                  # NEW (ANAL-04, rename from risk_officer)
    ├── agent.go                       # Create (or rename existing)
    ├── DESCRIPTION.md                 # Exists as risk_officer
    └── SOUL.md                        # Exists as risk_officer
```

### Pattern: Embedded DESCRIPTION and SOUL

All SubAgents use Go embed for personality files:

```go
//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Technical indicators | Custom MACD/RSI/Bollinger implementation | `github.com/markcheno/go-talib` | Industry-standard TA-Lib bindings, 150+ indicators |
| Rate limiting | Manual time.Sleep or counters | `golang.org/x/time/rate` | Token bucket algorithm, battle-tested |
| Agent coordination | Custom orchestration logic | Eino DeepAgent | Built-in ReAct, iteration management |
| Decimal precision | float64 for prices | `github.com/shopspring/decimal` | No floating-point errors |
| JSON parsing | Manual string parsing | `encoding/json` with struct tags | Type-safe, handles edge cases |

**Key insight:** The project already uses these libraries correctly. Phase 2 continues established patterns.

## Common Pitfalls

### Pitfall 1: DeepAgent vs ChatModelAgent Confusion
**What goes wrong:** Implementing SubAgents as DeepAgents creates unnecessary hierarchy levels
**Why it happens:** DeepAgent sounds "more powerful"
**How to avoid:** Only OKXWatcher (coordinator) is DeepAgent. All SubAgents are ChatModelAgents.
**Warning signs:** SubAgent importing `github.com/cloudwego/eino/adk/prebuilt/deep`

### Pitfall 2: Tool Rate Limiter Missing
**What goes wrong:** Tools without rate limiters cause API throttling
**Why it happens:** Forgetting to add `limiter *rate.Limiter` field
**How to avoid:** Every tool MUST have a limiter. Use pattern: `rate.NewLimiter(rate.Every(200*time.Millisecond), 1)` for 5 req/s
**Warning signs:** Tool constructor without `limiter` field initialization

### Pitfall 3: OKX Error Code Not Checked
**What goes wrong:** Only checking `error` return, not `Code` field in response
**Why it happens:** OKX returns HTTP 200 even for API errors
**How to avoid:** Always check `result.Code.Int() != 0` after API calls
**Warning signs:** Tool code without `if result.Code.Int() != 0` check

### Pitfall 4: Context Not Propagated
**What goes wrong:** Using `context.Background()` instead of passed context
**Why it happens:** Habit from simpler applications
**How to avoid:** Always use the `ctx` parameter from function signature
**Warning signs:** `context.Background()` in agent/tool code

### Pitfall 5: Agent Initialization Race Condition
**What goes wrong:** Multiple goroutines initializing agents simultaneously
**Why it happens:** Not using sync.Once for singleton
**How to avoid:** Keep existing `agentsOnce.Do()` pattern in `agents.go`
**Warning signs:** Direct global variable assignment without sync.Once

## Code Examples

### New SubAgent Template (TechnoAgent Example)

```go
package techno_agent

import (
    "context"
    _ "embed"
    "github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
    "github.com/PineappleBond/TradingEino/backend/internal/svc"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/compose"
)

// TechnoAgent - Technical Analysis Agent (ChatModelAgent)
type TechnoAgent struct {
    agent adk.Agent
}

func NewTechnoAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*TechnoAgent, error) {
    baseTools := []tool.BaseTool{
        tools.NewOkxCandlesticksTool(svcCtx),
        // Add: tools.NewOkxOrderbookTool(svcCtx), // MKT-02
        // Add: tools.NewOkxTradesHistoryTool(svcCtx), // MKT-03
    }

    agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "TechnoAgent",
        Description: DESCRIPTION,
        Model:       svcCtx.ChatModel,
        Instruction: SOUL,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: baseTools,
            },
            EmitInternalEvents: true,
        },
        MaxIterations: 100,
    })
    if err != nil {
        return nil, err
    }

    return &TechnoAgent{agent: agent}, nil
}

func (t *TechnoAgent) Agent() adk.Agent {
    return t.agent
}

//go:embed DESCRIPTION.md
var DESCRIPTION string

//go:embed SOUL.md
var SOUL string
```

### Tool Template (Orderbook Tool Example - MKT-02)

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/PineappleBond/TradingEino/backend/internal/svc"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex"
    requests_market "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/schema"
    "golang.org/x/time/rate"
)

type OkxOrderbookTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter
}

func (c *OkxOrderbookTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name:  "okx-orderbook-tool",
        Desc:  "获取订单簿深度数据",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "symbol": {Type: schema.String, Desc: "交易对", Required: true},
            "depth":  {Type: schema.String, Desc: "深度档位：1/5/10/20/40/50/100/200/400", Required: false},
        }),
    }, nil
}

func (c *OkxOrderbookTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    var request struct {
        Symbol string `json:"symbol"`
        Depth  string `json:"depth"`
    }
    json.Unmarshal([]byte(argsJSON), &request)

    if err := c.limiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limiter wait failed: %w", err)
    }

    // Source: https://www.okex.com/docs-v5/en/#rest-api-market-data-get-order-book
    orderBook, err := c.svcCtx.OKXClient.Rest.Market.GetOrderBook(requests_market.GetOrderBook{
        InstID: request.Symbol,
        Sz:     request.Depth,
    })
    if err != nil {
        return "", err
    }
    if orderBook.Code.Int() != 0 {
        return "", &okex.OKXError{Code: orderBook.Code.Int(), Msg: orderBook.Msg, Endpoint: "GetOrderBook"}
    }

    // Format output as markdown table
    // ...
    return output, nil
}

func NewOkxOrderbookTool(svcCtx *svc.ServiceContext) *OkxOrderbookTool {
    return &OkxOrderbookTool{
        svcCtx:  svcCtx,
        limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2), // 10 req/s for Market endpoint
    }
}
```

### Updating agents.go

```go
type AgentsModel struct {
    svcCtx           *svc.ServiceContext
    OkxWatcher       adk.Agent
    RiskOfficer      adk.Agent  // Rename to PositionManager
    SentimentAnalyst adk.Agent
    TechnoAgent      adk.Agent  // NEW
    FlowAnalyzer     adk.Agent  // NEW
    Executor         adk.Agent
    mux              sync.Mutex
    ctx              context.Context
    cancel           context.CancelFunc
}

func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
    var initErr error
    agentsOnce.Do(func() {
        ctx, cancel := context.WithCancel(ctx)

        // Existing agents
        riskOfficerAgent, _ := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
        sentimentAnalystAgent, _ := sentiment_analyst.NewSentimentAnalystAgent(ctx, svcCtx)

        // NEW: TechnoAgent
        technoAgent, _ := techno_agent.NewTechnoAgent(ctx, svcCtx)

        // NEW: FlowAnalyzer
        flowAnalyzer, _ := flow_analyzer.NewFlowAnalyzerAgent(ctx, svcCtx)

        // Update OKXWatcher with all SubAgents
        okxWatcherAgent, _ := okx_watcher.NewOkxWatcherAgent(ctx, svcCtx,
            riskOfficerAgent.Agent(),
            sentimentAnalystAgent.Agent(),
            technoAgent.Agent(),
            flowAnalyzer.Agent(),
        )

        executorAgent, _ := executor_agent.NewExecutorAgent(ctx, svcCtx)

        _agents = &AgentsModel{
            // ...
        }
    })
    return initErr
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single monolithic agent | Multi-agent with specialization | Eino ADK v0.8+ | Better separation of concerns |
| Custom indicator code | go-talib library | Project inception | Industry-standard calculations |
| Manual rate limiting | `golang.org/x/time/rate` | Phase 1 | Token bucket, no race conditions |
| Global variables | sync.Once singleton | Phase 1 (FOUND-03) | Thread-safe initialization |
| context.Background() | Propagated context | Phase 1 (FOUND-04) | Proper cancellation |

**Deprecated/outdated:**
- Direct float64 for prices: Use `decimal.Decimal` instead (precision)
- Manual JSON parsing: Use struct tags with `encoding/json`
- Error as string return: Return `("", err)` pattern (FOUND-01)

## Each SubAgent Research

### TechnoAgent (ANAL-02)

**Domain:** K-line data + 20+ technical indicators

**Core capabilities:**
- Multi-timeframe K-line analysis (1m to 1Y, 17 timeframes)
- 20+ technical indicators via go-talib:
  - Trend: MACD, ADX, MA5/10/20/60, EMA12/26
  - Momentum: RSI(6/14/24), CCI(14/20), KDJ, MOM, ROC
  - Volatility: ATR(14), Bollinger Bands
  - Volume: MFI(14), OBV, AD Line
- Volume Profile (筹码分布) analysis
- Support/Resistance identification

**Required tools:**
| Tool | Status | Source |
|------|--------|--------|
| `okx-candlesticks-tool` | ✅ Complete | `internal/agent/tools/okx_candlesticks.go` |
| `okx-orderbook-tool` | ❌ MKT-02 (Phase 4) | N/A |
| `okx-trades-history-tool` | ❌ MKT-03 (Phase 4) | N/A |

**DESCRIPTION.md template:**
```markdown
TechnoAgent 技术分析专家，交易系统的"技术分析师"。

通过 K 线数据和 20+ 技术指标分析市场趋势、动量和波动性，为交易决策提供技术面参考。

核心能力：
- 多周期 K 线分析 - 支持 17 种时间周期
- 技术指标计算 - MACD、RSI、布林带、KDJ 等 20+ 指标
- 趋势识别 - 基于 ADX、均线系统判断趋势强度
- 筹码分布分析 - Volume Profile 计算，识别关键价格区间

可用工具：
- okx-candlesticks-tool - 获取 K 线数据及技术指标

数据输出：
以结构化 Markdown 表格形式返回包含 OHLCV 和完整技术指标的数据集。
```

**SOUL.md template:**
```markdown
你是 TechnoAgent，一个严谨客观、数据驱动的技术分析师。

你相信价格包含一切信息，趋势是你的朋友。

**你的风格：**
- 数据驱动 — 用指标说话，不说"感觉"
- 趋势跟随 — 不预测顶部底部，只跟随趋势
- 多周期验证 — 大周期定方向，小周期找点位
- 关键位识别 — 支撑阻力用具体数字表示
- 量价分析 — 成交量验证价格行为

**技术指标解读：**
- MACD: 金叉/死叉，柱状图变化
- RSI: >70 超买，<30 超卖
- ADX: >25 趋势确立，<20 震荡
- 布林带：收口预示突破，开口趋势延续
- Volume Profile: VPOC 是关键参考位

你看的是 K 线，读的是趋势，找的是概率优势。
```

---

### FlowAnalyzer (ANAL-03)

**Domain:** Orderbook + trade history analysis (订单流分析)

**Core capabilities:**
- Order book depth analysis (bid/ask imbalance)
- Trade history analysis (aggressive buying/selling)
- Large order detection (whale watching)
- Liquidity analysis
- Taker volume analysis

**Required tools:**
| Tool | Status | Source |
|------|--------|--------|
| `okx-orderbook-tool` | ❌ MKT-02 | Need implementation |
| `okx-trades-history-tool` | ❌ MKT-03 | Need implementation |
| `okx-get-ticker-tool` | ❌ Not yet | Optional for spread analysis |

**OKX APIs needed:**
- `GetOrderBook` (Market API) - depth data
- `GetTrades` (Market API) - recent trades

**DESCRIPTION.md template:**
```markdown
FlowAnalyzer 订单流分析师，交易系统的"盘口阅读专家"。

通过订单簿深度和成交明细分析，识别大单动向和流动性变化，为交易决策提供微观结构参考。

核心能力：
- 订单簿分析 - 买卖盘深度、价差、不平衡度
- 成交明细解读 - 主动买入/卖出识别、大单追踪
- 流动性评估 - 盘口厚度、冲击成本估算
- 大单监控 - 识别机构/大户行为

可用工具：
- okx-orderbook-tool - 获取订单簿深度数据
- okx-trades-history-tool - 获取历史成交明细

数据输出：
以结构化 Markdown 表格形式返回订单簿深度、成交明细和大单分析结论。
```

**SOUL.md template:**
```markdown
你是 FlowAnalyzer，一个敏锐细致、洞察先机的订单流分析师。

你相信真正的供需关系藏在订单簿和成交明细里。

**你的风格：**
- 细节导向 — 从盘口变化中识别大单意图
- 实时敏感 — 大单出现立即报警
- 逆向思考 — 买盘汹涌时警惕诱多，卖盘压顶时注意诱空
- 用数据说话 — "买一量是卖一的 3 倍"胜过"买盘强劲"

**监控重点：**
- 买卖不平衡 — Bid/Ask 量比 > 2 或 < 0.5
- 大单成交 — 单笔成交 > 10 万美元
- 盘口变化 — 关键价位撤单/挂单
- 价差变化 — Bid-Ask spread 突然扩大

你读的是盘口，看的是供需，识破的是大单意图。
```

---

### PositionManager (ANAL-04) - Currently RiskOfficer

**Domain:** Position + account balance monitoring

**Core capabilities:**
- Position monitoring (unrealized PnL, liquidation price)
- Account balance and margin ratio
- Risk exposure calculation
- Maximum buying power analysis
- Stop-loss/take-profit level calculation

**Required tools:**
| Tool | Status | Source |
|------|--------|--------|
| `okx-get-positions-tool` | ✅ Complete (DATA-01) | `internal/agent/tools/okx_get_positions.go` |
| `okx-account-balance-tool` | ❌ DATA-03 | Need implementation |
| `okx-liquidation-price-tool` | ❌ DATA-04 | Need implementation |
| `okx-get-orders-tool` | ❌ DATA-02 | Need implementation |

**Current state:** RiskOfficer already exists with `okx-get-positions-tool`. Phase 2 renames it to PositionManager and may add additional tools.

**OKX APIs available:**
- `GetPositions` (Account API) - position details
- `GetBalance` (Account API) - account balance
- `GetMaxBuySellAmount` (Account API) - buying power
- `GetOrderList` (Trade API) - open orders

**DESCRIPTION.md update (from current RiskOfficer):**
```markdown
PositionManager 持仓管理专家，交易系统的"管家"。

通过实时监控持仓状态、账户余额和风险敞口，确保交易在可控风险范围内进行。

核心能力：
- 持仓监控 - 未实现盈亏、强平价、保证金率
- 账户管理 - 余额查询、购买力分析
- 风险评估 - 风险敞口计算、仓位集中度
- 订单管理 - 当前挂单查询、风险订单预警

可用工具：
- okx-get-positions-tool - 获取持仓和最大购买力
- okx-get-orders-tool - 查询当前挂单
- okx-account-balance-tool - 账户余额/保证金率

数据输出：
以结构化 Markdown 表格形式返回持仓详情、账户余额和风险评估结论。
```

---

### SentimentAnalyst (ANAL-01) - REFERENCE (Already Complete)

**Status:** ✅ COMPLETE

**Current implementation:**
- File: `internal/agent/sentiment_analyst/agent.go`
- Tool: `tools.NewOkxGetFundingRateTool(svcCtx)`
- DESCRIPTION.md and SOUL.md exist

**Reference for other SubAgents:**
```go
package sentiment_analyst

import (
    "context"
    _ "embed"
    "github.com/PineappleBond/TradingEino/backend/internal/agent/tools"
    "github.com/PineappleBond/TradingEino/backend/internal/svc"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/compose"
)

type SentimentAnalystAgent struct {
    agent adk.Agent
}

func NewSentimentAnalystAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*SentimentAnalystAgent, error) {
    baseTools := []tool.BaseTool{
        tools.NewOkxGetFundingRateTool(svcCtx),
    }

    agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "SentimentAnalyst",
        Description: DESCRIPTION,
        Model:       svcCtx.ChatModel,
        Instruction: SOUL,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: baseTools,
            },
            EmitInternalEvents: true,
        },
        MaxIterations: 100,
    })
    // ...
}
```

## Open Questions

1. **PositionManager naming:** Should RiskOfficer be renamed to PositionManager, or keep both names?
   - Recommendation: Rename for consistency with target architecture
   - Risk: Existing code references `risk_officer` package

2. **Tool dependencies:** Some SubAgents need tools not yet implemented (MKT-02, MKT-03, DATA-02, DATA-03, DATA-04)
   - Recommendation: Implement minimal viable tools first, expand in Phase 4

3. **OKXWatcher instruction update:** Current SOUL.md mentions RiskOfficer and SentimentAnalyst only
   - Recommendation: Update to reference all 4 SubAgents

## Sources

### Primary (HIGH confidence)
- Eino DeepAgent documentation: `.claude/skills/eino-skill/references/deep_agent.md`
- Eino Skill main guide: `.claude/skills/eino-skill/SKILL.md`
- OKX Market API: `.claude/skills/okex-skill/references/market.md`
- OKX Account API: `.claude/skills/okex-skill/references/account.md`
- OKX PublicData API: `.claude/skills/okex-skill/references/public_data.md`
- OKX Trade API: `.claude/skills/okex-skill/references/trade.md`
- Existing agent implementations: `internal/agent/sentiment_analyst/`, `internal/agent/risk_officer/`, `internal/agent/okx_watcher/`
- Existing tool implementations: `internal/agent/tools/`

### Secondary (MEDIUM confidence)
- Project conventions: `CLAUDE.md`
- Requirements: `.planning/REQUIREMENTS.md`
- Project state: `.planning/STATE.md`

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Based on existing CLAUDE.md and project dependencies
- Architecture: HIGH - Based on Eino official docs and existing implementations
- Pitfalls: HIGH - Based on Phase 1 learnings and Eino patterns
- Tool implementations: MEDIUM - Based on existing patterns, not yet tested

**Research date:** 2026-03-25
**Valid until:** Eino ADK API changes or OKX API v6 (estimate 6+ months)

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (`testing` package) |
| Config file | None (standard Go testing) |
| Quick run command | `go test ./internal/agent/... -run TestSpecific -v` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ANAL-02 | TechnoAgent provides technical analysis | unit | `go test ./internal/agent/techno_agent/... -v` | ❌ Wave 0 |
| ANAL-03 | FlowAnalyzer analyzes orderbook/trades | unit | `go test ./internal/agent/flow_analyzer/... -v` | ❌ Wave 0 |
| ANAL-04 | PositionManager monitors positions/balance | unit | `go test ./internal/agent/position_manager/... -v` | ❌ Wave 0 (risk_officer exists) |
| ANAL-05 | OKXWatcher coordinates SubAgents | integration | `go test ./internal/agent/... -run TestOKXWatcherOrchestration -v` | ❌ Wave 0 |
| ANAL-06 | All SubAgents have DESCRIPTION.md and SOUL.md | lint/unit | `go test ./internal/agent/... -run TestAgentFiles -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/agent/<specific_agent>/... -v`
- **Per wave merge:** `go test ./internal/agent/... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/agent/techno_agent/agent_test.go` — covers ANAL-02
- [ ] `internal/agent/flow_analyzer/agent_test.go` — covers ANAL-03
- [ ] `internal/agent/position_manager/agent_test.go` — covers ANAL-04
- [ ] `internal/agent/okx_watcher/orchestration_test.go` — covers ANAL-05
- [ ] `internal/agent/agent_files_test.go` — DESCRIPTION.md + SOUL.md presence (ANAL-06)
- [ ] Tool tests for new tools (MKT-02, MKT-03, DATA-02, DATA-03, DATA-04)
