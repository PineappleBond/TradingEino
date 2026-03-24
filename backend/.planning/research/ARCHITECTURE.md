# Architecture Patterns

**Domain:** Multi-Agent Cryptocurrency Trading System
**Researched:** 2026-03-24
**Confidence:** HIGH (based on existing architecture docs + Eino framework patterns)

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Coordination Layer                               │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │              OKXWatcher (DeepAgent - Coordinator)                   │ │
│  │  - Market analysis orchestration                                    │ │
│  │  - Sub-agent scheduling (Techno, Sentiment, Flow, Position)         │ │
│  │  - Trade signal generation                                          │ │
│  └────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
            ┌───────────────────────┼───────────────────────┐
            │                       │                       │
┌───────────▼───────────┐ ┌────────▼────────┐ ┌───────────▼───────────┐
│   Analysis Layer      │ │ Execution Layer │ │   Risk Layer          │
│  ┌─────────────────┐  │ │ ┌─────────────┐ │ │ ┌───────────────────┐ │
│  │ TechnoAgent     │  │ │ │ Executor    │ │ │ │ RiskMonitor       │ │
│  │ (Technical)     │  │ │ │ Agent       │ │ │ │ (Independent)     │ │
│  ├─────────────────┤  │ │ ├─────────────┤ │ │ ├───────────────────┤ │
│  │ SentimentAgent  │  │ │ │ Tools:      │ │ │ │ Tools:            │ │
│  │ (Funding Rate)  │  │ │ │ - place     │ │ │ │ - positions       │ │
│  ├─────────────────┤  │ │ │ - cancel    │ │ │ │ - balance         │ │
│  │ FlowAnalyzer    │  │ │ │ - get-order │ │ │ │ - liquidation     │ │
│  │ (Order Book)    │  │ │ │ - close     │ │ │ │                   │ │
│  ├─────────────────┤  │ │ └─────────────┘ │ │ ├───────────────────┤ │
│  │ PositionMgr     │  │ │                 │ │ │ Circuit Breaker   │ │
│  │ (Account State) │  │ │ Autonomy:       │ │ │ (Auto Stop-Loss)  │ │
│  └─────────────────┘  │ │ - Level 1-3     │ │ └───────────────────┘ │
└───────────────────────┘ └─────────────────┘ └─────────────────────────┘
            │                       │                       │
            └───────────────────────┼───────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────┐
│                         Infrastructure Layer                             │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────┐  │
│  │ RAG Vector Store │  │   OKX API        │  │   SQLite/Redis       │  │
│  │ (Decision Memory)│  │   Gateway        │  │   (Persistence)      │  │
│  │ - Redis Stack    │  │   (Rate Limited) │  │                      │  │
│  │ - m3e-base       │  │                  │  │                      │  │
│  └──────────────────┘  └──────────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **OKXWatcher** | Top-level coordinator, decision orchestration | Eino DeepAgent with SubAgents config |
| **TechnoAgent** | Technical analysis (K-line, indicators) | ChatModelAgent + okx-candlesticks-tool |
| **SentimentAnalyst** | Funding rate and market sentiment | ChatModelAgent + okx-get-funding-rate-tool |
| **FlowAnalyzer** | Order book depth and trade flow analysis | ChatModelAgent + orderbook/trades-history tools |
| **PositionManager** | Account position and margin monitoring | ChatModelAgent + positions/balance tools |
| **Executor** | Trade execution (place/cancel/close) | ChatModelAgent + trading tools (configurable autonomy Level 1-3) |
| **RiskMonitor** | Independent real-time risk monitoring | Parallel ChatModelAgent, triggers alerts/forced actions |
| **CircuitBreaker** | Auto stop-loss/take-profit enforcement | Interrupt/Resume pattern with checkpoint store |

## Recommended Project Structure

