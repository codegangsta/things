package db

import (
	"os"
	"testing"
)

func TestDefaultDBPath(t *testing.T) {
	path, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v", err)
	}
	if path == "" {
		t.Error("DefaultDBPath() returned empty string")
	}
	t.Logf("Default DB path: %s", path)
}

func TestOpenAndClose(t *testing.T) {
	path, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v", err)
	}

	// Check if the database file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Things 3 database not found at %s, skipping test", path)
	}

	db := &DB{}
	if err := db.Open(path); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Verify we can query the database
	conn := db.Conn()
	if conn == nil {
		t.Fatal("Conn() returned nil")
	}

	// Try a simple query to verify the connection works
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM TMTask").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query TMTask: %v", err)
	}
	t.Logf("Found %d tasks in the database", count)
}

func TestOpenNonExistentDB(t *testing.T) {
	db := &DB{}
	err := db.Open("/nonexistent/path/to/database.sqlite")
	if err == nil {
		db.Close()
		t.Error("Open() should fail for non-existent database")
	}
}
