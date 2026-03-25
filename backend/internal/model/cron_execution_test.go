package model

import (
	"database/sql"
	"testing"
	"time"
)

func TestExecutionStatus_Values(t *testing.T) {
	tests := []struct {
		name   string
		status ExecutionStatus
		want   string
	}{
		{"pending", ExecutionStatusPending, "pending"},
		{"running", ExecutionStatusRunning, "running"},
		{"success", ExecutionStatusSuccess, "success"},
		{"failed", ExecutionStatusFailed, "failed"},
		{"retried", ExecutionStatusRetried, "retried"},
		{"cancelled", ExecutionStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("ExecutionStatus(%v) = %v, want %v", tt.status, tt.status, tt.want)
			}
		})
	}
}

func TestCronExecution_TableName(t *testing.T) {
	exec := CronExecution{}
	want := "cron_execution"
	if got := exec.TableName(); got != want {
		t.Errorf("CronExecution.TableName() = %v, want %v", got, want)
	}
}

func TestCronExecution_Fields(t *testing.T) {
	now := time.Now()
	nullTime := sql.NullTime{Time: now, Valid: true}

	// 验证 Go 零值
	exec := CronExecution{}
	if exec.Status != "" {
		t.Errorf("zero value Status = %v, want empty string", exec.Status)
	}
	if exec.RetryCount != 0 {
		t.Errorf("zero value RetryCount = %v, want 0", exec.RetryCount)
	}

	// 验证字段可设置性（GORM default 标签在数据库层面生效）
	exec.TaskID = 1
	exec.ScheduledAt = now
	exec.Status = ExecutionStatusPending
	exec.StartedAt = nullTime
	exec.CompletedAt = nullTime
	exec.RetryCount = 1
	exec.Error = "test error"

	if exec.TaskID != 1 {
		t.Errorf("TaskID = %v, want 1", exec.TaskID)
	}
	if !exec.ScheduledAt.Equal(now) {
		t.Error("ScheduledAt should be settable")
	}
	if exec.Status != ExecutionStatusPending {
		t.Errorf("Status = %v, want %v", exec.Status, ExecutionStatusPending)
	}
	if !exec.StartedAt.Valid || !exec.StartedAt.Time.Equal(now) {
		t.Error("StartedAt should be settable")
	}
	if !exec.CompletedAt.Valid || !exec.CompletedAt.Time.Equal(now) {
		t.Error("CompletedAt should be settable")
	}
	if exec.RetryCount != 1 {
		t.Errorf("RetryCount = %v, want 1", exec.RetryCount)
	}
	if exec.Error != "test error" {
		t.Errorf("Error = %v, want 'test error'", exec.Error)
	}
}

func TestCronExecution_StatusTransitions(t *testing.T) {
	now := time.Now()
	nullTime := sql.NullTime{Time: now, Valid: true}

	exec := &CronExecution{
		TaskID:      1,
		ScheduledAt: now,
		Status:      ExecutionStatusPending,
	}

	// Pending -> Running
	exec.Status = ExecutionStatusRunning
	if exec.Status != ExecutionStatusRunning {
		t.Errorf("status transition to running failed: %v", exec.Status)
	}

	// Running -> Success
	exec.CompletedAt = nullTime
	exec.Status = ExecutionStatusSuccess
	if exec.Status != ExecutionStatusSuccess {
		t.Errorf("status transition to success failed: %v", exec.Status)
	}

	// Reset for retry test
	exec.Status = ExecutionStatusFailed
	exec.RetryCount = 1
	if exec.Status != ExecutionStatusFailed {
		t.Errorf("status transition to failed failed: %v", exec.Status)
	}
}

func TestCronExecution_WithRetry(t *testing.T) {
	now := time.Now()
	exec := CronExecution{
		TaskID:      1,
		ScheduledAt: now,
		RetryCount:  0,
	}

	// 模拟重试
	exec.RetryCount++
	if exec.RetryCount != 1 {
		t.Errorf("RetryCount after increment = %v, want 1", exec.RetryCount)
	}

	exec.RetryCount++
	if exec.RetryCount != 2 {
		t.Errorf("RetryCount after second increment = %v, want 2", exec.RetryCount)
	}
}

func TestCronExecution_DatabaseOperations(t *testing.T) {
	db := newTestDB(t)

	now := time.Now()
	nullTime := sql.NullTime{Time: now, Valid: true}

	// Create an execution record
	execution := CronExecution{
		TaskID:      1,
		ScheduledAt: now,
		Status:      ExecutionStatusPending,
		StartedAt:   nullTime,
		RetryCount:  0,
	}

		if err := db.Create(&execution).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was assigned
	if execution.ID == 0 {
		t.Error("ID should be auto-assigned after create")
	}

	// Query the execution
	var queried CronExecution
	if err := db.First(&queried, execution.ID).Error; err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Verify fields
	if queried.TaskID != execution.TaskID {
		t.Errorf("TaskID mismatch: got %v, want %v", queried.TaskID, execution.TaskID)
	}
	if queried.Status != execution.Status {
		t.Errorf("Status mismatch: got %v, want %v", queried.Status, execution.Status)
	}
	if queried.RetryCount != execution.RetryCount {
		t.Errorf("RetryCount mismatch: got %v, want %v", queried.RetryCount, execution.RetryCount)
	}

		// Update status
	execution.Status = ExecutionStatusRunning
	if err := db.Save(&execution).Error; err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	var updated CronExecution
	if err := db.First(&updated, execution.ID).Error; err != nil {
		t.Fatalf("Query after update failed: %v", err)
	}
	if updated.Status != ExecutionStatusRunning {
		t.Errorf("Status after update mismatch: got %v, want %v", updated.Status, ExecutionStatusRunning)
	}
}
