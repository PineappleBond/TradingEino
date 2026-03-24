# Testing Patterns

**Analysis Date:** 2026-03-24

## Test Framework

**Runner:**
- `testing` (Go standard library)
- `github.com/stretchr/testify/assert` for assertions

**Assertion Library:**
- testify/assert for assertions: `assert.NoError(t, err)`, `assert.Equal(t, expected, actual)`
- testify/require not currently used

**Run Commands:**
```bash
go test ./...                    # Run all tests
go test -v ./...                 # Run all tests with verbose output
go test -run TestName ./...      # Run specific test
go test -race ./...              # Run with race detector
go test -cover ./...             # Run with coverage
go test -coverprofile=cover.out  # Generate coverage profile
```

## Test File Organization

**Location:**
- Co-located with source files in same package
- Test files use `_test.go` suffix

**Directory Structure:**
```
internal/
├── model/
│   ├── cron_task.go
│   └── cron_task_test.go
├── repository/
│   ├── cron_task.go
│   └── cron_task_test.go
├── api/handler/
│   ├── cron_task_handler.go
│   └── api_test.go
└── service/scheduler/
    ├── scheduler.go
    └── scheduler_test.go
```

**Naming:**
- Test files: `<source>_test.go` (e.g., `cron_task_test.go`)
- Test functions: `Test<FunctionName>` or `Test<Scenario>` (e.g., `TestCronTaskHandler_ListTasks`)
- Table-driven test subtests: `t.Run(tt.name, ...)`

## Test Structure

**Package Declaration:**
- Tests use same package name as source (not `_test`):
  ```go
  package model      // Same as source
  package handler    // Same as source
  ```

**Test Function Pattern:**
```go
func TestCronTaskHandler_ListTasks(t *testing.T) {
    // 1. Setup
    svcCtx := newTestServiceContext(t)
    router := setupTestRouter(t, svcCtx)

    // 2. Create test data
    tasks := []model.CronTask{
        {Name: "Task 1", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http", Enabled: true},
    }
    for i := range tasks {
        createTestTask(t, svcCtx.DB, &tasks[i])
    }

    // 3. Define test cases
    tests := []struct {
        name          string
        url           string
        expectedLen   int
        expectedTotal int
    }{
        {
            name:          "default pagination",
            url:           "/api/crontask",
            expectedLen:   3,
            expectedTotal: 3,
        },
    }

    // 4. Run test cases
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := performRequest(router, "GET", tt.url, nil)
            resp := assertResponse(t, w, http.StatusOK)
            // ... assertions
        })
    }
}
```

**Table-Driven Tests:**
```go
tests := []struct {
    name     string
    status   TaskStatus
    want     string
}{
    {"pending", TaskStatusPending, "pending"},
    {"running", TaskStatusRunning, "running"},
    {"completed", TaskStatusCompleted, "completed"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        if string(tt.status) != tt.want {
            t.Errorf("TaskStatus(%v) = %v, want %v", tt.status, tt.status, tt.want)
        }
    })
}
```

**Setup/Teardown:**
```go
// Helper function for setup
func newTestServiceContext(t *testing.T) *svc.ServiceContext {
    t.Helper()
    db := newTestDB(t)
    return &svc.ServiceContext{DB: db}
}

func newTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")
    db, err := gorm.Open(gormlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open database: %v", err)
    }
    err = db.AutoMigrate(&model.CronTask{}, &model.CronExecution{}, &model.CronExecutionLog{})
    if err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }
    return db
}
```

**TestMain for Global Setup:**
```go
func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)
    m.Run()
}
```

## Mocking

**Pattern:** Manual mock implementations using interfaces

**Mock Handler Example:**
```go
// mockHandler 用于测试
type mockHandler struct {
    name      string
    execErr   error
    execCount int
    mu        sync.Mutex
}

func (h *mockHandler) Name() string {
    return h.name
}

func (h *mockHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
    h.mu.Lock()
    h.execCount++
    h.mu.Unlock()
    return h.execErr
}

func (h *mockHandler) GetExecCount() int {
    h.mu.Lock()
    defer h.mu.Unlock()
    return h.execCount
}
```

**Configurable Mock Handler:**
```go
// countingHandler 包装 mockHandler，允许自定义 Execute 逻辑
type countingHandler struct {
    name        string
    execCount   int
    mu          sync.Mutex
    executeFunc func(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error
}

func newCountingHandler(name string, executeFunc func(...) error) *countingHandler {
    return &countingHandler{
        name:        name,
        executeFunc: executeFunc,
    }
}
```

**What to Mock:**
- External services (OKX API, chat models)
- Slow operations (database - use in-memory SQLite)
- Time-dependent operations (use controlled time values)
- Concurrent operations (use controlled handlers)

**What NOT to Mock:**
- Business logic being tested
- Repository layer when testing handlers (use real in-memory DB)

## Fixtures and Factories

**Helper Functions:**
```go
// createTestTask creates a test task in the database
func createTestTask(t *testing.T, db *gorm.DB, task *model.CronTask) uint {
    t.Helper()
    if err := db.Create(task).Error; err != nil {
        t.Fatalf("failed to create test task: %v", err)
    }
    return task.ID
}

// createTestExecution creates a test execution in the database
func createTestExecution(t *testing.T, db *gorm.DB, execution *model.CronExecution) uint {
    t.Helper()
    if err := db.Create(execution).Error; err != nil {
        t.Fatalf("failed to create test execution: %v", err)
    }
    return execution.ID
}
```

