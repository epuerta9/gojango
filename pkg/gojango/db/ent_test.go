package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

// Mock implementations for testing

type MockEntClient struct {
	closeError  error
	schemaAPI   SchemaAPI
	txError     error
	transaction Transaction
}

func (m *MockEntClient) Close() error {
	return m.closeError
}

func (m *MockEntClient) Schema() SchemaAPI {
	return m.schemaAPI
}

func (m *MockEntClient) Tx(ctx context.Context) (Transaction, error) {
	if m.txError != nil {
		return nil, m.txError
	}
	return m.transaction, nil
}

type MockSchemaAPI struct {
	createError error
	writeError  error
}

func (m *MockSchemaAPI) Create(ctx context.Context, opts ...MigrateOption) error {
	return m.createError
}

func (m *MockSchemaAPI) WriteTo(ctx context.Context, w interface{}) error {
	return m.writeError
}

type MockTransaction struct {
	commitError   error
	rollbackError error
}

func (m *MockTransaction) Commit() error {
	return m.commitError
}

func (m *MockTransaction) Rollback() error {
	return m.rollbackError
}

func TestMigrateOptions(t *testing.T) {
	config := &MigrateConfig{}

	// Test WithDisableForeignKeys
	opt1 := WithDisableForeignKeys()
	opt1(config)
	if !config.DisableForeignKeys {
		t.Errorf("Expected DisableForeignKeys to be true")
	}

	// Test WithDropColumns
	config = &MigrateConfig{}
	opt2 := WithDropColumns()
	opt2(config)
	if !config.DropColumns {
		t.Errorf("Expected DropColumns to be true")
	}

	// Test WithDropIndexes
	config = &MigrateConfig{}
	opt3 := WithDropIndexes()
	opt3(config)
	if !config.DropIndexes {
		t.Errorf("Expected DropIndexes to be true")
	}

	// Test multiple options
	config = &MigrateConfig{}
	opt1(config)
	opt2(config)
	opt3(config)
	if !config.DisableForeignKeys || !config.DropColumns || !config.DropIndexes {
		t.Errorf("Expected all options to be true")
	}
}

func TestNewEntManager(t *testing.T) {
	manager := NewEntManager()

	if manager == nil {
		t.Fatal("Expected manager to be non-nil")
	}

	if manager.connections == nil {
		t.Error("Expected connections map to be initialized")
	}

	if manager.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if manager.defaultConn != "" {
		t.Errorf("Expected default connection to be empty, got %s", manager.defaultConn)
	}
}

func TestEntManagerAddConnection(t *testing.T) {
	manager := NewEntManager()

	// Test adding nil connection
	err := manager.AddConnection("test", nil)
	if err == nil {
		t.Error("Expected error when adding nil connection")
	}

	// Test adding valid connection
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Errorf("Failed to add connection: %v", err)
	}

	// Should be set as default
	if manager.defaultConn != "test" {
		t.Errorf("Expected default connection to be test, got %s", manager.defaultConn)
	}

	// Add second connection
	tempDir2 := t.TempDir()
	dbPath2 := filepath.Join(tempDir2, "test2.db")
	config2 := SQLiteConfig(dbPath2)
	conn2, err := Open(config2)
	if err != nil {
		t.Fatalf("Failed to create second test connection: %v", err)
	}
	defer conn2.Close()

	err = manager.AddConnection("test2", conn2)
	if err != nil {
		t.Errorf("Failed to add second connection: %v", err)
	}

	// Default should still be the first
	if manager.defaultConn != "test" {
		t.Errorf("Expected default connection to still be test, got %s", manager.defaultConn)
	}
}

