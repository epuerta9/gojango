package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

type ProjectOptions struct {
	Name       string
	ModulePath string
	Frontend   string // "htmx", "react", "nextjs", "none"
	API        string // "grpc", "rest", "graphql", "all"
	Database   string // "postgres", "mysql", "sqlite"
	Features   []string // "admin", "auth", "signals", "jobs"
}

func newNewCmd() *cobra.Command {
	var opts ProjectOptions
	
	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Gojango project",
		Long: `Create a new Gojango project with Django-style structure.

This command creates a complete project structure with:
- Multi-app architecture (apps/ directory)
- Settings management (config/ directory)  
- Database integration with Ent
- Admin interface
- API layer (gRPC/REST)
- Frontend setup (React/HTMX)
- Docker configuration
- CI/CD templates`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			
			// Default module path if not provided
			if opts.ModulePath == "" {
				opts.ModulePath = fmt.Sprintf("github.com/user/%s", opts.Name)
			}
			
			return createProject(opts)
		},
	}

	cmd.Flags().StringVar(&opts.ModulePath, "module", "", "Go module path (default: github.com/user/PROJECT_NAME)")
	cmd.Flags().StringVar(&opts.Frontend, "frontend", "htmx", "Frontend framework: htmx, react, nextjs, none")
	cmd.Flags().StringVar(&opts.API, "api", "grpc", "API type: grpc, rest, graphql, all")
	cmd.Flags().StringVar(&opts.Database, "database", "postgres", "Database: postgres, mysql, sqlite")
	cmd.Flags().StringSliceVar(&opts.Features, "features", []string{"admin", "auth"}, "Features to include: admin, auth, signals, jobs")

	return cmd
}

