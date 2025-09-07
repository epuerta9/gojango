package commands

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/epuerta9/gojango/internal/cli/ui"
	"github.com/epuerta9/gojango/internal/codegen"
)

// NewStartAppCmd creates the 'startapp' command for generating new apps
func NewStartAppCmd() *cobra.Command {
	var features []string

	cmd := &cobra.Command{
		Use:   "startapp [app-name]",
		Short: "Create a new app within the current Gojango project",
		Long: `Create a new application within the current Gojango project.

This command must be run from within a Gojango project directory.
It will create a new app directory structure with:

  • App configuration and registration
  • Schema directory for models
  • Views for HTTP handlers
  • Templates directory
  • Static files directory
  • Basic tests

The app will be automatically registered in your project's main.go file.`,
		Args: cobra.ExactArgs(1),
		Example: `  # Create a basic app
  gojango startapp blog

  # Create an app with specific features
  gojango startapp blog --features admin,api

  # Create an app with all features
  gojango startapp blog --features admin,api,tasks,signals`,
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]

			// Validate app name
			if !isValidAppName(appName) {
				return fmt.Errorf("invalid app name: %s\nApp name must be a valid Go package name", appName)
			}

			// Check if we're in a Gojango project
			if !isGojangoProject() {
				return fmt.Errorf("not in a Gojango project directory\nRun this command from the root of a Gojango project")
			}

			// Check if app already exists
			if appExists(appName) {
				return fmt.Errorf("app '%s' already exists in apps/%s", appName, appName)
			}

			// Get module path from go.mod
			modulePath, err := getModulePath()
			if err != nil {
				return fmt.Errorf("failed to read module path: %w", err)
			}

			// Create app configuration
			config := codegen.AppConfig{
				Name:      appName,
				Module:    modulePath,
				Features:  features,
			}

			// Generate the app
			generator := codegen.NewAppGenerator()
			if err := generator.Generate(config); err != nil {
				return fmt.Errorf("failed to create app: %w", err)
			}

			// Update project files
			if err := updateProjectFiles(appName, modulePath); err != nil {
				ui.Warning(fmt.Sprintf("App created but failed to update project files: %v", err))
				ui.Info("You may need to manually import the app in your main.go:")
				ui.Info(fmt.Sprintf(`  _ "%s/apps/%s"`, modulePath, appName))
			} else {
				ui.Success("App created and registered successfully")
			}

			// Success message
			ui.Info("")
			ui.Info(fmt.Sprintf("📁 Created app: apps/%s/", appName))
			ui.Info("  ├── app.go         # App configuration")
			ui.Info("  ├── schema/        # Database models")
			ui.Info("  ├── views.go       # HTTP handlers")
			ui.Info("  ├── templates/     # HTML templates")
			ui.Info("  └── static/        # Static files")
			ui.Info("")
			ui.Info("Next steps:")
			ui.Info(fmt.Sprintf("  1. Define models in apps/%s/schema/", appName))
			ui.Info("  2. Run 'make generate' to generate code")
			ui.Info("  3. Run 'make run' to start the server")

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringSliceVar(&features, "features", []string{}, 
		"App features to include (admin,api,tasks,signals)")

	return cmd
}

// isValidAppName validates the app name as a valid Go package name
func isValidAppName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Must start with a letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest can be letters, digits, or underscores
	for _, r := range name[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	// Cannot be Go keywords
	keywords := []string{
		"break", "case", "chan", "const", "continue", "default", "defer",
		"else", "fallthrough", "for", "func", "go", "goto", "if",
		"import", "interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var",
	}

	for _, keyword := range keywords {
		if name == keyword {
			return false
		}
	}

	return true
}

// isGojangoProject checks if the current directory is a Gojango project
func isGojangoProject() bool {
	// Check for gojango.yaml
	if _, err := os.Stat("gojango.yaml"); err == nil {
		return true
	}

	// Check for main.go with gojango import
	if _, err := os.Stat("main.go"); err == nil {
		if hasGojangoImport("main.go") {
			return true
		}
	}

	// Check for cmd/server/main.go
	if _, err := os.Stat("cmd/server/main.go"); err == nil {
		if hasGojangoImport("cmd/server/main.go") {
			return true
		}
	}

	return false
}

// hasGojangoImport checks if a Go file imports gojango
func hasGojangoImport(filename string) bool {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		return false
	}

	for _, imp := range node.Imports {
		if imp.Path != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(path, "github.com/epuerta9/gojango") {
				return true
			}
		}
	}

	return false
}

// appExists checks if an app directory already exists
func appExists(appName string) bool {
	_, err := os.Stat(fmt.Sprintf("apps/%s", appName))
	return err == nil
}

// getModulePath reads the module path from go.mod
func getModulePath() (string, error) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

// updateProjectFiles updates main.go to import the new app
func updateProjectFiles(appName, modulePath string) error {
	// Try to update main.go first
	if err := updateMainImports("main.go", appName, modulePath); err != nil {
		// Try cmd/server/main.go
		if err2 := updateMainImports("cmd/server/main.go", appName, modulePath); err2 != nil {
			return fmt.Errorf("failed to update imports in main.go or cmd/server/main.go: %v, %v", err, err2)
		}
	}

	return nil
}

// updateMainImports adds the app import to main.go
func updateMainImports(filename, appName, modulePath string) error {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse the file to check format
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return err
	}

	// Simple string replacement approach for now
	// In a more sophisticated implementation, we'd use AST manipulation
	lines := strings.Split(string(content), "\n")
	
	// Find the import block and add our import
	var importIndex = -1
	var inImportBlock = false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}
		
		if inImportBlock && trimmed == ")" {
			importIndex = i
			break
		}
	}
	
	if importIndex == -1 {
		return fmt.Errorf("could not find import block in %s", filename)
	}
	
	// Add our import before the closing )
	newImport := fmt.Sprintf("\t_ \"%s/apps/%s\"", modulePath, appName)
	
	// Check if already imported
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("%s/apps/%s", modulePath, appName)) {
			return nil // Already imported
		}
	}
	
	// Insert the new import
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:importIndex]...)
	newLines = append(newLines, newImport)
	newLines = append(newLines, lines[importIndex:]...)
	
	// Write back to file
	return os.WriteFile(filename, []byte(strings.Join(newLines, "\n")), 0644)
}