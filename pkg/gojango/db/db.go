// Package db provides database connection management and Ent integration for Gojango applications.
//
// This package handles:
//   - Database connection lifecycle
//   - Multi-driver support (PostgreSQL, SQLite, MySQL)
//   - Connection pooling and configuration
//   - Integration with Ent ORM
//   - Migration management
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Driver represents supported database drivers
type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite3"
	DriverMySQL    Driver = "mysql"
)

// Config holds database configuration
type Config struct {
	Driver   Driver `yaml:"driver" json:"driver"`
	DSN      string `yaml:"dsn" json:"dsn"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`

	// Connection pool settings
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		Driver:          DriverSQLite,
		Database:        "gojango.db",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// PostgresConfig returns a PostgreSQL configuration with common defaults
func PostgresConfig(host, database, username, password string) *Config {
	return &Config{
		Driver:          DriverPostgres,
		Host:            host,
		Port:            5432,
		Database:        database,
		Username:        username,
		Password:        password,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// SQLiteConfig returns a SQLite configuration
func SQLiteConfig(path string) *Config {
	return &Config{
		Driver:          DriverSQLite,
		Database:        path,
		MaxOpenConns:    1, // SQLite works better with single connection
		MaxIdleConns:    1,
		ConnMaxLifetime: 0, // No limit for SQLite
		ConnMaxIdleTime: 0, // No limit for SQLite
	}
}

// BuildDSN builds a Data Source Name from the configuration
func (c *Config) BuildDSN() (string, error) {
	if c.DSN != "" {
		return c.DSN, nil
	}

	switch c.Driver {
	case DriverPostgres:
		return c.buildPostgresDSN(), nil
	case DriverSQLite:
		return c.buildSQLiteDSN(), nil
	case DriverMySQL:
		return "", fmt.Errorf("MySQL driver not yet implemented")
	default:
		return "", fmt.Errorf("unsupported database driver: %s", c.Driver)
	}
}

func (c *Config) buildPostgresDSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Database, c.SSLMode)
	
	if c.Password != "" {
		dsn += fmt.Sprintf(" password=%s", c.Password)
	}
	
	return dsn
}

func (c *Config) buildSQLiteDSN() string {
	if c.Database == "" {
		return ":memory:"
	}
	return c.Database
}

// Connection wraps a database connection with additional functionality
type Connection struct {
	db     *sql.DB
	config *Config
}

// Open creates a new database connection
func Open(config *Config) (*Connection, error) {
	dsn, err := config.BuildDSN()
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	db, err := sql.Open(string(config.Driver), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connection established: driver=%s, database=%s", 
		config.Driver, config.Database)

	return &Connection{
		db:     db,
		config: config,
	}, nil
}

// DB returns the underlying sql.DB instance
func (c *Connection) DB() *sql.DB {
	return c.db
}

// Config returns the database configuration
func (c *Connection) Config() *Config {
	return c.config
}

// Driver returns the database driver name
func (c *Connection) Driver() Driver {
	return c.config.Driver
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.db != nil {
		log.Printf("Closing database connection: driver=%s", c.config.Driver)
		return c.db.Close()
	}
	return nil
}

// Ping tests the database connection
func (c *Connection) Ping() error {
	return c.db.Ping()
}

// Stats returns database connection statistics
func (c *Connection) Stats() sql.DBStats {
	return c.db.Stats()
}

// Manager handles multiple database connections
type Manager struct {
	connections map[string]*Connection
	defaultConn string
}

// NewManager creates a new database connection manager
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
	}
}

// AddConnection adds a named connection to the manager
func (m *Manager) AddConnection(name string, config *Config) error {
	conn, err := Open(config)
	if err != nil {
		return fmt.Errorf("failed to add connection '%s': %w", name, err)
	}

	m.connections[name] = conn

	// Set as default if it's the first connection
	if m.defaultConn == "" {
		m.defaultConn = name
	}

	return nil
}

// GetConnection returns a named connection
func (m *Manager) GetConnection(name string) (*Connection, error) {
	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}
	return conn, nil
}

// Default returns the default connection
func (m *Manager) Default() (*Connection, error) {
	if m.defaultConn == "" {
		return nil, fmt.Errorf("no default connection set")
	}
	return m.GetConnection(m.defaultConn)
}

// SetDefault sets the default connection
func (m *Manager) SetDefault(name string) error {
	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}
	m.defaultConn = name
	return nil
}

// CloseAll closes all connections
func (m *Manager) CloseAll() error {
	var firstError error
	
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil && firstError == nil {
			firstError = fmt.Errorf("failed to close connection '%s': %w", name, err)
		}
	}

	// Clear connections
	m.connections = make(map[string]*Connection)
	m.defaultConn = ""

	return firstError
}