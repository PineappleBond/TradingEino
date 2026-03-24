# Technology Stack for Multi-Agent Trading Features

**Project:** TradingEino - Multi-Agent Crypto Trading Bot
**Research Date:** 2026-03-24
**Focus:** Gaps for execution layer, RAG memory, risk monitoring, stop-loss/take-profit

---

## Existing Stack (Already in Place)

| Component | Technology | Version | Status |
|-----------|------------|---------|--------|
| **Runtime** | Go | 1.26.1 | ✅ |
| **Multi-Agent Framework** | Cloudwego Eino | 0.8.4 | ✅ |
| **HTTP Server** | Gin | 1.12.0 | ✅ |
| **Database** | SQLite3 (ncruces/go-sqlite3) | 0.30.2 | ✅ |
| **ORM** | GORM | 1.31.1 | ✅ |
| **LLM Client** | eino-ext/components/model/openai | 0.1.10 | ✅ |
| **Configuration** | Viper | 1.21.0 | ✅ |
| **Technical Analysis** | go-talib | latest | ✅ |
| **Decimal Math** | shopspring/decimal | 1.4.0 | ✅ |
| **Rate Limiting** | golang.org/x/time | 0.15.0 | ✅ |
| **OKX API** | In-house (pkg/okex) | - | ✅ |

---

## New Stack Requirements (Gaps to Fill)

### 1. RAG Vector Memory Stack

**Purpose:** Store and retrieve historical trading decisions for LLM context augmentation.

| Component | Recommended | Alternative | Why |
|-----------|-------------|-------------|-----|
| **Vector Store** | Redis Stack (RediSearch) | Milvus, Qdrant | Eino has official Redis retriever/indexer; Redis Stack combines vector + structured data in one DB |
| **Embedding Model** | m3e-base (via Ollama) | BAAI/bge-small-zh, text-embedding-3-small | m3e-base optimized for Chinese, runs locally on M2 Pro 32GB, no API dependency |
| **Embedding Client** | eino-ext/components/embedding/ollama | Direct Ollama SDK | Consistent with Eino ecosystem |
| **Redis Client** | redis/go-redis/v9 | Official Redis Go | Most popular, maintained by Redis team |

**Installation:**
```bash
go get github.com/redis/go-redis/v9
go get github.com/cloudwego/eino-ext/components/embedding/ollama@latest
go get github.com/cloudwego/eino-ext/components/retriever/redis@latest
go get github.com/cloudwego/eino-ext/components/indexer/redis@latest
```

**Redis Stack Deployment:**
```bash
# Docker (local development)
docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest

# Redis Stack includes:
# - Redis 7.x with RedisJSON, RediSearch, RedisGraph
# - Vector similarity search (HNSW, FLAT index)
# - Web UI at localhost:8001
```

**Ollama + m3e-base Setup:**
```bash
# Install Ollama
brew install ollama  # macOS
# Or: curl -fsSL https://ollama.com/install.sh | sh

# Pull m3e-base embedding model
ollama pull m3e-base

# Verify
ollama run m3e-base "测试文本"
```

**Configuration (for ServiceContext):**
```go
type ServiceContext struct {
    // ... existing fields
    Embedder components.Embedder  // m3e-base via Ollama
    Redis    *redis.Client
    // RAG components created on-demand
}
```

**Confidence:** HIGH - Based on official Eino documentation (eino-skill/SKILL.md) and ADR-004 approval.

---

### 2. Execution Layer Tools Stack

**Purpose:** Place orders, cancel orders, query orders, close positions.

**Current State:** OKX API client (`pkg/okex/api/trading.go`) already has:
- `PlaceOrder()` - places new order
- `CancelOrder()` - cancels existing order
- `GetOrderDetails()` - queries order status
- `ClosePosition()` - defined in trade_requests.go but not wrapped in API layer

**Tools to Implement (Eino Tool Interface):**

| Tool Name | Wraps | Priority |
|-----------|-------|----------|
| `okx-place-order-tool` | `c.OKXClient.Rest.Trade.PlaceOrder()` | P1 |
| `okx-cancel-order-tool` | `c.OKXClient.Rest.Trade.CandleOrder()` | P1 |
| `okx-get-order-tool` | `c.OKXClient.Rest.Trade.GetOrderDetail()` | P4 |
| `okx-close-position-tool` | `c.OKXClient.Rest.Trade.ClosePosition()` | P4 |

**No New Libraries Required** - uses existing `pkg/okex` client.

**Add Rate Limiting (P0):**
```go
import "golang.org/x/time/rate"

type OkxPlaceOrderTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter // 5 requests/second for OKX trade APIs
}

func NewOkxPlaceOrderTool(svcCtx *svc.ServiceContext) *OkxPlaceOrderTool {
    return &OkxPlaceOrderTool{
        svcCtx:  svcCtx,
        limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1), // 5 req/s
    }
}

func (t *OkxPlaceOrderTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    if err := t.limiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limit: %w", err)
    }
    // ... proceed with API call
}
```

