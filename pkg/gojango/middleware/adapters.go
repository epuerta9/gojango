package middleware

import (
	"github.com/gin-gonic/gin"
)

// AdapterConfig provides common configuration patterns
type AdapterConfig struct {
	// Skip middleware for certain paths
	SkipPaths []string
}

// WithConfig wraps a middleware with configuration options
func WithConfig(middleware gin.HandlerFunc, config AdapterConfig) gin.HandlerFunc {
	if len(config.SkipPaths) == 0 {
		return middleware
	}
	
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}
	
	return func(c *gin.Context) {
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}
		middleware(c)
	}
}

// Examples of how users can extend the middleware system:

// Example: Rate limiting middleware adapter (users would implement this)
// func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
//     // Implementation would use a rate limiting library like:
//     // - github.com/gin-contrib/timeout
//     // - github.com/ulule/limiter
//     // - Custom implementation
// }

// Example: Authentication middleware adapter  
// func JWT(secret string) gin.HandlerFunc {
//     // Implementation would use JWT library like:
//     // - github.com/golang-jwt/jwt
//     // - github.com/gin-gonic/contrib/jwt
// }

// Example: Metrics middleware adapter
// func Prometheus() gin.HandlerFunc {
//     // Implementation would use:
//     // - github.com/gin-contrib/pprof
//     // - github.com/prometheus/client_golang
// }

// Utility functions for common patterns

// ConditionalMiddleware applies middleware only when condition is true
func ConditionalMiddleware(condition func(*gin.Context) bool, middleware gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if condition(c) {
			middleware(c)
		} else {
			c.Next()
		}
	}
}

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(middlewares ...gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, middleware := range middlewares {
			middleware(c)
			if c.IsAborted() {
				return
			}
		}
	}
}