package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

// newTestServiceContext creates a ServiceContext with in-memory SQLite for testing
func newTestServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()

	db := newTestDB(t)

	cfg := config.DefaultConfig()
	cfg.Logger.FilePath = filepath.Join(t.TempDir(), "test.jsonl")

	log4Gin := logger.New(config.LoggerConfig{
		Level:     cfg.Logger.Level,
		Output:    cfg.Logger.Output,
		FilePath:  cfg.Logger.FilePath,
		AddSource: cfg.Logger.AddSource,
	}, 5)

	return &svc.ServiceContext{
		Config:     *cfg,
		Logger4Gin: log4Gin,
		DB:         db,
		// ChatModel and OKXClient are nil in tests - handlers don't use them
	}
}

// newTestDB creates an in-memory SQLite database for testing
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := gorm.Open(gormlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	err = db.AutoMigrate(
		&model.CronTask{},
		&model.CronExecution{},
		&model.CronExecutionLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

// setupTestRouter creates a test router with all handlers
func setupTestRouter(t *testing.T, svcCtx *svc.ServiceContext) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register handlers
	healthHandler := NewHealthCheckHandler(svcCtx)
	cronTaskHandler := NewCronTaskHandler(svcCtx)
	cronExecutionHandler := NewCronExecutionHandler(svcCtx)
	cronExecutionLogHandler := NewCronExecutionLogHandler(svcCtx)

	// Health check
	router.GET("/api/health", healthHandler.HealthCheck)

	// CronTask routes
	router.GET("/api/crontask", cronTaskHandler.ListTasks)
	router.GET("/api/crontask/:id", cronTaskHandler.GetTask)
	router.POST("/api/crontask", cronTaskHandler.CreateTask)
	router.PUT("/api/crontask/:id", cronTaskHandler.UpdateTask)
	router.DELETE("/api/crontask/:id", cronTaskHandler.DeleteTask)
	router.POST("/api/crontask/:id/enable", cronTaskHandler.EnableTask)
	router.POST("/api/crontask/:id/disable", cronTaskHandler.DisableTask)
	router.POST("/api/crontask/:id/start", cronTaskHandler.StartTask)
	router.POST("/api/crontask/:id/stop", cronTaskHandler.StopTask)

	// CronExecution routes
	router.GET("/api/cronexecution", cronExecutionHandler.ListExecutions)
	router.GET("/api/cronexecution/:id", cronExecutionHandler.GetExecution)
	router.GET("/api/cronexecution/task/:task_id", cronExecutionHandler.GetByTaskID)

	// CronExecutionLog routes
	router.GET("/api/cronexecutionlog", cronExecutionLogHandler.ListLogs)
	router.GET("/api/cronexecutionlog/:id", cronExecutionLogHandler.GetLog)
	router.GET("/api/cronexecutionlog/execution/:execution_id", cronExecutionLogHandler.GetByExecutionID)

	return router
}

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

// createTestLog creates a test log in the database
func createTestLog(t *testing.T, db *gorm.DB, log *model.CronExecutionLog) uint {
	t.Helper()

	if err := db.Create(log).Error; err != nil {
		t.Fatalf("failed to create test log: %v", err)
	}
	return log.ID
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// ==================== Health Check Tests ====================

func TestHealthCheckHandler_HealthCheck(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	w := performRequest(router, "GET", "/api/health", nil)

	resp := assertResponse(t, w, http.StatusOK)
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message 'success', got '%s'", resp.Message)
	}
	if resp.Data != "ok" {
		t.Errorf("expected data 'ok', got '%v'", resp.Data)
	}
}

// ==================== CronTask Tests ====================

func TestCronTaskHandler_ListTasks(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test tasks
	tasks := []model.CronTask{
		{Name: "Task 1", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http", Enabled: true, Status: model.TaskStatusPending},
		{Name: "Task 2", Spec: "0 0 * * *", Type: model.TaskTypeOnce, ExecType: "script", Enabled: false, Status: model.TaskStatusRunning},
		{Name: "Task 3", Spec: "*/5 * * * *", Type: model.TaskTypeRecurring, ExecType: "http", Enabled: true, Status: model.TaskStatusPending},
	}
	for i := range tasks {
		createTestTask(t, svcCtx.DB, &tasks[i])
	}

	tests := []struct {
		name         string
		url          string
		expectedLen  int
		expectedTotal int
	}{
		{
			name:          "default pagination",
			url:           "/api/crontask",
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "with page size",
			url:           "/api/crontask?page=1&pageSize=2",
			expectedLen:   2,
			expectedTotal: 3,
		},
		{
			name:          "filter by enabled",
			url:           "/api/crontask?enabled=true",
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "filter by disabled",
			url:           "/api/crontask?enabled=false",
			expectedLen:   1,
			expectedTotal: 1,
		},
		{
			name:          "filter by status",
			url:           "/api/crontask?status=pending",
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "filter by type",
			url:           "/api/crontask?type=recurring",
			expectedLen:   2,
			expectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "GET", tt.url, nil)
			resp := assertResponse(t, w, http.StatusOK)

			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected data format: %T", resp.Data)
			}

			items, ok := data["items"].([]interface{})
			if !ok {
				t.Fatalf("unexpected items format: %T", data["items"])
			}

			page, ok := data["page"].(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected page format: %T", data["page"])
			}

			if len(items) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(items))
			}

			total := int(page["total"].(float64))
			if total != tt.expectedTotal {
				t.Errorf("expected %d total, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestCronTaskHandler_GetTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test task
	task := &model.CronTask{
		Name:     "Test Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  true,
		Status:   model.TaskStatusPending,
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid task",
			url:          fmt.Sprintf("/api/crontask/%d", taskID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid id format",
			url:          "/api/crontask/abc",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "task not found",
			url:          "/api/crontask/99999",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "id less than 1",
			url:          "/api/crontask/0",
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

func TestCronTaskHandler_CreateTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	tests := []struct {
		name         string
		body         interface{}
		expectedCode int
	}{
		{
			name: "valid recurring task",
			body: map[string]interface{}{
				"name":       "New Task",
				"spec":       "0 * * * *",
				"type":       "recurring",
				"exec_type":  "http",
				"enabled":    true,
				"max_retries": 3,
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "valid once task",
			body: map[string]interface{}{
				"name":       "Once Task",
				"type":       "once",
				"exec_type":  "script",
				"enabled":    false,
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "missing required name",
			body: map[string]interface{}{
				"spec":      "0 * * * *",
				"type":      "recurring",
				"exec_type": "http",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid type",
			body: map[string]interface{}{
				"name":       "Invalid Task",
				"type":       "invalid",
				"exec_type":  "http",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty body",
			body:         nil,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "POST", "/api/crontask", tt.body)
			resp := assertResponse(t, w, tt.expectedCode)

			if tt.expectedCode == http.StatusOK {
				if resp.Code != 0 {
					t.Errorf("expected code 0, got %d", resp.Code)
				}
				data, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Fatalf("unexpected data format: %T", resp.Data)
				}
				if data["name"] == nil {
					t.Error("expected name in response")
				}
			}
		})
	}
}

func TestCronTaskHandler_UpdateTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test task
	task := &model.CronTask{
		Name:     "Original Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  true,
		Status:   model.TaskStatusPending,
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	tests := []struct {
		name         string
		url          string
		body         interface{}
		expectedCode int
	}{
		{
			name: "valid update",
			url:  fmt.Sprintf("/api/crontask/%d", taskID),
			body: map[string]interface{}{
				"name": "Updated Task",
				"spec": "*/5 * * * *",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "task not found",
			url:          "/api/crontask/99999",
			body:         map[string]interface{}{"name": "New Name"},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid id format",
			url:          "/api/crontask/abc",
			body:         map[string]interface{}{"name": "New Name"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "id less than 1",
			url:          "/api/crontask/0",
			body:         map[string]interface{}{"name": "New Name"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "PUT", tt.url, tt.body)
			resp := assertResponse(t, w, tt.expectedCode)

			if tt.expectedCode == http.StatusOK {
				data, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Fatalf("unexpected data format: %T", resp.Data)
				}
				if data["name"] != "Updated Task" {
					t.Errorf("expected updated name 'Updated Task', got '%v'", data["name"])
				}
			}
		})
	}
}

func TestCronTaskHandler_DeleteTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test task
	task := &model.CronTask{
		Name:     "To Delete",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid delete",
			url:          fmt.Sprintf("/api/crontask/%d", taskID),
			expectedCode: http.StatusOK,
		},
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
			w := performRequest(router, "DELETE", tt.url, nil)
			assertResponse(t, w, tt.expectedCode)
		})
	}
}

func TestCronTaskHandler_EnableTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create disabled task
	task := &model.CronTask{
		Name:     "Disabled Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  false,
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid enable",
			url:          fmt.Sprintf("/api/crontask/%d/enable", taskID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "task not found",
			url:          "/api/crontask/99999/enable",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid id format",
			url:          "/api/crontask/abc/enable",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "POST", tt.url, nil)
			resp := assertResponse(t, w, tt.expectedCode)

			if tt.expectedCode == http.StatusOK {
				if resp.Data != "enabled" {
					t.Errorf("expected data 'enabled', got '%v'", resp.Data)
				}
			}
		})
	}
}

