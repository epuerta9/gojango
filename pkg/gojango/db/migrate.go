// Package db provides database migration management for Gojango applications.
//
// This module handles:
//   - Migration file discovery and execution
//   - Migration state tracking
//   - Rollback capabilities
//   - Integration with Ent schema management
package db

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration represents a single database migration
type Migration struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Filename    string    `json:"filename"`
	AppliedAt   time.Time `json:"applied_at,omitempty"`
	SQL         string    `json:"-"`
	RollbackSQL string    `json:"-"`
}

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
	Applied   []Migration `json:"applied"`
	Pending   []Migration `json:"pending"`
	LastApplied *Migration `json:"last_applied,omitempty"`
}

// Migrator handles database migrations
type Migrator struct {
	conn           *Connection
	migrationsPath string
	tableName      string
}

// NewMigrator creates a new migration manager
func NewMigrator(conn *Connection, migrationsPath string) *Migrator {
	return &Migrator{
		conn:           conn,
		migrationsPath: migrationsPath,
		tableName:      "gojango_migrations",
	}
}

// SetMigrationsTable sets a custom migrations table name
func (m *Migrator) SetMigrationsTable(tableName string) {
	m.tableName = tableName
}

// Initialize creates the migrations table if it doesn't exist
func (m *Migrator) Initialize(ctx context.Context) error {
	var createTableSQL string
	
	switch m.conn.Driver() {
	case DriverPostgres:
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL UNIQUE,
				filename VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_%s_applied_at ON %s (applied_at);
		`, m.tableName, m.tableName, m.tableName)
	case DriverSQLite:
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				filename TEXT NOT NULL,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_%s_applied_at ON %s (applied_at);
		`, m.tableName, m.tableName, m.tableName)
	case DriverMySQL:
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) NOT NULL UNIQUE,
				filename VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_%s_applied_at (applied_at)
			);
		`, m.tableName, m.tableName)
	default:
		return fmt.Errorf("unsupported database driver: %s", m.conn.Driver())
	}

	_, err := m.conn.DB().ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	log.Printf("Initialized migrations table: %s", m.tableName)
	return nil
}

// DiscoverMigrations finds all migration files in the migrations directory
func (m *Migrator) DiscoverMigrations() ([]Migration, error) {
	var migrations []Migration

	if _, err := os.Stat(m.migrationsPath); os.IsNotExist(err) {
		log.Printf("Migrations directory does not exist: %s", m.migrationsPath)
		return migrations, nil
	}

	err := filepath.WalkDir(m.migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		filename := filepath.Base(path)
		// Only process up migrations or regular migrations (not down files)
		if strings.HasSuffix(filename, "_down.sql") {
			return nil // Skip down files - they'll be loaded when needed
		}

		migration, err := m.parseMigrationFile(filename, path)
		if err != nil {
			log.Printf("Warning: failed to parse migration file %s: %v", filename, err)
			return nil // Continue with other files
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover migrations: %w", err)
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file and extracts metadata
func (m *Migrator) parseMigrationFile(filename, path string) (Migration, error) {
	// Expected format: 0001_initial_migration.sql or 0001_initial_migration_up.sql
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	// Parse migration ID
	idStr := parts[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid migration ID in filename %s: %w", filename, err)
	}

	// Extract migration name (everything except ID and .sql extension)
	nameWithExt := strings.Join(parts[1:], "_")
	name := strings.TrimSuffix(nameWithExt, ".sql")
	
	// Handle _up and _down suffixes
	if strings.HasSuffix(name, "_up") {
		name = strings.TrimSuffix(name, "_up")
	}

	// Read migration content
	content, err := os.ReadFile(path)
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read migration file %s: %w", path, err)
	}

	migration := Migration{
		ID:       id,
		Name:     name,
		Filename: filename,
		SQL:      string(content),
	}

	// Look for corresponding rollback file
	rollbackPath := strings.Replace(path, "_up.sql", "_down.sql", 1)
	if rollbackPath == path {
		rollbackPath = strings.Replace(path, ".sql", "_down.sql", 1)
	}

	if rollbackContent, err := os.ReadFile(rollbackPath); err == nil {
		migration.RollbackSQL = string(rollbackContent)
	}

	return migration, nil
}

// GetAppliedMigrations returns all migrations that have been applied
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	query := fmt.Sprintf(`
		SELECT id, name, filename, applied_at 
		FROM %s 
		ORDER BY id ASC
	`, m.tableName)

	rows, err := m.conn.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.ID, &migration.Name, &migration.Filename, &migration.AppliedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		migrations = append(migrations, migration)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
	}

	return migrations, nil
}

// GetStatus returns the current migration status
func (m *Migrator) GetStatus(ctx context.Context) (*MigrationStatus, error) {
	allMigrations, err := m.DiscoverMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to discover migrations: %w", err)
	}

	appliedMigrations, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create a map of applied migrations for quick lookup
	appliedMap := make(map[int]Migration)
	for _, applied := range appliedMigrations {
		appliedMap[applied.ID] = applied
	}

	// Separate applied and pending migrations
	var applied, pending []Migration
	for _, migration := range allMigrations {
		if appliedMigration, exists := appliedMap[migration.ID]; exists {
			applied = append(applied, appliedMigration)
		} else {
			pending = append(pending, migration)
		}
	}

	status := &MigrationStatus{
		Applied: applied,
		Pending: pending,
	}

	if len(applied) > 0 {
		status.LastApplied = &applied[len(applied)-1]
	}

	return status, nil
}

// Apply runs all pending migrations
func (m *Migrator) Apply(ctx context.Context) error {
	status, err := m.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if len(status.Pending) == 0 {
		log.Println("No pending migrations to apply")
		return nil
	}

	log.Printf("Applying %d pending migrations", len(status.Pending))

	for _, migration := range status.Pending {
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d_%s: %w", migration.ID, migration.Name, err)
		}
		log.Printf("Applied migration: %d_%s", migration.ID, migration.Name)
	}

	log.Printf("Successfully applied %d migrations", len(status.Pending))
	return nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.conn.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if migration.SQL != "" {
		_, err = tx.ExecContext(ctx, migration.SQL)
		if err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}
	}

	// Record migration as applied
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (name, filename, applied_at) 
		VALUES ($1, $2, $3)
	`, m.tableName)

	// Adjust placeholder for different databases
	switch m.conn.Driver() {
	case DriverMySQL:
		insertQuery = strings.Replace(insertQuery, "$1", "?", -1)
		insertQuery = strings.Replace(insertQuery, "$2", "?", -1)
		insertQuery = strings.Replace(insertQuery, "$3", "?", -1)
	case DriverSQLite:
		// SQLite uses ? placeholders
		insertQuery = strings.Replace(insertQuery, "$1", "?", -1)
		insertQuery = strings.Replace(insertQuery, "$2", "?", -1)
		insertQuery = strings.Replace(insertQuery, "$3", "?", -1)
	}

	_, err = tx.ExecContext(ctx, insertQuery, migration.Name, migration.Filename, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// Rollback rolls back the last applied migration
func (m *Migrator) Rollback(ctx context.Context) error {
	status, err := m.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if status.LastApplied == nil {
		log.Println("No migrations to rollback")
		return nil
	}

	migration := *status.LastApplied

	// Find the migration file to get rollback SQL
	allMigrations, err := m.DiscoverMigrations()
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	var rollbackMigration *Migration
	for _, m := range allMigrations {
		if m.ID == migration.ID {
			rollbackMigration = &m
			break
		}
	}

	if rollbackMigration == nil {
		return fmt.Errorf("migration file not found for rollback: %d_%s", migration.ID, migration.Name)
	}

	if rollbackMigration.RollbackSQL == "" {
		return fmt.Errorf("no rollback SQL found for migration: %d_%s", migration.ID, migration.Name)
	}

	log.Printf("Rolling back migration: %d_%s", migration.ID, migration.Name)

	return m.rollbackMigration(ctx, migration, rollbackMigration.RollbackSQL)
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration Migration, rollbackSQL string) error {
	tx, err := m.conn.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	_, err = tx.ExecContext(ctx, rollbackSQL)
	if err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	deleteQuery := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, m.tableName)
	
	// Adjust placeholder for different databases
	switch m.conn.Driver() {
	case DriverMySQL, DriverSQLite:
		deleteQuery = strings.Replace(deleteQuery, "$1", "?", -1)
	}

	_, err = tx.ExecContext(ctx, deleteQuery, migration.ID)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	log.Printf("Rolled back migration: %d_%s", migration.ID, migration.Name)
	return tx.Commit()
}

// Reset rolls back all applied migrations
func (m *Migrator) Reset(ctx context.Context) error {
	status, err := m.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if len(status.Applied) == 0 {
		log.Println("No migrations to reset")
		return nil
	}

	log.Printf("Resetting %d migrations", len(status.Applied))

	// Rollback migrations in reverse order
	for i := len(status.Applied) - 1; i >= 0; i-- {
		if err := m.Rollback(ctx); err != nil {
			return fmt.Errorf("failed to rollback during reset: %w", err)
		}
	}

	log.Printf("Successfully reset %d migrations", len(status.Applied))
	return nil
}