```
internal/
├── agent/                    # Multi-agent system
│   ├── agents.go             # Agent initialization and wiring
│   ├── coordinator/          # OKXWatcher (DeepAgent)
│   │   ├── agent.go          # DeepAgent creation
│   │   ├── DESCRIPTION.md    # Agent description for LLM
│   │   └── SOUL.md           # Agent personality/instruction
│   ├── analysis/             # Analysis SubAgents
│   │   ├── techno/           # Technical analysis agent
│   │   ├── sentiment/        # Sentiment/funding rate agent
│   │   └── flow/             # Order flow analysis agent
│   ├── execution/            # Execution layer
│   │   ├── executor/         # Trade execution agent
│   │   └── circuit_breaker/  # Auto stop-loss/take-profit
│   ├── risk/                 # Risk management layer
│   │   ├── monitor/          # Independent risk monitoring
│   │   └── limits/           # Risk threshold definitions
│   └── tools/                # Atomic tools for agents
│       ├── data/             # Data retrieval tools
│       │   ├── candlesticks.go
│       │   ├── funding_rate.go
│       │   ├── orderbook.go
│       │   └── positions.go
│       └── trading/          # Trade execution tools
│           ├── place_order.go
│           ├── cancel_order.go
│           └── close_position.go
├── rag/                      # RAG infrastructure
│   ├── store.go              # Vector store interface
│   ├── redis_store.go        # Redis Stack implementation
│   ├── embedder.go           # m3e-base embedder setup
│   └── decision_memory.go    # Decision record schema
└── service/
    ├── scheduler/            # Task scheduling
    └── risk_engine/          # Risk calculation engine
```

### Structure Rationale

