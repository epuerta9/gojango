# Gojango Framework Structure & Code Generation Architecture

## Framework Repository Structure

The Gojango framework is organized as a Go module with clear separation between public APIs, internal implementation, and code generation tools.

```
github.com/yourusername/gojango/
├── cmd/
│   └── gojango/                    # Global CLI tool
│       ├── main.go
│       ├── commands/
│       │   ├── new.go             # Create new project
│       │   ├── startapp.go        # Create new app
│       │   ├── generate.go        # Code generation commands
│       │   ├── upgrade.go         # Upgrade project to new version
│       │   └── doctor.go          # Check project health
│       └── templates/             # Templates for code generation
│           ├── project/           # New project template
│           │   ├── main.go.tmpl
│           │   ├── go.mod.tmpl
│           │   ├── Makefile.tmpl
│           │   ├── docker-compose.yml.tmpl
│           │   ├── .env.example.tmpl
│           │   ├── .gitignore.tmpl
│           │   ├── README.md.tmpl
│           │   └── config/
│           │       ├── settings.star.tmpl
│           │       └── settings_dev.star.tmpl
│           ├── app/              # New app template
│           │   ├── app.go.tmpl
│           │   ├── views.go.tmpl
│           │   ├── urls.go.tmpl
│           │   ├── admin.go.tmpl
│           │   ├── signals.go.tmpl
│           │   ├── tasks.go.tmpl
│           │   ├── schema/
│           │   │   └── models.go.tmpl
│           │   └── templates/
│           │       ├── list.templ.tmpl
│           │       └── detail.templ.tmpl
│           └── model/            # Model generation templates
│               ├── schema.go.tmpl
│               ├── admin.go.tmpl
│               └── service.proto.tmpl
│
├── pkg/                           # Public packages (users import these)
│   ├── gojango/                  # Core framework
│   │   ├── app.go               # App interface and base implementation
│   │   ├── application.go       # Main application struct
│   │   ├── context.go           # Request and app contexts
│   │   ├── registry.go          # Global app registry
│   │   ├── signals.go           # Cross-app signals system
│   │   ├── settings.go          # Starlark settings loader
│   │   ├── version.go           # Version management
│   │   └── errors.go            # Common error types
│   │
│   ├── admin/                   # Admin framework
│   │   ├── site.go             # Admin site registry
│   │   ├── model.go            # Model admin base
│   │   ├── generator.go        # Auto-generation from Ent
│   │   ├── actions.go          # Bulk actions
│   │   ├── filters.go          # List filters
│   │   ├── widgets/            # Form widgets
│   │   │   ├── text.go
│   │   │   ├── select.go
│   │   │   ├── datetime.go
│   │   │   └── file.go
│   │   └── templates/          # Admin templates
│   │       ├── base.templ
│   │       ├── list.templ
│   │       ├── form.templ
│   │       ├── delete.templ
│   │       └── components/
│   │           ├── pagination.templ
│   │           └── filters.templ
│   │
│   ├── auth/                    # Authentication system
│   │   ├── middleware.go       # Auth middleware
│   │   ├── user.go            # User interface
│   │   ├── permissions.go     # Permission system
│   │   ├── decorators.go      # View decorators
│   │   ├── backends/          # Auth backends
│   │   │   ├── session.go
│   │   │   ├── jwt.go
│   │   │   ├── oauth.go
│   │   │   └── apikey.go
│   │   ├── password.go        # Password utilities
│   │   └── tokens.go          # Token generation
│   │
│   ├── db/                     # Database utilities
│   │   ├── connection.go      # Connection management
│   │   ├── migrations.go      # Migration runner
│   │   ├── fixtures.go        # Fixture loader
│   │   ├── backends/          # Database backends
│   │   │   ├── postgres.go
│   │   │   ├── mysql.go
│   │   │   └── sqlite.go
│   │   └── ent/              # Ent extensions
│   │       ├── mixins/       # Common mixins
│   │       │   ├── timestamp.go
│   │       │   ├── softdelete.go
│   │       │   └── uuid.go
│   │       ├── hooks.go      # Common hooks
│   │       └── annotations.go # Custom annotations
│   │
│   ├── routing/                # Routing system
│   │   ├── router.go         # Router wrapper
│   │   ├── middleware/       # Common middleware
│   │   │   ├── cors.go
│   │   │   ├── csrf.go
│   │   │   ├── ratelimit.go
│   │   │   ├── logging.go
│   │   │   └── recovery.go
│   │   ├── reverse.go        # URL reversal
│   │   ├── static.go         # Static file serving
│   │   └── websocket.go      # WebSocket support
│   │
│   ├── templates/              # Template system
│   │   ├── engine.go         # Templ wrapper
│   │   ├── funcs.go          # Template functions
│   │   ├── cache.go          # Template caching
│   │   ├── loader.go         # Template loader
│   │   └── components/       # Reusable components
│   │       ├── pagination.templ
│   │       ├── forms.templ
│   │       ├── messages.templ
│   │       └── breadcrumbs.templ
│   │
│   ├── cache/                  # Cache framework
│   │   ├── cache.go          # Cache interface
│   │   ├── backends/
│   │   │   ├── redis.go
│   │   │   ├── memory.go
│   │   │   ├── memcached.go
│   │   │   └── hybrid.go    # Multi-tier cache
│   │   ├── decorators.go     # Function decorators
│   │   └── tags.go           # Cache tag invalidation
│   │
│   ├── tasks/                  # Background tasks
│   │   ├── worker.go         # Task worker
│   │   ├── scheduler.go      # Cron scheduler
│   │   ├── queue.go          # Task queue
│   │   ├── backends/
│   │   │   ├── asynq.go     # Redis-based
│   │   │   └── database.go  # DB-based queue
│   │   └── monitoring.go     # Task monitoring
│   │
│   ├── api/                    # API utilities
│   │   ├── connect.go        # Connect/gRPC setup
│   │   ├── graphql.go        # GraphQL setup
│   │   ├── rest.go           # REST utilities
│   │   ├── openapi.go        # OpenAPI generation
│   │   ├── serializers.go    # Serialization helpers
│   │   └── validators.go     # Input validation
│   │
│   ├── forms/                  # Form handling
│   │   ├── form.go           # Form base class
│   │   ├── fields.go         # Form fields
│   │   ├── validation.go     # Validators
│   │   └── rendering.go      # Form rendering
│   │
│   ├── testing/                # Testing utilities
│   │   ├── client.go         # Test client
│   │   ├── fixtures.go       # Fixture management
│   │   ├── assertions.go     # Custom assertions
│   │   ├── factories.go      # Model factories
│   │   └── mocks.go          # Common mocks
│   │
│   ├── storage/                # File storage
│   │   ├── storage.go        # Storage interface
│   │   ├── backends/
│   │   │   ├── s3.go        # AWS S3
│   │   │   ├── gcs.go       # Google Cloud Storage
│   │   │   ├── azure.go     # Azure Blob
│   │   │   └── local.go     # Filesystem
│   │   └── media.go          # Media handling
│   │
│   ├── nats/                   # NATS integration
│   │   ├── server.go         # Embedded NATS server
│   │   ├── signals.go        # Signal pub/sub
│   │   ├── client.go         # NATS client wrapper
│   │   └── jetstream.go      # JetStream support
│   │
│   └── email/                  # Email system
│       ├── backend.go        # Email backend interface
│       ├── backends/
│       │   ├── smtp.go       # SMTP backend
│       │   ├── sendgrid.go  # SendGrid
│       │   ├── ses.go        # AWS SES
│       │   └── console.go   # Development
│       └── templates.go      # Email templates
│
├── internal/                    # Internal packages (not exported)
│   ├── codegen/                # Code generation logic
│   │   ├── project.go        # Project generator
│   │   ├── app.go            # App generator
│   │   ├── model.go          # Model generator
│   │   ├── parser.go         # Go AST parser
│   │   ├── analyzer.go       # Code analyzer
│   │   ├── templates.go      # Template engine
│   │   └── imports.go        # Import management
│   │
│   ├── cli/                    # CLI utilities
│   │   ├── colors.go         # Terminal colors
│   │   ├── prompts.go        # Interactive prompts
│   │   ├── spinner.go        # Progress indicators
│   │   ├── table.go          # Table formatting
│   │   └── output.go         # Output formatting
│   │
│   ├── build/                  # Build utilities
│   │   ├── embed.go          # Asset embedding
│   │   ├── bundle.go         # Frontend bundling
│   │   └── docker.go         # Docker utils
│   │
│   └── utils/                  # Internal utilities
│       ├── strings.go        # String helpers
│       ├── files.go          # File helpers
│       ├── reflection.go     # Reflection utils
│       └── validation.go     # Validation helpers
│
├── contrib/                     # Optional contrib apps
│   ├── blog/                   # Blog app
│   ├── shop/                   # E-commerce app
│   ├── cms/                    # CMS app
│   ├── forum/                  # Forum app
│   └── wiki/                   # Wiki app
│
├── examples/                    # Example projects
│   ├── blog/                   # Simple blog
│   ├── saas/                   # SaaS starter
│   ├── api/                    # API-only project
│   ├── marketplace/            # Marketplace
│   └── social/                 # Social network
│
├── tools/                       # Development tools
│   ├── protoc-gen-gojango/    # Proto generator plugin
│   ├── entc-gen-admin/        # Ent admin generator
│   └── gojango-lsp/           # Language server
│
├── scripts/                     # Build & release scripts
│   ├── install.sh              # Installation script
│   ├── release.sh              # Release script
│   ├── test.sh                 # Test runner
│   └── benchmark.sh            # Performance tests
│
├── docs/                        # Documentation
│   ├── getting-started.md
│   ├── tutorial/
│   ├── guides/
│   ├── api/
│   ├── deployment/
│   └── contributing.md
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── LICENSE
└── CHANGELOG.md
```

