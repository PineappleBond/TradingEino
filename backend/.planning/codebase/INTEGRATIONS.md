# External Integrations

**Analysis Date:** 2026-03-24

## APIs & External Services

**OKX Exchange:**
- OKX API v5 - Cryptocurrency exchange for market data and trading
  - SDK/Client: Custom wrapper in `pkg/okex/`
  - Auth: `TRADINGEINO_OKX_API_KEY`, `TRADINGEINO_OKX_SECRET_KEY`, `TRADINGEINO_OKX_PASSPHRASE`
  - Endpoints:
    - REST: `https://www.okx.com` (or `https://aws.okx.com` for AWS region)
    - Public WebSocket: `wss://ws.okx.com:8443/ws/v5/business`
    - Private WebSocket: `wss://ws.okx.com:8443/ws/v5/private`
  - Sandbox: `TRADINGEINO_OKX_SANDBOX=false` (uses Demo server when true)

**LLM Provider:**
- DashScope (阿里云通义千问) - AI chat model provider
  - SDK/Client: `github.com/cloudwego/eino-ext/components/model/openai`
  - Auth: `TRADINGEINO_CHAT_MODEL_API_KEY`
  - Base URL: `https://coding.dashscope.aliyuncs.com/v1`
  - Model: `qwen3.5-plus` (configurable)
  - Compatible with OpenAI API format

## Data Storage

**Databases:**
- SQLite (embedded)
  - Connection: `./data/TradingEino.db` (configurable via `TRADINGEINO_DB_DB_PATH`)
  - Client: GORM v1.31.1 with `ncruces/go-sqlite3/gormlite` driver
  - Pure Go implementation (no CGO required)
  - Tables: `cron_tasks`, `cron_executions`, `cron_execution_logs`

**File Storage:**
- Local filesystem only
  - Logs: `logs/TradingEino.log.jsonl`
  - Database: `data/TradingEino.db`

**Caching:**
- None (in-memory only)

## Authentication & Identity

**Auth Provider:**
- Custom API key authentication (for OKX integration)
- No user authentication system detected
- No session management

## Monitoring & Observability

**Error Tracking:**
- None (logging only)

**Logs:**
- Custom structured logging in `internal/logger/`
- Output formats: JSONL, stdout, stderr, or file
- Log levels: debug, info, warn, error
- GORM SQL logging to separate file: `logs/gorm.log.jsonl`
- Gin HTTP logging to separate file: `logs/gin.log.jsonl`

## CI/CD & Deployment

**Hosting:**
- Self-hosted binary deployment
- No platform-specific integration detected

**CI Pipeline:**
- None detected in codebase

## Environment Configuration

**Required env vars (or config.yaml):**
```
# Server
TRADINGEINO_SERVER_MODE=debug|release|test
TRADINGEINO_SERVER_LISTEN_ON=0.0.0.0:10098

# Logger
TRADINGEINO_LOGGER_LEVEL=debug|info|warn|error
TRADINGEINO_LOGGER_OUTPUT=stdout|stderr|file
TRADINGEINO_LOGGER_FILE_PATH=/path/to/log.jsonl

# Database
TRADINGEINO_DB_TYPE=sqlite
TRADINGEINO_DB_DB_PATH=./data/TradingEino.db

# Scheduler
TRADINGEINO_SCHEDULER_ENABLED=true
TRADINGEINO_SCHEDULER_MAX_CONCURRENCY=5
TRADINGEINO_SCHEDULER_CHECK_INTERVAL=10
TRADINGEINO_SCHEDULER_DEFAULT_TIMEOUT=300

# Chat Model (AI)
TRADINGEINO_CHAT_MODEL_API_KEY=sk-xxx
TRADINGEINO_CHAT_MODEL_BASE_URL=https://coding.dashscope.aliyuncs.com/v1
TRADINGEINO_CHAT_MODEL_MODEL=qwen3.5-plus

# OKX Exchange
TRADINGEINO_OKX_API_KEY=xxx
TRADINGEINO_OKX_SECRET_KEY=xxx
TRADINGEINO_OKX_PASSPHRASE=xxx
TRADINGEINO_OKX_SANDBOX=false
```

**Secrets location:**
- Configuration file: `etc/config.yaml` (not git-tracked, example in `config.example.yaml`)
- Environment variables (preferred for production)

## Webhooks & Callbacks

**Incoming:**
- None detected

**Outgoing:**
- None detected

## OKX Integration Details

**REST API Endpoints Used:**
- Market data: Get candlesticks history (`/api/v5/market/candles`)
- Public endpoints for ticker, orderbook, trades

**WebSocket Channels:**
- Business channel for market data subscriptions
- Private channel for account/order updates (when trading enabled)

**Rate Limiting:**
- Client-side rate limiter: 10 requests/second for candlestick API
- Implemented via `golang.org/x/time/rate`

## AI Agent Integration

**Eino Agent Tools:**
- `OkxCandlesticksTool` - Fetches K-line data with technical indicators
  - Uses OKX REST API
  - Calculates TA-Lib indicators (MACD, RSI, Bollinger Bands, etc.)
  - Rate limited to 10 req/s

**Multi-Agent Architecture:**
- `OKXWatcher` (DeepAgent orchestrator) - `internal/agent/okx_watcher/`
- `RiskOfficer` (sub-agent) - `internal/agent/risk_officer/`
- `SentimentAnalyst` (sub-agent) - `internal/agent/sentiment_analyst/`

---

*Integration audit: 2026-03-24*
