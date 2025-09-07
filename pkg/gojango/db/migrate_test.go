package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestMigrator(t *testing.T) (*Migrator, string, func()) {
	// Create temporary directory for test database and migrations
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	migrationsPath := filepath.Join(tempDir, "migrations")

	// Create database connection
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create migrator
	migrator := NewMigrator(conn, migrationsPath)

	// Return cleanup function
	cleanup := func() {
		conn.Close()
	}

	return migrator, migrationsPath, cleanup
}

func createTestMigration(t *testing.T, migrationsPath string, id int, name, upSQL, downSQL string) {
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		t.Fatalf("Failed to create migrations directory: %v", err)
	}

	upFile := filepath.Join(migrationsPath, fmt.Sprintf("%04d_%s_up.sql", id, name))
	downFile := filepath.Join(migrationsPath, fmt.Sprintf("%04d_%s_down.sql", id, name))

	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		t.Fatalf("Failed to create up migration: %v", err)
	}

	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		t.Fatalf("Failed to create down migration: %v", err)
	}
}

func TestNewMigrator(t *testing.T) {
	migrator, _, cleanup := setupTestMigrator(t)
	defer cleanup()

	if migrator == nil {
		t.Fatal("Expected migrator to be non-nil")
	}

	if migrator.tableName != "gojango_migrations" {
		t.Errorf("Expected default table name to be gojango_migrations, got %s", migrator.tableName)
	}
}

func TestSetMigrationsTable(t *testing.T) {
	migrator, _, cleanup := setupTestMigrator(t)
	defer cleanup()

	customTableName := "custom_migrations"
	migrator.SetMigrationsTable(customTableName)

	if migrator.tableName != customTableName {
		t.Errorf("Expected table name to be %s, got %s", customTableName, migrator.tableName)
	}
}

func TestMigratorInitialize(t *testing.T) {
	migrator, _, cleanup := setupTestMigrator(t)
	defer cleanup()

	ctx := context.Background()

	// Test initialization
	err := migrator.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Verify migrations table was created
	db := migrator.conn.DB()
	var count int
	query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	err = db.QueryRowContext(ctx, query, migrator.tableName).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check if migrations table exists: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected migrations table to exist, count: %d", count)
	}

	// Test multiple initializations (should not error)
	err = migrator.Initialize(ctx)
	if err != nil {
		t.Errorf("Expected multiple initializations to succeed: %v", err)
	}
}

func TestDiscoverMigrations(t *testing.T) {
	migrator, migrationsPath, cleanup := setupTestMigrator(t)
	defer cleanup()

	// Test with no migrations directory
	migrations, err := migrator.DiscoverMigrations()
	if err != nil {
		t.Errorf("Expected no error when migrations directory doesn't exist: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("Expected 0 migrations, got %d", len(migrations))
	}

	// Create test migrations (only up files for discovery test)
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		t.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Create migration files manually for precise control
	upFiles := []struct {
		id   int
		name string
		sql  string
	}{
		{1, "initial", "CREATE TABLE users (id INTEGER PRIMARY KEY);"},
		{3, "add_email", "ALTER TABLE users ADD COLUMN email TEXT;"},
		{2, "add_posts", "CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT);"},
	}

	for _, migration := range upFiles {
		upFile := filepath.Join(migrationsPath, fmt.Sprintf("%04d_%s_up.sql", migration.id, migration.name))
		downFile := filepath.Join(migrationsPath, fmt.Sprintf("%04d_%s_down.sql", migration.id, migration.name))
		
		if err := os.WriteFile(upFile, []byte(migration.sql), 0644); err != nil {
			t.Fatalf("Failed to create up migration: %v", err)
		}
		
		if err := os.WriteFile(downFile, []byte("-- Rollback"), 0644); err != nil {
			t.Fatalf("Failed to create down migration: %v", err)
		}
	}

	// Discover migrations
	migrations, err = migrator.DiscoverMigrations()
	if err != nil {
		t.Fatalf("Failed to discover migrations: %v", err)
	}

	if len(migrations) != 3 {
		t.Errorf("Expected 3 migrations, got %d", len(migrations))
	}

	// Check if migrations are sorted by ID
	expectedOrder := []int{1, 2, 3}
	expectedNames := []string{"initial", "add_posts", "add_email"}
	
	for i, migration := range migrations {
		if migration.ID != expectedOrder[i] {
			t.Errorf("Expected migration ID %d at position %d, got %d", expectedOrder[i], i, migration.ID)
		}
		if migration.Name != expectedNames[i] {
			t.Errorf("Expected migration name %s at position %d, got %s", expectedNames[i], i, migration.Name)
		}
		if migration.SQL == "" {
			t.Errorf("Expected migration %d to have SQL content", migration.ID)
		}
		if migration.RollbackSQL == "" {
			t.Errorf("Expected migration %d to have rollback SQL content", migration.ID)
		}
	}
}

func TestParseMigrationFile(t *testing.T) {
	migrator, migrationsPath, cleanup := setupTestMigrator(t)
	defer cleanup()

	tests := []struct {
		name      string
		filename  string
		shouldErr bool
		expectedID int
		expectedName string
	}{
		{"Valid up migration", "0001_initial_up.sql", false, 1, "initial"},
		{"Valid regular migration", "0002_add_users.sql", false, 2, "add_users"},
		{"Valid complex name", "0003_add_user_profile_table.sql", false, 3, "add_user_profile_table"},
		{"Invalid format", "invalid.sql", true, 0, ""},
		{"Invalid ID", "abc_invalid.sql", true, 0, ""},
		{"Missing parts", "001.sql", true, 0, ""},
	}

	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		t.Fatalf("Failed to create migrations directory: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			testFile := filepath.Join(migrationsPath, tt.filename)
			content := "-- Test migration content"
			err := os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(testFile)

			migration, err := migrator.parseMigrationFile(tt.filename, testFile)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for filename %s", tt.filename)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for filename %s: %v", tt.filename, err)
				return
			}

			if migration.ID != tt.expectedID {
				t.Errorf("Expected ID %d, got %d", tt.expectedID, migration.ID)
			}

			if migration.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, migration.Name)
			}

			if migration.SQL != content {
				t.Errorf("Expected SQL content to match")
			}
		})
	}
}