## Code Generation Implementation

### Project Generator

```go
// cmd/gojango/commands/new.go
package commands

import (
    "embed"
    "fmt"
    "os"
    "path/filepath"
    "text/template"
    
    "github.com/spf13/cobra"
    "github.com/yourusername/gojango/internal/cli"
    "github.com/yourusername/gojango/internal/codegen"
)

//go:embed all:templates/project
var projectTemplates embed.FS

type ProjectConfig struct {
    Name           string
    Module         string
    GoVersion      string
    GojangoVersion string
    Frontend       string // htmx, react, vue
    API            string // rest, grpc, graphql, all
    Database       string // postgres, mysql, sqlite
    Cache          string // redis, memcached, none
    MessageQueue   string // asynq, rabbitmq, none
    Features       []string
}

func NewProjectCmd() *cobra.Command {
    var config ProjectConfig
    
    cmd := &cobra.Command{
        Use:   "new [project-name]",
        Short: "Create a new Gojango project",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            config.Name = args[0]
            
            // Interactive mode if no flags provided
            if !cmd.Flags().Changed("module") {
                if err := interactiveProjectSetup(&config); err != nil {
                    return err
                }
            }
            
            // Create project
            generator := codegen.NewProjectGenerator(projectTemplates)
            if err := generator.Generate(config); err != nil {
                return fmt.Errorf("failed to generate project: %w", err)
            }
            
            // Post-generation steps
            if err := postGenerateSteps(config); err != nil {
                return err
            }
            
            // Print success message
            printSuccessMessage(config)
            
            return nil
        },
    }
    
    // Add flags
    cmd.Flags().StringVar(&config.Module, "module", "", "Go module path")
    cmd.Flags().StringVar(&config.Frontend, "frontend", "htmx", "Frontend framework")
    cmd.Flags().StringVar(&config.API, "api", "rest", "API style")
    cmd.Flags().StringVar(&config.Database, "database", "postgres", "Database")
    cmd.Flags().StringVar(&config.Cache, "cache", "redis", "Cache backend")
    cmd.Flags().StringSliceVar(&config.Features, "features", []string{}, "Additional features")
    
    return cmd
}

func interactiveProjectSetup(config *ProjectConfig) error {
    // Module path
    config.Module = cli.Prompt("Module path", fmt.Sprintf("github.com/yourusername/%s", config.Name))
    
    // Frontend selection
    frontend := cli.Select("Choose frontend framework", []string{
        "htmx - Server-side rendering with HTMX",
        "react - React with TypeScript",
        "vue - Vue 3 with TypeScript",
        "none - API only",
    })
    config.Frontend = extractChoice(frontend)
    
    // API style
    api := cli.Select("Choose API style", []string{
        "rest - RESTful API",
        "grpc - gRPC with Connect",
        "graphql - GraphQL API",
        "all - All API styles",
    })
    config.API = extractChoice(api)
    
    // Database
    database := cli.Select("Choose database", []string{
        "postgres - PostgreSQL",
        "mysql - MySQL/MariaDB",
        "sqlite - SQLite",
    })
    config.Database = extractChoice(database)
    
    // Additional features
    features := cli.MultiSelect("Select additional features", []string{
        "admin - Admin interface",
        "auth - Authentication system",
        "api-docs - API documentation",
        "docker - Docker support",
        "ci - CI/CD pipelines",
        "monitoring - Monitoring setup",
    })
    config.Features = features
    
    return nil
}

func postGenerateSteps(config ProjectConfig) error {
    projectPath := config.Name
    
    // Initialize git
    if err := initGit(projectPath); err != nil {
        return fmt.Errorf("failed to initialize git: %w", err)
    }
    
    // Run go mod tidy
    if err := runGoModTidy(projectPath); err != nil {
        return fmt.Errorf("failed to tidy go modules: %w", err)
    }
    
    // Install frontend dependencies if needed
    if config.Frontend != "htmx" && config.Frontend != "none" {
        if err := installFrontendDeps(projectPath, config.Frontend); err != nil {
            return fmt.Errorf("failed to install frontend dependencies: %w", err)
        }
    }
    
    // Generate .env file from .env.example
    if err := generateEnvFile(projectPath); err != nil {
        return fmt.Errorf("failed to generate .env file: %w", err)
    }
    
    return nil
}

func printSuccessMessage(config ProjectConfig) {
    fmt.Println(cli.Green("✨ Successfully created project: " + config.Name))
    fmt.Println()
    fmt.Println("📁 Project structure:")
    fmt.Println("   " + config.Name + "/")
    fmt.Println("   ├── apps/          # Your applications")
    fmt.Println("   ├── config/        # Configuration files")
    fmt.Println("   ├── static/        # Static files")
    fmt.Println("   ├── templates/     # Global templates")
    fmt.Println("   └── main.go        # Entry point")
    fmt.Println()
    fmt.Println("🚀 Next steps:")
    fmt.Println()
    fmt.Printf("   cd %s\n", config.Name)
    fmt.Println("   cp .env.example .env  # Configure environment")
    fmt.Println("   make setup            # Setup database")
    fmt.Println("   make migrate          # Run migrations")
    fmt.Println("   make run              # Start development server")
    fmt.Println()
    fmt.Println("📚 Documentation: https://gojango.dev")
}
```

