package repository

import (
	"context"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

func TestCronExecutionRepository_Create(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      model.ExecutionStatusPending,
		RetryCount:  0,
	}

	err := repo.Create(context.Background(), execution)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if execution.ID == 0 {
		t.Error("Create() failed to set ID")
	}
}

func TestCronExecutionRepository_GetByID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	// Create a test execution
	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      model.ExecutionStatusPending,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Test GetByID
	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != execution.ID {
		t.Errorf("GetByID() ID = %v, want %v", got.ID, execution.ID)
	}
	if got.TaskID != execution.TaskID {
		t.Errorf("GetByID() TaskID = %v, want %v", got.TaskID, execution.TaskID)
	}
}

func TestCronExecutionRepository_GetByID_NotFound(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	_, err := repo.GetByID(context.Background(), 999)
	if err == nil {
		t.Error("GetByID() expected error for non-existent ID")
	}
}

func TestCronExecutionRepository_GetByTaskID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	taskID := uint(1)
	now := time.Now()

	// Create multiple executions for the same task
	executions := []*model.CronExecution{
		{TaskID: taskID, ScheduledAt: now.Add(-2 * time.Hour), Status: model.ExecutionStatusSuccess},
		{TaskID: taskID, ScheduledAt: now.Add(-1 * time.Hour), Status: model.ExecutionStatusPending},
		{TaskID: taskID, ScheduledAt: now, Status: model.ExecutionStatusRunning},
	}

	for _, e := range executions {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetByTaskID
	got, err := repo.GetByTaskID(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetByTaskID() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("GetByTaskID() len = %d, want %d", len(got), 3)
	}

	// Should be ordered by scheduled_at DESC
	if got[0].ScheduledAt.Before(got[1].ScheduledAt) {
		t.Error("GetByTaskID() not ordered by scheduled_at DESC")
	}
}

