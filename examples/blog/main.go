// Example Gojango application demonstrating database integration with Ent ORM.
//
// This example shows how to:
//   - Configure database connections
//   - Set up Ent client integration  
//   - Use migrations for schema management
//   - Implement basic CRUD operations
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/epuerta9/gojango/pkg/gojango"
	"github.com/epuerta9/gojango/pkg/gojango/db"
	"github.com/gin-gonic/gin"
)

// BlogSettings implements the Gojango Settings interface
type BlogSettings struct {
	*gojango.BasicSettings
}

// NewBlogSettings creates a new blog settings instance
func NewBlogSettings() *BlogSettings {
	s := &BlogSettings{
		BasicSettings: gojango.NewBasicSettings(),
	}
	
	// Set defaults
	s.Set("DEBUG", true)
	s.Set("SECRET_KEY", "your-secret-key-here-in-production-use-env-var")
	s.Set("DATABASE_DRIVER", "sqlite3")
	s.Set("DATABASE_NAME", "blog.db")
	
	return s
}

// GetDatabaseConfig returns the database configuration
func (s *BlogSettings) GetDatabaseConfig() *db.Config {
	driver := s.GetString("DATABASE_DRIVER", "sqlite3")
	
	switch driver {
	case "postgres":
		return db.PostgresConfig(
			s.GetString("DATABASE_HOST", "localhost"),
			s.GetString("DATABASE_NAME", "blog"),
			s.GetString("DATABASE_USER", "user"),
			s.GetString("DATABASE_PASSWORD", ""),
		)
	default:
		return db.SQLiteConfig(s.GetString("DATABASE_NAME", "blog.db"))
	}
}

func main() {
	ctx := context.Background()

	// Create application settings
	settings := NewBlogSettings()
	settings.LoadFromEnv() // Load any environment variables

	// Create Gojango application with database support
	app := gojango.New(
		gojango.WithName("blog-example"),
		gojango.WithDebug(true),
		gojango.WithPort("8080"),
	)

	// Load settings
	if err := app.LoadSettings(settings); err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}

	// Register our blog app (using the registry directly)
	blogApp := &BlogApp{}
	registry := gojango.GetRegistry()
	registry.RegisterApp(blogApp)

	// Initialize and run
	log.Println("Starting Blog Example with Database Integration...")
	log.Println("This demonstrates Gojango's database layer with Ent ORM")
	log.Println("Available endpoints:")
	log.Println("  GET  / - Welcome page with app overview")
	log.Println("  GET  /health - Health check")
	log.Println("  GET  /blog/ - Blog posts (when implemented)")
	log.Println("  GET  /blog/posts - List all posts (when implemented)")

	if err := app.Run(ctx); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

// BlogApp represents our blog application
type BlogApp struct{}

// Config returns the app configuration
func (a *BlogApp) Config() gojango.AppConfig {
	return gojango.AppConfig{
		Name:    "blog",
		Label:   "Blog Application",
		Version: "1.0.0",
	}
}

// Initialize sets up the blog app
func (a *BlogApp) Initialize(appCtx *gojango.AppContext) error {
	log.Println("Initializing Blog app with database integration...")

	// In a real application, you would:
	// 1. Set up your Ent client here
	// 2. Run migrations
	// 3. Set up any required database seeding
	
	// Get database config (need to cast to our settings type)  
	if blogSettings, ok := appCtx.Settings.(*BlogSettings); ok {
		dbConfig := blogSettings.GetDatabaseConfig()
		log.Printf("Database configured: driver=%s, database=%s", dbConfig.Driver, dbConfig.Database)
		
		// Example of setting up database connection
		conn, err := db.Open(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		// Example of setting up migrator
		migrator := db.NewMigrator(conn, "migrations")
		if err := migrator.Initialize(appCtx.Context); err != nil {
			log.Printf("Warning: failed to initialize migrator: %v", err)
		} else {
			log.Println("Database migrator initialized successfully")
		}

		// In a real app, you would create your Ent client here:
		// driver, err := entsql.Open(string(dbConfig.Driver), dsn)
		// client := ent.NewClient(ent.Driver(driver))
	}

	return nil
}

// GetRoutes returns the blog app routes
func (a *BlogApp) GetRoutes() []gojango.Route {
	return []gojango.Route{
		{
			Method:  "GET",
			Path:    "/",
			Handler: a.indexHandler,
			Name:    "index",
		},
		{
			Method:  "GET", 
			Path:    "/posts",
			Handler: a.postsHandler,
			Name:    "posts",
		},
		{
			Method:  "GET",
			Path:    "/posts/:id",
			Handler: a.postDetailHandler,
			Name:    "post-detail",
		},
	}
}

// indexHandler handles the blog index page
func (a *BlogApp) indexHandler(c *gin.Context) {
	c.HTML(200, "blog/index.html", gin.H{
		"title": "Blog Home",
		"message": "Welcome to the Gojango Blog Example!",
		"features": []string{
			"Database integration with multiple drivers",
			"Ent ORM support", 
			"Migration management",
			"Connection pooling",
			"Transaction support",
		},
	})
}

// postsHandler handles listing blog posts
func (a *BlogApp) postsHandler(c *gin.Context) {
	// In a real application, you would query the database here
	// Example:
	// posts, err := blogClient.Post.Query().All(ctx)
	
	posts := []map[string]interface{}{
		{
			"id": 1,
			"title": "Getting Started with Gojango Database Layer",
			"content": "This post demonstrates how to use Gojango's database integration...",
			"created_at": "2024-01-01T00:00:00Z",
		},
		{
			"id": 2, 
			"title": "Advanced Migration Techniques",
			"content": "Learn how to manage complex database migrations...",
			"created_at": "2024-01-02T00:00:00Z",
		},
	}

	c.HTML(200, "blog/posts.html", gin.H{
		"title": "Blog Posts",
		"posts": posts,
	})
}

// postDetailHandler handles individual post details
func (a *BlogApp) postDetailHandler(c *gin.Context) {
	id := c.Param("id")
	
	// In a real application, you would query by ID:
	// post, err := blogClient.Post.Get(ctx, id)
	
	post := map[string]interface{}{
		"id": id,
		"title": "Sample Blog Post",
		"content": "This is a sample blog post demonstrating database integration with Gojango. In a real application, this would be loaded from your database using Ent ORM.",
		"created_at": "2024-01-01T00:00:00Z",
	}

	c.HTML(200, "blog/post_detail.html", gin.H{
		"title": "Post Detail",
		"post": post,
	})
}