### App Generator

```go
// cmd/gojango/commands/startapp.go
package commands

import (
    "embed"
    "fmt"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/spf13/cobra"
    "github.com/yourusername/gojango/internal/codegen"
)

//go:embed all:templates/app
var appTemplates embed.FS

type AppConfig struct {
    Name       string
    NameTitle  string
    Module     string
    Package    string
    Features   []string
}

func StartAppCmd() *cobra.Command {
    var features []string
    
    cmd := &cobra.Command{
        Use:   "startapp [app-name]",
        Short: "Create a new app in the current project",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            appName := args[0]
            
            // Validate we're in a Gojango project
            if !isGojangoProject() {
                return fmt.Errorf("not in a Gojango project directory")
            }
            
            // Check if app already exists
            if appExists(appName) {
                return fmt.Errorf("app '%s' already exists", appName)
            }
            
            config := AppConfig{
                Name:      appName,
                NameTitle: strings.Title(appName),
                Module:    getModulePath(),
                Package:   appName,
                Features:  features,
            }
            
            // Generate app structure
            generator := codegen.NewAppGenerator(appTemplates)
            if err := generator.Generate(config); err != nil {
                return fmt.Errorf("failed to generate app: %w", err)
            }
            
            // Update project files
            if err := updateProjectFiles(appName); err != nil {
                return fmt.Errorf("failed to update project files: %w", err)
            }
            
            printAppSuccess(appName)
            
            return nil
        },
    }
    
    cmd.Flags().StringSliceVar(&features, "features", []string{}, 
        "App features (admin,api,tasks,signals)")
    
    return cmd
}

func isGojangoProject() bool {
    // Check for gojango.yaml or main.go with gojango import
    if _, err := os.Stat("gojango.yaml"); err == nil {
        return true
    }
    
    if _, err := os.Stat("main.go"); err == nil {
        // Parse main.go and check for gojango import
        fset := token.NewFileSet()
        node, err := parser.ParseFile(fset, "main.go", nil, parser.ImportsOnly)
        if err == nil {
            for _, imp := range node.Imports {
                if strings.Contains(imp.Path.Value, "gojango") {
                    return true
                }
            }
        }
    }
    
    return false
}

func updateProjectFiles(appName string) error {
    // Update main.go to import the new app
    if err := updateMainImports(appName); err != nil {
        return err
    }
    
    // Add to settings.star INSTALLED_APPS
    if err := updateSettings(appName); err != nil {
        return err
    }
    
    // Update gojango.yaml if it exists
    if err := updateProjectConfig(appName); err != nil {
        // Non-fatal error
        fmt.Printf("Warning: Could not update gojango.yaml: %v\n", err)
    }
    
    return nil
}

func updateMainImports(appName string) error {
    mainPath := "main.go"
    content, err := os.ReadFile(mainPath)
    if err != nil {
        return err
    }
    
    // Parse the file
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, mainPath, content, parser.ParseComments)
    if err != nil {
        return err
    }
    
    // Add import using AST manipulation or simple string replacement
    // For simplicity, using string replacement here
    lines := strings.Split(string(content), "\n")
    
    // Find the import block
    importIdx := -1
    for i, line := range lines {
        if strings.Contains(line, `_ "github.com/yourusername/gojango/contrib`) {
            importIdx = i
            break
        }
    }
    
    if importIdx != -1 {
        // Insert new import after contrib imports
        module := getModulePath()
        newImport := fmt.Sprintf(`    _ "%s/apps/%s"`, module, appName)
        lines = append(lines[:importIdx+1], 
            append([]string{newImport}, lines[importIdx+1:]...)...)
        
        // Write back
        return os.WriteFile(mainPath, []byte(strings.Join(lines, "\n")), 0644)
    }
    
    return fmt.Errorf("could not find import block in main.go")
}

