package db

import (
	"os"
	"testing"
	"time"
)

func TestThingsDateCode(t *testing.T) {
	tests := []struct {
		date time.Time
		want int64
		desc string
	}{
		{time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC), 132793600, "Apr 10, 2026"},
		{time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC), 132794368, "Apr 16, 2026"},
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 132780160, "Jan 1, 2026"},
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), 132763520, "Dec 31, 2025"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := thingsDateCode(tt.date)
			if got != tt.want {
				t.Errorf("thingsDateCode(%v) = %d, want %d", tt.date, got, tt.want)
			}
		})
	}
}

func TestDecodeThingsDateCode(t *testing.T) {
	tests := []struct {
		code int64
		want time.Time
		desc string
	}{
		{132793600, time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC), "Apr 10, 2026"},
		{132794368, time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC), "Apr 16, 2026"},
		{132780160, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "Jan 1, 2026"},
		{132763520, time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "Dec 31, 2025"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := decodeThingsDateCode(tt.code)
			if !got.Equal(tt.want) {
				t.Errorf("decodeThingsDateCode(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	dates := []time.Time{
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	for _, d := range dates {
		code := thingsDateCode(d)
		decoded := decodeThingsDateCode(code)
		if !decoded.Equal(d) {
			t.Errorf("round trip failed: %v -> %d -> %v", d, code, decoded)
		}
	}
}

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
