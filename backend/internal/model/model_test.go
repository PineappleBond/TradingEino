package model

import (
	"path/filepath"
	"testing"

	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

// newTestDB creates a new in-memory SQLite database for testing and auto-migrates all models
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := gorm.Open(gormlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	err = db.AutoMigrate(
		&CronTask{},
		&CronExecution{},
		&CronExecutionLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}