func printAppSuccess(appName string) {
    fmt.Printf("✅ Created app '%s'\n", appName)
    fmt.Println()
    fmt.Println("📁 App structure:")
    fmt.Printf("   apps/%s/\n", appName)
    fmt.Println("   ├── app.go         # App configuration")
    fmt.Println("   ├── schema/        # Ent models")
    fmt.Println("   ├── views.go       # HTTP handlers")
    fmt.Println("   ├── admin.go       # Admin configuration")
    fmt.Println("   └── templates/     # App templates")
    fmt.Println()
    fmt.Println("📝 Next steps:")
    fmt.Println()
    fmt.Printf("   1. Define your models in apps/%s/schema/\n", appName)
    fmt.Println("   2. Run 'make generate' to generate code")
    fmt.Println("   3. Run 'make migrate' to create database tables")
    fmt.Println()
    fmt.Printf("💡 Example model: apps/%s/schema/example.go\n", appName)
}
```

### Template Files

#### Project main.go Template

```go
// cmd/gojango/templates/project/main.go.tmpl
package main

import (
    "log"
    "os"
    
    "github.com/yourusername/gojango/pkg/gojango"
    
    // Core contrib apps
    _ "github.com/yourusername/gojango/contrib/auth"
    {{- if has .Features "admin"}}
    _ "github.com/yourusername/gojango/contrib/admin"
    {{- end}}
    
    // Your apps will be imported here
    // _ "{{.Module}}/apps/myapp"
)