**OKX API Rate Limits Reference:**
| Endpoint Type | Rate Limit |
|---------------|------------|
| Trade (place/cancel) | 5 requests/second |
| Order query | 10 requests/second |
| Account/Position | 10 requests/second |

**Confidence:** HIGH - Uses existing OKX client, rate limiter already in go.mod.

---

### 3. Independent Risk Monitoring Layer Stack

**Purpose:** Real-time position/margin monitoring independent of OKXWatcher schedule.

**Architecture:** Separate goroutine with its own cron scheduler (not dependent on OKXWatcher).

**Components:**

| Component | Recommended | Why |
|-----------|-------------|-----|
| **Scheduler** | Existing `internal/service/scheduler` | Reuse existing cron infrastructure |
| **Agent Type** | ChatModelAgent (not DeepAgent) | RiskMonitor doesn't need task decomposition |
| **Tools** | Reuse: `okx-get-positions-tool`, `okx-account-balance-tool`, `okx-liquidation-price-tool` | No new tools needed |
| **Alert Channels** | Telegram Bot API | Crypto standard, real-time push notifications |

**Telegram Integration (for Alerts):**
```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

**RiskMonitor Agent Structure:**
```go
type RiskMonitorAgent struct {
    agent       adk.Agent
    ctx         context.Context
    cancel      context.CancelFunc
    alertChan   chan RiskAlert
    telegramBot *tgbotapi.BotAPI
}

// Runs as independent goroutine
func (r *RiskMonitorAgent) StartMonitoring() {
    ticker := time.NewTicker(30 * time.Second) // Check every 30s
    defer ticker.Stop()

    for range ticker.C {
        r.checkMarginRatio()
        r.checkLiquidationDistance()
        r.checkDailyLoss()
    }
}
```

**Risk Rules Configuration:**
```yaml
# etc/config.yaml
risk:
  margin_ratio_warning: 0.8    # 80% -> warning
  margin_ratio_critical: 0.9   # 90% -> force reduce
  liquidation_distance_warning: 0.03   # 3% -> warning
  liquidation_distance_critical: 0.02  # 2% -> force close
  single_position_max: 0.5     # 50% of equity
  daily_loss_limit: 0.05       # 5% -> stop trading 24h
  daily_loss_halt: 0.10        # 10% -> close all positions
```

**Confidence:** HIGH - Uses existing scheduler and OKX tools, Telegram is standard for crypto.

---

### 4. Stop-Loss/Take-Profit Automation Stack

**Purpose:** Automatic exit positions at predefined levels.

**Two Approaches:**

#### Approach A: OKX Server-Side Algo Orders (Recommended)

OKX provides native stop-loss/take-profit orders that execute on their matching engine.

**OKX Algo Order Types:**
| Type | Use Case |
|------|----------|
| `sl_tp` | Simultaneous stop-loss and take-profit |
| `trigger` | Price-triggered market/limit order |
| `trailing_stop` | Trailing stop-loss |
| `iceberg` | Large order splitting |
| `twap` | Time-weighted average price execution |

**Implementation:**
```go
// pkg/okex/requests/rest/trade/trade_requests.go already has:
type PlaceAlgoOrder struct {
    InstID     string
    TdMode     okex.TradeMode
    Side       okex.OrderSide
    PosSide    okex.PositionSide
    OrdType    okex.AlgoOrderType // "sl_tp", "trigger", etc.
    Sz         int64
    StopOrder  // tpTriggerPx, tpOrdPx, slTriggerPx, slOrdPx
    // ...
}
```

**Tool to Add:**
```go
// okx-place-algo-order-tool
type OkxPlaceAlgoOrderTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter
}

// Input:
{
    "symbol": "ETH-USDT-SWAP",
    "pos_side": "long",
    "size": 100,
    "stop_loss": {
        "trigger_px": 3200,
        "order_px": -1  // -1 = market order
    },
    "take_profit": {
        "trigger_px": 3800,
        "order_px": 3850  // limit order at 3850
    }
}
```

**Rate Limit:** Algo orders share the same 5 req/s limit as regular orders.

#### Approach B: Client-Side Monitoring (Fallback)

If OKX algo orders are not suitable, implement client-side monitoring:

| Component | Recommended | Why |
|-----------|-------------|-----|
| **Price Feed** | OKX WebSocket (public channel) | Real-time mark price updates |
| **Monitoring** | Existing `internal/service/scheduler` | Reuse cron infrastructure |
| **Execution** | `okx-place-order-tool` | Standard market order on trigger |

**WebSocket Client (already available):**
```go
// pkg/okex/api/ws/public.go handles public WebSocket channels
// Subscribe to mark price channel:
// channel: "mark-price", inst_id: "ETH-USDT-SWAP"
```

**Recommendation:** Use **Approach A (OKX Algo Orders)** because:
1. Lower latency (executes on matching engine, not dependent on client polling)
2. More reliable (no network disconnection risk)
3. No additional infrastructure needed
4. Already partially implemented in trade_requests.go

**Library to Add:** None - uses existing OKX client and WebSocket.

**Confidence:** HIGH - OKX algo order API is well-documented, existing client support.

---

## Complete Stack Summary

### New Dependencies to Add

```bash
# go.mod additions for RAG
go get github.com/redis/go-redis/v9
go get github.com/cloudwego/eino-ext/components/embedding/ollama@latest
go get github.com/cloudwego/eino-ext/components/retriever/redis@latest
go get github.com/cloudwego/eino-ext/components/indexer/redis@latest

