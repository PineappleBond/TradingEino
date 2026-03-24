# Technology Stack

**Analysis Date:** 2026-03-24

## Languages

**Primary:**
- Go 1.26.1 - Primary backend language for all server-side logic

**Secondary:**
- None detected (pure Go backend)

## Runtime

**Environment:**
- Go 1.26.1 runtime

**Package Manager:**
- Go modules (go mod)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Gin v1.12.0 - HTTP web framework for REST API server
- Cloudwego Eino v0.8.4 - AI agent framework for multi-agent orchestration
- GORM v1.31.1 - ORM library for database operations

**Testing:**
- testify v1.11.1 - Assertion library and test utilities

**Build/Dev:**
- go build - Standard Go compilation (no additional build tools detected)

## Key Dependencies

**Critical:**
- `github.com/cloudwego/eino` v0.8.4 - AI agent framework for LLM orchestration
- `github.com/cloudwego/eino-ext/components/model/openai` v0.1.10 - OpenAI-compatible chat model client
- `github.com/gin-gonic/gin` v1.12.0 - Web framework
- `github.com/gorilla/websocket` v1.5.3 - WebSocket client for real-time connections
- `github.com/ncruces/go-sqlite3` v0.30.2 - SQLite3 driver (pure Go, no CGO)
- `gorm.io/gorm` v1.31.1 - ORM for database operations

**Infrastructure:**
- `github.com/spf13/viper` v1.21.0 - Configuration management with env var support
- `github.com/robfig/cron/v3` v3.0.1 - Cron scheduler for background jobs
- `github.com/shopspring/decimal` v1.4.0 - Precise decimal arithmetic for trading
- `github.com/markcheno/go-talib` v0.0.0 - Technical analysis library (TA-Lib Go wrapper)
- `github.com/chromedp/chromedp` v0.15.0 - Headless Chrome automation (local override in `pkg/chromedp-v0.15.0`)

**Utilities:**
- `golang.org/x/time` v0.15.0 - Rate limiter for API calls
- `github.com/sirupsen/logrus` v1.9.3 - Structured logging (used by Eino)

## Configuration

**Environment:**
- YAML configuration via Viper: `etc/config.yaml`
- Environment variable overrides: `TRADINGEINO_<SECTION>_<KEY>` format
- Example config: `etc/config.example.yaml`
- No `.env` files detected - config via YAML + env vars

**Build:**
- Standard Go build: `go build -o server ./cmd/server`
- No Makefile detected
- Compiled binaries: `main`, `server` in root directory

## Platform Requirements

**Development:**
- Go 1.26.1+
- Access to OKX API (or sandbox)
- Access to LLM API (DashScope/Aliyun or OpenAI-compatible)
- Filesystem access for SQLite database and logs

**Production:**
- Linux/macOS deployment target
- TCP port for HTTP server (default: 10098)
- Persistent storage for SQLite database (`data/TradingEino.db`)
- Log directory access (`logs/`)
- Frontend static files embedded via `//go:embed web/dist`

## Web Frontend

**Build:**
- Pre-built static files in `web/dist/`
- Embedded into Go binary via `web/embed.go`
- SPA (Single Page Application) architecture

---

*Stack analysis: 2026-03-24*