func TestCronTaskHandler_DisableTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create enabled task
	task := &model.CronTask{
		Name:     "Enabled Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  true,
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid disable",
			url:          fmt.Sprintf("/api/crontask/%d/disable", taskID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "task not found",
			url:          "/api/crontask/99999/disable",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid id format",
			url:          "/api/crontask/abc/disable",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "POST", tt.url, nil)
			resp := assertResponse(t, w, tt.expectedCode)

			if tt.expectedCode == http.StatusOK {
				if resp.Data != "disabled" {
					t.Errorf("expected data 'disabled', got '%v'", resp.Data)
				}
			}
		})
	}
}

func TestCronTaskHandler_StartTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test task
	task := &model.CronTask{
		Name:     "Stopped Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  false,
		Status:   model.TaskStatusStopped,
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	nextExecTime := time.Now().Add(time.Hour).Format("2006-01-02 15:04:05")

	tests := []struct {
		name         string
		url          string
		body         interface{}
		expectedCode int
	}{
		{
			name: "valid start",
			url:  fmt.Sprintf("/api/crontask/%d/start", taskID),
			body: map[string]interface{}{
				"next_execution_time": nextExecTime,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "missing next_execution_time",
			url:          fmt.Sprintf("/api/crontask/%d/start", taskID),
			body:         nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid time format",
			url:  fmt.Sprintf("/api/crontask/%d/start", taskID),
			body: map[string]interface{}{
				"next_execution_time": "invalid-time",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "task not found",
			url:          "/api/crontask/99999/start",
			body:         map[string]interface{}{"next_execution_time": nextExecTime},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "POST", tt.url, tt.body)
			resp := assertResponse(t, w, tt.expectedCode)

			if tt.expectedCode == http.StatusOK {
				if resp.Data != "started" {
					t.Errorf("expected data 'started', got '%v'", resp.Data)
				}
			}
		})
	}
}

