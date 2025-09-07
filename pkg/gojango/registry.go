package gojango

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Registry manages all registered apps in the Gojango application.
// It handles app registration, dependency resolution, and initialization.
type Registry struct {
	mu       sync.RWMutex
	apps     map[string]App
	order    []string              // Registration order for dependency resolution
	models   map[string]ModelMeta  // All models across apps
	routes   map[string][]Route    // Routes grouped by app
	services map[string]Service    // gRPC/Connect services
	
	// Lifecycle hooks
	preInit  []func() error
	postInit []func() error
	
	initialized bool
}

// ModelMeta contains metadata about registered models
type ModelMeta struct {
	App       string
	Name      string
	FullName  string // app.model format
	TableName string
	Type      interface{}
}

// Global registry instance
var (
	globalRegistry *Registry
	registryOnce   sync.Once
)

// GetRegistry returns the global registry instance
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		globalRegistry = &Registry{
			apps:     make(map[string]App),
			models:   make(map[string]ModelMeta),
			routes:   make(map[string][]Route),
			services: make(map[string]Service),
		}
	})
	return globalRegistry
}

// Register adds an app to the global registry.
// This is typically called from app init() functions.
func Register(app App) {
	registry := GetRegistry()
	registry.RegisterApp(app)
}

// RegisterApp adds an app to the registry
func (r *Registry) RegisterApp(app App) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	config := app.Config()
	
	// Validate app name
	if config.Name == "" {
		panic("app name cannot be empty")
	}
	
	// Check for duplicate registration
	if _, exists := r.apps[config.Name]; exists {
		panic(fmt.Sprintf("app '%s' is already registered", config.Name))
	}
	
	// Store the app
	r.apps[config.Name] = app
	r.order = append(r.order, config.Name)
	
	// Register models if app provides them
	if provider, ok := app.(ModelProvider); ok {
		for _, model := range provider.Models() {
			r.registerModel(config.Name, model)
		}
	}
	
	// Register routes if app provides them
	if provider, ok := app.(RouterProvider); ok {
		r.routes[config.Name] = provider.Routes()
	}
	
	// Register services if app provides them
	if provider, ok := app.(ServiceProvider); ok {
		for _, service := range provider.Services() {
			r.services[service.Name()] = service
		}
	}
}

// HasApp checks if an app with the given name is registered
func (r *Registry) HasApp(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.apps[name]
	return exists
}

// GetApp returns the app with the given name
func (r *Registry) GetApp(name string) (App, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	app, exists := r.apps[name]
	return app, exists
}

// GetApps returns all registered apps
func (r *Registry) GetApps() map[string]App {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	apps := make(map[string]App, len(r.apps))
	for name, app := range r.apps {
		apps[name] = app
	}
	return apps
}

// GetRoutes returns all routes for a given app
func (r *Registry) GetRoutes(appName string) []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.routes[appName]
}

// GetAllRoutes returns routes from all apps
func (r *Registry) GetAllRoutes() map[string][]Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	routes := make(map[string][]Route, len(r.routes))
	for app, appRoutes := range r.routes {
		routes[app] = make([]Route, len(appRoutes))
		copy(routes[app], appRoutes)
	}
	return routes
}

// Initialize initializes all registered apps in dependency order
func (r *Registry) Initialize(ctx context.Context, settings Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.initialized {
		// Allow reinitialization by resetting the flag
		r.initialized = false
	}
	
	// Run pre-init hooks
	for _, hook := range r.preInit {
		if err := hook(); err != nil {
			return fmt.Errorf("pre-init hook failed: %w", err)
		}
	}
	
	// Sort apps by dependencies
	sorted, err := r.topologicalSort()
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}
	
	// Initialize apps in dependency order
	for _, appName := range sorted {
		app := r.apps[appName]
		
		// Create app context
		appCtx := &AppContext{
			Context:  ctx,
			Name:     appName,
			Settings: settings,
			Registry: r,
		}
		
		// Initialize the app
		if err := app.Initialize(appCtx); err != nil {
			return fmt.Errorf("failed to initialize app '%s': %w", appName, err)
		}
	}
	
	// Run post-init hooks
	for _, hook := range r.postInit {
		if err := hook(); err != nil {
			return fmt.Errorf("post-init hook failed: %w", err)
		}
	}
	
	r.initialized = true
	return nil
}

// AddPreInitHook adds a hook that runs before app initialization
func (r *Registry) AddPreInitHook(hook func() error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preInit = append(r.preInit, hook)
}

// AddPostInitHook adds a hook that runs after app initialization
func (r *Registry) AddPostInitHook(hook func() error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.postInit = append(r.postInit, hook)
}

// registerModel adds a model to the registry
func (r *Registry) registerModel(appName string, model interface{}) {
	// For now, we'll do basic model registration
	// In the future, this will extract metadata from Ent schemas
	modelName := fmt.Sprintf("%T", model)
	
	meta := ModelMeta{
		App:       appName,
		Name:      modelName,
		FullName:  fmt.Sprintf("%s.%s", appName, modelName),
		TableName: toSnakeCase(fmt.Sprintf("%s_%s", appName, modelName)),
		Type:      model,
	}
	
	r.models[meta.FullName] = meta
}

// topologicalSort sorts apps by their dependencies
func (r *Registry) topologicalSort() ([]string, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	
	// Initialize all apps with zero in-degree
	for appName := range r.apps {
		inDegree[appName] = 0
		graph[appName] = []string{}
	}
	
	// Build the graph
	for appName, app := range r.apps {
		config := app.Config()
		for _, dep := range config.Dependencies {
			// Check if dependency exists
			if _, exists := r.apps[dep]; !exists {
				return nil, fmt.Errorf("app '%s' depends on '%s' which is not registered", appName, dep)
			}
			
			// Add edge: dep -> appName
			graph[dep] = append(graph[dep], appName)
			inDegree[appName]++
		}
	}
	
	// Kahn's algorithm for topological sorting
	var queue []string
	for appName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, appName)
		}
	}
	
	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	
	// Check for circular dependencies
	if len(result) != len(r.apps) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	
	return result, nil
}

// GetAppNames returns all registered app names sorted alphabetically
func (r *Registry) GetAppNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.apps))
	for name := range r.apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(input string) string {
	if len(input) == 0 {
		return ""
	}
	
	var result []rune
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r-'A'+'a')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}