func TestEntManagerSetClient(t *testing.T) {
	manager := NewEntManager()

	// Test setting client for non-existent connection
	mockClient := &MockEntClient{}
	err := manager.SetClient("nonexistent", mockClient)
	if err == nil {
		t.Error("Expected error when setting client for non-existent connection")
	}

	// Add connection first
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Test setting nil client
	err = manager.SetClient("test", nil)
	if err == nil {
		t.Error("Expected error when setting nil client")
	}

	// Test setting valid client
	err = manager.SetClient("test", mockClient)
	if err != nil {
		t.Errorf("Failed to set client: %v", err)
	}

	// Verify client was set
	client, err := manager.GetClient("test")
	if err != nil {
		t.Errorf("Failed to get client: %v", err)
	}
	if client != mockClient {
		t.Error("Expected client to be the mock client")
	}
}

func TestEntManagerGetClient(t *testing.T) {
	manager := NewEntManager()

	// Test getting non-existent client
	_, err := manager.GetClient("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent client")
	}

	// Add connection and client
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	mockClient := &MockEntClient{}
	err = manager.SetClient("test", mockClient)
	if err != nil {
		t.Fatalf("Failed to set client: %v", err)
	}

	// Test getting existing client
	client, err := manager.GetClient("test")
	if err != nil {
		t.Errorf("Failed to get client: %v", err)
	}
	if client != mockClient {
		t.Error("Expected client to be the mock client")
	}
}

func TestEntManagerDefault(t *testing.T) {
	manager := NewEntManager()

	// Test getting default from empty manager
	_, err := manager.Default()
	if err == nil {
		t.Error("Expected error when getting default from empty manager")
	}

	// Add connection and client
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	mockClient := &MockEntClient{}
	err = manager.SetClient("test", mockClient)
	if err != nil {
		t.Fatalf("Failed to set client: %v", err)
	}

	// Test getting default
	client, err := manager.Default()
	if err != nil {
		t.Errorf("Failed to get default client: %v", err)
	}
	if client != mockClient {
		t.Error("Expected default client to be the mock client")
	}
}

func TestEntManagerCreateDriver(t *testing.T) {
	manager := NewEntManager()

	// Test creating driver for non-existent connection
	_, err := manager.CreateDriver("nonexistent")
	if err == nil {
		t.Error("Expected error when creating driver for non-existent connection")
	}

	// Test SQLite driver
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("sqlite", conn)
	if err != nil {
		t.Fatalf("Failed to add SQLite connection: %v", err)
	}

	driver, err := manager.CreateDriver("sqlite")
	if err != nil {
		t.Errorf("Failed to create SQLite driver: %v", err)
	}
	if driver == nil {
		t.Error("Expected driver to be non-nil")
	}

	// Test PostgreSQL driver (simulated)
	pgConfig := PostgresConfig("localhost", "test", "user", "pass")
	// We can't actually connect to PostgreSQL in tests, so we'll just test the config setup
	pgConn := &Connection{
		db:     nil, // This would normally be set by Open()
		config: pgConfig,
	}

	err = manager.AddConnection("postgres", pgConn)
	if err != nil {
		t.Fatalf("Failed to add PostgreSQL connection: %v", err)
	}

	// This will fail because we don't have a real DB connection, but it tests the driver selection logic
	_, err = manager.CreateDriver("postgres")
	// We expect an error here because the db is nil, but not because of unsupported driver
	if err != nil && err.Error() == "unsupported driver for Ent: postgres" {
		t.Errorf("PostgreSQL driver should be supported")
	}

	// Test unsupported driver
	unsupportedConfig := &Config{Driver: Driver("unsupported")}
	unsupportedConn := &Connection{config: unsupportedConfig}
	err = manager.AddConnection("unsupported", unsupportedConn)
	if err != nil {
		t.Fatalf("Failed to add unsupported connection: %v", err)
	}

	_, err = manager.CreateDriver("unsupported")
	if err == nil {
		t.Error("Expected error for unsupported driver")
	}
}

