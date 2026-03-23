package model

import (
	"database/sql"
	"testing"
	"time"
)

func TestTaskStatus_Values(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   string
	}{
		{"pending", TaskStatusPending, "pending"},
		{"running", TaskStatusRunning, "running"},
		{"completed", TaskStatusCompleted, "completed"},
		{"stopped", TaskStatusStopped, "stopped"},
		{"failed", TaskStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("TaskStatus(%v) = %v, want %v", tt.status, tt.status, tt.want)
			}
		})
	}
}

func TestTaskType_Values(t *testing.T) {
	tests := []struct {
		name string
		typ  TaskType
		want string
	}{
		{"once", TaskTypeOnce, "once"},
		{"recurring", TaskTypeRecurring, "recurring"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.typ) != tt.want {
				t.Errorf("TaskType(%v) = %v, want %v", tt.typ, tt.typ, tt.want)
			}
		})
	}
}

func TestCronTask_TableName(t *testing.T) {
	task := CronTask{}
	want := "cron_task"
	if got := task.TableName(); got != want {
		t.Errorf("CronTask.TableName() = %v, want %v", got, want)
	}
}

func TestCronTask_Fields(t *testing.T) {
	now := time.Now()
	nullTime := sql.NullTime{Time: now, Valid: true}

	// 验证 Go 零值
	task := CronTask{}
	if task.Status != "" {
		t.Errorf("zero value Status = %v, want empty string", task.Status)
	}
	if task.MaxRetries != 0 {
		t.Errorf("zero value MaxRetries = %v, want 0", task.MaxRetries)
	}
	if task.TotalExecutions != 0 {
		t.Errorf("zero value TotalExecutions = %v, want 0", task.TotalExecutions)
	}
	if task.Enabled != false {
		t.Errorf("zero value Enabled = %v, want false", task.Enabled)
	}

	// 验证字段可设置性（GORM default 标签在数据库层面生效）
	task.Name = "test-task"
	task.Spec = "0 * * * *"
	task.Type = TaskTypeRecurring
	task.ExecType = "http"
	task.Raw = `{"url": "http://example.com"}`
	task.Enabled = true
	task.Status = TaskStatusPending
	task.MaxRetries = 3
	task.ValidFrom = nullTime
	task.ValidUntil = nullTime
	task.LastExecutedAt = nullTime
	task.NextExecutionAt = nullTime

	if task.Name != "test-task" {
		t.Error("Name should be settable")
	}
	if task.Type != TaskTypeRecurring {
		t.Errorf("Type = %v, want %v", task.Type, TaskTypeRecurring)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusPending)
	}
	if task.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want 3", task.MaxRetries)
	}
	if !task.ValidFrom.Valid || !task.ValidFrom.Time.Equal(now) {
		t.Error("ValidFrom should be settable")
	}
	if !task.ValidUntil.Valid || !task.ValidUntil.Time.Equal(now) {
		t.Error("ValidUntil should be settable")
	}
}

func TestCronTask_WithType(t *testing.T) {
	tests := []struct {
		name string
		typ  TaskType
	}{
		{"once task", TaskTypeOnce},
		{"recurring task", TaskTypeRecurring},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := CronTask{
				Name:     "test-task",
				Type:     tt.typ,
				ExecType: "http",
			}
			if task.Type != tt.typ {
				t.Errorf("Task.Type = %v, want %v", task.Type, tt.typ)
			}
		})
	}
}

func TestSqlNullTimeUsage(t *testing.T) {
	now := time.Now()

	// 测试创建 Valid NullTime
	validNullTime := sql.NullTime{Time: now, Valid: true}
	if !validNullTime.Valid {
		t.Error("Valid should be true")
	}
	if !validNullTime.Time.Equal(now) {
		t.Error("Time should equal now")
	}

	// 测试创建 Invalid NullTime（表示 NULL）
	invalidNullTime := sql.NullTime{Valid: false}
	if invalidNullTime.Valid {
		t.Error("Valid should be false")
	}
}

func TestCronTask_DatabaseOperations(t *testing.T) {
	db := newTestDB(t)

	// Create a task
	task := CronTask{
		Name:     "test-task",
		Spec:     "0 * * * *",
		Type:     TaskTypeRecurring,
		ExecType: "http",
		Raw:      `{"url": "http://example.com"}`,
		Enabled:  true,
		Status:   TaskStatusPending,
		MaxRetries: 3,
	}

		if err := db.Create(&task).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was assigned
	if task.ID == 0 {
		t.Error("ID should be auto-assigned after create")
	}

	// Query the task
	var queried CronTask
	if err := db.First(&queried, task.ID).Error; err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Verify fields
	if queried.Name != task.Name {
		t.Errorf("Name mismatch: got %v, want %v", queried.Name, task.Name)
	}
	if queried.Spec != task.Spec {
		t.Errorf("Spec mismatch: got %v, want %v", queried.Spec, task.Spec)
	}
	if queried.Type != task.Type {
		t.Errorf("Type mismatch: got %v, want %v", queried.Type, task.Type)
	}
	if queried.Enabled != task.Enabled {
		t.Errorf("Enabled mismatch: got %v, want %v", queried.Enabled, task.Enabled)
	}
	if queried.MaxRetries != task.MaxRetries {
		t.Errorf("MaxRetries mismatch: got %v, want %v", queried.MaxRetries, task.MaxRetries)
	}
}
