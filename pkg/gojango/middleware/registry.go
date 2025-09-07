package middleware

import (
	"github.com/gin-gonic/gin"
)

// MiddlewareFunc is a function that returns a Gin middleware
type MiddlewareFunc func() gin.HandlerFunc

// Registry holds middleware functions that can be applied to applications
type Registry struct {
	middlewares []MiddlewareFunc
}

// NewRegistry creates a new middleware registry
func NewRegistry() *Registry {
	return &Registry{
		middlewares: make([]MiddlewareFunc, 0),
	}
}

// Add adds a middleware function to the registry
func (r *Registry) Add(middleware MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middleware)
}

// AddGin adds a Gin HandlerFunc directly to the registry
func (r *Registry) AddGin(handler gin.HandlerFunc) {
	r.middlewares = append(r.middlewares, func() gin.HandlerFunc {
		return handler
	})
}

// Apply applies all registered middleware to a Gin engine
func (r *Registry) Apply(engine *gin.Engine) {
	for _, middleware := range r.middlewares {
		engine.Use(middleware())
	}
}

// GetDefaults returns a registry with sensible default middleware
func GetDefaults() *Registry {
	registry := NewRegistry()
	
	// Core middleware in recommended order
	registry.Add(RequestID)
	registry.Add(Logger)
	registry.Add(Recovery)
	registry.Add(SecurityHeaders)
	registry.Add(CORS)
	
	return registry
}

// GetDevelopment returns a registry optimized for development
func GetDevelopment() *Registry {
	registry := NewRegistry()
	
	// Development-friendly middleware
	registry.Add(RequestID)
	registry.Add(Logger)
	registry.Add(Recovery)
	registry.Add(CORSPermissive) // More permissive CORS for development
	
	return registry
}

// Common middleware presets that users can extend

// WithoutCORS returns default middleware without CORS (so users can add their own)
func WithoutCORS() *Registry {
	registry := NewRegistry()
	
	registry.Add(RequestID)
	registry.Add(Logger)
	registry.Add(Recovery)
	registry.Add(SecurityHeaders)
	
	return registry
}

// Minimal returns only essential middleware
func Minimal() *Registry {
	registry := NewRegistry()
	
	registry.Add(RequestID)
	registry.Add(Recovery)
	
	return registry
}

// Count returns the number of middleware functions in the registry
func (r *Registry) Count() int {
	return len(r.middlewares)
}