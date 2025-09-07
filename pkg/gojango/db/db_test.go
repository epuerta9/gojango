package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDriverConstants(t *testing.T) {
	tests := []struct {
		driver   Driver
		expected string
	}{
		{DriverPostgres, "postgres"},
		{DriverSQLite, "sqlite3"},
		{DriverMySQL, "mysql"},
	}

	for _, tt := range tests {
		if string(tt.driver) != tt.expected {
			t.Errorf("Driver constant mismatch: got %s, want %s", tt.driver, tt.expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Driver != DriverSQLite {
		t.Errorf("Expected default driver to be sqlite3, got %s", config.Driver)
	}

	if config.Database != "gojango.db" {
		t.Errorf("Expected default database to be gojango.db, got %s", config.Database)
	}

	if config.MaxOpenConns != 25 {
		t.Errorf("Expected MaxOpenConns to be 25, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 25 {
		t.Errorf("Expected MaxIdleConns to be 25, got %d", config.MaxIdleConns)
	}

	if config.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("Expected ConnMaxLifetime to be 5m, got %v", config.ConnMaxLifetime)
	}
}

func TestPostgresConfig(t *testing.T) {
	config := PostgresConfig("localhost", "testdb", "testuser", "testpass")

	if config.Driver != DriverPostgres {
		t.Errorf("Expected driver to be postgres, got %s", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host to be localhost, got %s", config.Host)
	}

	if config.Port != 5432 {
		t.Errorf("Expected port to be 5432, got %d", config.Port)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected database to be testdb, got %s", config.Database)
	}

	if config.Username != "testuser" {
		t.Errorf("Expected username to be testuser, got %s", config.Username)
	}

	if config.Password != "testpass" {
		t.Errorf("Expected password to be testpass, got %s", config.Password)
	}

	if config.SSLMode != "disable" {
		t.Errorf("Expected SSLMode to be disable, got %s", config.SSLMode)
	}
}

func TestSQLiteConfig(t *testing.T) {
	path := "/tmp/test.db"
	config := SQLiteConfig(path)

	if config.Driver != DriverSQLite {
		t.Errorf("Expected driver to be sqlite3, got %s", config.Driver)
	}

	if config.Database != path {
		t.Errorf("Expected database path to be %s, got %s", path, config.Database)
	}

	if config.MaxOpenConns != 1 {
		t.Errorf("Expected MaxOpenConns to be 1 for SQLite, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 1 {
		t.Errorf("Expected MaxIdleConns to be 1 for SQLite, got %d", config.MaxIdleConns)
	}

	if config.ConnMaxLifetime != 0 {
		t.Errorf("Expected ConnMaxLifetime to be 0 for SQLite, got %v", config.ConnMaxLifetime)
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
		hasError bool
	}{
		{
			name: "PostgreSQL DSN",
			config: &Config{
				Driver:   DriverPostgres,
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user dbname=testdb sslmode=disable password=pass",
			hasError: false,
		},
		{
			name: "PostgreSQL DSN without password",
			config: &Config{
				Driver:   DriverPostgres,
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user dbname=testdb sslmode=disable",
			hasError: false,
		},
		{
			name: "SQLite DSN with file",
			config: &Config{
				Driver:   DriverSQLite,
				Database: "/tmp/test.db",
			},
			expected: "/tmp/test.db",
			hasError: false,
		},
		{
			name: "SQLite DSN memory",
			config: &Config{
				Driver:   DriverSQLite,
				Database: "",
			},
			expected: ":memory:",
			hasError: false,
		},
		{
			name: "Custom DSN",
			config: &Config{
				Driver: DriverPostgres,
				DSN:    "custom://dsn/string",
			},
			expected: "custom://dsn/string",
			hasError: false,
		},
		{
			name: "MySQL not implemented",
			config: &Config{
				Driver: DriverMySQL,
			},
			expected: "",
			hasError: true,
		},
		{
			name: "Unsupported driver",
			config: &Config{
				Driver: Driver("unsupported"),
			},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, err := tt.config.BuildDSN()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if dsn != tt.expected {
				t.Errorf("DSN mismatch: got %s, want %s", dsn, tt.expected)
			}
		})
	}
}

func TestSQLiteConnection(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := SQLiteConfig(dbPath)

	// Test connection
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to open SQLite connection: %v", err)
	}
	defer conn.Close()

	// Test ping
	if err := conn.Ping(); err != nil {
		t.Errorf("Failed to ping SQLite database: %v", err)
	}

	// Test driver
	if conn.Driver() != DriverSQLite {
		t.Errorf("Expected driver to be sqlite3, got %s", conn.Driver())
	}

	// Test config
	if conn.Config().Database != dbPath {
		t.Errorf("Expected database path to be %s, got %s", dbPath, conn.Config().Database)
	}

	// Test stats
	stats := conn.Stats()
	if stats.MaxOpenConnections != 1 {
		t.Errorf("Expected MaxOpenConnections to be 1, got %d", stats.MaxOpenConnections)
	}

	// Test database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created: %s", dbPath)
	}
}

func TestSQLiteInMemoryConnection(t *testing.T) {
	config := &Config{
		Driver:          DriverSQLite,
		Database:        "",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
	}

	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite connection: %v", err)
	}
	defer conn.Close()

	// Test basic operations
	db := conn.DB()
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Errorf("Failed to create test table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test")
	if err != nil {
		t.Errorf("Failed to insert test data: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query test data: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}
}

func TestManager(t *testing.T) {
	manager := NewManager()

	// Test empty manager
	if manager.defaultConn != "" {
		t.Errorf("Expected default connection to be empty, got %s", manager.defaultConn)
	}

	_, err := manager.Default()
	if err == nil {
		t.Errorf("Expected error when getting default from empty manager")
	}

	// Add first connection
	config1 := SQLiteConfig(":memory:")
	err = manager.AddConnection("db1", config1)
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	// Should be set as default
	if manager.defaultConn != "db1" {
		t.Errorf("Expected default connection to be db1, got %s", manager.defaultConn)
	}

	// Test getting default
	conn, err := manager.Default()
	if err != nil {
		t.Errorf("Failed to get default connection: %v", err)
	}
	if conn == nil {
		t.Errorf("Expected connection to be non-nil")
	}

	// Add second connection
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test2.db")
	config2 := SQLiteConfig(dbPath)
	err = manager.AddConnection("db2", config2)
	if err != nil {
		t.Fatalf("Failed to add second connection: %v", err)
	}

	// Default should still be db1
	if manager.defaultConn != "db1" {
		t.Errorf("Expected default connection to still be db1, got %s", manager.defaultConn)
	}

	// Test getting specific connection
	conn2, err := manager.GetConnection("db2")
	if err != nil {
		t.Errorf("Failed to get db2 connection: %v", err)
	}
	if conn2 == nil {
		t.Errorf("Expected db2 connection to be non-nil")
	}

	// Test setting new default
	err = manager.SetDefault("db2")
	if err != nil {
		t.Errorf("Failed to set new default: %v", err)
	}

	if manager.defaultConn != "db2" {
		t.Errorf("Expected default connection to be db2, got %s", manager.defaultConn)
	}

	// Test getting non-existent connection
	_, err = manager.GetConnection("nonexistent")
	if err == nil {
		t.Errorf("Expected error when getting non-existent connection")
	}

	// Test setting non-existent default
	err = manager.SetDefault("nonexistent")
	if err == nil {
		t.Errorf("Expected error when setting non-existent default")
	}

	// Test close all
	err = manager.CloseAll()
	if err != nil {
		t.Errorf("Failed to close all connections: %v", err)
	}

	// Should be empty after close
	if len(manager.connections) != 0 {
		t.Errorf("Expected connections to be empty after close, got %d", len(manager.connections))
	}

	if manager.defaultConn != "" {
		t.Errorf("Expected default connection to be empty after close, got %s", manager.defaultConn)
	}
}

func TestConnectionFailure(t *testing.T) {
	// Test with invalid PostgreSQL connection
	config := &Config{
		Driver:   DriverPostgres,
		Host:     "nonexistent-host",
		Port:     9999,
		Username: "invalid",
		Password: "invalid",
		Database: "invalid",
		SSLMode:  "disable",
	}

	_, err := Open(config)
	if err == nil {
		t.Errorf("Expected error when connecting to invalid PostgreSQL server")
	}
}

func BenchmarkSQLiteConnection(b *testing.B) {
	config := SQLiteConfig(":memory:")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := Open(config)
		if err != nil {
			b.Fatalf("Failed to open connection: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkSQLiteQuery(b *testing.B) {
	config := SQLiteConfig(":memory:")
	conn, err := Open(config)
	if err != nil {
		b.Fatalf("Failed to open connection: %v", err)
	}
	defer conn.Close()

	db := conn.DB()
	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT INTO bench (value) VALUES (?)", fmt.Sprintf("value_%d", i))
		if err != nil {
			b.Errorf("Failed to insert: %v", err)
		}
	}
}