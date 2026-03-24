# Architecture

**Analysis Date:** 2026-03-24

## Pattern Overview

**Overall:** Layered Architecture with Multi-Agent System

**Key Characteristics:**
- Clean separation of concerns across distinct layers
- Service Context pattern for dependency injection
- Repository pattern for data access abstraction
- Multi-Agent architecture using Cloudwego Eino framework
- Scheduler-based task execution with concurrency control

## Layers

**Entry Layer (`cmd/server/`):**
- Purpose: Application bootstrap and initialization
- Location: `cmd/server/main.go`
- Contains: Configuration loading, logger setup, agent initialization, scheduler startup
- Depends on: `internal/config`, `internal/logger`, `internal/svc`, `internal/agent`, `internal/server`
- Used by: None (application entry point)

**Presentation Layer (`internal/server/`, `internal/api/`):**
- Purpose: HTTP server and API routing
- Location: `internal/server/server.go`, `internal/api/route.go`
- Contains: Gin engine setup, middleware (CORS, logging), route definitions, request handlers
- Depends on: `internal/svc`, `internal/api/handler`, `internal/api/middleware`
- Used by: External clients (web frontend, API consumers)

**Handler Layer (`internal/api/handler/`):**
- Purpose: Request processing and response formatting
- Location: `internal/api/handler/cron_task_handler.go`
- Contains: HTTP handlers that bind requests, call services/repositories, return responses
- Depends on: `internal/repository`, `internal/api/request`, `internal/api/response`
- Used by: `internal/server` via route handlers

**Service/Agent Layer (`internal/agent/`, `internal/service/`):**
- Purpose: Business logic and AI agent orchestration
- Location: `internal/agent/agents.go`, `internal/service/scheduler/scheduler.go`
- Contains: Multi-agent system (OKXWatcher, RiskOfficer, SentimentAnalyst), task scheduling
- Depends on: `internal/svc`, `internal/agent/tools`, `pkg/okex`
- Used by: Entry layer and scheduler

**Repository Layer (`internal/repository/`):**
- Purpose: Data access abstraction
- Location: `internal/repository/cron_task.go`
- Contains: CRUD operations for domain entities using GORM
- Depends on: `internal/svc`, `internal/model`
- Used by: Handler layer and service layer

**Model Layer (`internal/model/`):**
- Purpose: Domain entity definitions
- Location: `internal/model/cron_task.go`, `internal/model/cron_execution.go`
- Contains: GORM models with business semantics
- Depends on: GORM, standard library
- Used by: All layers

**Infrastructure Layer (`pkg/okex/`, `internal/svc/`):**
- Purpose: External integrations and shared services
- Location: `pkg/okex/api/client.go`, `internal/svc/servicecontext.go`
- Contains: OKX API client, database connection, chat model, service context
- Depends on: External SDKs (Eino, GORM, OKX API)
- Used by: All internal layers

## Data Flow

**HTTP Request Flow:**

1. Request arrives at Gin server (`internal/server/server.go`)
2. Middleware processes request (CORS, logging)
3. Router matches path to handler (`internal/api/route.go`)
4. Handler binds request to struct (`internal/api/request/`)
5. Handler calls repository or agent
6. Repository queries database via GORM
7. Response formatted (`internal/api/response/`)
8. JSON response returned to client

**Scheduler Task Flow:**

1. Scheduler loads enabled tasks from database (`internal/service/scheduler/scheduler.go:loadTasks`)
2. Tasks registered with cron parser (6-field cron expressions)
3. On trigger, `onTaskTrigger` creates execution record
4. Concurrency control via semaphore channel
5. Handler executes task with timeout context
6. Execution result logged to `cron_execution_log`
7. Task status updated in database

**Agent Invocation Flow:**

1. OKXWatcher (DeepAgent) receives prompt from scheduler
2. Sub-agents invoked based on task requirements:
   - TechnoAgent for technical analysis
   - SentimentAnalyst for funding rate analysis
   - RiskOfficer for position risk assessment
3. Each agent uses assigned tools (e.g., `okx-candlesticks-tool`)
4. Tools call OKX API via `pkg/okex/api/client.go`
5. Results aggregated and returned as analysis

## Key Abstractions

**ServiceContext (`internal/svc/servicecontext.go`):**
- Purpose: Centralized dependency container
- Examples: `svc.ServiceContext`
- Pattern: Composition of Config, DB, Logger, ChatModel, OKXClient

**CronTask (`internal/model/cron_task.go`):**
- Purpose: Scheduled task definition
- Examples: One-time tasks, recurring tasks
- Pattern: Domain entity with status lifecycle (Pending -> Running -> Completed/Failed)

**TaskHandler (`internal/service/scheduler/scheduler.go`):**
- Purpose: Pluggable task execution interface
- Examples: `OkxWatcherHandler`
- Pattern: Strategy pattern for different task types

**AgentsModel (`internal/agent/agents.go`):**
- Purpose: Multi-agent system container
- Examples: OKXWatcher (coordinator), RiskOfficer, SentimentAnalyst
- Pattern: Composite of Eino agents with shared context

**Tool (`internal/agent/tools/`):**
- Purpose: Atomic functions exposed to LLM agents
- Examples: `OkxCandlesticksTool`, `OkxGetPositionsTool`
- Pattern: Tool interface with `Info()` and `InvokableRun()` methods

## Entry Points

**Main Application (`cmd/server/main.go`):**
- Location: `cmd/server/main.go`
- Triggers: Direct execution or process spawn
- Responsibilities: Config load, logger init, DB migration, agent setup, server start

**Scheduler (`internal/service/scheduler/scheduler.go`):**
- Location: `internal/service/scheduler/scheduler.go:Start`
- Triggers: Application startup
- Responsibilities: Task loading, cron registration, execution dispatch

**HTTP Server (`internal/server/server.go`):**
- Location: `internal/server/server.go:Start`
- Triggers: Application startup
- Responsibilities: Gin engine initialization, route registration, HTTP listening

## Error Handling

**Strategy:** Centralized error handling with structured logging

**Patterns:**
- Repository layer returns raw errors to handlers
- Handlers convert to standardized response codes (`internal/api/response/response.go:26-63`)
- Service layer uses context with timeout for task execution
- Logger supports stack trace capture for errors (`internal/logger/logger.go:227-258`)
- Error codes categorized: 1xxx auth, 2xxx parameter, 3xxx business, 4xxx system

## Cross-Cutting Concerns

**Logging:** Custom slog-based logger in `internal/logger/logger.go`
- JSON output format
- Source file/line injection
- Global default logger pattern
- Separate log files for Gin and GORM

**Validation:**
- Request binding via Gin's `ShouldBindJSON`, `ShouldBindQuery`, `ShouldBindUri`
- Business validation in handlers before repository calls

**Authentication:** Not implemented (API is open, relies on network isolation)

**Configuration:** Viper-based in `internal/config/config.go`
- YAML file loading
- Environment variable override (TRADINGEINO_ prefix)
- Default values for all settings

**Database:** SQLite via GORM with ncruces/go-sqlite3 (pure Go, no CGO)
- Auto-migration on startup
- Soft delete support
- Row-level locking for concurrent task execution

---

*Architecture analysis: 2026-03-24*