func TestCronExecutionRepository_GetPendingExecutions(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	now := time.Now()

	// Create executions with different statuses
	executions := []*model.CronExecution{
		{TaskID: 1, ScheduledAt: now.Add(-2 * time.Hour), Status: model.ExecutionStatusPending},
		{TaskID: 2, ScheduledAt: now.Add(-1 * time.Hour), Status: model.ExecutionStatusPending},
		{TaskID: 3, ScheduledAt: now.Add(1 * time.Hour), Status: model.ExecutionStatusPending},
		{TaskID: 4, ScheduledAt: now.Add(-30 * time.Minute), Status: model.ExecutionStatusRunning},
	}

	for _, e := range executions {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetPendingExecutions - should get pending executions scheduled before now
	got, err := repo.GetPendingExecutions(context.Background(), now)
	if err != nil {
		t.Fatalf("GetPendingExecutions() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetPendingExecutions() len = %d, want %d", len(got), 2)
	}

	// Should be ordered by scheduled_at ASC
	if got[0].ScheduledAt.After(got[1].ScheduledAt) {
		t.Error("GetPendingExecutions() not ordered by scheduled_at ASC")
	}
}

func TestCronExecutionRepository_GetRunningExecutions(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	now := time.Now()

	executions := []*model.CronExecution{
		{TaskID: 1, ScheduledAt: now.Add(-1 * time.Hour), Status: model.ExecutionStatusRunning},
		{TaskID: 2, ScheduledAt: now.Add(-2 * time.Hour), Status: model.ExecutionStatusPending},
		{TaskID: 3, ScheduledAt: now.Add(-3 * time.Hour), Status: model.ExecutionStatusSuccess},
	}

	for _, e := range executions {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := repo.GetRunningExecutions(context.Background())
	if err != nil {
		t.Fatalf("GetRunningExecutions() error = %v", err)
	}

	if len(got) != 1 {
		t.Errorf("GetRunningExecutions() len = %d, want %d", len(got), 1)
	}
	if got[0].Status != model.ExecutionStatusRunning {
		t.Errorf("GetRunningExecutions() Status = %v, want %v", got[0].Status, model.ExecutionStatusRunning)
	}
}

func TestCronExecutionRepository_Update(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
		Error:       "initial error",
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the execution
	execution.Status = model.ExecutionStatusFailed
	execution.Error = "updated error"

	err := repo.Update(context.Background(), execution)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.ExecutionStatusFailed {
		t.Errorf("Update() Status = %v, want %v", got.Status, model.ExecutionStatusFailed)
	}
	if got.Error != "updated error" {
		t.Errorf("Update() Error = %v, want %v", got.Error, "updated error")
	}
}

func TestCronExecutionRepository_MarkAsRunning(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsRunning(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("MarkAsRunning() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.ExecutionStatusRunning {
		t.Errorf("MarkAsRunning() Status = %v, want %v", got.Status, model.ExecutionStatusRunning)
	}
	if !got.StartedAt.Valid {
		t.Error("MarkAsRunning() StartedAt should be set")
	}
}

func TestCronExecutionRepository_MarkAsSuccess(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusRunning,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	completedAt := time.Now()
	err := repo.MarkAsSuccess(context.Background(), execution.ID, completedAt)
	if err != nil {
		t.Fatalf("MarkAsSuccess() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.ExecutionStatusSuccess {
		t.Errorf("MarkAsSuccess() Status = %v, want %v", got.Status, model.ExecutionStatusSuccess)
	}
	if !got.CompletedAt.Valid || got.CompletedAt.Time.IsZero() {
		t.Error("MarkAsSuccess() CompletedAt should be set")
	}
}

func TestCronExecutionRepository_MarkAsFailed(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusRunning,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	completedAt := time.Now()
	errMsg := "test error message"
	err := repo.MarkAsFailed(context.Background(), execution.ID, completedAt, errMsg)
	if err != nil {
		t.Fatalf("MarkAsFailed() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.ExecutionStatusFailed {
		t.Errorf("MarkAsFailed() Status = %v, want %v", got.Status, model.ExecutionStatusFailed)
	}
	if got.Error != errMsg {
		t.Errorf("MarkAsFailed() Error = %v, want %v", got.Error, errMsg)
	}
}

func TestCronExecutionRepository_IncrementRetryCount(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
		RetryCount:  0,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.IncrementRetryCount(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("IncrementRetryCount() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.RetryCount != 1 {
		t.Errorf("IncrementRetryCount() RetryCount = %d, want %d", got.RetryCount, 1)
	}

	// Increment again
	err = repo.IncrementRetryCount(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("IncrementRetryCount() second call error = %v", err)
	}

	got2, _ := repo.GetByID(context.Background(), execution.ID)
	if got2.RetryCount != 2 {
		t.Errorf("IncrementRetryCount() RetryCount after second call = %d, want %d", got2.RetryCount, 2)
	}
}

func TestCronExecutionRepository_MarkAsRetried(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusFailed,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsRetried(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("MarkAsRetried() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.ExecutionStatusRetried {
		t.Errorf("MarkAsRetried() Status = %v, want %v", got.Status, model.ExecutionStatusRetried)
	}
}

func TestCronExecutionRepository_MarkAsCancelled(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsCancelled(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("MarkAsCancelled() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.ExecutionStatusCancelled {
		t.Errorf("MarkAsCancelled() Status = %v, want %v", got.Status, model.ExecutionStatusCancelled)
	}
	if !got.CompletedAt.Valid {
		t.Error("MarkAsCancelled() CompletedAt should be set")
	}
}

func TestCronExecutionRepository_Delete(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	execution := &model.CronExecution{
		TaskID:      1,
		ScheduledAt: time.Now(),
		Status:      model.ExecutionStatusPending,
	}
	if err := repo.Create(context.Background(), execution); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.Delete(context.Background(), execution.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should return error for soft-deleted record
	_, err = repo.GetByID(context.Background(), execution.ID)
	if err == nil {
		t.Error("Delete() should have soft-deleted the record")
	}
}

func TestCronExecutionRepository_DeleteByTaskID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	taskID := uint(1)

	// Create multiple executions for the same task
	executions := []*model.CronExecution{
		{TaskID: taskID, ScheduledAt: time.Now(), Status: model.ExecutionStatusPending},
		{TaskID: taskID, ScheduledAt: time.Now().Add(time.Hour), Status: model.ExecutionStatusPending},
		{TaskID: taskID, ScheduledAt: time.Now().Add(2 * time.Hour), Status: model.ExecutionStatusPending},
	}

	for _, e := range executions {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	err := repo.DeleteByTaskID(context.Background(), taskID)
	if err != nil {
		t.Fatalf("DeleteByTaskID() error = %v", err)
	}

	// All should be soft-deleted
	got, err := repo.GetByTaskID(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetByTaskID() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("DeleteByTaskID() should have deleted all records, got %d remaining", len(got))
	}
}

func TestCronExecutionRepository_GetOverdueRunning(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronExecutionRepository(svcCtx)

	now := time.Now()
	timeout := 30 * time.Minute

	// Create executions with different start times
	executions := []*model.CronExecution{
		{TaskID: 1, ScheduledAt: now.Add(-2 * time.Hour), Status: model.ExecutionStatusRunning},
		{TaskID: 2, ScheduledAt: now.Add(-1 * time.Hour), Status: model.ExecutionStatusRunning},
		{TaskID: 3, ScheduledAt: now.Add(-10 * time.Minute), Status: model.ExecutionStatusRunning},
		{TaskID: 4, ScheduledAt: now.Add(-5 * time.Minute), Status: model.ExecutionStatusPending},
	}

	for _, e := range executions {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Manually set started_at for running executions to simulate overdue state
	// Use raw DB update to set specific started_at values
	for _, e := range executions {
		if e.Status == model.ExecutionStatusRunning {
			// Set started_at based on task ID to control timing
			var startedAt time.Time
			switch e.TaskID {
			case 1:
				startedAt = now.Add(-2 * time.Hour) // Overdue
			case 2:
				startedAt = now.Add(-1 * time.Hour) // Overdue
			case 3:
				startedAt = now.Add(-10 * time.Minute) // Not overdue
			}
			// Update started_at directly
			r := svcCtx.DB.Model(&model.CronExecution{}).Where("id = ?", e.ID).Update("started_at", startedAt)
			if r.Error != nil {
				t.Fatalf("Update started_at error = %v", r.Error)
			}
		}
	}

	got, err := repo.GetOverdueRunning(context.Background(), timeout)
	if err != nil {
		t.Fatalf("GetOverdueRunning() error = %v", err)
	}

	// Should get executions that have been running longer than timeout (task 1 and 2)
	if len(got) != 2 {
		t.Errorf("GetOverdueRunning() len = %d, want %d", len(got), 2)
	}
}
