// Package db provides Ent ORM integration for Gojango applications.
//
// This module handles:
//   - Ent client lifecycle and management
//   - Schema migration management
//   - Database seeding and fixtures
//   - Transaction management helpers
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

// EntClient represents an interface that Ent clients should implement
// This allows for dependency injection and testing
type EntClient interface {
	Close() error
	Schema() SchemaAPI
	Tx(context.Context) (Transaction, error)
}

// SchemaAPI represents the schema operations available
type SchemaAPI interface {
	Create(context.Context, ...MigrateOption) error
	WriteTo(context.Context, interface{}) error
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
}

// MigrateOption configures migration behavior
type MigrateOption func(*MigrateConfig)

// MigrateConfig holds migration configuration
type MigrateConfig struct {
	DisableForeignKeys bool
	DropColumns        bool
	DropIndexes        bool
}

// WithDisableForeignKeys disables foreign key constraints during migration
func WithDisableForeignKeys() MigrateOption {
	return func(c *MigrateConfig) {
		c.DisableForeignKeys = true
	}
}

// WithDropColumns allows dropping columns during migration
func WithDropColumns() MigrateOption {
	return func(c *MigrateConfig) {
		c.DropColumns = true
	}
}

// WithDropIndexes allows dropping indexes during migration
func WithDropIndexes() MigrateOption {
	return func(c *MigrateConfig) {
		c.DropIndexes = true
	}
}

// EntManager manages Ent client instances with database connections
type EntManager struct {
	connections map[string]*Connection
	clients     map[string]EntClient
	defaultConn string
}

// NewEntManager creates a new Ent client manager
func NewEntManager() *EntManager {
	return &EntManager{
		connections: make(map[string]*Connection),
		clients:     make(map[string]EntClient),
	}
}

// AddConnection adds a database connection for Ent usage
func (m *EntManager) AddConnection(name string, conn *Connection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	m.connections[name] = conn

	// Set as default if it's the first connection
	if m.defaultConn == "" {
		m.defaultConn = name
	}

	log.Printf("Added database connection '%s' for Ent usage", name)
	return nil
}

// SetClient sets an Ent client for a named connection
func (m *EntManager) SetClient(name string, client EntClient) error {
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}

	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}

	m.clients[name] = client
	log.Printf("Set Ent client for connection '%s'", name)
	return nil
}

// GetClient returns an Ent client for the named connection
func (m *EntManager) GetClient(name string) (EntClient, error) {
	client, exists := m.clients[name]
	if !exists {
		return nil, fmt.Errorf("Ent client '%s' not found", name)
	}
	return client, nil
}

// Default returns the default Ent client
func (m *EntManager) Default() (EntClient, error) {
	if m.defaultConn == "" {
		return nil, fmt.Errorf("no default connection set")
	}
	return m.GetClient(m.defaultConn)
}

// GetConnection returns the underlying database connection
func (m *EntManager) GetConnection(name string) (*Connection, error) {
	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}
	return conn, nil
}

// CreateDriver creates an Ent SQL driver from a database connection
func (m *EntManager) CreateDriver(name string) (*entsql.Driver, error) {
	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	var dialectName string
	switch conn.Driver() {
	case DriverPostgres:
		dialectName = dialect.Postgres
	case DriverSQLite:
		dialectName = dialect.SQLite
	case DriverMySQL:
		dialectName = dialect.MySQL
	default:
		return nil, fmt.Errorf("unsupported driver for Ent: %s", conn.Driver())
	}

	return entsql.OpenDB(dialectName, conn.DB()), nil
}

// SetDefault sets the default connection
func (m *EntManager) SetDefault(name string) error {
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}
	m.defaultConn = name
	return nil
}

// Migrate runs schema migrations for a specific client
func (m *EntManager) Migrate(ctx context.Context, name string, opts ...MigrateOption) error {
	client, err := m.GetClient(name)
	if err != nil {
		return fmt.Errorf("failed to get client '%s': %w", name, err)
	}

	config := &MigrateConfig{}
	for _, opt := range opts {
		opt(config)
	}

	log.Printf("Running migrations for connection '%s'", name)
	
	if err := client.Schema().Create(ctx, opts...); err != nil {
		return fmt.Errorf("failed to create schema for '%s': %w", name, err)
	}

	log.Printf("Successfully ran migrations for connection '%s'", name)
	return nil
}

// MigrateAll runs schema migrations for all registered clients
func (m *EntManager) MigrateAll(ctx context.Context, opts ...MigrateOption) error {
	if len(m.clients) == 0 {
		log.Println("No Ent clients registered, skipping migrations")
		return nil
	}

	var firstError error
	for name := range m.clients {
		if err := m.Migrate(ctx, name, opts...); err != nil && firstError == nil {
			firstError = err
		}
	}

	return firstError
}

// CloseAll closes all Ent clients and database connections
func (m *EntManager) CloseAll() error {
	var firstError error

	// Close all Ent clients
	for name, client := range m.clients {
		if err := client.Close(); err != nil && firstError == nil {
			firstError = fmt.Errorf("failed to close Ent client '%s': %w", name, err)
		}
	}

	// Close all database connections
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil && firstError == nil {
			firstError = fmt.Errorf("failed to close connection '%s': %w", name, err)
		}
	}

	// Clear all references
	m.clients = make(map[string]EntClient)
	m.connections = make(map[string]*Connection)
	m.defaultConn = ""

	return firstError
}

// WithTransaction executes a function within a database transaction
func (m *EntManager) WithTransaction(ctx context.Context, name string, fn func(ctx context.Context, tx Transaction) error) error {
	client, err := m.GetClient(name)
	if err != nil {
		return fmt.Errorf("failed to get client '%s': %w", name, err)
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Execute function within transaction
	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Stats returns database statistics for a connection
func (m *EntManager) Stats(name string) (sql.DBStats, error) {
	conn, exists := m.connections[name]
	if !exists {
		return sql.DBStats{}, fmt.Errorf("connection '%s' not found", name)
	}
	return conn.Stats(), nil
}

// ListConnections returns a list of all registered connection names
func (m *EntManager) ListConnections() []string {
	names := make([]string, 0, len(m.connections))
	for name := range m.connections {
		names = append(names, name)
	}
	return names
}

// ListClients returns a list of all registered client names
func (m *EntManager) ListClients() []string {
	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}