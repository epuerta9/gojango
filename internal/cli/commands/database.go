package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/epuerta9/gojango/pkg/gojango/db"
	"github.com/spf13/cobra"
)

// NewDatabaseCmd creates the database management command
func NewDatabaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "db",
		Aliases: []string{"database"},
		Short:   "Database management commands",
		Long: `Database management commands for Gojango applications.
		
This command provides utilities for managing database connections,
running migrations, and interacting with your database.`,
		Example: `  # Run all pending migrations
  gojango db migrate

  # Create a new migration
  gojango db makemigration create_users

  # Check migration status
  gojango db showmigrations

  # Rollback last migration
  gojango db rollback

  # Open database shell
  gojango db dbshell`,
	}

	// Add subcommands
	cmd.AddCommand(newMigrateCmd())
	cmd.AddCommand(newMakeMigrationCmd())
	cmd.AddCommand(newShowMigrationsCmd())
	cmd.AddCommand(newRollbackCmd())
	cmd.AddCommand(newResetCmd())
	cmd.AddCommand(newDBShellCmd())

	return cmd
}

// newMigrateCmd creates the migrate command
func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long: `Run all pending database migrations.
		
This command will apply all migrations that haven't been run yet,
in order from lowest to highest migration number.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrations(cmd.Context())
		},
	}
}

// newMakeMigrationCmd creates the makemigration command
func newMakeMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "makemigration [name]",
		Short: "Create a new database migration",
		Long: `Create a new database migration file.
		
The migration file will be created in the migrations/ directory
with a sequential ID and the provided name.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createMigration(args[0])
		},
	}

	cmd.Flags().Bool("empty", false, "Create an empty migration file")
	cmd.Flags().String("sql", "", "SQL content for the migration")

	return cmd
}

// newShowMigrationsCmd creates the showmigrations command
func newShowMigrationsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "showmigrations",
		Short: "Show migration status",
		Long: `Show the status of all migrations.
		
This command displays which migrations have been applied
and which are still pending.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showMigrations(cmd.Context())
		},
	}
}

// newRollbackCmd creates the rollback command
func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last migration",
		Long: `Rollback the most recently applied migration.
		
This will execute the rollback SQL (if available) and mark
the migration as not applied.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rollbackMigration(cmd.Context())
		},
	}
}

// newResetCmd creates the reset command
func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset all migrations",
		Long: `Reset all applied migrations.
		
WARNING: This will rollback ALL migrations in reverse order.
Use with caution in production environments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return resetMigrations(cmd.Context())
		},
	}
}

// newDBShellCmd creates the dbshell command
func newDBShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dbshell",
		Short: "Open database shell",
		Long: `Open an interactive database shell.
		
This will launch the appropriate database client (psql, sqlite3, mysql)
with the connection parameters from your configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return openDatabaseShell()
		},
	}
}

// runMigrations executes all pending migrations
func runMigrations(ctx context.Context) error {
	config, err := loadDatabaseConfig()
	if err != nil {
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	conn, err := db.Open(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	migrator := db.NewMigrator(conn, "migrations")
	
	if err := migrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return migrator.Apply(ctx)
}

// createMigration creates a new migration file
func createMigration(name string) error {
	// Ensure migrations directory exists
	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Find next migration ID
	nextID, err := getNextMigrationID(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to determine next migration ID: %w", err)
	}

	// Clean name for filename
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "-", "_")

	// Create migration files
	timestamp := time.Now().Format("20060102_150405")
	baseFilename := fmt.Sprintf("%04d_%s_%s", nextID, cleanName, timestamp)
	
	upFile := filepath.Join(migrationsDir, baseFilename+"_up.sql")
	downFile := filepath.Join(migrationsDir, baseFilename+"_down.sql")

	// Create up migration file
	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: %s

-- Add your SQL statements here
-- Example: CREATE TABLE users (id SERIAL PRIMARY KEY, email VARCHAR(255));
`, name, time.Now().Format("2006-01-02 15:04:05"), name)

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downContent := fmt.Sprintf(`-- Rollback for: %s
-- Created: %s
-- Description: Rollback %s

-- Add your rollback SQL statements here
-- Example: DROP TABLE IF EXISTS users;
`, name, time.Now().Format("2006-01-02 15:04:05"), name)

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)

	return nil
}

// getNextMigrationID finds the next available migration ID
func getNextMigrationID(migrationsDir string) (int, error) {
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return 1, nil
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	maxID := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		if id, err := strconv.Atoi(parts[0]); err == nil && id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}

// showMigrations displays the migration status
func showMigrations(ctx context.Context) error {
	config, err := loadDatabaseConfig()
	if err != nil {
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	conn, err := db.Open(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	migrator := db.NewMigrator(conn, "migrations")
	
	if err := migrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	status, err := migrator.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("Migration Status\n")
	fmt.Printf("================\n\n")

	if len(status.Applied) > 0 {
		fmt.Printf("Applied Migrations (%d):\n", len(status.Applied))
		for _, migration := range status.Applied {
			fmt.Printf("  ✓ %04d_%s (applied: %s)\n", 
				migration.ID, 
				migration.Name, 
				migration.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	if len(status.Pending) > 0 {
		fmt.Printf("Pending Migrations (%d):\n", len(status.Pending))
		for _, migration := range status.Pending {
			fmt.Printf("  ○ %04d_%s\n", migration.ID, migration.Name)
		}
		fmt.Println()
	}

	if len(status.Applied) == 0 && len(status.Pending) == 0 {
		fmt.Println("No migrations found.")
	} else {
		fmt.Printf("Total: %d applied, %d pending\n", len(status.Applied), len(status.Pending))
	}

	return nil
}

// rollbackMigration rolls back the last migration
func rollbackMigration(ctx context.Context) error {
	config, err := loadDatabaseConfig()
	if err != nil {
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	conn, err := db.Open(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	migrator := db.NewMigrator(conn, "migrations")
	
	if err := migrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return migrator.Rollback(ctx)
}

// resetMigrations rolls back all migrations
func resetMigrations(ctx context.Context) error {
	config, err := loadDatabaseConfig()
	if err != nil {
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	conn, err := db.Open(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	migrator := db.NewMigrator(conn, "migrations")
	
	if err := migrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return migrator.Reset(ctx)
}

// openDatabaseShell opens an interactive database shell
func openDatabaseShell() error {
	fmt.Println("Database shell functionality will be implemented based on your database configuration.")
	fmt.Println("This would typically launch psql, sqlite3, or mysql client.")
	return nil
}

// loadDatabaseConfig loads database configuration from the current project
func loadDatabaseConfig() (*db.Config, error) {
	// This is a simplified version - in a real implementation,
	// you would load from a configuration file or environment
	// For now, return a default SQLite configuration
	config := db.DefaultConfig()
	
	// Check if we're in a Gojango project directory
	if _, err := os.Stat("go.mod"); err != nil {
		return nil, fmt.Errorf("not in a Gojango project directory")
	}

	// Try to find database configuration in common locations
	configFiles := []string{
		"database.yml",
		"config/database.yml", 
		"db.yml",
		".env",
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			log.Printf("Found configuration file: %s", configFile)
			// TODO: Parse configuration file
			break
		}
	}

	return config, nil
}