func TestEntManagerMigrate(t *testing.T) {
	manager := NewEntManager()
	ctx := context.Background()

	// Test migrate with non-existent client
	err := manager.Migrate(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when migrating non-existent client")
	}

	// Set up connection and mock client
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Test successful migration
	mockSchema := &MockSchemaAPI{createError: nil}
	mockClient := &MockEntClient{schemaAPI: mockSchema}
	err = manager.SetClient("test", mockClient)
	if err != nil {
		t.Fatalf("Failed to set client: %v", err)
	}

	err = manager.Migrate(ctx, "test")
	if err != nil {
		t.Errorf("Failed to migrate: %v", err)
	}

	// Test migration failure
	mockSchemaFail := &MockSchemaAPI{createError: errors.New("migration failed")}
	mockClientFail := &MockEntClient{schemaAPI: mockSchemaFail}
	err = manager.SetClient("test", mockClientFail)
	if err != nil {
		t.Fatalf("Failed to set failing client: %v", err)
	}

	err = manager.Migrate(ctx, "test")
	if err == nil {
		t.Error("Expected migration to fail")
	}
}

func TestEntManagerMigrateAll(t *testing.T) {
	manager := NewEntManager()
	ctx := context.Background()

	// Test migrate all with no clients
	err := manager.MigrateAll(ctx)
	if err != nil {
		t.Errorf("Expected no error when no clients registered: %v", err)
	}

	// Set up multiple connections and clients
	for i := 1; i <= 3; i++ {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")
		config := SQLiteConfig(dbPath)
		conn, err := Open(config)
		if err != nil {
			t.Fatalf("Failed to create test connection %d: %v", i, err)
		}
		defer conn.Close()

		connName := fmt.Sprintf("test%d", i)
		err = manager.AddConnection(connName, conn)
		if err != nil {
			t.Fatalf("Failed to add connection %s: %v", connName, err)
		}

		mockSchema := &MockSchemaAPI{createError: nil}
		mockClient := &MockEntClient{schemaAPI: mockSchema}
		err = manager.SetClient(connName, mockClient)
		if err != nil {
			t.Fatalf("Failed to set client %s: %v", connName, err)
		}
	}

	// Test successful migration of all
	err = manager.MigrateAll(ctx)
	if err != nil {
		t.Errorf("Failed to migrate all: %v", err)
	}
}

func TestEntManagerWithTransaction(t *testing.T) {
	manager := NewEntManager()
	ctx := context.Background()

	// Test transaction with non-existent client
	err := manager.WithTransaction(ctx, "nonexistent", func(ctx context.Context, tx Transaction) error {
		return nil
	})
	if err == nil {
		t.Error("Expected error when using transaction with non-existent client")
	}

	// Set up connection and client
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Test successful transaction
	mockTx := &MockTransaction{}
	mockClient := &MockEntClient{transaction: mockTx}
	err = manager.SetClient("test", mockClient)
	if err != nil {
		t.Fatalf("Failed to set client: %v", err)
	}

	executed := false
	err = manager.WithTransaction(ctx, "test", func(ctx context.Context, tx Transaction) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Errorf("Failed to execute transaction: %v", err)
	}
	if !executed {
		t.Error("Transaction function was not executed")
	}

	// Test transaction creation failure
	mockClientTxFail := &MockEntClient{txError: errors.New("tx failed")}
	err = manager.SetClient("test", mockClientTxFail)
	if err != nil {
		t.Fatalf("Failed to set failing client: %v", err)
	}

	err = manager.WithTransaction(ctx, "test", func(ctx context.Context, tx Transaction) error {
		return nil
	})
	if err == nil {
		t.Error("Expected error when transaction creation fails")
	}

	// Test transaction function failure with rollback
	mockTxRollback := &MockTransaction{rollbackError: nil}
	mockClientRollback := &MockEntClient{transaction: mockTxRollback}
	err = manager.SetClient("test", mockClientRollback)
	if err != nil {
		t.Fatalf("Failed to set rollback client: %v", err)
	}

	err = manager.WithTransaction(ctx, "test", func(ctx context.Context, tx Transaction) error {
		return errors.New("function failed")
	})
	if err == nil {
		t.Error("Expected error when transaction function fails")
	}

	// Test commit failure
	mockTxCommitFail := &MockTransaction{commitError: errors.New("commit failed")}
	mockClientCommitFail := &MockEntClient{transaction: mockTxCommitFail}
	err = manager.SetClient("test", mockClientCommitFail)
	if err != nil {
		t.Fatalf("Failed to set commit fail client: %v", err)
	}

	err = manager.WithTransaction(ctx, "test", func(ctx context.Context, tx Transaction) error {
		return nil
	})
	if err == nil {
		t.Error("Expected error when commit fails")
	}
}