- **agent/**: All LLM agent code grouped by responsibility (coordinator, analysis, execution, risk)
- **agent/tools/**: Atomic tools separated by domain (data vs trading) for clear capability boundaries
- **rag/**: Dedicated package for vector store and decision memory logic
- **risk** as parallel layer: Risk monitoring independent from OKXWatcher execution chain

## Architectural Patterns

### Pattern 1: DeepAgent Coordinator with ChatModelAgent SubAgents

**What:** Use DeepAgent only for the top-level coordinator (OKXWatcher), with all SubAgents implemented as ChatModelAgents.

**When to use:** Complex tasks requiring dynamic sub-agent scheduling and task decomposition.

**Trade-offs:**
- Pros: Clear hierarchy, dynamic routing, sub-agents are lightweight
- Cons: Coordinator becomes potential bottleneck, requires clear sub-agent descriptions

**Example:**
```go
// Coordinator (DeepAgent)
agent, err := deep.New(ctx, &deep.Config{
    Name:        "OKXWatcher",
    Description: "Cryptocurrency market analysis coordinator",
    ChatModel:   svcCtx.ChatModel,
    Instruction: SOUL, // Embedded personality
    SubAgents:   []adk.Agent{technoAgent, sentimentAgent, flowAgent},
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{candlesticksTool},
        },
    },
    MaxIteration: 100,
})

// SubAgent (ChatModelAgent - lightweight, no task decomposition)
technoAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "TechnoAgent",
    Description: "Technical analysis expert specializing in K-line patterns and indicators",
    Model:       svcCtx.ChatModel,
    Instruction: "You analyze trends, support/resistance, and technical signals...",
    ToolsConfig: adk.ToolsConfig{
        Tools: []tool.BaseTool{candlesticksTool},
    },
    MaxIterations: 10,
})
```

### Pattern 2: Independent Risk Layer (Parallel Monitoring)

**What:** Risk monitoring runs as an independent parallel process, not dependent on OKXWatcher's scheduling cycle.

**When to use:** When risk checks need to happen at higher frequency than trading decisions, or when risk actions need to override trading decisions.

**Trade-offs:**
- Pros: Real-time protection, can force-stop trades, independent failure domain
- Cons: Requires separate scheduler, potential race conditions with executor

**Implementation:**
```go
// RiskMonitor runs on separate ticker (e.g., every 30 seconds)
func StartRiskMonitor(ctx context.Context, svcCtx *svc.ServiceContext) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            checkRiskLimits(ctx, svcCtx)
        }
    }
}

func checkRiskLimits(ctx context.Context, svcCtx *svc.ServiceContext) {
    // 1. Check margin ratio
    balance := getAccountBalance(ctx)
    if balance.MarginRatio > 0.80 {
        triggerRiskAlert(ctx, "Margin ratio exceeds 80%")
    }
    if balance.MarginRatio > 0.90 {
        triggerForcedReduction(ctx) // Override trading decisions
    }

    // 2. Check liquidation distance
    positions := getPositions(ctx)
    for _, pos := range positions {
        distance := calculateLiquidationDistance(pos)
        if distance < 0.03 {
            triggerRiskAlert(ctx, fmt.Sprintf("Liquidation distance < 3%% for %s", pos.Symbol))
        }
    }
}
```

### Pattern 3: Interrupt/Resume for Circuit Breaker

**What:** Use Eino's Interrupt/Resume mechanism to implement circuit breaker pattern for trading safety.

**When to use:** When trades require human approval, or when risk conditions trigger automatic trade suspension.

**Trade-offs:**
- Pros: Clean state management, automatic checkpoint persistence, resumable workflows
- Cons: Requires CheckpointStore implementation, adds complexity to flow

**Example:**
```go
// Circuit breaker interrupt before trade execution
func (e *ExecutorAgent) executeTrade(ctx context.Context, signal *TradeSignal) error {
    state := compose.GetState[*ExecutorState](ctx)

    // Check if this is a resume after interruption
    isResume, hasData, data := compose.GetResumeContext[ApprovalData](ctx)
    if isResume && hasData {
        if data.Approved {
            return e.placeOrder(ctx, signal)
        }
        return fmt.Errorf("trade rejected by circuit breaker")
    }

    // Check circuit breaker conditions
    if shouldTriggerCircuitBreaker(ctx, signal) {
        // Trigger interrupt and wait for approval
        compose.Interrupt(ctx, map[string]any{
            "type":    "circuit_breaker",
            "reason":  "Large position or unusual market condition",
            "signal":  signal,
            "riskLevel": calculateRiskLevel(signal),
        })
        return nil // Will resume after approval
    }

    return e.placeOrder(ctx, signal)
}

// Compile with checkpoint store
runner, err := graph.Compile(ctx,
    compose.WithCheckPointStore(redisStore),
    compose.WithInterruptBeforeNodes([]string{"ExecutorNode"}),
)
```

### Pattern 4: RAG for Decision Memory

**What:** Store historical trading decisions in vector store, retrieve similar situations during new analysis.

**When to use:** When LLM needs to reference past decisions and outcomes for consistency and learning.

**Trade-offs:**
- Pros: Decision consistency, outcome tracking, pattern recognition over time
- Cons: Embedding latency, requires vector infrastructure (Redis Stack), storage management

**Example:**
```go
// Decision record schema
type DecisionRecord struct {
    ID        string                 `json:"id"`
    Timestamp time.Time              `json:"timestamp"`
    Symbol    string                 `json:"symbol"`
    Type      string                 `json:"type"` // analysis/decision/execution
    Input     map[string]any         `json:"input"`
    Analysis  map[string]any         `json:"analysis"` // Per-agent results
    Decision  Decision               `json:"decision"`
    Execution *ExecutionResult       `json:"execution,omitempty"`
    Outcome   *Outcome               `json:"outcome,omitempty"`
}

// Save decision to vector store
func (r *RAGStore) SaveDecision(ctx context.Context, record *DecisionRecord) error {
    // Generate embedding from decision content
    embedding, err := r.embedder.EmbedStrings(ctx, []string{record.Content()})

    // Store in Redis Stack with metadata
    doc := &schema.Document{
        ID:      record.ID,
        Content: record.Content(),
        MetaData: map[string]any{
            "symbol":    record.Symbol,
            "timestamp": record.Timestamp.Unix(),
            "type":      record.Type,
            "outcome":   record.Outcome.Rating, // profitable/loss
        },
    }
    return r.indexer.Store(ctx, []*schema.Document{doc})
}

// Retrieve similar historical decisions
func (r *RAGStore) SearchSimilar(ctx context.Context, query string, symbol string, topK int) ([]*DecisionRecord, error) {
    docs, err := r.retriever.Retrieve(ctx, query)
    // Filter by symbol and time range
    // Return top K similar decisions
}
```

### Pattern 5: Tool Atomicity with Pre/Post Processing

**What:** Each tool does exactly one thing. Wrap tools with pre-processing (input repair) and post-processing (formatting).

**When to use:** For all API-facing tools to ensure robustness and LLM compatibility.

**Trade-offs:**
- Pros: Tools are testable, LLM understands single-purpose tools, flexible composition
- Cons: More tool files, requires wrapper pattern

**Example:**
```go
// Atomic tool: Get positions only
type OkxGetPositionsTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter
}

func (t *OkxGetPositionsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "okx_get_positions",
        Desc: "Get current account positions and max buy/sell power",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "symbol": {Type: "string", Desc: "Symbol (optional, empty for all)", Required: false},
        }),
    }, nil
}

func (t *OkxGetPositionsTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    // Rate limiting
    if err := t.limiter.Wait(ctx); err != nil {
        return "", err
    }

    // Parse input
    var input struct{ Symbol string }
    json.Unmarshal([]byte(argsJSON), &input)

    // Call OKX API
    positions, err := t.svcCtx.OKXClient.GetPositions(ctx, input.Symbol)
    if err != nil {
        return "", err // Return error properly, not as string
    }

    // Format output
    return formatPositions(positions), nil
}

// Wrap with pre/post processing
wrappedTool := tools.NewWrapTool(
    baseTool,
    []tools.ToolRequestPreprocess{tools.ToolRequestRepairJSON}, // Fix malformed JSON
    []tools.ToolResponsePostprocess{tools.FilePostProcess},     // Format output
)
```

## Data Flow

### Analysis Mode Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│  1. Scheduler triggers OKXWatcher (via cron expression)                  │
│                    │                                                     │
│                    ▼                                                     │
│  2. OKXWatcher generates analysis prompt                                 │
│     - Symbol: BTC-USDT-SWAP                                              │
│     - Context: Current price, 24h change                                 │
│                    │                                                     │
│                    ▼                                                     │
│  3. OKXWatcher calls SubAgents in parallel/sequence                      │
│     ┌──────────────┬──────────────┬──────────────┬──────────────┐       │
│     │ TechnoAgent  │ Sentiment    │ FlowAnalyzer │ PositionMgr  │       │
│     │ (candlesticks│ (funding     │ (orderbook)  │ (positions)  │       │
│     │  + indicators)│  rate)       │              │              │       │
│     └──────┬───────┴──────┬───────┴──────┬───────┴──────┬───────┘       │
│            │              │              │              │                │
│            ▼              ▼              ▼              ▼                │
│  4. Results aggregated into analysis context                             │
│                    │                                                     │
│                    ▼                                                     │
│  5. OKXWatcher generates trade signal or hold recommendation             │
│                    │                                                     │
│                    ▼                                                     │
│  6. Decision saved to RAG vector store (for future reference)            │
└─────────────────────────────────────────────────────────────────────────┘
```

### Execution Mode Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│  1-5. Same as Analysis Mode (generates trade signal)                     │
│                    │                                                     │
│                    ▼                                                     │
│  6. OKXWatcher calls Executor Agent with trade signal                    │
│                    │                                                     │
│                    ▼                                                     │
│  7. Executor checks Circuit Breaker conditions                           │
│     ├── If triggered ──► Interrupt, wait for approval                   │
│     └── If clear ──► Continue                                           │
│                    │                                                     │
│                    ▼                                                     │
│  8. Executor calls trading tools                                         │
│     - okx_place_order (limit/market)                                     │
│     - Returns order_id, status                                           │
│                    │                                                     │
│                    ▼                                                     │
│  9. Executor monitors order status                                       │
│     - okx_get_order (poll until filled/cancelled)                        │
│                    │                                                     │
│                    ▼                                                     │
│  10. Execution result saved to RAG + database                            │
│                    │                                                     │
│                    ▼                                                     │
│  11. RiskMonitor validates post-execution state                          │
│     - Margin ratio check                                                 │
│     - Position limit check                                               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Risk Monitoring Flow (Parallel)

```
┌─────────────────────────────────────────────────────────────────────────┐
│  RiskMonitor ticker (every 30 seconds, independent of OKXWatcher)        │
│                    │                                                     │
│                    ▼                                                     │
│  1. Fetch real-time account state                                        │
│     - Positions, balance, margin ratio                                   │
│     - Liquidation prices                                                 │
│                    │                                                     │
│                    ▼                                                     │
│  2. Check risk thresholds                                                │
│     ├── Margin > 80% ──► Warning alert                                  │
│     ├── Margin > 90% ──► Trigger forced reduction                       │
│     ├── Liq. distance < 3% ──► Warning alert                            │
│     └── Liq. distance < 2% ──► Trigger forced close                     │
│                    │                                                     │
│                    ▼                                                     │
│  3. Risk actions (if triggered)                                          │
│     - Send alert (Telegram/DingTalk)                                     │
│     - Call Executor to reduce/close position                             │
│     - Log risk event to audit trail                                      │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Data Flows

1. **Decision Persistence Flow:**
   ```
   OKXWatcher decision → Serialize to DecisionRecord → Embedding (m3e-base)
   → Store in Redis Stack → Tagged with symbol/outcome/timestamp
   ```

2. **RAG Retrieval Flow:**
   ```
   New analysis request → Generate query embedding → Vector similarity search
   → Filter by symbol + time range → Return Top-K similar historical decisions
   → Inject into LLM prompt as context
   ```

3. **Circuit Breaker Flow:**
   ```
   Executor receives trade signal → Check risk thresholds → If exceeded:
   Interrupt + save checkpoint → Wait for approval → Resume with decision
   → Execute or reject trade
   ```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| **0-100 trades/day** | Single OKXWatcher instance, SQLite for persistence, simple cron scheduling |
| **100-1000 trades/day** | Redis for RAG + session management, rate limiting on all tools, separate RiskMonitor scheduler |
| **1000+ trades/day** | Multiple OKXWatcher instances (sharded by symbol), PostgreSQL for audit trail, dedicated execution service |

### Scaling Priorities

1. **First bottleneck:** OKX API rate limits
   - Fix: Implement per-tool rate.Limiter with token bucket algorithm
   - Priority: P0 (before any trading)

2. **Second bottleneck:** RiskMonitor frequency vs OKXWatcher frequency
   - Fix: Run RiskMonitor on separate ticker (30s vs 5min for OKXWatcher)
   - Priority: P0 (before enabling real trading)

3. **Third bottleneck:** Single OKXWatcher becomes bottleneck for multi-symbol tracking
   - Fix: Shard by symbol (BTC-USDT-SWAP watcher, ETH-USDT-SWAP watcher)
   - Priority: P2 (after execution layer is stable)

## Anti-Patterns

### Anti-Pattern 1: SubAgents as DeepAgents

**What people do:** Making all SubAgents (TechnoAgent, RiskOfficer) as DeepAgents.

**Why it's wrong:** DeepAgents are for task decomposition. SubAgents with single responsibility don't need this overhead. Causes unnecessary complexity and slower execution.

**Do this instead:** Use DeepAgent only for OKXWatcher (coordinator). Implement all SubAgents as ChatModelAgents.

### Anti-Pattern 2: Risk Monitoring Inside OKXWatcher Chain

**What people do:** Calling RiskOfficer as a SubAgent within OKXWatcher's analysis chain.

**Why it's wrong:** Risk checks only happen when OKXWatcher runs. If OKXWatcher runs every 5 minutes, risk issues can go undetected for 5 minutes. Risk should be real-time.

**Do this instead:** Run RiskMonitor as independent parallel process with higher frequency (30s). Allow RiskMonitor to override trading decisions.

### Anti-Pattern 3: Tools Returning Errors as Success Strings

**What people do:** `return err.Error(), nil` instead of `return "", err`

**Why it's wrong:** LLM interprets error message as successful tool output, leading to incorrect decisions.

**Do this instead:** Always return `return "", err` for tool errors. Eino framework handles error propagation correctly.

### Anti-Pattern 4: Global Agent State

**What people do:** Storing agents in global variables without proper lifecycle management.

**Why it's wrong:** Context cancellation not propagated, resources not cleaned up on failure, testing becomes difficult.

**Do this instead:** Use AgentsModel struct with explicit ctx/cancel, inject via ServiceContext.

### Anti-Pattern 5: Executor Without Autonomy Levels

**What people do:** Executor always runs trades immediately without configurable autonomy.

**Why it's wrong:** No safety gradient for production rollout. Bugs cause immediate real trades.

**Do this instead:** Implement autonomy levels:
- Level 1: Execute only explicit OKXWatcher commands
- Level 2: Optimize timing/price within bounds
- Level 3: Auto stop-loss/take-profit

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **OKX Exchange** | REST API via `pkg/okex/api/client.go` | All tools must use rate limiter, handle 429/5xx gracefully |
| **Redis Stack** | RAG vector store + CheckpointStore | Required for RAG + Interrupt/Resume features |
| **Ollama (Local)** | m3e-base embedder via HTTP | Run `ollama pull m3e-base`, no external API dependency |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| **OKXWatcher ↔ SubAgents** | DeepAgent.Invoke() | SubAgents receive prompt, return analysis string |
| **OKXWatcher ↔ Executor** | DeepAgent.Invoke() | Executor receives trade signal, returns execution result |
| **RiskMonitor → Executor** | Direct function call (override) | RiskMonitor can force Executor actions |
| **Agents ↔ Tools** | Tool.InvokableRun() | Atomic function calls, JSON input/output |
| **RAG ↔ Agents** | Search/Save functions | Decision memory persistence and retrieval |

## Build Order Implications

Based on component dependencies:

```
Phase 1: Infrastructure
├── OKX API client (existing)
├── Rate limiter for all tools
├── Tool error handling fix
└── Basic logging/auditing

Phase 2: Analysis Layer
├── Refactor RiskOfficer to ChatModelAgent
├── Refactor SentimentAnalyst (existing)
├── Add TechnoAgent
└── OKXWatcher integration with SubAgents

Phase 3: Execution Layer
├── Trading tools (place-order, cancel-order, close-position)
├── Executor Agent (Level 1 autonomy)
├── Circuit breaker with Interrupt/Resume
└── Paper trading mode

Phase 4: Risk Layer
├── Independent RiskMonitor scheduler
├── Risk threshold configuration
├── Auto stop-loss/take-profit tools
└── Forced reduction/close logic

Phase 5: Memory (RAG)
├── Redis Stack deployment
├── Ollama + m3e-base setup
├── Decision record schema
├── RAG save/search tools
└── OKXWatcher RAG integration

Phase 6: Analysis Enhancement
├── FlowAnalyzer (orderbook + trades-history)
├── PositionManager refactor
├── Additional data tools (OI, long/short ratio)
└── Multi-symbol OKXWatcher sharding
```

## Sources

- Existing Architecture: `.planning/codebase/ARCHITECTURE.md`
- Multi-Agent Requirements: `.planning/03-multi-agent/03-CONTEXT.md`
- Eino DeepAgent Pattern: `.claude/skills/eino-skill/references/deep_agent.md`
- Eino Interrupt/Resume: `.claude/skills/eino-skill/references/interrupt.md`
- Eino RAG Pattern: `.claude/skills/eino-skill/references/rag.md`
- Eino Framework Overview: `.claude/skills/eino-skill/SKILL.md`

---

*Architecture research for: Multi-Agent Cryptocurrency Trading System*
*Researched: 2026-03-24*