func main() {
    // Create application with options
    opts := []gojango.Option{
        gojango.WithName("{{.Name}}"),
        {{- if eq .Frontend "react"}}
        gojango.WithFrontend(gojango.React),
        {{- else if eq .Frontend "vue"}}
        gojango.WithFrontend(gojango.Vue),
        {{- else if eq .Frontend "htmx"}}
        gojango.WithFrontend(gojango.HTMX),
        {{- end}}
        {{- if or (eq .API "grpc") (eq .API "all")}}
        gojango.WithGRPC(),
        {{- end}}
        {{- if or (eq .API "graphql") (eq .API "all")}}
        gojango.WithGraphQL(),
        {{- end}}
        {{- if eq .Database "postgres"}}
        gojango.WithDatabase(gojango.Postgres),
        {{- else if eq .Database "mysql"}}
        gojango.WithDatabase(gojango.MySQL),
        {{- else if eq .Database "sqlite"}}
        gojango.WithDatabase(gojango.SQLite),
        {{- end}}
    }
    
    app := gojango.New(opts...)
    
    // Load settings based on environment
    env := os.Getenv("GOJANGO_ENV")
    if env == "" {
        env = "development"
    }
    
    settingsFile := "config/settings.star"
    if env != "development" {
        settingsFile = fmt.Sprintf("config/settings_%s.star", env)
    }
    
    if err := app.LoadSettings(settingsFile); err != nil {
        log.Fatal("Failed to load settings:", err)
    }
    
    // Execute CLI commands
    if err := app.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

#### App Template

```go
// cmd/gojango/templates/app/app.go.tmpl
package {{.Package}}

import (
    "embed"
    
    "github.com/yourusername/gojango/pkg/gojango"
    "{{.Module}}/internal/ent"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed migrations/*.sql
var migrationFS embed.FS

func init() {
    // Register app on import
    gojango.Register(&{{.NameTitle}}App{})
}

type {{.NameTitle}}App struct {
    gojango.BaseApp
    ent *ent.Client
}

func (app *{{.NameTitle}}App) Config() gojango.AppConfig {
    return gojango.AppConfig{
        Name:    "{{.Name}}",
        Label:   "{{.NameTitle}} Application",
        Version: "1.0.0",
        Dependencies: []string{
            "auth",  // Depends on auth app
        },
    }
}

func (app *{{.NameTitle}}App) Initialize(ctx *gojango.AppContext) error {
    // Store references
    app.ent = ctx.DB
    
    // Register templates
    ctx.Templates.Register("{{.Name}}", templateFS)
    
    // Register static files  
    ctx.Static.Register("{{.Name}}", staticFS)
    
    // Register migrations
    ctx.Migrations.Register("{{.Name}}", migrationFS)
    
    {{- if has .Features "signals"}}
    // Setup signals
    app.setupSignals(ctx)
    {{- end}}
    
    {{- if has .Features "tasks"}}
    // Register background tasks
    app.registerTasks(ctx)
    {{- end}}
    
    return nil
}

{{- if has .Features "admin"}}
func (app *{{.NameTitle}}App) AdminSite() *admin.AdminSite {
    site := admin.NewAdminSite("{{.Name}}")
    
    // Register models with admin
    // site.Register(&PostAdmin{})
    
    return site
}
{{- end}}

{{- if has .Features "api"}}
func (app *{{.NameTitle}}App) Services() []gojango.Service {
    return []gojango.Service{
        // &{{.NameTitle}}Service{app: app},
    }
}
{{- end}}
```

## User Project Structure

The generated user project follows Django's app-based architecture adapted for Go:

```
myproject/
├── apps/                        # Application modules
│   ├── core/                   # Core/common functionality
│   │   ├── app.go             # App registration and config
│   │   ├── schema/            # Ent schemas (models)
│   │   │   ├── user.go       # User model
│   │   │   ├── mixins.go     # Shared mixins
│   │   │   └── README.md     # Schema documentation
│   │   ├── views.go          # HTTP handlers (views)
│   │   ├── urls.go           # URL routing patterns
│   │   ├── admin.go          # Admin interface config
│   │   ├── api.go            # API endpoints
│   │   ├── services.go       # gRPC service implementations
│   │   ├── tasks.go          # Background task definitions
│   │   ├── signals.go        # Signal handlers
│   │   ├── forms.go          # Form definitions
│   │   ├── validators.go     # Custom validators
│   │   ├── templates/        # Templ templates
│   │   │   ├── base.templ   # Base template
│   │   │   ├── home.templ   # Homepage
│   │   │   └── components/  # Reusable components
│   │   ├── static/           # App-specific static files
│   │   │   ├── css/
│   │   │   ├── js/
│   │   │   └── img/
│   │   ├── migrations/       # Database migrations
│   │   │   ├── 001_initial.sql
│   │   │   └── 002_add_fields.sql
│   │   ├── fixtures/         # Test data
│   │   │   └── initial_data.json
│   │   └── tests/            # App tests
│   │       ├── models_test.go
│   │       ├── views_test.go
│   │       └── api_test.go
│   │
│   ├── blog/                  # Blog application
│   │   ├── app.go
│   │   ├── schema/
│   │   │   ├── post.go      # Post model
│   │   │   ├── comment.go   # Comment model
│   │   │   └── category.go  # Category model
│   │   ├── views.go
│   │   ├── api.go
│   │   ├── admin.go
│   │   ├── templates/
│   │   │   ├── list.templ
│   │   │   ├── detail.templ
│   │   │   └── components/
│   │   │       └── post_card.templ
│   │   └── tests/
│   │
│   └── shop/                  # E-commerce application
│       ├── app.go
│       ├── schema/
│       │   ├── product.go
│       │   ├── order.go
│       │   └── cart.go
│       └── ...
│
├── cmd/                        # Custom management commands
│   └── custom/
│       ├── import_data.go    # Data import command
│       └── cleanup.go        # Cleanup command
│
├── config/                     # Configuration files
│   ├── settings.star          # Main settings file
│   ├── settings_dev.star      # Development overrides
│   ├── settings_test.star     # Test settings
│   ├── settings_staging.star  # Staging settings
│   └── settings_prod.star     # Production settings
│
├── internal/                   # Private packages
│   ├── ent/                  # Generated Ent code
│   │   ├── entc.go          # Ent configuration
│   │   ├── schema/          # Combined schemas
│   │   ├── migrate/         # Migration schemas
│   │   ├── client.go        # Generated client
│   │   ├── user.go          # Generated user code
│   │   └── ...              # Other generated files
│   │
│   ├── proto/                # Generated protobuf code
│   │   └── gen/
│   │       ├── blog/
│   │       │   └── v1/
│   │       │       ├── blog.pb.go
│   │       │       └── blog_grpc.pb.go
│   │       └── shop/
│   │
│   ├── graphql/              # Generated GraphQL code
│   │   ├── generated.go
│   │   ├── resolver.go
│   │   └── schema.graphql
│   │
│   └── utils/                # Internal utilities
│       ├── helpers.go
│       ├── validators.go
│       └── converters.go
│
├── web/                       # Frontend application
│   ├── src/
│   │   ├── api/             # Generated TypeScript client
│   │   │   ├── blog.ts
│   │   │   └── shop.ts
│   │   ├── components/      # React/Vue components
│   │   ├── pages/          # Page components
│   │   ├── hooks/          # Custom hooks
│   │   ├── store/          # State management
│   │   ├── styles/         # Global styles
│   │   ├── utils/          # Frontend utilities
│   │   └── App.tsx         # Main app component
│   ├── public/
│   │   ├── index.html
│   │   └── favicon.ico
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── README.md
│
├── static/                    # Global static files
│   ├── css/
│   │   └── global.css
│   ├── js/
│   │   └── app.js
│   ├── img/
│   └── vendor/               # Third-party assets
│
├── media/                     # User-uploaded files
│   ├── uploads/
│   └── temp/
│
├── templates/                 # Global templates
│   ├── base.templ           # Global base template
│   ├── errors/              # Error pages
│   │   ├── 404.templ
│   │   ├── 403.templ
│   │   └── 500.templ
│   └── email/               # Email templates
│       ├── welcome.templ
│       └── reset_password.templ
│
├── proto/                     # Protocol buffer definitions
│   ├── blog/
│   │   └── v1/
│   │       ├── blog.proto
│   │       └── blog_service.proto
│   └── shop/
│       └── v1/
│           └── shop.proto
│
├── migrations/                # Global migrations
│   ├── 001_initial.sql
│   └── README.md
│
├── fixtures/                  # Global test data
│   ├── users.json
│   └── sample_data.json
│
├── tests/                     # Integration tests
│   ├── integration/
│   │   ├── api_test.go
│   │   └── e2e_test.go
│   └── benchmarks/
│       └── performance_test.go
│
├── scripts/                   # Utility scripts
│   ├── deploy.sh            # Deployment script
│   ├── backup.sh            # Backup script
│   ├── migrate_data.py      # Data migration
│   └── health_check.sh      # Health check
│
├── deployments/              # Deployment configurations
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── ingress.yaml
│   │   └── configmap.yaml
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── Dockerfile.dev
│   │   └── docker-compose.prod.yml
│   └── terraform/
│       └── infrastructure.tf
│
├── docs/                      # Project documentation
│   ├── README.md
│   ├── API.md
│   ├── DEPLOYMENT.md
│   └── CONTRIBUTING.md
│
├── .github/                   # GitHub specific files
│   ├── workflows/
│   │   ├── ci.yml
│   │   └── deploy.yml
│   └── ISSUE_TEMPLATE/
│
├── main.go                   # Application entry point
├── go.mod                    # Go module file
├── go.sum                    # Go dependencies
├── package.json              # Frontend dependencies (if applicable)
├── Makefile                  # Build automation
├── docker-compose.yml        # Local development setup
├── .env.example              # Environment variables template
├── .gitignore               # Git ignore rules
├── .dockerignore            # Docker ignore rules
├── .editorconfig            # Editor configuration
├── README.md                # Project documentation
├── LICENSE                  # License file
└── gojango.yaml            # Gojango project metadata
```

## Project Metadata File

```yaml
# gojango.yaml - Project configuration
name: myproject
version: 1.0.0
gojango_version: 1.0.0

apps:
  - core
  - blog
  - shop

frontend:
  framework: react
  bundler: vite
  
api:
  styles:
    - rest
    - grpc
    - graphql
    
database:
  engine: postgres
  
cache:
  backend: redis
  
storage:
  backend: s3
  
features:
  - admin
  - auth
  - api-docs
  - monitoring
  
deployment:
  platform: kubernetes
  registry: gcr.io/myproject
```

## CLI Tool Separation

### Global CLI (Installed via `go install`)

The global `gojango` CLI handles project scaffolding and code generation:

```bash
# Install globally
go install github.com/yourusername/gojango/cmd/gojango@latest

# Commands available globally
gojango new myproject          # Create new project
gojango startapp blog          # Create new app (in project)
gojango generate model Post    # Generate model
gojango generate admin         # Generate admin
gojango generate typescript    # Generate TS types
gojango upgrade               # Upgrade Gojango version
gojango doctor                # Check project health
```

### Project Binary (Compiled from main.go)

The project-specific binary handles runtime operations:

```bash
# Development (using go run)
go run main.go runserver       # Start dev server
go run main.go migrate         # Run migrations
go run main.go shell          # Interactive shell
go run main.go test           # Run tests
go run main.go createsuperuser # Create admin user

# Production (compiled binary)
go build -o myapp
./myapp runserver --production
./myapp worker                # Start background worker
./myapp scheduler             # Start task scheduler
```

## Makefile for Better DX

```makefile
# Makefile - Developer convenience commands
.PHONY: help run migrate test generate build clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Development
run: ## Run development server
	@go run main.go runserver

shell: ## Open interactive shell
	@go run main.go shell

migrate: ## Run database migrations
	@go run main.go migrate

makemigrations: ## Create new migrations
	@gojango generate migrations
	@go run main.go migrate --check

seed: ## Load fixture data
	@go run main.go loaddata fixtures/*.json

# Code Generation
generate: ## Generate all code
	@gojango generate all
	@go generate ./...

generate-ent: ## Generate Ent code
	@go generate ./internal/ent

generate-proto: ## Generate protobuf code
	@buf generate

generate-ts: ## Generate TypeScript client
	@gojango generate typescript

# Testing
test: ## Run all tests
	@go test ./...
	@go run main.go test

test-unit: ## Run unit tests
	@go test ./apps/...

test-integration: ## Run integration tests
	@go test ./tests/integration/...

coverage: ## Generate test coverage
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Building
build: ## Build production binary
	@echo "Building production binary..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/myapp main.go
	@echo "Binary created at bin/myapp"

build-docker: ## Build Docker image
	@docker build -t myapp:latest .

# Database
db-shell: ## Open database shell
	@go run main.go dbshell

db-dump: ## Dump database
	@scripts/backup.sh

db-restore: ## Restore database
	@scripts/restore.sh

# Frontend (if applicable)
frontend-dev: ## Run frontend dev server
	@cd web && npm run dev

frontend-build: ## Build frontend
	@cd web && npm run build

frontend-test: ## Test frontend
	@cd web && npm test

# Deployment
deploy-staging: ## Deploy to staging
	@scripts/deploy.sh staging

deploy-production: ## Deploy to production
	@scripts/deploy.sh production

# Utilities
clean: ## Clean build artifacts
	@rm -rf bin/ dist/ coverage.out
	@go clean -testcache

setup: ## Initial project setup
	@cp .env.example .env
	@docker-compose up -d
	@sleep 5
	@go run main.go migrate
	@go run main.go createsuperuser
	@echo "Setup complete! Run 'make run' to start the server"

fmt: ## Format code
	@go fmt ./...
	@gojango fmt

lint: ## Run linters
	@golangci-lint run
	@gojango lint

upgrade: ## Upgrade Gojango
	@gojango upgrade
	@go get -u github.com/yourusername/gojango
	@go mod tidy
```

## Summary

This structure provides:

1. **Clear separation** between framework code and user code
2. **Comprehensive code generation** for rapid development
3. **Familiar project structure** for Django developers
4. **Type safety** throughout the stack
5. **Single binary deployment** with embedded assets
6. **Excellent developer experience** with helpful CLI tools
7. **Scalable architecture** supporting monolith to microservices evolution

The framework handles the complexity while exposing simple, composable interfaces to users, maintaining Go's philosophy while providing Django's productivity.