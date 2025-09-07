// Package gojango provides the core framework interfaces and types.
//
// This package defines the fundamental App interface that all Gojango applications
// must implement, along with supporting types for configuration and context.
package gojango

import (
	"context"
	
	"github.com/gin-gonic/gin"
)

// App is the core interface that all Gojango applications must implement.
// It defines the minimal contract for application lifecycle management.
type App interface {
	// Config returns the application configuration metadata
	Config() AppConfig

	// Initialize is called during application startup to set up the app
	Initialize(ctx *AppContext) error
}

// Route represents an HTTP route definition
type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc // Gin handler function
	Name    string
}

// Optional interfaces that apps can implement for additional functionality

// RouterProvider allows apps to define HTTP routes
type RouterProvider interface {
	Routes() []Route
}

// ModelProvider allows apps to register database models
type ModelProvider interface {
	Models() []interface{}
}

// ServiceProvider allows apps to register gRPC/Connect services
type ServiceProvider interface {
	Services() []Service
}

// AdminProvider allows apps to customize the admin interface
type AdminProvider interface {
	AdminSite() *AdminSite
}

// SignalProvider allows apps to define signal handlers
type SignalProvider interface {
	Signals() []SignalHandler
}

// AppConfig defines application metadata and configuration
type AppConfig struct {
	// Name is the unique identifier for the app
	Name string

	// Label is a human-readable name for the app
	Label string

	// Version is the app version
	Version string

	// Dependencies lists other apps this app depends on
	Dependencies []string

	// Settings contains app-specific settings
	Settings map[string]interface{}
}

// AppContext provides access to shared resources during app initialization
type AppContext struct {
	// Context for cancellation and timeouts
	Context context.Context

	// Name of the app being initialized
	Name string

	// Settings provides access to application configuration
	Settings Settings

	// Registry provides access to the global app registry
	Registry *Registry
}

// BaseApp provides a basic implementation that apps can embed
// to reduce boilerplate code
type BaseApp struct {
	name     string
	settings Settings
}

// Config returns a basic configuration - apps should override this
func (b *BaseApp) Config() AppConfig {
	return AppConfig{
		Name:    b.name,
		Label:   b.name,
		Version: "1.0.0",
	}
}

// Initialize provides a no-op implementation - apps can override if needed
func (b *BaseApp) Initialize(ctx *AppContext) error {
	b.name = ctx.Name
	b.settings = ctx.Settings
	return nil
}


// Service represents a gRPC/Connect service
type Service interface {
	Name() string
}

// AdminSite represents admin interface configuration
type AdminSite struct {
	Name   string
	Models []interface{}
}

// SignalHandler represents a signal handler
type SignalHandler struct {
	Signal  string
	Handler func(data interface{}) error
}

// Settings interface provides access to application configuration
type Settings interface {
	Get(key string, defaultValue ...interface{}) interface{}
	GetString(key string, defaultValue ...string) string
	GetInt(key string, defaultValue ...int) int
	GetBool(key string, defaultValue ...bool) bool
}