// Example showing the gj alias package for cleaner imports
//
// Instead of:
//   import "github.com/epuerta9/gojango/pkg/gojango"
//   app := gojango.New(gojango.WithName("app"))
//
// You can use:
//   import "github.com/epuerta9/gojango"
//   app := gj.New(gj.WithName("app"))
//
// This provides a Django-like experience where the main package
// can be imported with a short alias.
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	gj "github.com/epuerta9/gojango"
)

// ExampleSettings demonstrates custom settings with database support
type ExampleSettings struct {
	*gj.BasicSettings
}

func NewExampleSettings() *ExampleSettings {
	s := &ExampleSettings{
		BasicSettings: gj.NewBasicSettings(),
	}
	
	// Configure defaults
	s.Set("DEBUG", true)
	s.Set("SECRET_KEY", "example-secret-key")
	s.Set("DATABASE_DRIVER", "sqlite3")
	s.Set("DATABASE_NAME", "example.db")
	
	return s
}

func (s *ExampleSettings) GetDatabaseConfig() *gj.Config {
	driver := s.GetString("DATABASE_DRIVER", "sqlite3")
	
	switch driver {
	case "postgres":
		return gj.PostgresConfig(
			s.GetString("DATABASE_HOST", "localhost"),
			s.GetString("DATABASE_NAME", "example"),
			s.GetString("DATABASE_USER", "user"),
			s.GetString("DATABASE_PASSWORD", ""),
		)
	default:
		return gj.SQLiteConfig(s.GetString("DATABASE_NAME", "example.db"))
	}
}

// ExampleApp demonstrates an app using the gj alias
type ExampleApp struct{}

func (a *ExampleApp) Config() gj.AppConfig {
	return gj.AppConfig{
		Name:    "example",
		Label:   "Example App with GJ Alias",
		Version: "1.0.0",
	}
}

func (a *ExampleApp) Initialize(ctx *gj.AppContext) error {
	log.Println("Initializing Example app using gj alias...")
	
	// Demonstrate database setup with gj alias
	if exampleSettings, ok := ctx.Settings.(*ExampleSettings); ok {
		dbConfig := exampleSettings.GetDatabaseConfig()
		log.Printf("Database: driver=%s, database=%s", dbConfig.Driver, dbConfig.Database)
		
		// Setup database connection
		conn, err := gj.Open(dbConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		// Setup migrator
		migrator := gj.NewMigrator(conn, "migrations")
		if err := migrator.Initialize(ctx.Context); err != nil {
			log.Printf("Warning: failed to initialize migrator: %v", err)
		} else {
			log.Println("Database migrator initialized successfully using gj alias")
		}
	}
	
	return nil
}

func (a *ExampleApp) GetRoutes() []gj.AppRoute {
	return []gj.AppRoute{
		{
			Method:  "GET",
			Path:    "/",
			Handler: a.homeHandler,
			Name:    "home",
		},
		{
			Method:  "GET",
			Path:    "/about",
			Handler: a.aboutHandler,
			Name:    "about",
		},
		{
			Method:  "GET",
			Path:    "/api/status",
			Handler: a.statusHandler,
			Name:    "api-status",
		},
	}
}

func (a *ExampleApp) homeHandler(c *gin.Context) {
	c.HTML(200, "example/home.html", gin.H{
		"title":   "GJ Alias Example",
		"message": "This app was built using the gj alias package!",
		"features": []string{
			"Cleaner imports: gj instead of gojango",
			"All functionality available through gj.*",
			"Django-like import experience",
			"Full database integration",
			"Complete middleware stack",
		},
	})
}

func (a *ExampleApp) aboutHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"app":        "GJ Alias Example", 
		"framework":  "Gojango",
		"version":    "0.2.0",
		"alias":      "gj",
		"description": "Demonstrating the gj alias package for cleaner Gojango imports",
		"database":   "Integrated with database layer",
		"middleware": "Full middleware stack",
	})
}

func (a *ExampleApp) statusHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"using_alias": "gj",
		"timestamp": "2024-01-01T00:00:00Z",
		"components": gin.H{
			"database":   "available",
			"migrations": "ready", 
			"middleware": "active",
			"routing":    "functional",
			"templates":  "loaded",
		},
	})
}

func main() {
	ctx := context.Background()

	log.Println("Starting GJ Alias Example")
	log.Println("This demonstrates using 'gj' as an alias for cleaner imports")

	// Create settings using gj alias
	settings := NewExampleSettings()
	settings.LoadFromEnv()

	// Create application using gj alias
	app := gj.New(
		gj.WithName("gj-example"),
		gj.WithDebug(true),
		gj.WithPort("8080"),
	)

	// Load settings
	if err := app.LoadSettings(settings); err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}

	// Register app using gj alias
	exampleApp := &ExampleApp{}
	registry := gj.GetRegistry()
	registry.RegisterApp(exampleApp)

	log.Println("Available endpoints:")
	log.Println("  GET  / - Home page")
	log.Println("  GET  /example/about - About endpoint")
	log.Println("  GET  /example/api/status - Status endpoint")
	log.Println("  GET  /health - Framework health check")

	// Run application
	if err := app.Run(ctx); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}