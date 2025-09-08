package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MigrationManager handles Django-style database migrations
type MigrationManager struct {
	db            *sql.DB
	migrationsDir string
	tableName     string
}

// Migration represents a single database migration
type Migration struct {
	ID          string
	Name        string
	App         string
	Operations  []Operation
	Dependencies []string
	Applied     bool
	AppliedAt   *time.Time
}

// Operation represents a migration operation
type Operation interface {
	Forward(tx *sql.Tx) error
	Reverse(tx *sql.Tx) error
	Description() string
}

// CreateTable operation
type CreateTable struct {
	Name    string
	Columns []Column
	Indexes []Index
}

// Column definition
type Column struct {
	Name     string
	Type     string
	Null     bool
	Default  interface{}
	Primary  bool
	Unique   bool
	ForeignKey *ForeignKey
}

// Index definition
type Index struct {
	Name    string
	Columns []string
	Unique  bool
}

// ForeignKey definition
type ForeignKey struct {
	Table  string
	Column string
	OnDelete string
	OnUpdate string
}

// DropTable operation
type DropTable struct {
	Name string
}

// AddColumn operation
type AddColumn struct {
	Table  string
	Column Column
}

// DropColumn operation
type DropColumn struct {
	Table  string
	Column string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB, migrationsDir string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationsDir: migrationsDir,
		tableName:     "django_migrations",
	}
}

// Initialize creates the migration tracking table
func (m *MigrationManager) Initialize() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			applied DATETIME NOT NULL,
			UNIQUE(app, name)
		)
	`, m.tableName)

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// LoadMigrations loads all available migrations from the migrations directory
func (m *MigrationManager) LoadMigrations() ([]*Migration, error) {
	if err := os.MkdirAll(m.migrationsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create migrations directory: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(m.migrationsDir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to find migration files: %w", err)
	}

	var migrations []*Migration

	for _, file := range files {
		migration, err := m.parseMigrationFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration %s: %w", file, err)
		}

		// Check if migration is applied
		applied, appliedAt, err := m.isMigrationApplied(migration.App, migration.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check migration status: %w", err)
		}

		migration.Applied = applied
		migration.AppliedAt = appliedAt

		migrations = append(migrations, migration)
	}

	// Sort by migration ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// parseMigrationFile parses a migration SQL file
func (m *MigrationManager) parseMigrationFile(filename string) (*Migration, error) {
	basename := filepath.Base(filename)
	name := strings.TrimSuffix(basename, ".sql")
	
	// Parse migration name format: 0001_initial.sql
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid migration filename format: %s", basename)
	}

	migration := &Migration{
		ID:   parts[0],
		Name: name,
		App:  "default", // For now, assume default app
	}

	return migration, nil
}

// isMigrationApplied checks if a migration has been applied
func (m *MigrationManager) isMigrationApplied(app, name string) (bool, *time.Time, error) {
	query := fmt.Sprintf("SELECT applied FROM %s WHERE app = ? AND name = ?", m.tableName)
	
	var appliedStr string
	err := m.db.QueryRow(query, app, name).Scan(&appliedStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, err
	}

	applied, err := time.Parse(time.RFC3339, appliedStr)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse applied time: %w", err)
	}

	return true, &applied, nil
}

// ApplyMigrations applies all pending migrations
func (m *MigrationManager) ApplyMigrations() error {
	migrations, err := m.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	var applied int
	for _, migration := range migrations {
		if !migration.Applied {
			if err := m.ApplyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
			}
			applied++
			fmt.Printf("✅ Applied migration: %s\n", migration.Name)
		}
	}

	if applied == 0 {
		fmt.Println("✅ No migrations to apply")
	} else {
		fmt.Printf("✅ Applied %d migrations\n", applied)
	}

	return nil
}

// ApplyMigration applies a single migration
func (m *MigrationManager) ApplyMigration(migration *Migration) error {
	// Read migration file
	filename := filepath.Join(m.migrationsDir, migration.Name+".sql")
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	_, err = tx.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration as applied
	now := time.Now().Format(time.RFC3339)
	recordQuery := fmt.Sprintf("INSERT INTO %s (app, name, applied) VALUES (?, ?, ?)", m.tableName)
	_, err = tx.Exec(recordQuery, migration.App, migration.Name, now)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// RollbackMigration rolls back a single migration
func (m *MigrationManager) RollbackMigration(migrationName string) error {
	// Check if migration is applied
	migration := &Migration{App: "default", Name: migrationName}
	applied, _, err := m.isMigrationApplied(migration.App, migration.Name)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !applied {
		return fmt.Errorf("migration %s is not applied", migrationName)
	}

	// Read rollback file
	rollbackFilename := filepath.Join(m.migrationsDir, "rollback_"+migrationName+".sql")
	if _, err := os.Stat(rollbackFilename); os.IsNotExist(err) {
		return fmt.Errorf("rollback file not found for migration %s", migrationName)
	}

	content, err := os.ReadFile(rollbackFilename)
	if err != nil {
		return fmt.Errorf("failed to read rollback file: %w", err)
	}

	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	_, err = tx.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	// Remove migration record
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE app = ? AND name = ?", m.tableName)
	_, err = tx.Exec(deleteQuery, migration.App, migration.Name)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	fmt.Printf("✅ Rolled back migration: %s\n", migrationName)
	return nil
}

// ShowMigrations displays the status of all migrations
func (m *MigrationManager) ShowMigrations() error {
	migrations, err := m.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("=================")
	
	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	for _, migration := range migrations {
		status := "❌ Not Applied"
		appliedInfo := ""
		
		if migration.Applied {
			status = "✅ Applied"
			if migration.AppliedAt != nil {
				appliedInfo = fmt.Sprintf(" (%s)", migration.AppliedAt.Format("2006-01-02 15:04:05"))
			}
		}

		fmt.Printf("%s %s%s\n", status, migration.Name, appliedInfo)
	}

	return nil
}

// GenerateMigration creates a new migration file
func (m *MigrationManager) GenerateMigration(name, operation string) error {
	if err := os.MkdirAll(m.migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Get next migration number
	nextID, err := m.getNextMigrationID()
	if err != nil {
		return fmt.Errorf("failed to get next migration ID: %w", err)
	}

	migrationName := fmt.Sprintf("%04d_%s", nextID, name)
	filename := filepath.Join(m.migrationsDir, migrationName+".sql")

	// Generate migration template
	template := m.generateMigrationTemplate(operation)

	if err := os.WriteFile(filename, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	// Generate rollback file
	rollbackFilename := filepath.Join(m.migrationsDir, "rollback_"+migrationName+".sql")
	rollbackTemplate := m.generateRollbackTemplate(operation)

	if err := os.WriteFile(rollbackFilename, []byte(rollbackTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write rollback file: %w", err)
	}

	fmt.Printf("✅ Generated migration: %s\n", migrationName)
	fmt.Printf("📝 Edit the migration file: %s\n", filename)
	fmt.Printf("📝 Edit the rollback file: %s\n", rollbackFilename)

	return nil
}

// getNextMigrationID gets the next available migration ID
func (m *MigrationManager) getNextMigrationID() (int, error) {
	files, err := filepath.Glob(filepath.Join(m.migrationsDir, "*.sql"))
	if err != nil {
		return 1, nil // Start with 1 if no files found
	}

	maxID := 0
	for _, file := range files {
		basename := filepath.Base(file)
		if strings.HasPrefix(basename, "rollback_") {
			continue // Skip rollback files
		}
		
		name := strings.TrimSuffix(basename, ".sql")
		parts := strings.SplitN(name, "_", 2)
		if len(parts) > 0 {
			var id int
			if _, err := fmt.Sscanf(parts[0], "%04d", &id); err == nil {
				if id > maxID {
					maxID = id
				}
			}
		}
	}

	return maxID + 1, nil
}

// generateMigrationTemplate generates a template for a new migration
func (m *MigrationManager) generateMigrationTemplate(operation string) string {
	switch operation {
	case "create_table":
		return `-- Create table migration
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