func TestMigratorFullWorkflow(t *testing.T) {
	migrator, migrationsPath, cleanup := setupTestMigrator(t)
	defer cleanup()

	ctx := context.Background()

	// Initialize migrator
	err := migrator.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Create test migrations
	createTestMigration(t, migrationsPath, 1, "create_users", 
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);",
		"DROP TABLE users;")
	
	createTestMigration(t, migrationsPath, 2, "create_posts", 
		"CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER);",
		"DROP TABLE posts;")

	// Test initial status - all pending
	status, err := migrator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get initial status: %v", err)
	}

	if len(status.Applied) != 0 {
		t.Errorf("Expected 0 applied migrations initially, got %d", len(status.Applied))
	}

	if len(status.Pending) != 2 {
		t.Errorf("Expected 2 pending migrations initially, got %d", len(status.Pending))
	}

	if status.LastApplied != nil {
		t.Errorf("Expected no last applied migration initially")
	}

	// Apply migrations
	err = migrator.Apply(ctx)
	if err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	// Verify tables were created
	db := migrator.conn.DB()
	var userTableCount, postTableCount int
	
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&userTableCount)
	if err != nil {
		t.Fatalf("Failed to check users table: %v", err)
	}
	if userTableCount != 1 {
		t.Errorf("Expected users table to exist")
	}

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&postTableCount)
	if err != nil {
		t.Fatalf("Failed to check posts table: %v", err)
	}
	if postTableCount != 1 {
		t.Errorf("Expected posts table to exist")
	}

	// Test status after applying - all applied
	status, err = migrator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get status after apply: %v", err)
	}

	if len(status.Applied) != 2 {
		t.Errorf("Expected 2 applied migrations, got %d", len(status.Applied))
	}

	if len(status.Pending) != 0 {
		t.Errorf("Expected 0 pending migrations, got %d", len(status.Pending))
	}

	if status.LastApplied == nil {
		t.Errorf("Expected last applied migration to be set")
	} else if status.LastApplied.ID != 2 {
		t.Errorf("Expected last applied migration ID to be 2, got %d", status.LastApplied.ID)
	}

	// Test rollback
	err = migrator.Rollback(ctx)
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify posts table was dropped
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&postTableCount)
	if err != nil {
		t.Fatalf("Failed to check posts table after rollback: %v", err)
	}
	if postTableCount != 0 {
		t.Errorf("Expected posts table to be dropped after rollback")
	}

	// Verify users table still exists
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&userTableCount)
	if err != nil {
		t.Fatalf("Failed to check users table after rollback: %v", err)
	}
	if userTableCount != 1 {
		t.Errorf("Expected users table to still exist after rollback")
	}

	// Test status after rollback
	status, err = migrator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get status after rollback: %v", err)
	}

	if len(status.Applied) != 1 {
		t.Errorf("Expected 1 applied migration after rollback, got %d", len(status.Applied))
	}

	if len(status.Pending) != 1 {
		t.Errorf("Expected 1 pending migration after rollback, got %d", len(status.Pending))
	}

	if status.LastApplied == nil || status.LastApplied.ID != 1 {
		t.Errorf("Expected last applied migration to be ID 1 after rollback")
	}

	// Test reset (rollback all)
	err = migrator.Reset(ctx)
	if err != nil {
		t.Fatalf("Failed to reset migrations: %v", err)
	}

	// Verify all tables were dropped
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&userTableCount)
	if err != nil {
		t.Fatalf("Failed to check users table after reset: %v", err)
	}
	if userTableCount != 0 {
		t.Errorf("Expected users table to be dropped after reset")
	}

	// Test status after reset - all pending
	status, err = migrator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get status after reset: %v", err)
	}

	if len(status.Applied) != 0 {
		t.Errorf("Expected 0 applied migrations after reset, got %d", len(status.Applied))
	}

	if len(status.Pending) != 2 {
		t.Errorf("Expected 2 pending migrations after reset, got %d", len(status.Pending))
	}

	if status.LastApplied != nil {
		t.Errorf("Expected no last applied migration after reset")
	}
}