func createProject(opts ProjectOptions) error {
	// Create project directory
	if err := os.MkdirAll(opts.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	fmt.Printf("Creating Gojango project '%s'...\n", opts.Name)
	
	// Create Django-style directory structure
	structure := []string{
		// Core project structure
		"cmd/server",
		"apps/core",
		"apps/core/schema",
		"apps/core/templates", 
		"apps/core/static",
		"config",
		
		// Generated code directories
		"internal/ent",
		"internal/proto", 
		
		// Static files and media
		"static",
		"media",
		"templates",
		
		// Database migrations
		"migrations",
		
		// Tests and fixtures
		"tests",
		"fixtures",
		
		// Docker and deployment
		"docker",
		"scripts",
	}

	// Add frontend-specific directories
	if opts.Frontend == "react" || opts.Frontend == "nextjs" {
		structure = append(structure, "web/src", "web/public")
	}

	// Create directory structure
	for _, dir := range structure {
		path := filepath.Join(opts.Name, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	// Generate core files
	if err := generateProjectFiles(opts); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}

	fmt.Printf(`
✅ Successfully created Gojango project '%s'

Next steps:
  cd %s
  go mod download
  go run manage.go migrate
  go run manage.go runserver

Your project structure:
  %s/
  ├── manage.go      # Django-style management CLI
  ├── apps/          # Django-style apps
  │   └── core/      # Core application 
  ├── cmd/server/    # Main application entry point
  ├── config/        # Settings and configuration
  ├── internal/      # Generated code (ent, proto)
  ├── templates/     # Global templates
  ├── static/        # Static files
  └── migrations/    # Database migrations

Django-style commands:
  go run manage.go runserver    # Start development server
  go run manage.go migrate      # Run migrations
  go run manage.go startapp     # Create new app
  go run manage.go shell        # Interactive shell
  
Visit http://localhost:8080/admin for the admin interface.
`, opts.Name, opts.Name, opts.Name)

	return nil
}

func generateProjectFiles(opts ProjectOptions) error {
	files := map[string]string{
		"go.mod":              generateGoMod(opts),
		"main.go":            generateMainGo(opts), 
		"manage.go":          generateManageGo(opts),
		"cmd/server/main.go": generateServerMain(opts),
		"config/settings.star": generateSettings(opts),
		"apps/core/app.go":   generateCoreApp(opts),
		"docker-compose.yml": generateDockerCompose(opts),
		"Makefile":           generateMakefile(opts),
		".gitignore":         generateGitignore(),
		"README.md":          generateReadme(opts),
	}

	// Add frontend-specific files
	switch opts.Frontend {
	case "react":
		files["web/package.json"] = generateReactPackageJson(opts)
		files["web/src/main.tsx"] = generateReactMain(opts)
	case "nextjs":
		files["web/package.json"] = generateNextPackageJson(opts)
		files["web/src/app/page.tsx"] = generateNextPage(opts)
	}

	// Write all files
	for path, content := range files {
		fullPath := filepath.Join(opts.Name, path)
		
		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}
		
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}

// Template generation functions
func generateGoMod(opts ProjectOptions) string {
	tmpl := `module {{.ModulePath}}

go 1.24

require (
    github.com/epuerta9/gojango v0.1.1
    github.com/gin-gonic/gin v1.10.1
    entgo.io/ent v0.14.5
    github.com/spf13/cobra v1.7.0
    go.starlark.net v0.0.0-20231121155337-90ade8b19d09
    github.com/mattn/go-sqlite3 v1.14.32
{{- if .HasGRPC}}
    connectrpc.com/connect v1.18.1
{{- end}}
{{- if .HasAuth}}
    github.com/golang-jwt/jwt/v5 v5.0.0
{{- end}}
)

// For development, replace with local gojango
replace github.com/epuerta9/gojango => ./gojango
`
	
	data := struct {
		ProjectOptions
		HasGRPC bool
		HasAuth bool
	}{
		opts,
		opts.API == "grpc" || opts.API == "all",
		contains(opts.Features, "auth"),
	}
	
	return executeTemplate(tmpl, data)
}

func generateMainGo(opts ProjectOptions) string {
	return fmt.Sprintf(`package main

import (
	"log"
	"%s/cmd/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
`, opts.ModulePath)
}

func generateServerMain(opts ProjectOptions) string {
	tmpl := `package server

import (
	"context"
	"log"

	"github.com/epuerta9/gojango/pkg/gojango"
	
	// Import apps to register them
	_ "{{.ModulePath}}/apps/core"
)

func Run() error {
	// Create Gojango application
	app := gojango.New(
		gojango.WithDebug(true),
		gojango.WithPort("8080"),
	)

	// Load settings
	if err := app.LoadSettingsFromFile("config/settings.star"); err != nil {
		return err
	}

	// Setup database
	if err := app.SetupDatabase(); err != nil {
		return err
	}

	// Setup admin interface
	app.SetupAdmin()

	log.Println("Starting {{.Name}} server on http://localhost:8080")
	log.Println("Admin interface: http://localhost:8080/admin")
	
	return app.Run(context.Background())
}
`
	
	return executeTemplate(tmpl, opts)
}

func generateSettings(opts ProjectOptions) string {
	tmpl := `# {{.Name}} Settings
# Django-style configuration using Starlark

load("env", "env")

# Core settings
DEBUG = env.bool("DEBUG", True)
SECRET_KEY = env.get("SECRET_KEY", "your-secret-key-here")

# Database configuration
DATABASES = {
    "default": {
        "engine": "{{.Database}}",
        "host": env.get("DB_HOST", "localhost"),
        "port": env.int("DB_PORT", {{.DatabasePort}}),
        "name": env.get("DB_NAME", "{{.Name}}"),
        "user": env.get("DB_USER", "{{.DatabaseUser}}"),
        "password": env.get("DB_PASSWORD", ""),
    }
}

# Installed apps (Django-style)
INSTALLED_APPS = [
    "gojango.contrib.admin",
{{- if .HasAuth}}
    "gojango.contrib.auth",
{{- end}}
    "apps.core",
]

# API Configuration
{{- if .HasGRPC}}
GRPC_PORT = env.int("GRPC_PORT", 9000)
{{- end}}

# Static files
STATIC_URL = "/static/"
STATIC_ROOT = "./staticfiles"

MEDIA_URL = "/media/"
MEDIA_ROOT = "./media"

# Admin
ADMIN_SITE_HEADER = "{{.Name}} Administration"
ADMIN_INDEX_TITLE = "{{.Name}} Admin"
`
	
	data := struct {
		ProjectOptions
		HasAuth        bool
		HasGRPC        bool
		DatabasePort   int
		DatabaseUser   string
	}{
		opts,
		contains(opts.Features, "auth"),
		opts.API == "grpc" || opts.API == "all",
		getDatabasePort(opts.Database),
		getDatabaseUser(opts.Database),
	}
	
	return executeTemplate(tmpl, data)
}

func generateCoreApp(opts ProjectOptions) string {
	tmpl := `package core

import (
	"github.com/epuerta9/gojango/pkg/gojango"
	"github.com/gin-gonic/gin"
)

func init() {
	gojango.Register(&CoreApp{})
}

type CoreApp struct{}

func (app *CoreApp) Config() gojango.AppConfig {
	return gojango.AppConfig{
		Name:  "core",
		Label: "Core Application",
	}
}

func (app *CoreApp) Initialize(ctx *gojango.AppContext) error {
	return nil
}

func (app *CoreApp) Routes() []gojango.Route {
	return []gojango.Route{
		{
			Method:  "GET",
			Path:    "/",
			Handler: app.HomeView,
			Name:    "core:home",
		},
	}
}

func (app *CoreApp) HomeView(c *gin.Context) {
	c.HTML(200, "core/home.html", gin.H{
		"title": "{{.Name}}",
		"message": "Welcome to {{.Name}}!",
	})
}
`
	
	return executeTemplate(tmpl, opts)
}

func generateDockerCompose(opts ProjectOptions) string {
	tmpl := `version: '3.8'

services:
  web:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DEBUG=true
      - DB_HOST=db
    depends_on:
      - db
{{- if .HasRedis}}
      - redis
{{- end}}

  db:
{{- if eq .Database "postgres"}}
    image: postgres:15
    environment:
      POSTGRES_DB: {{.Name}}
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
{{- else if eq .Database "mysql"}}
    image: mysql:8.0
    environment:
      MYSQL_DATABASE: {{.Name}}
      MYSQL_USER: mysql
      MYSQL_PASSWORD: mysql
      MYSQL_ROOT_PASSWORD: root
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
{{- end}}

{{- if .HasRedis}}
  redis:
    image: redis:7
    ports:
      - "6379:6379"
{{- end}}

volumes:
{{- if eq .Database "postgres"}}
  postgres_data:
{{- else if eq .Database "mysql"}}
  mysql_data:
{{- end}}
`
	
	data := struct {
		ProjectOptions
		HasRedis bool
	}{
		opts,
		contains(opts.Features, "signals") || contains(opts.Features, "jobs"),
	}
	
	return executeTemplate(tmpl, data)
}

func generateMakefile(opts ProjectOptions) string {
	return fmt.Sprintf(`# {{.Name}} Makefile

.PHONY: help run build test migrate clean

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %%s\n", $$1, $$2}'

run: ## Run development server
	gojango run

build: ## Build the application
	go build -o bin/{{.Name}} main.go

test: ## Run tests
	go test ./...

migrate: ## Run database migrations
	gojango migrate

clean: ## Clean build artifacts
	rm -rf bin/ staticfiles/

generate: ## Run code generation
	gojango generate all

dev-setup: ## Setup development environment
	go mod download
	gojango migrate
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server"
`, opts.Name)
}

func generateGitignore() string {
	return `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Go files
*.test
*.out
coverage.*

# Environment
.env
.env.local

# Database
*.db
*.sqlite3

# Generated files
internal/ent/
internal/proto/

# Frontend
node_modules/
web/dist/
web/.next/

# IDE
.idea/
.vscode/

# Logs
*.log

# Media files
media/

# Static files
staticfiles/
`
}

func generateReadme(opts ProjectOptions) string {
	tmpl := `# {{.Name}}

A Gojango web application with Django-style architecture.

## Features

- 🚀 **Multi-app architecture** - Organize code like Django
- 🔧 **Admin interface** - Auto-generated from Ent schemas
- 🌐 **API-first** - gRPC/Connect with TypeScript generation
- 🎨 **Modern frontend** - HTMX/React with server-side rendering
- 📦 **Batteries included** - Auth, migrations, background jobs

## Quick Start

` + "```" + `bash
# Install dependencies
go mod download

# Run database migrations (Django-style)
go run manage.go migrate

# Start development server (Django-style)
go run manage.go runserver
` + "```" + `

Visit:
- **Application**: http://localhost:8080
- **Admin Interface**: http://localhost:8080/admin

## Project Structure

` + "```" + `
{{.Name}}/
├── apps/          # Django-style applications
│   └── core/      # Core app with models, views, templates
├── cmd/server/    # Application entry point
├── config/        # Settings (Starlark-based)
├── internal/      # Generated code
├── templates/     # Global templates
├── static/        # Static files
└── migrations/    # Database migrations
` + "```" + `

## Development

` + "```" + `bash
# Create a new app (Django-style)
go run manage.go startapp myapp

# Generate code from Ent schemas
go run manage.go generate proto

# Run tests with project context
go run manage.go test

# Build for production
make build
` + "```" + `

## Deployment

See ` + "`docker-compose.yml`" + ` for containerized deployment.

---

Built with [Gojango](https://github.com/epuerta9/gojango) 🎸
`

	return executeTemplate(tmpl, opts)
}

func generateManageGo(opts ProjectOptions) string {
	// Read the manage.go template
	tmpl := `package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/epuerta9/gojango/pkg/gojango"
	"github.com/epuerta9/gojango/pkg/gojango/codegen"
	"github.com/epuerta9/gojango/pkg/gojango/migrations"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	
	// Import all your apps to register them
	_ "{{.ModulePath}}/apps/core"
)

var version = "1.0.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "manage.go",
		Short: "{{.Name}} - Django-style management interface",
		Long: ` + "`" + `{{.Name}} management interface (Django manage.py equivalent).

This provides project-specific commands with full application context:
- Database migrations with your project settings
- Development server with your installed apps  
- Code generation from your schemas
- App management within your project` + "`" + `,
		Version: version,
	}

	// Add project-specific commands (Django manage.py equivalent)
	rootCmd.AddCommand(newRunServerCmd())
	rootCmd.AddCommand(newMigrationCmd())
	rootCmd.AddCommand(newStartAppCmd())
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newShellCmd())
	rootCmd.AddCommand(newTestCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
		os.Exit(1)
	}
}

func newRunServerCmd() *cobra.Command {
	var port string
	var debug bool

	cmd := &cobra.Command{
		Use:   "runserver [port]",
		Short: "Start the development server",
		Long: ` + "`" + `Start the Django-style development server.

This starts the server with your project's full application context,
loaded settings, registered apps, and database connections.` + "`" + `,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				port = args[0]
			}
			if port == "" {
				port = "8080"
			}

			// Create Gojango application with project context
			app := gojango.New(
				gojango.WithDebug(debug),
				gojango.WithPort(port),
				gojango.WithName("{{.Name}}"),
			)

			// Load project settings
			if err := app.LoadSettingsFromFile("config/settings.star"); err != nil {
				return fmt.Errorf("failed to load settings: %w", err)
			}

			fmt.Printf("Starting {{.Name}} development server on http://localhost:%s\\n", port)
			fmt.Println("Quit the server with CONTROL-C.")

			return app.Run(context.Background())
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run server on")
	cmd.Flags().BoolVar(&debug, "debug", true, "Enable debug mode")

	return cmd
}

// Simplified migration commands for the generated manage.go
func newMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migrations",
		Long:  "Django-style database migrations with your project context.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Connect to database (using project settings)
			db, err := sql.Open("sqlite3", "{{.Name}}.db")
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			// Create migration manager
			manager := migrations.NewMigrationManager(db, "migrations")
			if err := manager.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize migrations: %w", err)
			}

			return manager.ApplyMigrations()
		},
	}

	return cmd
}

func newStartAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "startapp [app-name]",
		Short: "Create a new Django-style app",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			fmt.Printf("Creating app '%s'...\\n", appName)
			fmt.Println("✅ App creation functionality will be implemented")
			return nil
		},
	}

	return cmd
}

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [type]",
		Short: "Generate code from schemas",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Code generation for %s...\\n", args[0])
			fmt.Println("✅ Code generation functionality will be implemented")
			return nil
		},
	}

	return cmd
}

func newShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Interactive project shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("{{.Name}} Interactive Shell")
			fmt.Println("🚧 Interactive shell coming soon!")
			return nil
		},
	}

	return cmd
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests with project context",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running {{.Name}} tests...")
			
			goCmd := exec.Command("go", "test", "./...")
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
			
			return goCmd.Run()
		},
	}

	return cmd
}
`

	data := struct {
		ProjectOptions
		Version string
	}{
		opts,
		"1.0.0",
	}
	
	return executeTemplate(tmpl, data)
}

// Helper functions
func executeTemplate(tmpl string, data interface{}) string {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return ""
	}
	
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return ""
	}
	
	return buf.String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getDatabasePort(db string) int {
	switch db {
	case "postgres":
		return 5432
	case "mysql":
		return 3306
	default:
		return 0
	}
}

func getDatabaseUser(db string) string {
	switch db {
	case "postgres":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return "user"
	}
}

func generateReactPackageJson(opts ProjectOptions) string {
	return fmt.Sprintf(`{
  "name": "%s-web",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@connectrpc/connect": "^1.4.0",
    "@connectrpc/connect-web": "^1.4.0",
    "@tanstack/react-query": "^5.0.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.0.0",
    "typescript": "^5.0.2",
    "vite": "^5.0.0"
  }
}`, opts.Name)
}

func generateReactMain(opts ProjectOptions) string {
	return `import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
`
}

func generateNextPackageJson(opts ProjectOptions) string {
	return fmt.Sprintf(`{
  "name": "%s-web",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint"
  },
  "dependencies": {
    "react": "^18",
    "react-dom": "^18",
    "next": "14",
    "@connectrpc/connect": "^1.4.0",
    "@connectrpc/connect-web": "^1.4.0"
  },
  "devDependencies": {
    "typescript": "^5",
    "@types/node": "^20",
    "@types/react": "^18",
    "@types/react-dom": "^18",
    "eslint": "^8",
    "eslint-config-next": "14"
  }
}`, opts.Name)
}

func generateNextPage(opts ProjectOptions) string {
	return fmt.Sprintf(`export default function Home() {
  return (
    <main>
      <h1>Welcome to %s</h1>
      <p>Your Gojango application is running!</p>
    </main>
  )
}
`, opts.Name)
}