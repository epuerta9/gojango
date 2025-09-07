package main

import (
	"context"
	"log"
	"time"

	"github.com/epuerta9/gojango/pkg/gojango"
	"github.com/epuerta9/gojango/pkg/gojango/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Example custom middleware
func CustomAPIVersionMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("API-Version", version)
		c.Header("X-Custom-Framework", "Gojango")
		c.Next()
	}
}

func main() {
	// Example 1: Using preset middleware configurations
	log.Println("=== Example 1: Preset Configurations ===")
	
	// Start with minimal middleware
	minimalRegistry := middleware.Minimal()
	minimalApp := gojango.New(
		gojango.WithName("minimal-app"),
		gojango.WithMiddleware(minimalRegistry),
	)
	log.Printf("Minimal app has %d middleware functions", minimalRegistry.Count())

	// Development preset
	devRegistry := middleware.GetDevelopment()
	devApp := gojango.New(
		gojango.WithName("dev-app"),
		gojango.WithMiddleware(devRegistry),
	)
	log.Printf("Development app has %d middleware functions", devRegistry.Count())

	// Example 2: Building custom middleware stack
	log.Println("\n=== Example 2: Custom Middleware Stack ===")
	
	customRegistry := middleware.NewRegistry()
	
	// Add built-in middleware
	customRegistry.Add(middleware.RequestID)
	customRegistry.Add(middleware.Logger)
	customRegistry.Add(middleware.Recovery)
	
	// Add custom CORS configuration
	customRegistry.AddGin(cors.New(cors.Config{
		AllowOrigins:     []string{"https://myapp.com", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	// Add custom middleware
	customRegistry.AddGin(CustomAPIVersionMiddleware("v1.0"))
	
	customApp := gojango.New(
		gojango.WithName("custom-app"),
		gojango.WithMiddleware(customRegistry),
	)
	
	// Example 3: Adding middleware after app creation
	log.Println("\n=== Example 3: Adding Middleware After Creation ===")
	
	app := gojango.New(
		gojango.WithName("extensible-app"),
		gojango.WithDebug(true),
	)
	
	// Add a custom middleware function
	app.AddMiddleware(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			start := time.Now()
			c.Next()
			duration := time.Since(start)
			log.Printf("Request to %s took %s", c.Request.URL.Path, duration)
		}
	})
	
	// Add Gin middleware directly
	app.AddGinMiddleware(func(c *gin.Context) {
		c.Header("X-Processing-Time", time.Now().Format(time.RFC3339))
		c.Next()
	})

	// Load settings for all apps
	settings := gojango.NewBasicSettings()
	settings.Set("PROJECT_NAME", "middleware-example")
	
	for _, exampleApp := range []*gojango.Application{minimalApp, devApp, customApp, app} {
		if err := exampleApp.LoadSettings(settings); err != nil {
			log.Fatalf("Failed to load settings: %v", err)
		}
		
		if err := exampleApp.Initialize(context.Background()); err != nil {
			log.Fatalf("Failed to initialize app: %v", err)
		}
	}
	
	log.Println("\n=== All Examples Initialized Successfully! ===")
	log.Println("This demonstrates the flexibility of Gojango's middleware system:")
	log.Println("- Use sensible defaults")
	log.Println("- Customize with presets") 
	log.Println("- Build completely custom stacks")
	log.Println("- Extend after app creation")
	log.Println("- Leverage the rich Gin ecosystem")
}