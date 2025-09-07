package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/epuerta9/gojango/internal/cli/ui"
	"github.com/epuerta9/gojango/internal/codegen"
)

// ProjectConfig holds configuration for new project generation
type ProjectConfig struct {
	Name       string
	Module     string
	Directory  string
	GoVersion  string
	Frontend   string
	Database   string
	Features   []string
	Interactive bool
}

// NewProjectCmd creates the 'new' command for generating new projects
func NewProjectCmd() *cobra.Command {
	var config ProjectConfig

	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Gojango project",
		Long: `Create a new Gojango project with the specified name.

This command creates a new directory with the project name and generates
a complete Gojango project structure including:

  • Go module setup
  • Basic application structure
  • Configuration files
  • Development tools (Makefile, docker-compose)
  • Example code and documentation

The project will be ready to run immediately after creation.`,
		Args: cobra.ExactArgs(1),
		Example: `  # Create a new project with default settings
  gojango new myblog

  # Create with custom module path
  gojango new myblog --module github.com/myuser/myblog

  # Interactive setup with prompts
  gojango new myblog --interactive

  # Create with specific features
  gojango new myblog --features admin,api`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Name = args[0]

			// Validate project name
			if !isValidProjectName(config.Name) {
				return fmt.Errorf("invalid project name: %s\nProject name must be a valid directory name", config.Name)
			}

			// Check if directory already exists
			if _, err := os.Stat(config.Name); err == nil {
				return fmt.Errorf("directory '%s' already exists", config.Name)
			}

			// Set default module path if not provided
			if config.Module == "" {
				config.Module = fmt.Sprintf("github.com/yourusername/%s", config.Name)
			}

			// Interactive mode
			if config.Interactive {
				if err := interactiveSetup(&config); err != nil {
					return fmt.Errorf("interactive setup failed: %w", err)
				}
			}

			// Create the project
			generator := codegen.NewProjectGenerator()
			if err := generator.Generate(config.toCodegenConfig()); err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			// Success message
			ui.Success(fmt.Sprintf("Successfully created project: %s", config.Name))
			ui.Info("")
			ui.Info("Next steps:")
			ui.Info("  cd " + config.Name)
			ui.Info("  make setup     # Setup development environment")
			ui.Info("  make run       # Start development server")
			ui.Info("")
			ui.Info("Documentation: https://gojango.dev/docs")

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&config.Module, "module", "m", "", "Go module path (default: github.com/yourusername/PROJECT-NAME)")
	cmd.Flags().StringVar(&config.Frontend, "frontend", "htmx", "Frontend framework (htmx, react, vue, none)")
	cmd.Flags().StringVar(&config.Database, "database", "postgres", "Database type (postgres, mysql, sqlite)")
	cmd.Flags().StringSliceVar(&config.Features, "features", []string{}, "Additional features (admin,auth,api,docker)")
	cmd.Flags().BoolVarP(&config.Interactive, "interactive", "i", false, "Interactive setup with prompts")

	return cmd
}

// interactiveSetup guides the user through project configuration
func interactiveSetup(config *ProjectConfig) error {
	ui.Header("🚀 Gojango Project Setup")
	ui.Info("")

	// Confirm project name
	if !ui.Confirm(fmt.Sprintf("Create project '%s'?", config.Name), true) {
		return fmt.Errorf("cancelled by user")
	}

	// Module path
	config.Module = ui.Input("Go module path", config.Module)

	// Frontend framework
	frontend := ui.Select("Choose frontend framework", []string{
		"htmx - Server-side rendering with HTMX (recommended)",
		"react - React with TypeScript",
		"vue - Vue 3 with TypeScript", 
		"none - API only, no frontend",
	})
	config.Frontend = extractChoice(frontend)

	// Database
	database := ui.Select("Choose database", []string{
		"postgres - PostgreSQL (recommended)",
		"mysql - MySQL/MariaDB",
		"sqlite - SQLite (development only)",
	})
	config.Database = extractChoice(database)

	// Features
	features := ui.MultiSelect("Select additional features", []string{
		"admin - Admin interface",
		"auth - User authentication",
		"api - REST/gRPC API endpoints",
		"docker - Docker development setup",
	})
	config.Features = features

	return nil
}

// isValidProjectName validates the project name
func isValidProjectName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Must not start with a dot or hyphen
	if name[0] == '.' || name[0] == '-' {
		return false
	}

	// Only allow alphanumeric, hyphens, and underscores
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	// Reserved names
	reserved := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", 
					   "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3",
					   "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
	
	lowerName := strings.ToLower(name)
	for _, res := range reserved {
		if lowerName == res {
			return false
		}
	}

	return true
}

// extractChoice extracts the key from a "key - description" string
func extractChoice(choice string) string {
	parts := strings.SplitN(choice, " - ", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return choice
}

// toCodegenConfig converts ProjectConfig to codegen.ProjectConfig
func (c *ProjectConfig) toCodegenConfig() codegen.ProjectConfig {
	return codegen.ProjectConfig{
		Name:      c.Name,
		Module:    c.Module,
		Directory: filepath.Join(".", c.Name),
		Frontend:  c.Frontend,
		Database:  c.Database,
		Features:  c.Features,
	}
}