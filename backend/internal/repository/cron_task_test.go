package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
)

func TestCronTaskRepository_Create(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Raw:      `{"url": "http://example.com"}`,
		Enabled:  true,
	}

	err := repo.Create(context.Background(), task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if task.ID == 0 {
		t.Error("Create() failed to set ID")
	}
}

func TestCronTaskRepository_GetByID(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	// Create a test task
	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Test GetByID
	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != task.ID {
		t.Errorf("GetByID() ID = %v, want %v", got.ID, task.ID)
	}
	if got.Name != task.Name {
		t.Errorf("GetByID() Name = %v, want %v", got.Name, task.Name)
	}
}

func TestCronTaskRepository_GetByID_NotFound(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	_, err := repo.GetByID(context.Background(), 999)
	if err == nil {
		t.Error("GetByID() expected error for non-existent ID")
	}
}

func TestCronTaskRepository_GetAll(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	// Create multiple tasks
	tasks := []*model.CronTask{
		{Name: "task-1", Spec: "0 * * * *", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "task-2", Spec: "0 0 * * *", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "task-3", Spec: "0 0 0 * *", Type: model.TaskTypeOnce, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
	}

	for _, task := range tasks {
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetAll
	got, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("GetAll() len = %d, want %d", len(got), 3)
	}
}

func TestCronTaskRepository_GetEnabledTasks(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	// Create tasks with different enabled states and statuses
	taskData := []struct {
		Name     string
		Type     model.TaskType
		Status   model.TaskStatus
		ExecType string
		Enabled  bool
	}{
		{Name: "enabled-pending", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "enabled-running", Type: model.TaskTypeRecurring, Status: model.TaskStatusRunning, ExecType: "http", Enabled: true},
		{Name: "disabled", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: false},
		{Name: "enabled-stopped", Type: model.TaskTypeRecurring, Status: model.TaskStatusStopped, ExecType: "http", Enabled: true},
	}

	for _, td := range taskData {
		task := &model.CronTask{
			Name:     td.Name,
			Type:     td.Type,
			Status:   td.Status,
			ExecType: td.ExecType,
			Enabled:  td.Enabled,
		}
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		// Update enabled using SQL to handle false value correctly
		enabledValue := 0
		if td.Enabled {
			enabledValue = 1
		}
		r := svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("enabled", enabledValue)
		if r.Error != nil {
			t.Fatalf("Update enabled error = %v", r.Error)
		}
		// Update status
		r = svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("status", td.Status)
		if r.Error != nil {
			t.Fatalf("Update status error = %v", r.Error)
		}
	}

	// Test GetEnabledTasks - should get enabled tasks with pending or running status
	got, err := repo.GetEnabledTasks(context.Background())
	if err != nil {
		t.Fatalf("GetEnabledTasks() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetEnabledTasks() len = %d, want %d", len(got), 2)
	}
}

func TestCronTaskRepository_GetRecurringTasks(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	// Create tasks with different types
	tasks := []*model.CronTask{
		{Name: "recurring-1", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "recurring-2", Type: model.TaskTypeRecurring, Status: model.TaskStatusRunning, ExecType: "http", Enabled: true},
		{Name: "recurring-disabled", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: false},
		{Name: "once-1", Type: model.TaskTypeOnce, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
	}

	for _, task := range tasks {
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		// Update enabled using raw SQL (GORM ignores boolean false)
		enabledValue := 0
		if task.Enabled {
			enabledValue = 1
		}
		r := svcCtx.DB.Exec("UPDATE cron_task SET enabled = ? WHERE id = ?", enabledValue, task.ID)
		if r.Error != nil {
			t.Fatalf("Update enabled error = %v", r.Error)
		}
	}

	// Test GetRecurringTasks - should return 2 enabled recurring tasks
	got, err := repo.GetRecurringTasks(context.Background())
	if err != nil {
		t.Fatalf("GetRecurringTasks() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetRecurringTasks() len = %d, want %d", len(got), 2)
	}
}

func TestCronTaskRepository_GetPendingTasks(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	// Create tasks with different statuses
	tasks := []*model.CronTask{
		{Name: "pending-1", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "pending-2", Type: model.TaskTypeOnce, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "running-1", Type: model.TaskTypeRecurring, Status: model.TaskStatusRunning, ExecType: "http", Enabled: true},
		{Name: "stopped-1", Type: model.TaskTypeRecurring, Status: model.TaskStatusStopped, ExecType: "http", Enabled: true},
	}

	for _, task := range tasks {
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetPendingTasks
	got, err := repo.GetPendingTasks(context.Background())
	if err != nil {
		t.Fatalf("GetPendingTasks() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetPendingTasks() len = %d, want %d", len(got), 2)
	}
}

func TestCronTaskRepository_GetRunningTasks(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	// Create tasks with different statuses
	tasks := []*model.CronTask{
		{Name: "running-1", Type: model.TaskTypeRecurring, Status: model.TaskStatusRunning, ExecType: "http", Enabled: true},
		{Name: "running-2", Type: model.TaskTypeOnce, Status: model.TaskStatusRunning, ExecType: "http", Enabled: false},
		{Name: "pending-1", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
	}

	for _, task := range tasks {
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Test GetRunningTasks
	got, err := repo.GetRunningTasks(context.Background())
	if err != nil {
		t.Fatalf("GetRunningTasks() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetRunningTasks() len = %d, want %d", len(got), 2)
	}
}

func TestCronTaskRepository_Update(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the task
	task.Spec = "0 0 * * *"
	task.MaxRetries = 3

	err := repo.Update(context.Background(), task)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Spec != "0 0 * * *" {
		t.Errorf("Update() Spec = %v, want %v", got.Spec, "0 0 * * *")
	}
	if got.MaxRetries != 3 {
		t.Errorf("Update() MaxRetries = %d, want %d", got.MaxRetries, 3)
	}
}

func TestCronTaskRepository_UpdateStatus(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.UpdateStatus(context.Background(), task.ID, model.TaskStatusRunning)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.TaskStatusRunning {
		t.Errorf("UpdateStatus() Status = %v, want %v", got.Status, model.TaskStatusRunning)
	}
}

func TestCronTaskRepository_MarkAsRunning(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsRunning(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("MarkAsRunning() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.TaskStatusRunning {
		t.Errorf("MarkAsRunning() Status = %v, want %v", got.Status, model.TaskStatusRunning)
	}
	if !got.LastExecutedAt.Valid {
		t.Error("MarkAsRunning() LastExecutedAt should be set")
	}
}

func TestCronTaskRepository_MarkAsCompleted(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeOnce,
		Status:   model.TaskStatusRunning,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsCompleted(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("MarkAsCompleted() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.TaskStatusCompleted {
		t.Errorf("MarkAsCompleted() Status = %v, want %v", got.Status, model.TaskStatusCompleted)
	}
}

func TestCronTaskRepository_MarkAsStopped(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusRunning,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsStopped(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("MarkAsStopped() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.TaskStatusStopped {
		t.Errorf("MarkAsStopped() Status = %v, want %v", got.Status, model.TaskStatusStopped)
	}
}

func TestCronTaskRepository_MarkAsFailed(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusRunning,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.MarkAsFailed(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("MarkAsFailed() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != model.TaskStatusFailed {
		t.Errorf("MarkAsFailed() Status = %v, want %v", got.Status, model.TaskStatusFailed)
	}
}

func TestCronTaskRepository_UpdateNextExecution(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	nextAt := time.Now().Add(time.Hour)
	err := repo.UpdateNextExecution(context.Background(), task.ID, nextAt)
	if err != nil {
		t.Fatalf("UpdateNextExecution() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if !got.NextExecutionAt.Valid {
		t.Error("UpdateNextExecution() NextExecutionAt should be set")
	}
}

func TestCronTaskRepository_IncrementTotalExecutions(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.IncrementTotalExecutions(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("IncrementTotalExecutions() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.TotalExecutions != 1 {
		t.Errorf("IncrementTotalExecutions() TotalExecutions = %d, want %d", got.TotalExecutions, 1)
	}

	// Increment again
	err = repo.IncrementTotalExecutions(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("IncrementTotalExecutions() second call error = %v", err)
	}

	got2, _ := repo.GetByID(context.Background(), task.ID)
	if got2.TotalExecutions != 2 {
		t.Errorf("IncrementTotalExecutions() TotalExecutions after second call = %d, want %d", got2.TotalExecutions, 2)
	}
}

func TestCronTaskRepository_Enable(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  false,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.Enable(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if !got.Enabled {
		t.Error("Enable() Enabled should be true")
	}
}

func TestCronTaskRepository_Disable(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.Disable(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Disable() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Enabled {
		t.Error("Disable() Enabled should be false")
	}
}

func TestCronTaskRepository_Delete(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	task := &model.CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     model.TaskTypeRecurring,
		Status:   model.TaskStatusPending,
		ExecType: "http",
		Enabled:  true,
	}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.Delete(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should return error for soft-deleted record
	_, err = repo.GetByID(context.Background(), task.ID)
	if err == nil {
		t.Error("Delete() should have soft-deleted the record")
	}
}

func TestCronTaskRepository_GetDueTasks(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	now := time.Now()

	// Create tasks with different next execution times
	taskData := []struct {
		Name    string
		Enabled bool
	}{
		{Name: "due-1", Enabled: true},
		{Name: "due-2", Enabled: true},
		{Name: "future", Enabled: true},
		{Name: "disabled", Enabled: false},
	}

	// Set next execution times
	dueTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	var tasks []*model.CronTask
	for i, td := range taskData {
		task := &model.CronTask{
			Name:     td.Name,
			Type:     model.TaskTypeOnce,
			Status:   model.TaskStatusPending,
			ExecType: "http",
			Enabled:  td.Enabled,
		}
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		tasks = append(tasks, task)

		// Update enabled using SQL to handle false value correctly
		enabledValue := 0
		if td.Enabled {
			enabledValue = 1
		}
		r := svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("enabled", enabledValue)
		if r.Error != nil {
			t.Fatalf("Update enabled error = %v", r.Error)
		}

		// Update NextExecutionAt after creation
		var nextExecAt sql.NullTime
		if i < 2 {
			nextExecAt = sql.NullTime{Time: dueTime, Valid: true}
		} else if i == 2 {
			nextExecAt = sql.NullTime{Time: futureTime, Valid: true}
		} else {
			nextExecAt = sql.NullTime{Time: dueTime, Valid: true}
		}
		r = svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("next_execution_at", nextExecAt)
		if r.Error != nil {
			t.Fatalf("Update NextExecutionAt error = %v", r.Error)
		}
	}

	// Test GetDueTasks
	got, err := repo.GetDueTasks(context.Background(), now)
	if err != nil {
		t.Fatalf("GetDueTasks() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetDueTasks() len = %d, want %d", len(got), 2)
	}
}

func TestCronTaskRepository_GetTasksDueForExecution(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	now := time.Now()

	// Create recurring tasks with different next execution times
	taskData := []struct {
		Name    string
		Status  model.TaskStatus
		Enabled bool
	}{
		{Name: "due-pending", Status: model.TaskStatusPending, Enabled: true},
		{Name: "due-running", Status: model.TaskStatusRunning, Enabled: true},
		{Name: "future", Status: model.TaskStatusPending, Enabled: true},
		{Name: "disabled", Status: model.TaskStatusPending, Enabled: false},
	}

	// Set next execution times
	dueTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	for i, td := range taskData {
		task := &model.CronTask{
			Name:     td.Name,
			Type:     model.TaskTypeRecurring,
			Status:   td.Status,
			ExecType: "http",
			Enabled:  td.Enabled,
		}
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Update enabled using SQL to handle false value correctly
		enabledValue := 0
		if td.Enabled {
			enabledValue = 1
		}
		r := svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("enabled", enabledValue)
		if r.Error != nil {
			t.Fatalf("Update enabled error = %v", r.Error)
		}

		// Update status
		r = svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("status", td.Status)
		if r.Error != nil {
			t.Fatalf("Update status error = %v", r.Error)
		}

		// Update NextExecutionAt after creation
		var nextExecAt sql.NullTime
		if i < 2 {
			nextExecAt = sql.NullTime{Time: dueTime, Valid: true}
		} else if i == 2 {
			nextExecAt = sql.NullTime{Time: futureTime, Valid: true}
		} else {
			nextExecAt = sql.NullTime{Time: dueTime, Valid: true}
		}
		r = svcCtx.DB.Model(&model.CronTask{}).Where("id = ?", task.ID).
			Update("next_execution_at", nextExecAt)
		if r.Error != nil {
			t.Fatalf("Update NextExecutionAt error = %v", r.Error)
		}
	}

	// Test GetTasksDueForExecution
	got, err := repo.GetTasksDueForExecution(context.Background(), now)
	if err != nil {
		t.Fatalf("GetTasksDueForExecution() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetTasksDueForExecution() len = %d, want %d", len(got), 2)
	}
}

func TestCronTaskRepository_GetDueTasksWithLock(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	repo := NewCronTaskRepository(svcCtx)

	now := time.Now()
	dueTime := now.Add(-1 * time.Hour)

	// Create tasks that are due
	tasks := []*model.CronTask{
		{Name: "due-once", Type: model.TaskTypeOnce, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "due-recurring", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "future-task", Type: model.TaskTypeRecurring, Status: model.TaskStatusPending, ExecType: "http", Enabled: true},
		{Name: "disabled-task", Type: model.TaskTypeOnce, Status: model.TaskStatusPending, ExecType: "http", Enabled: false},
	}

	for _, task := range tasks {
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		taskID := task.ID

		// Update NextExecutionAt and enabled using raw SQL (GORM ignores boolean false)
		var nextExecAt sql.NullTime
		if task.Name == "future-task" {
			nextExecAt = sql.NullTime{Time: now.Add(1 * time.Hour), Valid: true}
		} else {
			nextExecAt = sql.NullTime{Time: dueTime, Valid: true}
		}

		enabledValue := 0
		if task.Enabled {
			enabledValue = 1
		}

		// Use a single raw SQL update to avoid GORM boolean issues
		r := svcCtx.DB.Exec("UPDATE cron_task SET next_execution_at = ?, enabled = ? WHERE id = ?",
			nextExecAt.Time, enabledValue, taskID)
		if r.Error != nil {
			t.Fatalf("Update error = %v", r.Error)
		}
		if r.RowsAffected != 1 {
			t.Fatalf("Update rows affected = %d, want 1", r.RowsAffected)
		}
	}

	// Test GetDueTasksWithLock - should return 2 enabled tasks that are due
	got, err := repo.GetDueTasksWithLock(context.Background(), now)
	if err != nil {
		t.Fatalf("GetDueTasksWithLock() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetDueTasksWithLock() len = %d, want %d", len(got), 2)
	}

	// Verify the correct tasks were returned
	names := make(map[string]bool)
	for _, task := range got {
		names[task.Name] = true
	}

	if !names["due-once"] {
		t.Error("GetDueTasksWithLock() should include due-once task")
	}
	if !names["due-recurring"] {
		t.Error("GetDueTasksWithLock() should include due-recurring task")
	}
	if names["future-task"] {
		t.Error("GetDueTasksWithLock() should not include future-task")
	}
	if names["disabled-task"] {
		t.Error("GetDueTasksWithLock() should not include disabled-task")
	}
}
