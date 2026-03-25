package model

import (
	"testing"
)

func TestCronExecutionLog_TableName(t *testing.T) {
	log := CronExecutionLog{}
	want := "cron_execution_log"
	if got := log.TableName(); got != want {
		t.Errorf("CronExecutionLog.TableName() = %v, want %v", got, want)
	}
}

func TestCronExecutionLog_Fields(t *testing.T) {
	log := CronExecutionLog{
		ExecutionID: 1,
		From:        "executor",
		Level:       "info",
		Message:     "task started",
	}

	// 验证字段值
	if log.ExecutionID != 1 {
		t.Errorf("ExecutionID = %v, want 1", log.ExecutionID)
	}
	if log.From != "executor" {
		t.Errorf("From = %v, want 'executor'", log.From)
	}
	if log.Level != "info" {
		t.Errorf("Level = %v, want 'info'", log.Level)
	}
	if log.Message != "task started" {
		t.Errorf("Message = %v, want 'task started'", log.Message)
	}
	// CreatedAt 由 GORM 自动设置，只需验证字段存在
	_ = log.CreatedAt
}

func TestCronExecutionLog_Levels(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := CronExecutionLog{
				ExecutionID: 1,
				From:        "scheduler",
				Level:       tt.level,
				Message:     "test message",
			}
			if log.Level != tt.level {
				t.Errorf("Level = %v, want %v", log.Level, tt.level)
			}
		})
	}
}

func TestCronExecutionLog_FromField(t *testing.T) {
	tests := []struct {
		name string
		from string
	}{
		{"scheduler", "scheduler"},
		{"executor", "executor"},
		{"retry", "retry"},
		{"validator", "validator"},
		{"callback", "callback"},
		{"custom_stage", "custom_stage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := CronExecutionLog{
				ExecutionID: 1,
				From:        tt.from,
				Level:       "info",
				Message:     "test message",
			}
			if log.From != tt.from {
				t.Errorf("From = %v, want %v", log.From, tt.from)
			}
		})
	}
}

func TestCronExecutionLog_LongMessage(t *testing.T) {
	longMessage := "This is a very long log message that contains a lot of details about the execution process, including various parameters, state information, and debugging data that might be useful for troubleshooting issues."

	log := CronExecutionLog{
		ExecutionID: 1,
		From:        "executor",
		Level:       "debug",
		Message:     longMessage,
	}

	if log.Message != longMessage {
		t.Error("long message should be preserved")
	}
}

func TestCronExecutionLog_DatabaseOperations(t *testing.T) {
	db := newTestDB(t)

	// Create a log entry
	logEntry := CronExecutionLog{
		ExecutionID: 1,
		From:        "test_executor",
		Level:       "info",
		Message:     "test log message",
	}

		if err := db.Create(&logEntry).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was assigned
	if logEntry.ID == 0 {
		t.Error("ID should be auto-assigned after create")
	}

	// Verify CreatedAt was set
	if logEntry.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	// Query the log entry
	var queried CronExecutionLog
	if err := db.First(&queried, logEntry.ID).Error; err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Verify fields
	if queried.ExecutionID != logEntry.ExecutionID {
		t.Errorf("ExecutionID mismatch: got %v, want %v", queried.ExecutionID, logEntry.ExecutionID)
	}
	if queried.From != logEntry.From {
		t.Errorf("From mismatch: got %v, want %v", queried.From, logEntry.From)
	}
	if queried.Level != logEntry.Level {
		t.Errorf("Level mismatch: got %v, want %v", queried.Level, logEntry.Level)
	}
	if queried.Message != logEntry.Message {
		t.Errorf("Message mismatch: got %v, want %v", queried.Message, logEntry.Message)
	}
}
