# Middleware in Gojango

Gojango provides a flexible and extensible middleware system built on top of Gin's middleware architecture. You can use our sensible defaults, customize them, or build your own middleware stack.

## Default Middleware Stack

Gojango comes with a production-ready middleware stack:

- **RequestID**: Adds unique request IDs for tracing
- **Logger**: Structured request/response logging  
- **Recovery**: Graceful panic recovery
- **SecurityHeaders**: Basic security headers (X-Frame-Options, etc.)
- **CORS**: Cross-origin resource sharing (using gin-contrib/cors)

## Using Default Middleware

```go
// Default production middleware
app := gojango.New(
    gojango.WithName("myapp"),
)

// Development middleware (more permissive CORS)
app := gojango.New(
    gojango.WithName("myapp"),
    gojango.WithDebug(true),
)
```

## Customizing Middleware

### Option 1: Use Preset Configurations

```go
import "github.com/epuerta9/gojango/pkg/gojango/middleware"

// Start with minimal middleware (RequestID + Recovery only)
app := gojango.New(
    gojango.WithMiddleware(middleware.Minimal()),
)

// Use defaults without CORS (add your own CORS config)
app := gojango.New(
    gojango.WithMiddleware(middleware.WithoutCORS()),
)

// Use development preset
app := gojango.New(
    gojango.WithMiddleware(middleware.GetDevelopment()),
)
```

### Option 2: Build Custom Middleware Stack

```go
import (
    "github.com/epuerta9/gojango/pkg/gojango/middleware"
    "github.com/gin-contrib/cors"
    "time"
)

// Create custom middleware registry
registry := middleware.NewRegistry()

// Add built-in middleware
registry.Add(middleware.RequestID)
registry.Add(middleware.Logger)

// Add custom CORS configuration
registry.AddGin(cors.New(cors.Config{
    AllowOrigins:     []string{"https://myapp.com", "https://admin.myapp.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))

app := gojango.New(
    gojango.WithMiddleware(registry),
)
```

### Option 3: Add Middleware After Creation

```go
app := gojango.New()

// Add your own middleware functions
app.AddMiddleware(func() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Custom middleware logic
        c.Header("X-Custom", "value")
        c.Next()
    }
})

// Add Gin middleware directly
app.AddGinMiddleware(gin.BasicAuth(gin.Accounts{
    "admin": "secret",
}))
```

## Popular Gin Middleware Integration

Gojango encourages using the rich Gin ecosystem:

### Authentication
```go
import "github.com/gin-gonic/contrib/jwt"

registry := middleware.Minimal()
registry.AddGin(jwt.Auth("secret"))
```

### Rate Limiting  
```go
import "github.com/ulule/limiter/v3/drivers/middleware/gin"

registry := middleware.GetDefaults()
registry.AddGin(ginlimiter.NewMiddleware(limiter))
```

### Prometheus Metrics
```go
import "github.com/gin-contrib/pprof"

registry := middleware.GetDefaults()
registry.AddGin(pprof.Register)
```

### Request Compression
```go
import "github.com/gin-contrib/gzip"

registry := middleware.GetDefaults()
registry.AddGin(gzip.Gzip(gzip.DefaultCompression))
```

## Custom Middleware Examples

### Simple Header Middleware
```go
func APIVersionMiddleware(version string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("API-Version", version)
        c.Next()
    }
}

// Usage
app.AddGinMiddleware(APIVersionMiddleware("v1.0"))
```

### Request Validation Middleware
```go
func ValidateAPIKey() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "API key required"})
            c.Abort()
            return
        }
        
        // Validate API key logic here
        c.Next()
    }
}
```

### Conditional Middleware
```go
import "github.com/epuerta9/gojango/pkg/gojango/middleware"

// Only apply auth to certain paths
authMiddleware := middleware.ConditionalMiddleware(
    func(c *gin.Context) bool {
        return strings.HasPrefix(c.Request.URL.Path, "/admin")
    },
    ValidateAPIKey(),
)

app.AddGinMiddleware(authMiddleware)
```

## Built-in CORS Options

### Default CORS (Production)
```go
// Allows all origins, common headers
middleware.CORS()
```

### Permissive CORS (Development)
```go
// Very permissive for development
middleware.CORSPermissive() 
```

### Custom CORS
```go
import "github.com/gin-contrib/cors"

corsConfig := cors.Config{
    AllowOrigins:     []string{"https://myapp.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Origin", "Content-Type"},
    AllowCredentials: false,
    MaxAge:           12 * time.Hour,
}

app.AddGinMiddleware(cors.New(corsConfig))
```

## Middleware Order

Middleware order matters! Gojango applies middleware in this recommended order:

1. **RequestID** - First, to track requests
2. **Logger** - Early, to log everything  
3. **Recovery** - Early, to catch panics
4. **SecurityHeaders** - Before business logic
5. **CORS** - Before route handlers
6. **Custom Auth/Rate Limiting** - Before business logic
7. **Route Handlers**

## Best Practices

1. **Use Gin ecosystem** - Don't reinvent the wheel, use proven Gin middleware
2. **Start minimal** - Use `middleware.Minimal()` and add what you need
3. **Order matters** - Put security and logging middleware early
4. **Environment-specific** - Use different middleware for dev/prod
5. **Make it configurable** - Let users customize your middleware choices

## Philosophy

Gojango provides conventions and sensible defaults while staying out of your way. We use the Gin ecosystem instead of creating custom middleware, making it easy to extend and familiar to Gin users.