package repository

import (
	"context"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

func TestCronExecutionLogRepository_Create(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	log := &model.CronExecutionLog{
		ExecutionID: 1,
		From:        "scheduler",
		Level:       "info",
		Message:     "Test log message",
	}

	err := repo.Create(context.Background(), log)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if log.ID == 0 {
		t.Error("Create() failed to set ID")
	}
}

func TestCronExecutionLogRepository_GetByID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	// Create a test log
	log := &model.CronExecutionLog{
		ExecutionID: 1,
		From:        "scheduler",
		Level:       "info",
		Message:     "Test log message",
	}
	if err := repo.Create(context.Background(), log); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Test GetByID
	got, err := repo.GetByID(context.Background(), log.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != log.ID {
		t.Errorf("GetByID() ID = %v, want %v", got.ID, log.ID)
	}
	if got.ExecutionID != log.ExecutionID {
		t.Errorf("GetByID() ExecutionID = %v, want %v", got.ExecutionID, log.ExecutionID)
	}
	if got.Level != log.Level {
		t.Errorf("GetByID() Level = %v, want %v", got.Level, log.Level)
	}
}

func TestCronExecutionLogRepository_GetByID_NotFound(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	_, err := repo.GetByID(context.Background(), 999)
	if err == nil {
		t.Error("GetByID() expected error for non-existent ID")
	}
}

func TestCronExecutionLogRepository_GetByExecutionID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	executionID := uint(1)
	now := time.Now()

	// Create multiple logs for the same execution
	logs := []*model.CronExecutionLog{
		{ExecutionID: executionID, From: "scheduler", Level: "info", Message: "First log", CreatedAt: now.Add(-2 * time.Hour)},
		{ExecutionID: executionID, From: "executor", Level: "warn", Message: "Second log", CreatedAt: now.Add(-1 * time.Hour)},
		{ExecutionID: executionID, From: "scheduler", Level: "error", Message: "Third log", CreatedAt: now},
	}

	for _, l := range logs {
		if err := repo.Create(context.Background(), l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetByExecutionID
	got, err := repo.GetByExecutionID(context.Background(), executionID)
	if err != nil {
		t.Fatalf("GetByExecutionID() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("GetByExecutionID() len = %d, want %d", len(got), 3)
	}

	// Should be ordered by created_at ASC
	for i := 1; i < len(got); i++ {
		if got[i].CreatedAt.Before(got[i-1].CreatedAt) {
			t.Error("GetByExecutionID() not ordered by created_at ASC")
			break
		}
	}
}

func TestCronExecutionLogRepository_GetByLevel(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	// Create logs with different levels
	logs := []*model.CronExecutionLog{
		{ExecutionID: 1, From: "scheduler", Level: "info", Message: "Info log 1"},
		{ExecutionID: 2, From: "executor", Level: "error", Message: "Error log 1"},
		{ExecutionID: 3, From: "scheduler", Level: "info", Message: "Info log 2"},
		{ExecutionID: 4, From: "executor", Level: "warn", Message: "Warn log 1"},
		{ExecutionID: 5, From: "scheduler", Level: "error", Message: "Error log 2"},
	}

	for _, l := range logs {
		if err := repo.Create(context.Background(), l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetByLevel for error logs
	got, err := repo.GetByLevel(context.Background(), "error", 0)
	if err != nil {
		t.Fatalf("GetByLevel() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetByLevel() error len = %d, want %d", len(got), 2)
	}

	// Test with limit
	gotLimited, err := repo.GetByLevel(context.Background(), "info", 1)
	if err != nil {
		t.Fatalf("GetByLevel() with limit error = %v", err)
	}

	if len(gotLimited) != 1 {
		t.Errorf("GetByLevel() with limit len = %d, want %d", len(gotLimited), 1)
	}
}

func TestCronExecutionLogRepository_GetRecentLogs(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	// Create multiple logs
	for i := 0; i < 15; i++ {
		log := &model.CronExecutionLog{
			ExecutionID: uint(i + 1),
			From:        "scheduler",
			Level:       "info",
			Message:     "Log message",
		}
		if err := repo.Create(context.Background(), log); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetRecentLogs with default limit (100)
	got, err := repo.GetRecentLogs(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetRecentLogs() error = %v", err)
	}

	if len(got) != 15 {
		t.Errorf("GetRecentLogs() default limit len = %d, want %d", len(got), 15)
	}

	// Test with custom limit
	gotLimited, err := repo.GetRecentLogs(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetRecentLogs() custom limit error = %v", err)
	}

	if len(gotLimited) != 5 {
		t.Errorf("GetRecentLogs() custom limit len = %d, want %d", len(gotLimited), 5)
	}
}

func TestCronExecutionLogRepository_GetErrorLogs(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	now := time.Now()

	// Create logs with different levels and times
	logs := []*model.CronExecutionLog{
		{ExecutionID: 1, From: "scheduler", Level: "error", Message: "Old error", CreatedAt: now.Add(-2 * time.Hour)},
		{ExecutionID: 2, From: "executor", Level: "warn", Message: "Recent warn", CreatedAt: now.Add(-30 * time.Minute)},
		{ExecutionID: 3, From: "scheduler", Level: "error", Message: "Recent error", CreatedAt: now.Add(-15 * time.Minute)},
		{ExecutionID: 4, From: "executor", Level: "info", Message: "Recent info", CreatedAt: now.Add(-10 * time.Minute)},
	}

	for _, l := range logs {
		if err := repo.Create(context.Background(), l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetErrorLogs - get error and warn logs since 1 hour ago
	since := now.Add(-1 * time.Hour)
	got, err := repo.GetErrorLogs(context.Background(), since)
	if err != nil {
		t.Fatalf("GetErrorLogs() error = %v", err)
	}

	// Should get recent warn and recent error (2 logs)
	if len(got) != 2 {
		t.Errorf("GetErrorLogs() len = %d, want %d", len(got), 2)
	}
}

func TestCronExecutionLogRepository_CreateBatch(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	// Create batch of logs
	logs := []*model.CronExecutionLog{
		{ExecutionID: 1, From: "scheduler", Level: "info", Message: "Batch log 1"},
		{ExecutionID: 2, From: "executor", Level: "warn", Message: "Batch log 2"},
		{ExecutionID: 3, From: "scheduler", Level: "error", Message: "Batch log 3"},
	}

	err := repo.CreateBatch(context.Background(), logs)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}

	// Verify all logs were created
	for _, log := range logs {
		if log.ID == 0 {
			t.Errorf("CreateBatch() failed to set ID for log %+v", log)
		}
	}
}

func TestCronExecutionLogRepository_Delete(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	log := &model.CronExecutionLog{
		ExecutionID: 1,
		From:        "scheduler",
		Level:       "info",
		Message:     "Test log message",
	}
	if err := repo.Create(context.Background(), log); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.Delete(context.Background(), log.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should return error for soft-deleted record
	_, err = repo.GetByID(context.Background(), log.ID)
	if err == nil {
		t.Error("Delete() should have soft-deleted the record")
	}
}

func TestCronExecutionLogRepository_DeleteByExecutionID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	executionID := uint(1)

	// Create multiple logs for the same execution
	logs := []*model.CronExecutionLog{
		{ExecutionID: executionID, From: "scheduler", Level: "info", Message: "Log 1"},
		{ExecutionID: executionID, From: "executor", Level: "warn", Message: "Log 2"},
		{ExecutionID: executionID, From: "scheduler", Level: "error", Message: "Log 3"},
	}

	for _, l := range logs {
		if err := repo.Create(context.Background(), l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	err := repo.DeleteByExecutionID(context.Background(), executionID)
	if err != nil {
		t.Fatalf("DeleteByExecutionID() error = %v", err)
	}

	// All should be soft-deleted
	got, err := repo.GetByExecutionID(context.Background(), executionID)
	if err != nil {
		t.Fatalf("GetByExecutionID() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("DeleteByExecutionID() should have deleted all records, got %d remaining", len(got))
	}
}

func TestCronExecutionLogRepository_DeleteOlderThan(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	now := time.Now()

	// Create logs with different times
	logs := []*model.CronExecutionLog{
		{ExecutionID: 1, From: "scheduler", Level: "info", Message: "Old log 1", CreatedAt: now.Add(-2 * time.Hour)},
		{ExecutionID: 2, From: "executor", Level: "warn", Message: "Old log 2", CreatedAt: now.Add(-90 * time.Minute)},
		{ExecutionID: 3, From: "scheduler", Level: "error", Message: "Recent log", CreatedAt: now.Add(-30 * time.Minute)},
		{ExecutionID: 4, From: "executor", Level: "info", Message: "Recent log 2", CreatedAt: now},
	}

	for _, l := range logs {
		if err := repo.Create(context.Background(), l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Delete logs older than 1 hour ago
	cutoff := now.Add(-1 * time.Hour)
	err := repo.DeleteOlderThan(context.Background(), cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}

	// Get all remaining logs
	got, err := repo.GetRecentLogs(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetRecentLogs() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("DeleteOlderThan() should have deleted 2 old logs, got %d remaining", len(got))
	}
}

func TestCronExecutionLogRepository_GetLogsByFrom(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionLogRepository(svcCtx)

	// Create logs from different sources
	logs := []*model.CronExecutionLog{
		{ExecutionID: 1, From: "scheduler", Level: "info", Message: "Scheduler log 1"},
		{ExecutionID: 2, From: "executor", Level: "warn", Message: "Executor log 1"},
		{ExecutionID: 3, From: "scheduler", Level: "error", Message: "Scheduler log 2"},
		{ExecutionID: 4, From: "api", Level: "info", Message: "API log 1"},
		{ExecutionID: 5, From: "scheduler", Level: "debug", Message: "Scheduler log 3"},
	}

	for _, l := range logs {
		if err := repo.Create(context.Background(), l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetLogsByFrom for scheduler logs
	got, err := repo.GetLogsByFrom(context.Background(), "scheduler", 0)
	if err != nil {
		t.Fatalf("GetLogsByFrom() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("GetLogsByFrom() scheduler len = %d, want %d", len(got), 3)
	}

	// Test with limit
	gotLimited, err := repo.GetLogsByFrom(context.Background(), "scheduler", 2)
	if err != nil {
		t.Fatalf("GetLogsByFrom() with limit error = %v", err)
	}

	if len(gotLimited) != 2 {
		t.Errorf("GetLogsByFrom() with limit len = %d, want %d", len(gotLimited), 2)
	}
}