func TestMigratorEdgeCases(t *testing.T) {
	migrator, migrationsPath, cleanup := setupTestMigrator(t)
	defer cleanup()

	ctx := context.Background()

	// Initialize migrator
	err := migrator.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Test apply with no migrations
	err = migrator.Apply(ctx)
	if err != nil {
		t.Errorf("Expected no error when applying with no migrations: %v", err)
	}

	// Test rollback with no applied migrations
	err = migrator.Rollback(ctx)
	if err != nil {
		t.Errorf("Expected no error when rolling back with no applied migrations: %v", err)
	}

	// Test reset with no applied migrations
	err = migrator.Reset(ctx)
	if err != nil {
		t.Errorf("Expected no error when resetting with no applied migrations: %v", err)
	}

	// Create migration without rollback SQL
	createTestMigration(t, migrationsPath, 1, "no_rollback", 
		"CREATE TABLE test (id INTEGER);",
		"")

	// Apply migration
	err = migrator.Apply(ctx)
	if err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Try to rollback migration without rollback SQL
	err = migrator.Rollback(ctx)
	if err == nil {
		t.Error("Expected error when rolling back migration without rollback SQL")
	}
}

func TestMigratorSQLErrors(t *testing.T) {
	migrator, migrationsPath, cleanup := setupTestMigrator(t)
	defer cleanup()

	ctx := context.Background()

	// Initialize migrator
	err := migrator.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Create migration with invalid SQL
	createTestMigration(t, migrationsPath, 1, "invalid_sql", 
		"INVALID SQL STATEMENT;",
		"DROP TABLE nonexistent;")

	// Try to apply migration - should fail
	err = migrator.Apply(ctx)
	if err == nil {
		t.Error("Expected error when applying migration with invalid SQL")
	}

	// Verify migration was not recorded as applied
	status, err := migrator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status.Applied) != 0 {
		t.Errorf("Expected 0 applied migrations after failed apply, got %d", len(status.Applied))
	}
}

func TestMigratorDifferentDrivers(t *testing.T) {
	// Test PostgreSQL-style queries (can't actually connect to PostgreSQL in tests)
	pgConfig := PostgresConfig("localhost", "test", "user", "pass")
	
	// This is a simplified test - we can't actually test PostgreSQL without a real connection
	// But we can test that the migrator accepts PostgreSQL connections
	migrator := &Migrator{
		conn: &Connection{config: pgConfig},
		migrationsPath: "/tmp/migrations",
		tableName: "test_migrations",
	}

	if migrator.conn.config.Driver != DriverPostgres {
		t.Errorf("Expected PostgreSQL driver")
	}

	// Test MySQL (not implemented yet, should return error)
	mysqlConfig := &Config{Driver: DriverMySQL}
	_, err := mysqlConfig.BuildDSN()
	if err == nil {
		t.Error("Expected error for MySQL driver (not implemented)")
	}
}

func TestGetAppliedMigrations(t *testing.T) {
	migrator, _, cleanup := setupTestMigrator(t)
	defer cleanup()

	ctx := context.Background()

	// Initialize migrator
	err := migrator.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Test with no applied migrations
	applied, err := migrator.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("Expected 0 applied migrations, got %d", len(applied))
	}

	// Manually insert a migration record for testing
	db := migrator.conn.DB()
	_, err = db.ExecContext(ctx, 
		fmt.Sprintf("INSERT INTO %s (name, filename, applied_at) VALUES (?, ?, ?)", migrator.tableName),
		"test_migration", "0001_test_migration.sql", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test migration record: %v", err)
	}

	// Test with applied migrations
	applied, err = migrator.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}
	if len(applied) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(applied))
	}

	migration := applied[0]
	if migration.Name != "test_migration" {
		t.Errorf("Expected migration name to be test_migration, got %s", migration.Name)
	}
	if migration.Filename != "0001_test_migration.sql" {
		t.Errorf("Expected filename to be 0001_test_migration.sql, got %s", migration.Filename)
	}
	if migration.AppliedAt.IsZero() {
		t.Errorf("Expected applied_at to be set")
	}
}