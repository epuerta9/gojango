package gojango

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Application represents the main Gojango application instance.
// It orchestrates all registered apps and provides the main entry point.
type Application struct {
	name     string
	settings Settings
	registry *Registry
	server   *http.Server
	
	// Options
	debug bool
	port  string
}

// Option is a function that configures the Application
type Option func(*Application)

// WithName sets the application name
func WithName(name string) Option {
	return func(app *Application) {
		app.name = name
	}
}

// WithDebug enables debug mode
func WithDebug(debug bool) Option {
	return func(app *Application) {
		app.debug = debug
	}
}

// WithPort sets the server port
func WithPort(port string) Option {
	return func(app *Application) {
		app.port = port
	}
}

// New creates a new Gojango application
func New(opts ...Option) *Application {
	app := &Application{
		name:     "gojango-app",
		registry: GetRegistry(),
		debug:    false,
		port:     "8080",
	}
	
	// Apply options
	for _, opt := range opts {
		opt(app)
	}
	
	return app
}

// LoadSettings loads configuration from the provided settings implementation
func (app *Application) LoadSettings(settings Settings) error {
	app.settings = settings
	return nil
}

// Initialize initializes all registered apps
func (app *Application) Initialize(ctx context.Context) error {
	if app.settings == nil {
		return fmt.Errorf("settings not loaded - call LoadSettings() first")
	}
	
	log.Printf("Initializing Gojango application: %s", app.name)
	
	// Initialize the registry with all registered apps
	if err := app.registry.Initialize(ctx, app.settings); err != nil {
		return fmt.Errorf("failed to initialize app registry: %w", err)
	}
	
	log.Printf("Successfully initialized %d apps", len(app.registry.GetApps()))
	
	return nil
}

// setupHTTPServer sets up the HTTP server with all app routes
func (app *Application) setupHTTPServer() error {
	// For now, create a basic HTTP server
	// In Phase 2, we'll integrate Gin properly
	
	mux := http.NewServeMux()
	
	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "app": "%s"}`, app.name)
	})
	
	// Add a root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		apps := app.registry.GetAppNames()
		
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><head><title>%s</title></head><body>", app.name)
		fmt.Fprintf(w, "<h1>Welcome to %s</h1>", app.name)
		fmt.Fprintf(w, "<p>Gojango application is running!</p>")
		fmt.Fprintf(w, "<h2>Registered Apps (%d)</h2><ul>", len(apps))
		
		for _, appName := range apps {
			fmt.Fprintf(w, "<li>%s</li>", appName)
		}
		
		fmt.Fprintf(w, "</ul>")
		fmt.Fprintf(w, "<p><a href=\"/health\">Health Check</a></p>")
		fmt.Fprintf(w, "</body></html>")
	})
	
	// TODO: In Phase 2, we'll add proper route registration from apps
	// For now, just log registered routes
	allRoutes := app.registry.GetAllRoutes()
	for appName, routes := range allRoutes {
		log.Printf("App '%s' registered %d routes", appName, len(routes))
		for _, route := range routes {
			log.Printf("  %s %s -> %s", route.Method, route.Path, route.Name)
		}
	}
	
	app.server = &http.Server{
		Addr:    ":" + app.port,
		Handler: mux,
	}
	
	return nil
}

// Run starts the application server
func (app *Application) Run(ctx context.Context) error {
	// Initialize the application
	if err := app.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}
	
	// Setup HTTP server
	if err := app.setupHTTPServer(); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}
	
	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on http://localhost:%s", app.port)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	
	log.Println("Server exited")
	return nil
}

// RunCommand runs a specific command (for CLI usage)
func (app *Application) RunCommand(ctx context.Context, command string, args []string) error {
	switch command {
	case "runserver":
		return app.Run(ctx)
	case "version":
		// Initialize only for commands that need it
		if err := app.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		fmt.Println("Gojango application version 0.1.0")
		return nil
	case "apps":
		// Initialize only for commands that need it
		if err := app.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		apps := app.registry.GetAppNames()
		fmt.Printf("Registered apps (%d):\n", len(apps))
		for _, appName := range apps {
			if app, exists := app.registry.GetApp(appName); exists {
				config := app.Config()
				fmt.Printf("  %s - %s (v%s)\n", appName, config.Label, config.Version)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}