CREATE TABLE example_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

	case "add_column":
		return `-- Add column migration
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

ALTER TABLE example_table 
ADD COLUMN new_column VARCHAR(255);`

	case "drop_column":
		return `-- Drop column migration
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

-- SQLite doesn't support DROP COLUMN directly
-- You may need to create a new table and migrate data`

	default:
		return `-- Custom migration
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

-- Add your migration SQL here`
	}
}

// generateRollbackTemplate generates a rollback template
func (m *MigrationManager) generateRollbackTemplate(operation string) string {
	switch operation {
	case "create_table":
		return `-- Rollback: Drop table
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

DROP TABLE IF EXISTS example_table;`

	case "add_column":
		return `-- Rollback: Remove column
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

-- SQLite doesn't support DROP COLUMN directly
-- You may need to create a new table without the column and migrate data`

	case "drop_column":
		return `-- Rollback: Add column back
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

ALTER TABLE example_table 
ADD COLUMN restored_column VARCHAR(255);`

	default:
		return `-- Custom rollback migration
-- Generated at: ` + time.Now().Format(time.RFC3339) + `

-- Add your rollback SQL here`
	}
}

// Implementation of Operation interface methods

func (op *CreateTable) Forward(tx *sql.Tx) error {
	// Implementation would generate CREATE TABLE SQL
	return nil
}

func (op *CreateTable) Reverse(tx *sql.Tx) error {
	// Implementation would generate DROP TABLE SQL
	return nil
}

func (op *CreateTable) Description() string {
	return fmt.Sprintf("Create table %s", op.Name)
}

func (op *DropTable) Forward(tx *sql.Tx) error {
	_, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", op.Name))
	return err
}

func (op *DropTable) Reverse(tx *sql.Tx) error {
	// Would need to recreate the table - this is complex
	return fmt.Errorf("cannot reverse DROP TABLE operation")
}

func (op *DropTable) Description() string {
	return fmt.Sprintf("Drop table %s", op.Name)
}

func (op *AddColumn) Forward(tx *sql.Tx) error {
	// Implementation would generate ALTER TABLE ADD COLUMN SQL
	return nil
}

func (op *AddColumn) Reverse(tx *sql.Tx) error {
	// Implementation would generate ALTER TABLE DROP COLUMN SQL
	return nil
}

func (op *AddColumn) Description() string {
	return fmt.Sprintf("Add column %s to table %s", op.Column.Name, op.Table)
}

func (op *DropColumn) Forward(tx *sql.Tx) error {
	// Implementation would generate ALTER TABLE DROP COLUMN SQL
	return nil
}

func (op *DropColumn) Reverse(tx *sql.Tx) error {
	// Implementation would generate ALTER TABLE ADD COLUMN SQL
	return nil
}

func (op *DropColumn) Description() string {
	return fmt.Sprintf("Drop column %s from table %s", op.Column, op.Table)
}