**SQL NullTime Helper:**
```go
func sqlNullTime(t time.Time) sql.NullTime {
    return sql.NullTime{Time: t, Valid: true}
}
```

**HTTP Test Helpers:**
```go
// performRequest performs an HTTP request on the router
func performRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
    var w *httptest.ResponseRecorder
    var req *http.Request
    if body != nil {
        jsonBody, _ := json.Marshal(body)
        req = httptest.NewRequest(method, path, bytes.NewReader(jsonBody))
        req.Header.Set("Content-Type", "application/json")
    } else {
        req = httptest.NewRequest(method, path, nil)
    }
    w = httptest.NewRecorder()
    r.ServeHTTP(w, req)
    return w
}

// assertResponse asserts the response code and body
func assertResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) *response.Response[any] {
    t.Helper()
    if w.Code != expectedCode {
        t.Errorf("expected status %d, got %d. Body: %s", expectedCode, w.Code, w.Body.String())
    }
    var resp response.Response[any]
    if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }
    return &resp
}
```

**Location:**
- Helper functions defined at top of test files
- Shared helpers in test files where multiple tests need them

## Coverage

**Requirements:** No explicit coverage threshold enforced

**View Coverage:**
```bash
go test -cover ./...                          # Show coverage percentage
go test -coverprofile=cover.out ./...         # Generate profile
go tool cover -html=cover.out -o cover.html   # View HTML report
```

## Test Types

**Unit Tests:**
- Test individual functions and methods
- Test model methods: `TestTaskStatus_Values`, `TestCronTask_TableName`
- Test repository methods with in-memory DB
- Test handler methods with mock HTTP requests

**Integration Tests:**
- API handler tests in `api_test.go` test full HTTP flow
- Scheduler tests test complete task execution flow
- Use in-memory SQLite for database integration

**E2E Tests:**
- Not currently implemented
- Would require external dependencies (OKX API, etc.)

## Common Patterns

**Async Testing:**
```go
func TestScenario_DirectTaskExecution(t *testing.T) {
    svcCtx := setupTestServiceContext(t)
    scheduler := NewScheduler(svcCtx)
    // ... setup

    // Direct call to trigger
    scheduler.onTaskTrigger(task)

    // Wait for async execution
    time.Sleep(500 * time.Millisecond)

    // Verify
    assert.GreaterOrEqual(t, handler.GetExecCount(), 1)
}
```

**Concurrency Testing:**
```go
func TestScenario_ConcurrencyLimit(t *testing.T) {
    var mu sync.Mutex
    maxConcurrent := 0
    currentConcurrent := 0

    slowHandler := newCountingHandler("slow_handler", func(...) error {
        mu.Lock()
        currentConcurrent++
        if currentConcurrent > maxConcurrent {
            maxConcurrent = currentConcurrent
        }
        mu.Unlock()

        time.Sleep(200 * time.Millisecond)

        mu.Lock()
        currentConcurrent--
        mu.Unlock()
        return nil
    })

    // ... run concurrent tasks

    assert.LessOrEqual(t, maxConcurrent, 2)
}
```

**Error Testing:**
```go
func TestCronTaskHandler_GetTask(t *testing.T) {
    tests := []struct {
        name         string
        url          string
        expectedCode int
    }{
        {
            name:         "task not found",
            url:          "/api/crontask/99999",
            expectedCode: http.StatusNotFound,
        },
        {
            name:         "invalid id format",
            url:          "/api/crontask/abc",
            expectedCode: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := performRequest(router, "GET", tt.url, nil)
            assertResponse(t, w, tt.expectedCode)
        })
    }
}
```

**t.Helper() Usage:**
- Used in all helper functions to improve error reporting
- Points errors to actual test code, not helper code

**Context Testing:**
```go
func TestScenario_Timeout(t *testing.T) {
    timeoutHandler := newCountingHandler("timeout_handler", func(ctx context.Context, ...) error {
        select {
        case <-time.After(2 * time.Second):
            return nil
        case <-ctx.Done():
            return ctx.Err() // Timeout cancellation
        }
    })
    // ... test with 1 second timeout
}
```

## Test Files Summary

| File | Purpose | Test Count |
|------|---------|------------|
| `internal/model/model_test.go` | Model validation tests | Multiple |
| `internal/model/cron_task_test.go` | CronTask model tests | 6 |
| `internal/model/cron_execution_test.go` | CronExecution tests | Multiple |
| `internal/model/cron_execution_log_test.go` | CronExecutionLog tests | Multiple |
| `internal/repository/repository_test.go` | Repository helpers | N/A |
| `internal/repository/cron_task_test.go` | Repository CRUD tests | Multiple |
| `internal/api/handler/api_test.go` | API handler integration tests | 15+ |
| `internal/service/scheduler/scheduler_test.go` | Scheduler tests | 10+ |
| `internal/config/config_test.go` | Configuration tests | 7 |
| `internal/logger/logger_test.go` | Logger tests | 15+ |

---

*Testing analysis: 2026-03-24*
