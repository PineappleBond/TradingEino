# Codebase Structure

**Analysis Date:** 2026-03-24

## Directory Layout

```
backend/
├── cmd/                    # Application entry points
│   └── server/             # Main server binary
├── internal/               # Private application code (Go internal pattern)
│   ├── agent/              # Multi-agent AI system
│   │   ├── okx_watcher/    # Coordinator agent (DeepAgent)
│   │   ├── risk_officer/   # Risk analysis sub-agent
│   │   ├── sentiment_analyst/  # Sentiment analysis sub-agent
│   │   └── tools/          # Agent tools (OKX API wrappers)
│   ├── api/                # HTTP API layer
│   │   ├── handler/        # Request handlers
│   │   ├── middleware/     # Gin middleware
│   │   ├── request/        # Request DTOs
│   │   └── response/       # Response DTOs and utilities
│   ├── config/             # Configuration loading
│   ├── logger/             # Structured logging
│   ├── model/              # Domain models (GORM entities)
│   ├── notify/             # Notification interfaces
│   ├── repository/         # Data access layer
│   ├── server/             # HTTP server setup
│   ├── service/            # Business logic
│   │   └── scheduler/      # Task scheduling system
│   ├── svc/                # Service context (dependency container)
│   └── utils/              # Internal utilities
│       └── xmd/            # Markdown utilities
├── pkg/                    # Public library code
│   └── okex/               # OKX exchange API client
│       ├── api/            # REST and WebSocket clients
│       ├── models/         # OKX API response models
│       ├── requests/       # OKX API request models
│       └── responses/      # OKX API response parsers
├── web/                    # Frontend assets
│   └── dist/               # Built SPA (embedded via go:embed)
├── data/                   # Runtime data (SQLite database)
├── logs/                   # Log file output
├── etc/                    # Configuration files
│   └── config.yaml         # Application configuration
├── test/                   # Test utilities
└── .planning/              # Project planning documents
    ├── 02-optimization-eino-using/   # Eino optimization notes
    └── 03-multi-agent/               # Multi-agent architecture docs
```

## Directory Purposes

**cmd/server/:**
- Purpose: Application bootstrap code
- Contains: `main.go` with initialization sequence
- Key files: `cmd/server/main.go`

**internal/agent/:**
- Purpose: AI agent system using Cloudwego Eino framework
- Contains: Agent definitions, tool implementations, agent orchestration
- Key files: `internal/agent/agents.go` (agent container), `internal/agent/okx_watcher/agent.go` (coordinator)

**internal/api/:**
- Purpose: HTTP API implementation
- Contains: Route definitions, handlers, middleware, request/response types
- Key files: `internal/api/route.go` (routing), `internal/api/handler/cron_task_handler.go` (example handler)

**internal/config/:**
- Purpose: Configuration management
- Contains: Config struct definitions, Viper-based loading
- Key files: `internal/config/config.go`

**internal/logger/:**
- Purpose: Structured logging infrastructure
- Contains: slog wrapper with source tracking
- Key files: `internal/logger/logger.go`

**internal/model/:**
- Purpose: Domain entity definitions
- Contains: GORM models for database entities
- Key files: `internal/model/cron_task.go`, `internal/model/cron_execution.go`, `internal/model/cron_execution_log.go`

**internal/repository/:**
- Purpose: Data access abstraction
- Contains: Repository implementations for each entity
- Key files: `internal/repository/cron_task.go`, `internal/repository/cron_execution.go`

**internal/service/scheduler/:**
- Purpose: Task scheduling and execution
- Contains: Scheduler, executor, task handlers
- Key files: `internal/service/scheduler/scheduler.go`

**internal/svc/:**
- Purpose: Service context / dependency injection
- Contains: ServiceContext struct, initialization functions
- Key files: `internal/svc/servicecontext.go`, `internal/svc/database.go`, `internal/svc/okx.go`

**pkg/okex/:**
- Purpose: OKX exchange API client library
- Contains: REST client, WebSocket client, request/response models
- Key files: `pkg/okex/api/client.go`, `pkg/okex/api/api.go`

