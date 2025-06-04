package checks

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// Mock database for testing
type mockDB struct {
	pingError  error
	queryError error
	pingDelay  time.Duration
	queryDelay time.Duration
}

func (m *mockDB) PingContext(ctx context.Context) error {
	if m.pingDelay > 0 {
		select {
		case <-time.After(m.pingDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return m.pingError
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.queryDelay > 0 {
		select {
		case <-time.After(m.queryDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, m.queryError
}

func TestDatabaseCheck(t *testing.T) {
	t.Run("NewDatabaseCheck", func(t *testing.T) {
		// We can't easily test with a real database in unit tests,
		// so we'll test the constructor and basic functionality
		db := &sql.DB{} // This will fail on actual operations, but constructor should work
		check := NewDatabaseCheck("test-db", db)

		if check.Name() != "test-db" {
			t.Errorf("Expected name 'test-db', got '%s'", check.Name())
		}
	})

	t.Run("NewDatabaseCheckWithOptions", func(t *testing.T) {
		db := &sql.DB{}
		opts := DatabaseOptions{
			PingTimeout: 10 * time.Second,
			Query:       "SELECT 1",
		}
		check := NewDatabaseCheckWithOptions("test-db", db, opts)

		if check.Name() != "test-db" {
			t.Errorf("Expected name 'test-db', got '%s'", check.Name())
		}

		if check.opts.PingTimeout != 10*time.Second {
			t.Errorf("Expected timeout 10s, got %v", check.opts.PingTimeout)
		}

		if check.opts.Query != "SELECT 1" {
			t.Errorf("Expected query 'SELECT 1', got '%s'", check.opts.Query)
		}
	})

	t.Run("DefaultTimeout", func(t *testing.T) {
		db := &sql.DB{}
		opts := DatabaseOptions{} // No timeout specified
		check := NewDatabaseCheckWithOptions("test-db", db, opts)

		if check.opts.PingTimeout != 5*time.Second {
			t.Errorf("Expected default timeout 5s, got %v", check.opts.PingTimeout)
		}
	})
}

// Note: For more comprehensive database testing, we would need:
// 1. A test database (like SQLite in-memory)
// 2. Or database mocking library
// 3. Integration tests with real database connections
//
// The current tests focus on the constructor and basic structure.
// In a real project, you might want to add integration tests that:
// - Test with a real database connection
// - Test timeout scenarios
// - Test custom query execution
// - Test connection failure scenarios