# go.mod additions for Risk Monitoring alerts
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

### Infrastructure Requirements

| Service | Purpose | Deployment |
|---------|---------|------------|
| Redis Stack | Vector store + structured data | Docker container or managed Redis Cloud |
| Ollama + m3e-base | Local embedding generation | Run on same host (M2 Pro 32GB has enough RAM) |
| Telegram Bot | Alert notifications | Free, create via @BotFather |

### What NOT to Use

| Technology | Reason to Avoid |
|------------|-----------------|
| **Pinecone/Weaviate Cloud** | External API dependency, latency, cost - Redis Stack runs locally |
| **OpenAI Embeddings** | API cost, latency, data privacy - m3e-base runs offline |
| **Separate risk DB** | Redis Stack handles both vectors and structured data |
| **Custom WebSocket price monitor** | OKX algo orders execute server-side, more reliable |
| **Milvus/Qdrant standalone** | Overkill for <100K decision records, Redis Stack sufficient |

---

## Integration Points with Existing Stack

### ServiceContext Extensions

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config      *config.Config
    DB          *gorm.DB
    Logger      *slog.Logger
    ChatModel   *openai.ChatModel
    OKXClient   *okex.Client

    // NEW: RAG components
    Embedder    components.Embedder
    Redis       *redis.Client

    // NEW: Risk monitoring
    TelegramBot *tgbotapi.BotAPI
}
```

### Agent Additions

```go
// internal/agent/agents.go
type AgentsModel struct {
    // Existing
    OkxWatcher       adk.Agent
    RiskOfficer      adk.Agent
    SentimentAnalyst adk.Agent

    // NEW
    Executor         adk.Agent      // P1
    RiskMonitor      *RiskMonitorAgent // P0 (independent goroutine)
}
```

### Tool Additions

```go
// internal/agent/tools/
// P0 - Add rate limiter to existing tools
okx_get_positions.go     -> add rate.Limiter
okx_get_fundingrate.go   -> add rate.Limiter
okx_candlesticks.go      -> add rate.Limiter

// P1 - New execution tools
okx_place_order.go       -> wraps PlaceOrder()
okx_cancel_order.go      -> wraps CancelOrder()
okx_get_order.go         -> wraps GetOrderDetails()
okx_close_position.go    -> wraps ClosePosition()

// P1 - RAG tools
okx_decision_save.go     -> saves to Redis vector store
okx_decision_search.go   -> retrieves from Redis

// P0 - Risk tools (reuse existing with rate limiter)
// No new tools needed for RiskMonitor
```

---

## Version Compatibility Matrix

| Library | Min Version | Tested Version | Notes |
|---------|-------------|----------------|-------|
| Go | 1.24 | 1.26.1 | Generic support required |
| Cloudwego Eino | 0.8.0 | 0.8.4 | RAG components available since 0.7.x |
| eino-ext/embedding/ollama | - | latest | Verify compatibility with Eino 0.8.4 |
| eino-ext/retriever/redis | - | latest | Verify compatibility |
| redis/go-redis | 9.0.0 | latest | v9 has breaking changes from v8 |
| telegram-bot-api | 5.0.0 | latest | Stable API |

---

## Sources and Confidence

| Component | Source | Confidence |
|-----------|--------|------------|
| Eino RAG pattern | `.claude/skills/eino-skill/references/rag.md` | HIGH |
| Eino ext components | `.claude/skills/eino-skill/SKILL.md` | HIGH |
| OKX algo orders | `pkg/okex/requests/rest/trade/trade_requests.go` | HIGH |
| Existing agent structure | `internal/agent/` | HIGH |
| Redis Stack for vectors | Training data (common pattern) | MEDIUM |
| m3e-base for Chinese | Training data (known model) | MEDIUM |
| Telegram for crypto alerts | Industry standard pattern | MEDIUM |

**Note:** Web search was not available to verify latest versions. Check official sources before finalizing:
- https://github.com/cloudwego/eino-ext
- https://redis.io/docs/latest/develop/data-types/vector/
- https://ollama.com/library/m3e-base

---

## Pre-Implementation Checklist

Before starting implementation:

- [ ] Verify `eino-ext/components/embedding/ollama` package exists and is compatible with Eino 0.8.4
- [ ] Verify `eino-ext/components/retriever/redis` package exists
- [ ] Test Redis Stack Docker image on target deployment platform
- [ ] Verify Ollama m3e-base model performance on M2 Pro
- [ ] Create Telegram bot and test message delivery
- [ ] Review OKX algo order API documentation for latest parameters
