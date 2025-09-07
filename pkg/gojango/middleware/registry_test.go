package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("Expected registry to be created")
	}
	
	if len(registry.middlewares) != 0 {
		t.Errorf("Expected empty registry, got: %d middlewares", len(registry.middlewares))
	}
}

func TestRegistryAdd(t *testing.T) {
	registry := NewRegistry()
	
	// Add a middleware function
	registry.Add(RequestID)
	
	if len(registry.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got: %d", len(registry.middlewares))
	}
}

func TestRegistryAddGin(t *testing.T) {
	registry := NewRegistry()
	
	// Add a Gin HandlerFunc directly
	testMiddleware := func(c *gin.Context) {
		c.Header("X-Test", "true")
		c.Next()
	}
	
	registry.AddGin(testMiddleware)
	
	if len(registry.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got: %d", len(registry.middlewares))
	}
}

func TestRegistryApply(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := NewRegistry()
	
	// Add test middleware that sets a header
	testMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Header("X-Test-Applied", "true")
			c.Next()
		}
	}
	
	registry.Add(testMiddleware)
	registry.Apply(engine)
	
	// Add a test route
	engine.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})
	
	// Test that middleware was applied
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)
	
	if w.Header().Get("X-Test-Applied") != "true" {
		t.Error("Expected middleware to be applied")
	}
}

func TestGetDefaults(t *testing.T) {
	registry := GetDefaults()
	if registry == nil {
		t.Fatal("Expected default registry to be created")
	}
	
	// Should have 5 default middleware: RequestID, Logger, Recovery, SecurityHeaders, CORS
	if len(registry.middlewares) != 5 {
		t.Errorf("Expected 5 default middlewares, got: %d", len(registry.middlewares))
	}
}

func TestGetDevelopment(t *testing.T) {
	registry := GetDevelopment()
	if registry == nil {
		t.Fatal("Expected development registry to be created")
	}
	
	// Should have 4 development middleware: RequestID, Logger, Recovery, CORSPermissive
	if len(registry.middlewares) != 4 {
		t.Errorf("Expected 4 development middlewares, got: %d", len(registry.middlewares))
	}
}

func TestWithoutCORS(t *testing.T) {
	registry := WithoutCORS()
	if registry == nil {
		t.Fatal("Expected registry without CORS to be created")
	}
	
	// Should have 4 middleware without CORS: RequestID, Logger, Recovery, SecurityHeaders
	if len(registry.middlewares) != 4 {
		t.Errorf("Expected 4 middlewares without CORS, got: %d", len(registry.middlewares))
	}
}

func TestMinimal(t *testing.T) {
	registry := Minimal()
	if registry == nil {
		t.Fatal("Expected minimal registry to be created")
	}
	
	// Should have 2 minimal middleware: RequestID, Recovery
	if len(registry.middlewares) != 2 {
		t.Errorf("Expected 2 minimal middlewares, got: %d", len(registry.middlewares))
	}
}

func TestRegistryExtensibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	
	// Start with minimal registry
	registry := Minimal()
	
	// Add custom middleware
	customMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Header("X-Custom", "added")
			c.Next()
		}
	}
	
	registry.Add(customMiddleware)
	
	// Add another middleware directly as Gin HandlerFunc
	registry.AddGin(func(c *gin.Context) {
		c.Header("X-Direct", "added")
		c.Next()
	})
	
	// Should now have 4 middleware total
	if len(registry.middlewares) != 4 {
		t.Errorf("Expected 4 middlewares after additions, got: %d", len(registry.middlewares))
	}
	
	// Apply to engine
	registry.Apply(engine)
	
	engine.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})
	
	// Test that custom middleware were applied
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)
	
	if w.Header().Get("X-Custom") != "added" {
		t.Error("Expected custom middleware to be applied")
	}
	
	if w.Header().Get("X-Direct") != "added" {
		t.Error("Expected direct Gin middleware to be applied")
	}
	
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected RequestID middleware from minimal registry")
	}
}