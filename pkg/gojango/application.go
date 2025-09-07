package gojango

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/epuerta9/gojango/pkg/gojango/middleware"
	"github.com/epuerta9/gojango/pkg/gojango/routing"
	"github.com/epuerta9/gojango/pkg/gojango/templates"
	"github.com/gin-gonic/gin"
)

// Application represents the main Gojango application instance.
// It orchestrates all registered apps and provides the main entry point.
type Application struct {
	name     string
	settings Settings
	registry *Registry
	router   *routing.Router
	templates *templates.Engine
	server   *http.Server
	middleware *middleware.Registry
	
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

// WithMiddleware sets a custom middleware registry
func WithMiddleware(middlewareRegistry *middleware.Registry) Option {
	return func(app *Application) {
		app.middleware = middlewareRegistry
	}
}

// New creates a new Gojango application
func New(opts ...Option) *Application {
	app := &Application{
		name:      "gojango-app",
		registry:  GetRegistry(),
		router:    routing.NewRouter(),
		templates: templates.NewEngine(),
		debug:     false,
		port:      "8080",
	}
	
	// Apply options
	for _, opt := range opts {
		opt(app)
	}
	
	// Set default middleware if none provided
	if app.middleware == nil {
		if app.debug {
			app.middleware = middleware.GetDevelopment()
		} else {
			app.middleware = middleware.GetDefaults()
		}
	}
	
	// Configure Gin based on debug mode
	if app.debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	return app
}

// LoadSettings loads configuration from the provided settings implementation
func (app *Application) LoadSettings(settings Settings) error {
	app.settings = settings
	return nil
}

// AddMiddleware adds a middleware function to the application
func (app *Application) AddMiddleware(middleware middleware.MiddlewareFunc) {
	app.middleware.Add(middleware)
}

// AddGinMiddleware adds a Gin HandlerFunc directly as middleware
func (app *Application) AddGinMiddleware(handler gin.HandlerFunc) {
	app.middleware.AddGin(handler)
}

// Initialize initializes all registered apps
func (app *Application) Initialize(ctx context.Context) error {
	if app.settings == nil {
		return fmt.Errorf("settings not loaded - call LoadSettings() first")
	}
	
	log.Printf("Initializing Gojango application: %s", app.name)
	
	// Setup middleware
	app.setupMiddleware()
	
	// Setup template functions (needs to be before app initialization)
	app.templates.AddFuncs(app.router.TemplateFuncs())
	
	// Initialize the registry with all registered apps
	if err := app.registry.Initialize(ctx, app.settings); err != nil {
		return fmt.Errorf("failed to initialize app registry: %w", err)
	}
	
	// Setup routing and templates after apps are initialized
	if err := app.setupRouting(); err != nil {
		return fmt.Errorf("failed to setup routing: %w", err)
	}
	
	if err := app.setupTemplates(); err != nil {
		return fmt.Errorf("failed to setup templates: %w", err)
	}
	
	log.Printf("Successfully initialized %d apps", len(app.registry.GetApps()))
	
	return nil
}

// setupMiddleware configures the middleware stack
func (app *Application) setupMiddleware() {
	// Apply middleware from the registry
	app.middleware.Apply(app.router.GetEngine())
}

// setupRouting registers routes from all apps
func (app *Application) setupRouting() error {
	// Add built-in routes first
	app.addBuiltinRoutes()
	
	// Register routes from all apps
	allRoutes := app.registry.GetAllRoutes()
	for appName, routes := range allRoutes {
		if len(routes) > 0 {
			// Convert gojango.Route to routing.Route
			routingRoutes := make([]routing.Route, len(routes))
			for i, route := range routes {
				routingRoutes[i] = routing.Route{
					Method:  route.Method,
					Path:    route.Path,
					Handler: route.Handler,
					Name:    route.Name,
				}
			}
			
			err := app.router.RegisterRoutes(appName, routingRoutes)
			if err != nil {
				return fmt.Errorf("failed to register routes for app '%s': %w", appName, err)
			}
			
			log.Printf("App '%s' registered %d routes", appName, len(routes))
			for _, route := range routes {
				log.Printf("  %s /%s%s -> %s:%s", route.Method, appName, route.Path, appName, route.Name)
			}
		}
	}
	
	// Setup static file serving
	app.setupStaticFiles()
	
	return nil
}

// setupTemplates loads templates from all apps
func (app *Application) setupTemplates() error {
	// Load global templates if they exist
	if err := app.templates.LoadGlobalTemplates("templates"); err != nil {
		log.Printf("Warning: failed to load global templates: %v", err)
	}
	
	// Load templates from each app
	for _, appName := range app.registry.GetAppNames() {
		templateDir := filepath.Join("apps", appName, "templates")
		if err := app.templates.LoadAppTemplates(appName, templateDir); err != nil {
			log.Printf("Warning: failed to load templates for app '%s': %v", appName, err)
		}
	}
	
	return nil
}

// addBuiltinRoutes adds framework built-in routes
func (app *Application) addBuiltinRoutes() {
	engine := app.router.GetEngine()
	
	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"app":    app.name,
		})
	})
	
	// Root welcome page
	engine.GET("/", func(c *gin.Context) {
		apps := app.registry.GetAppNames()
		routes := app.router.GetRoutes()
		
		// Try to render template, fall back to basic HTML
		if app.templates.Has("index.html") {
			html, err := app.templates.Render("index.html", gin.H{
				"AppName":    app.name,
				"Apps":       apps,
				"Routes":     routes,
				"RouteCount": len(routes),
			})
			if err == nil {
				c.Header("Content-Type", "text/html")
				c.String(200, html)
				return
			}
		}
		
		// Fallback HTML response
		c.Header("Content-Type", "text/html")
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 2rem; }
        .apps { margin-top: 1rem; }
        .routes { margin-top: 1rem; font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <h1>Welcome to %s</h1>
    <p>Your Gojango application is running successfully!</p>
    <div class="apps">
        <h2>Registered Apps (%d)</h2>
        <ul>`, app.name, app.name, len(apps))
        
		for _, appName := range apps {
			html += fmt.Sprintf("<li><a href=\"/%s/\">%s</a></li>", appName, appName)
		}
		
		html += fmt.Sprintf(`</ul>
    </div>
    <div class="routes">
        <p>%d total routes registered</p>
        <p><a href="/health">Health Check</a></p>
    </div>
</body>
</html>`, len(routes))
		
		c.String(200, html)
	})
}

// setupStaticFiles configures static file serving
func (app *Application) setupStaticFiles() {
	engine := app.router.GetEngine()
	
	// Serve global static files
	engine.Static("/static", "./static")
	
	// Serve app-specific static files
	for _, appName := range app.registry.GetAppNames() {
		staticPath := filepath.Join("apps", appName, "static")
		if _, err := os.Stat(staticPath); err == nil {
			engine.Static("/"+appName+"/static", staticPath)
		}
	}
}

// setupHTTPServer sets up the HTTP server with Gin
func (app *Application) setupHTTPServer() error {
	app.server = &http.Server{
		Addr:    ":" + app.port,
		Handler: app.router,
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