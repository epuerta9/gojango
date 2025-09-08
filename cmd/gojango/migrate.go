package main

import (
	"database/sql"
	"fmt"

	"github.com/epuerta9/gojango/pkg/gojango/migrations"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	var dbPath string
	var showStatus bool
	var rollback string
	var makeMigration string
	var operation string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Django-style database migrations",
		Long: `Django-style database migration system for Gojango.

This command provides Django-like migration functionality:
- Apply pending migrations
- Show migration status  
- Rollback migrations
- Generate new migrations

Examples:
  gojango migrate                           # Apply all pending migrations
  gojango migrate --show                   # Show migration status
  gojango migrate --rollback 0001_initial  # Rollback a migration
  gojango migrate --make initial --op create_table  # Generate new migration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default database path
			if dbPath == "" {
				dbPath = "app.db"
			}

			// Connect to database
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			// Create migration manager
			manager := migrations.NewMigrationManager(db, "migrations")
			
			// Initialize migration system
			if err := manager.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize migrations: %w", err)
			}

			// Handle different commands
			if makeMigration != "" {
				return manager.GenerateMigration(makeMigration, operation)
			}

			if rollback != "" {
				return manager.RollbackMigration(rollback)
			}

			if showStatus {
				return manager.ShowMigrations()
			}

			// Default: apply migrations
			return manager.ApplyMigrations()
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database file path (default: app.db)")
	cmd.Flags().BoolVar(&showStatus, "show", false, "Show migration status")
	cmd.Flags().StringVar(&rollback, "rollback", "", "Rollback migration by name")
	cmd.Flags().StringVar(&makeMigration, "make", "", "Generate new migration with name")
	cmd.Flags().StringVar(&operation, "op", "custom", "Migration operation type (create_table, add_column, drop_column, custom)")

	return cmd
}