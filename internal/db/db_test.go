package db

import (
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer database.Close()

	// Verify tasks table exists
	var name string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tasks'").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query tasks table: %v", err)
	}
	if name != "tasks" {
		t.Errorf("expected table name 'tasks', got '%s'", name)
	}

	// Verify foreign_keys is ON
	var foreignKeys int
	err = database.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("failed to check foreign_keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("expected foreign_keys to be 1, got %d", foreignKeys)
	}
}