func TestCronTaskHandler_StopTask(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create running task
	task := &model.CronTask{
		Name:     "Running Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  true,
		Status:   model.TaskStatusRunning,
	}
	taskID := createTestTask(t, svcCtx.DB, task)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid stop",
			url:          fmt.Sprintf("/api/crontask/%d/stop", taskID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "task not found",
			url:          "/api/crontask/99999/stop",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid id format",
			url:          "/api/crontask/abc/stop",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "POST", tt.url, nil)
			resp := assertResponse(t, w, tt.expectedCode)

			if tt.expectedCode == http.StatusOK {
				if resp.Data != "stopped" {
					t.Errorf("expected data 'stopped', got '%v'", resp.Data)
				}
			}
		})
	}
}

// ==================== CronExecution Tests ====================

func TestCronExecutionHandler_ListExecutions(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create a test task first
	task := &model.CronTask{
		Name:     "Parent Task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		ExecType: "http",
		Enabled:  true,
	}
	createTestTask(t, svcCtx.DB, task)

	// Create test executions
	executions := []model.CronExecution{
		{TaskID: task.ID, Status: model.ExecutionStatusSuccess, StartedAt: sqlNullTime(time.Now().Add(-2 * time.Hour))},
		{TaskID: task.ID, Status: model.ExecutionStatusRunning, StartedAt: sqlNullTime(time.Now().Add(-1 * time.Hour))},
		{TaskID: task.ID, Status: model.ExecutionStatusFailed, StartedAt: sqlNullTime(time.Now())},
	}
	for i := range executions {
		createTestExecution(t, svcCtx.DB, &executions[i])
	}

	tests := []struct {
		name          string
		url           string
		expectedLen   int
		expectedTotal int
	}{
		{
			name:          "default pagination",
			url:           "/api/cronexecution",
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "with page size",
			url:           "/api/cronexecution?page=1&pageSize=2",
			expectedLen:   2,
			expectedTotal: 3,
		},
		{
			name:          "filter by task_id",
			url:           fmt.Sprintf("/api/cronexecution?task_id=%d", task.ID),
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "filter by status",
			url:           "/api/cronexecution?status=success",
			expectedLen:   1,
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "GET", tt.url, nil)
			resp := assertResponse(t, w, http.StatusOK)

			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected data format: %T", resp.Data)
			}

			items, ok := data["items"].([]interface{})
			if !ok {
				t.Fatalf("unexpected items format: %T", data["items"])
			}

			page, ok := data["page"].(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected page format: %T", data["page"])
			}

			if len(items) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(items))
			}

			total := int(page["total"].(float64))
			if total != tt.expectedTotal {
				t.Errorf("expected %d total, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestCronExecutionHandler_GetExecution(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test execution
	task := &model.CronTask{Name: "Test Task", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http"}
	createTestTask(t, svcCtx.DB, task)

	execution := &model.CronExecution{
		TaskID:    task.ID,
		Status:    model.ExecutionStatusSuccess,
		StartedAt: sqlNullTime(time.Now()),
	}
	executionID := createTestExecution(t, svcCtx.DB, execution)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid execution",
			url:          fmt.Sprintf("/api/cronexecution/%d", executionID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid id format",
			url:          "/api/cronexecution/abc",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "execution not found",
			url:          "/api/cronexecution/99999",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "id less than 1",
			url:          "/api/cronexecution/0",
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

func TestCronExecutionHandler_GetByTaskID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test task and executions
	task := &model.CronTask{Name: "Parent Task", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http"}
	taskID := createTestTask(t, svcCtx.DB, task)

	executions := []model.CronExecution{
		{TaskID: task.ID, Status: model.ExecutionStatusSuccess, StartedAt: sqlNullTime(time.Now().Add(-2 * time.Hour))},
		{TaskID: task.ID, Status: model.ExecutionStatusRunning, StartedAt: sqlNullTime(time.Now().Add(-1 * time.Hour))},
	}
	for i := range executions {
		createTestExecution(t, svcCtx.DB, &executions[i])
	}

	tests := []struct {
		name          string
		url           string
		expectedLen   int
		expectedTotal int
	}{
		{
			name:          "valid task with executions",
			url:           fmt.Sprintf("/api/cronexecution/task/%d", taskID),
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "with pagination",
			url:           fmt.Sprintf("/api/cronexecution/task/%d?page=1&pageSize=1", taskID),
			expectedLen:   1,
			expectedTotal: 2,
		},
		{
			name:          "non-existent task",
			url:           "/api/cronexecution/task/99999",
			expectedLen:   0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "GET", tt.url, nil)
			resp := assertResponse(t, w, http.StatusOK)

			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected data format: %T", resp.Data)
			}

			items, ok := data["items"].([]interface{})
			if !ok {
				t.Fatalf("unexpected items format: %T", data["items"])
			}

			page, ok := data["page"].(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected page format: %T", data["page"])
			}

			if len(items) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(items))
			}

			total := int(page["total"].(float64))
			if total != tt.expectedTotal {
				t.Errorf("expected %d total, got %d", tt.expectedTotal, total)
			}
		})
	}
}