**web/:**
- Purpose: Frontend SPA assets (embedded)
- Contains: Built React/Vue/etc. application
- Key files: `web/embed.go` (Go embed directive)

**etc/:**
- Purpose: Configuration files
- Contains: YAML configuration templates
- Key files: `etc/config.yaml`

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Main application entry point
- `internal/server/server.go`: HTTP server initialization
- `internal/service/scheduler/scheduler.go`: Task scheduler

**Configuration:**
- `etc/config.yaml`: Application configuration
- `internal/config/config.go`: Config struct and loader

**Core Logic:**
- `internal/agent/agents.go`: Multi-agent system initialization
- `internal/service/scheduler/scheduler.go`: Task scheduling engine
- `internal/svc/servicecontext.go`: Dependency container

**Data Access:**
- `internal/repository/cron_task.go`: Task repository
- `internal/repository/cron_execution.go`: Execution repository
- `internal/repository/cron_execution_log.go`: Log repository

**API:**
- `internal/api/route.go`: Route definitions
- `internal/api/handler/*.go`: Request handlers
- `internal/api/response/response.go`: Response utilities

**External Integration:**
- `pkg/okex/api/client.go`: OKX API client
- `internal/agent/tools/okx_candlesticks.go`: Agent tools

**Models:**
- `internal/model/cron_task.go`: Task entity
- `internal/model/cron_execution.go`: Execution entity
- `internal/model/cron_execution_log.go`: Log entity

## Naming Conventions

**Files:**
- Snake case for Go files: `cron_task.go`, `service_context.go`
- Descriptive names matching content: `okx_candlesticks.go` for candlestick tool

**Directories:**
- Lowercase with underscores: `cron_task/`, `okx_watcher/`
- Plural for collections: `handlers/`, `tools/`, `models/`

**Types:**
- PascalCase for exported types: `CronTask`, `ServiceContext`
- camelCase for unexported: `serviceContext`, `taskHandler`

**Functions:**
- Constructor pattern: `NewServiceContext()`, `NewServer()`
- Repository pattern: `GetByID()`, `Create()`, `Update()`

## Where to Add New Code

**New API Endpoint:**
1. Add request struct to `internal/api/request/` (e.g., `internal/api/request/new_feature_request.go`)
2. Add response struct to `internal/api/response/` (e.g., `internal/api/response/new_feature_response.go`)
3. Create handler in `internal/api/handler/` (e.g., `internal/api/handler/new_feature_handler.go`)
4. Register route in `internal/api/route.go`

**New Agent Tool:**
1. Create tool in `internal/agent/tools/` (e.g., `internal/agent/tools/okx_new_feature_tool.go`)
2. Implement `Info()` returning `*schema.ToolInfo`
3. Implement `InvokableRun()` for execution
4. Register tool in agent's `ToolsConfig`

**New Repository:**
1. Add model to `internal/model/` (e.g., `internal/model/new_entity.go`)
2. Create repository in `internal/repository/` (e.g., `internal/repository/new_entity_repository.go`)
3. Add to service context in `internal/svc/servicecontext.go` if needed

**New Scheduler Task Handler:**
1. Implement `TaskHandler` interface in `internal/service/scheduler/handlers/`
2. Register handler with scheduler via `RegisterHandler()`

**New OKX API Wrapper:**
1. Add request model to `pkg/okex/requests/` subdirectory
2. Add response model to `pkg/okex/responses/` subdirectory
3. Add method to appropriate client in `pkg/okex/api/`

## Special Directories

**data/:**
- Purpose: SQLite database storage
- Generated: Yes (created on first run)
- Committed: No (in .gitignore)

**logs/:**
- Purpose: Log file output
- Generated: Yes (created on first run)
- Committed: No (in .gitignore)

**web/dist/:**
- Purpose: Built frontend SPA
- Generated: Yes (by frontend build process)
- Committed: Yes (embedded in binary via `go:embed`)

**.planning/:**
- Purpose: Project documentation and architecture notes
- Generated: Manually
- Committed: Yes

**test/:**
- Purpose: Test utilities and helpers
- Generated: Manually
- Committed: Yes

---

*Structure analysis: 2026-03-24*
