package repository

import (
	"path/filepath"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

// newTestServiceContext creates a new ServiceContext with in-memory SQLite database for testing
func newTestServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()

	db := newTestDB(t)
	return &svc.ServiceContext{DB: db}
}

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
		&model.CronTask{},
		&model.CronExecution{},
		&model.CronExecutionLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}