// ==================== CronExecutionLog Tests ====================

func TestCronExecutionLogHandler_ListLogs(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test task and execution
	task := &model.CronTask{Name: "Test Task", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http"}
	createTestTask(t, svcCtx.DB, task)

	execution := &model.CronExecution{
		TaskID:    task.ID,
		Status:    model.ExecutionStatusSuccess,
		StartedAt: sqlNullTime(time.Now()),
	}
	createTestExecution(t, svcCtx.DB, execution)

	// Create test logs
	logs := []model.CronExecutionLog{
		{ExecutionID: execution.ID, Level: "info", Message: "Log 1"},
		{ExecutionID: execution.ID, Level: "error", Message: "Log 2"},
		{ExecutionID: execution.ID, Level: "warn", Message: "Log 3"},
	}
	for i := range logs {
		createTestLog(t, svcCtx.DB, &logs[i])
	}

	tests := []struct {
		name          string
		url           string
		expectedLen   int
		expectedTotal int
	}{
		{
			name:          "default pagination",
			url:           "/api/cronexecutionlog",
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "with page size",
			url:           "/api/cronexecutionlog?page=1&pageSize=2",
			expectedLen:   2,
			expectedTotal: 3,
		},
		{
			name:          "filter by execution_id",
			url:           fmt.Sprintf("/api/cronexecutionlog?execution_id=%d", execution.ID),
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "filter by level",
			url:           "/api/cronexecutionlog?level=error",
			expectedLen:   1,
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "GET", tt.url, nil)
			resp := assertResponse(t, w, http.StatusOK)

			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected data format: %T", resp.Data)
			}

			items, ok := data["items"].([]interface{})
			if !ok {
				t.Fatalf("unexpected items format: %T", data["items"])
			}

			page, ok := data["page"].(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected page format: %T", data["page"])
			}

			if len(items) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(items))
			}

			total := int(page["total"].(float64))
			if total != tt.expectedTotal {
				t.Errorf("expected %d total, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestCronExecutionLogHandler_GetLog(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test data
	task := &model.CronTask{Name: "Test Task", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http"}
	createTestTask(t, svcCtx.DB, task)

	execution := &model.CronExecution{TaskID: task.ID, Status: model.ExecutionStatusSuccess, StartedAt: sqlNullTime(time.Now())}
	createTestExecution(t, svcCtx.DB, execution)

	log := &model.CronExecutionLog{ExecutionID: execution.ID, Level: "info", Message: "Test Log"}
	logID := createTestLog(t, svcCtx.DB, log)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "valid log",
			url:          fmt.Sprintf("/api/cronexecutionlog/%d", logID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid id format",
			url:          "/api/cronexecutionlog/abc",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "log not found",
			url:          "/api/cronexecutionlog/99999",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "id less than 1",
			url:          "/api/cronexecutionlog/0",
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

func TestCronExecutionLogHandler_GetByExecutionID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	router := setupTestRouter(t, svcCtx)

	// Create test data
	task := &model.CronTask{Name: "Test Task", Spec: "0 * * * *", Type: model.TaskTypeRecurring, ExecType: "http"}
	createTestTask(t, svcCtx.DB, task)

	execution := &model.CronExecution{TaskID: task.ID, Status: model.ExecutionStatusSuccess, StartedAt: sqlNullTime(time.Now())}
	executionID := createTestExecution(t, svcCtx.DB, execution)

	logs := []model.CronExecutionLog{
		{ExecutionID: execution.ID, Level: "info", Message: "Log 1"},
		{ExecutionID: execution.ID, Level: "error", Message: "Log 2"},
	}
	for i := range logs {
		createTestLog(t, svcCtx.DB, &logs[i])
	}

	tests := []struct {
		name          string
		url           string
		expectedLen   int
		expectedTotal int
	}{
		{
			name:          "valid execution with logs",
			url:           fmt.Sprintf("/api/cronexecutionlog/execution/%d", executionID),
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "with pagination",
			url:           fmt.Sprintf("/api/cronexecutionlog/execution/%d?page=1&pageSize=1", executionID),
			expectedLen:   1,
			expectedTotal: 2,
		},
		{
			name:          "non-existent execution",
			url:           "/api/cronexecutionlog/execution/99999",
			expectedLen:   0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "GET", tt.url, nil)
			resp := assertResponse(t, w, http.StatusOK)

			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected data format: %T", resp.Data)
			}

			items, ok := data["items"].([]interface{})
			if !ok {
				t.Fatalf("unexpected items format: %T", data["items"])
			}

			page, ok := data["page"].(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected page format: %T", data["page"])
			}

			if len(items) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(items))
			}

			total := int(page["total"].(float64))
			if total != tt.expectedTotal {
				t.Errorf("expected %d total, got %d", tt.expectedTotal, total)
			}
		})
	}
}

// Helper function to create sql.NullTime
func sqlNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}
