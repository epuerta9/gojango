package routing

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Router manages URL routing and reversal
type Router struct {
	engine *gin.Engine
	routes map[string]*RegisteredRoute
}

// Route represents a URL route configuration (matches gojango.Route)
type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
	Name    string
}

// RegisteredRoute contains a route and its metadata
type RegisteredRoute struct {
	Route
	AppName  string
	FullName string // app:name format
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	gin.SetMode(gin.ReleaseMode) // Start in release mode, can be overridden
	engine := gin.New()
	
	return &Router{
		engine: engine,
		routes: make(map[string]*RegisteredRoute),
	}
}

// GetEngine returns the underlying Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// RegisterRoutes registers routes for an app
func (r *Router) RegisterRoutes(appName string, routes []Route) error {
	// Create app route group
	group := r.engine.Group("/" + appName)
	
	for _, route := range routes {
		// Create full route name: app:name
		fullName := fmt.Sprintf("%s:%s", appName, route.Name)
		
		// Check for duplicate route names
		if _, exists := r.routes[fullName]; exists {
			return fmt.Errorf("route '%s' already exists", fullName)
		}
		
		// Register the route
		registeredRoute := &RegisteredRoute{
			Route:    route,
			AppName:  appName,
			FullName: fullName,
		}
		r.routes[fullName] = registeredRoute
		
		// Register with Gin engine
		switch strings.ToUpper(route.Method) {
		case "GET":
			group.GET(route.Path, route.Handler)
		case "POST":
			group.POST(route.Path, route.Handler)
		case "PUT":
			group.PUT(route.Path, route.Handler)
		case "DELETE":
			group.DELETE(route.Path, route.Handler)
		case "PATCH":
			group.PATCH(route.Path, route.Handler)
		default:
			return fmt.Errorf("unsupported HTTP method: %s", route.Method)
		}
	}
	
	return nil
}

// Reverse performs URL reversal - converts route name to URL
func (r *Router) Reverse(routeName string, params ...interface{}) string {
	route, exists := r.routes[routeName]
	if !exists {
		// Return empty string for missing routes (fail gracefully in templates)
		return ""
	}
	
	// Build the full path
	path := "/" + route.AppName + route.Path
	
	// Simple parameter substitution (basic implementation)
	// In a full implementation, this would handle URL parameters properly
	if len(params) > 0 {
		// Replace path parameters with provided values
		for i, param := range params {
			placeholder := fmt.Sprintf("{%d}", i)
			if paramStr, ok := param.(string); ok {
				path = strings.ReplaceAll(path, placeholder, paramStr)
			}
		}
	}
	
	return path
}

// GetRoutes returns all registered routes
func (r *Router) GetRoutes() map[string]*RegisteredRoute {
	routes := make(map[string]*RegisteredRoute)
	for name, route := range r.routes {
		routes[name] = route
	}
	return routes
}

// TemplateFuncs returns template functions for URL reversal and static files
func (r *Router) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"url": r.Reverse,
		"static": func(path string) string {
			// Static file URL generation
			return "/static/" + strings.TrimPrefix(path, "/")
		},
	}
}

// Use adds middleware to the router
func (r *Router) Use(middleware ...gin.HandlerFunc) {
	r.engine.Use(middleware...)
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}