func TestEntManagerStats(t *testing.T) {
	manager := NewEntManager()

	// Test stats for non-existent connection
	_, err := manager.Stats("nonexistent")
	if err == nil {
		t.Error("Expected error when getting stats for non-existent connection")
	}

	// Set up connection
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Test getting stats
	stats, err := manager.Stats("test")
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
	}

	// Verify some basic stats fields exist (actual values depend on driver and config)
	_ = stats.MaxOpenConnections
	_ = stats.OpenConnections
	_ = stats.Idle
	_ = stats.InUse
}

func TestEntManagerListMethods(t *testing.T) {
	manager := NewEntManager()

	// Test empty lists
	connections := manager.ListConnections()
	if len(connections) != 0 {
		t.Errorf("Expected empty connections list, got %d", len(connections))
	}

	clients := manager.ListClients()
	if len(clients) != 0 {
		t.Errorf("Expected empty clients list, got %d", len(clients))
	}

	// Add connections and clients
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	mockClient := &MockEntClient{}
	err = manager.SetClient("test", mockClient)
	if err != nil {
		t.Fatalf("Failed to set client: %v", err)
	}

	// Test populated lists
	connections = manager.ListConnections()
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(connections))
	}
	if connections[0] != "test" {
		t.Errorf("Expected connection name to be test, got %s", connections[0])
	}

	clients = manager.ListClients()
	if len(clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients))
	}
	if clients[0] != "test" {
		t.Errorf("Expected client name to be test, got %s", clients[0])
	}
}

func TestEntManagerCloseAll(t *testing.T) {
	manager := NewEntManager()

	// Set up multiple connections and clients
	for i := 1; i <= 2; i++ {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")
		config := SQLiteConfig(dbPath)
		conn, err := Open(config)
		if err != nil {
			t.Fatalf("Failed to create test connection %d: %v", i, err)
		}
		// Don't defer close here since CloseAll will handle it

		connName := fmt.Sprintf("test%d", i)
		err = manager.AddConnection(connName, conn)
		if err != nil {
			t.Fatalf("Failed to add connection %s: %v", connName, err)
		}

		mockClient := &MockEntClient{closeError: nil}
		err = manager.SetClient(connName, mockClient)
		if err != nil {
			t.Fatalf("Failed to set client %s: %v", connName, err)
		}
	}

	// Verify connections and clients exist
	if len(manager.connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(manager.connections))
	}
	if len(manager.clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(manager.clients))
	}

	// Test close all
	err := manager.CloseAll()
	if err != nil {
		t.Errorf("Failed to close all: %v", err)
	}

	// Verify everything was cleared
	if len(manager.connections) != 0 {
		t.Errorf("Expected 0 connections after close all, got %d", len(manager.connections))
	}
	if len(manager.clients) != 0 {
		t.Errorf("Expected 0 clients after close all, got %d", len(manager.clients))
	}
	if manager.defaultConn != "" {
		t.Errorf("Expected empty default connection after close all, got %s", manager.defaultConn)
	}

	// Test close all with client close error
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := SQLiteConfig(dbPath)
	conn, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}

	err = manager.AddConnection("test", conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	mockClientError := &MockEntClient{closeError: errors.New("close failed")}
	err = manager.SetClient("test", mockClientError)
	if err != nil {
		t.Fatalf("Failed to set client: %v", err)
	}

	err = manager.CloseAll()
	if err == nil {
		t.Error("Expected error when client close fails")
	}
}