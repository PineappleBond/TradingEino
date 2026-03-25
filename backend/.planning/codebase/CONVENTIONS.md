# Coding Conventions

**Analysis Date:** 2026-03-24

## Naming Patterns

**Files:**
- Lowercase with underscores for multi-word files: `cron_task.go`, `cron_execution_log.go`
- Test files append `_test.go`: `cron_task_test.go`, `scheduler_test.go`
- Request/response files use descriptive suffixes: `cron_task_request.go`, `cron_task_response.go`

**Functions:**
- PascalCase for exported functions: `NewCronTaskHandler()`, `Load()`
- camelCase for unexported functions: `newTestDB()`, `setupTestRouter()`
- Constructor pattern: `New<Service/Repository/Handler>()` returns pointer to struct

**Variables:**
- camelCase for local variables: `svcCtx`, `cfg`, `taskID`
- Descriptive names: `executionTimes`, `maxConcurrent`
- Short loop variables acceptable: `i`, `err`, `task`

**Types:**
- PascalCase for exported types: `CronTask`, `TaskStatus`, `Scheduler`
- camelCase for struct fields: `MaxRetries`, `NextExecutionAt`
- Type aliases use PascalCase: `TaskStatus`, `TaskType`

**Constants:**
- PascalCase with descriptive prefixes for grouped constants:
  ```go
  // TaskStatus constants
  const (
      TaskStatusPending   TaskStatus = "pending"
      TaskStatusRunning   TaskStatus = "running"
      TaskStatusCompleted TaskStatus = "completed"
  )
  ```

## Code Style

**Formatting:**
- Standard `gofmt` formatting (4-space indentation from gofmt)
- Line length: No explicit limit, but functions are kept readable
- Import groups: Standard library, blank line, third-party, blank line, internal packages

**Linting:**
- No explicit linter configuration detected
- Code follows standard Go conventions

**Import Organization:**
```go
import (
    // Standard library first
    "context"
    "database/sql"
    "errors"
    "fmt"
    "sync"
    "time"

    // Third-party packages
    "github.com/gin-gonic/gin"
    "github.com/robfig/cron/v3"
    "github.com/spf13/viper"
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"

    // Internal packages (project-specific)
    "github.com/PineappleBond/TradingEino/backend/internal/config"
    "github.com/PineappleBond/TradingEino/backend/internal/logger"
    "github.com/PineappleBond/TradingEino/backend/internal/model"
)
```

## Error Handling

**Patterns:**
- Return errors as last return value: `func Create(...) error`
- Use `errors.New()` for simple errors: `errors.New("scheduler is not running")`
- Use `fmt.Errorf()` for formatted errors: `fmt.Errorf("failed to load config: %w", err)`
- Error wrapping with `%w` for error chains: `fmt.Errorf("failed to read config file: %w", err)`
- Check errors immediately after function calls:
  ```go
  task, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
  if err != nil {
      ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
      return
  }
  ```

**Error Types:**
- Custom error types for specific scenarios:
  ```go
  // RetryableError 可重试错误接口
  type RetryableError interface {
      error
      IsRetryable() bool
  }

  // RetryableErrorImpl 可重试错误实现
  type RetryableErrorImpl struct {
      Err error
  }
  ```

**Error Handling in Handlers:**
- Map errors to HTTP status codes via response codes:
  ```go
  ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
  ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
  ```

## Logging

**Framework:** Custom wrapper around `log/slog` with JSON output

**Location:** `internal/logger/logger.go`

**Patterns:**
- Context-aware logging: `logger.Info(ctx, "message", "key", value)`
- Source location tracking: `AddSource: true` in config
- Stack traces for errors: `logger.Error(ctx, "message", err, "key", value)`
- Convenience functions:
  ```go
  logger.Info(ctx, "Scheduler: started")
  logger.Warn(ctx, "Scheduler: task already running, skipping", "task_id", task.ID)
  logger.Error(ctx, "Scheduler: failed to get due tasks", err)
  ```

**Log Levels:**
- `Debug`: Detailed debugging information
- `Info`: General operational messages
- `Warn`: Warning conditions
- `Error`: Error conditions with error object

**Usage Pattern:**
```go
logger.Info(ctx, "Executor: starting execution",
    "task_id", task.ID, "task_name", task.Name, "execution_id", execution.ID)
```

## Comments

**When to Comment:**
- Package-level documentation: `// Package okex is generally a golang Api wrapper...`
- Type/constant explanations in model files:
  ```go
  // TaskStatus 定时任务状态
  type TaskStatus string

  const (
      TaskStatusPending   TaskStatus = "pending"   // 待执行
      TaskStatusRunning   TaskStatus = "running"   // 执行中
  )
  ```
- Function purpose for exported functions via comments above

**JSDoc/TSDoc:**
- Not used (Go codebase)
- Go doc comments used for exported symbols

## Function Design

**Size:**
- Single responsibility principle
- Most functions under 50 lines
- Complex functions broken into helper methods

**Parameters:**
- Use context as first parameter for functions that need it: `func Execute(ctx context.Context, ...)`
- Use pointer to struct for multiple related parameters: `func Create(ctx context.Context, task *model.CronTask) error`
- Named parameters in struct literals for clarity

**Return Values:**
- Single error return for operations that can fail
- Multiple returns: `(result Type, err error)`
- Pointer returns for database operations: `(*model.CronTask, error)`

## Module Design

**Exports:**
- Exported symbols follow Go conventions (PascalCase = exported)
- Internal packages in `internal/` directory (not importable outside module)
- Public packages in `pkg/` directory (importable by external modules)

**Package Structure:**
```
internal/
├── api/        # HTTP handlers, routes, middleware
├── config/     # Configuration loading
├── logger/     # Logging wrapper
├── model/      # Data models
├── repository/ # Database access layer
├── scheduler/  # Task scheduling
├── server/     # HTTP server setup
├── service/    # Business logic
└── svc/        # Service context (dependency container)
```

**Dependency Injection:**
- ServiceContext pattern for dependency management:
  ```go
  type ServiceContext struct {
      Config    config.Config
      Logger4Gin *logger.Logger
      DB        *gorm.DB
      ChatModel *openai.ChatModel
      OKXClient *api.Client
  }

  func NewCronTaskHandler(svcCtx *svc.ServiceContext) *CronTaskHandler {
      return &CronTaskHandler{
          svcCtx:     svcCtx,
          repository: repository.NewCronTaskRepository(svcCtx),
      }
  }
  ```

## API Response Pattern

**Unified Response Format:**
```go
// Response represents the unified API response format.
// Format: {"code": 0, "message": "success", "data": {...}}
type Response[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}
```

**Error Code Categories:**
- 1xxx: Authentication errors
- 2xxx: Parameter errors
- 3xxx: Business errors
- 4xxx: System errors

**Helper Functions:**
```go
func Success[T any](data T) *Response[T]
func Error[T any](code int, message string) *Response[T]
```

## Configuration

**Pattern:** YAML configuration with environment variable override

**Location:** `internal/config/config.go`

**Usage:**
```go
cfg, err := config.Load(*configPath)
// Environment variables take precedence: TRADINGEINO_LOGGER_LEVEL=debug
```

**Conventions:**
- Struct tags use `mapstructure` for viper binding
- Nested config sections map to nested structs
- Environment variable format: `TRADINGEINO_<SECTION>_<KEY>`

---

*Convention analysis: 2026-03-24*
