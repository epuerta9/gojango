package gojango

import (
	"context"
	"testing"
)

func TestApplicationCreation(t *testing.T) {
	// Test default application
	app := New()
	if app == nil {
		t.Fatal("New() should return a valid application")
	}

	if app.name != "gojango-app" {
		t.Errorf("Expected default name 'gojango-app', got '%s'", app.name)
	}

	if app.port != "8080" {
		t.Errorf("Expected default port '8080', got '%s'", app.port)
	}

	if app.debug != false {
		t.Error("Expected debug to be false by default")
	}
}

func TestApplicationOptions(t *testing.T) {
	app := New(
		WithName("test-app"),
		WithPort("9000"),
		WithDebug(true),
	)

	if app.name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", app.name)
	}

	if app.port != "9000" {
		t.Errorf("Expected port '9000', got '%s'", app.port)
	}

	if app.debug != true {
		t.Error("Expected debug to be true")
	}
}

func TestApplicationSettingsRequired(t *testing.T) {
	app := New()

	// Try to initialize without settings
	ctx := context.Background()
	err := app.Initialize(ctx)
	if err == nil {
		t.Error("Expected error when initializing without settings")
	}

	if err.Error() != "settings not loaded - call LoadSettings() first" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestApplicationInitialization(t *testing.T) {
	app := New(WithName("test-app"))
	// Use a fresh registry to avoid conflicts with other tests
	app.registry = &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Create settings
	settings := NewBasicSettings()
	settings.Set("TEST", true)

	// Load settings
	err := app.LoadSettings(settings)
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Register a test app
	testApp := &TestApp{name: "test"}
	app.registry.RegisterApp(testApp)

	// Initialize
	ctx := context.Background()
	err = app.Initialize(ctx)
	if err != nil {
		t.Fatalf("Application initialization failed: %v", err)
	}

	// Check that the test app was initialized
	if !testApp.initialized {
		t.Error("Test app should be initialized")
	}
}

func TestApplicationCommands(t *testing.T) {
	app := New()
	// Use a fresh registry to avoid conflicts with other tests
	app.registry = &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}
	
	settings := NewBasicSettings()
	if err := app.LoadSettings(settings); err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	ctx := context.Background()

	// Test version command
	err := app.RunCommand(ctx, "version", []string{})
	if err != nil {
		t.Errorf("Version command failed: %v", err)
	}

	// Test apps command with fresh app
	app2 := New()
	app2.registry = &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}
	if err := app2.LoadSettings(settings); err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}
	testApp := &TestApp{name: "test"}
	app2.registry.RegisterApp(testApp)
	
	err = app2.RunCommand(ctx, "apps", []string{})
	if err != nil {
		t.Errorf("Apps command failed: %v", err)
	}

	// Test unknown command
	err = app.RunCommand(ctx, "unknown", []string{})
	if err == nil {
		t.Error("Expected error for unknown command")
	}
}

func TestApplicationWithApps(t *testing.T) {
	app := New(WithName("test-with-apps"))
	// Use a fresh registry to avoid conflicts with other tests
	app.registry = &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}
	
	// Create settings
	settings := NewBasicSettings()
	settings.Set("DEBUG", true)
	if err := app.LoadSettings(settings); err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Register multiple apps with dependencies
	coreApp := &TestApp{name: "core"}
	blogApp := &TestApp{name: "blog", deps: []string{"core"}}
	
	app.registry.RegisterApp(coreApp)
	app.registry.RegisterApp(blogApp)

	// Initialize
	ctx := context.Background()
	err := app.Initialize(ctx)
	if err != nil {
		t.Fatalf("Application initialization failed: %v", err)
	}

	// Verify both apps are initialized
	if !coreApp.initialized {
		t.Error("Core app should be initialized")
	}

	if !blogApp.initialized {
		t.Error("Blog app should be initialized")
	}

	// Verify registry has correct state
	apps := app.registry.GetApps()
	if len(apps) != 2 {
		t.Errorf("Expected 2 registered apps, got %d", len(apps))
	}

	if !app.registry.HasApp("core") {
		t.Error("Registry should have core app")
	}

	if !app.registry.HasApp("blog") {
		t.Error("Registry should have blog app")
	}
}

// RoutedTestApp is a test app that implements RouterProvider
type RoutedTestApp struct {
	TestApp
}

func (app *RoutedTestApp) Routes() []Route {
	return []Route{
		{Method: "GET", Path: "/", Name: "index"},
		{Method: "POST", Path: "/create", Name: "create"},
	}
}

// TestAppWithRoutes tests route registration
func TestAppWithRoutes(t *testing.T) {

	// Test route registration
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	routedApp := &RoutedTestApp{TestApp: TestApp{name: "routed"}}
	registry.RegisterApp(routedApp)

	// Check routes were registered
	routes := registry.GetRoutes("routed")
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}

	// Check specific routes
	indexFound := false
	createFound := false
	for _, route := range routes {
		switch route.Name {
		case "index":
			indexFound = true
			if route.Method != "GET" || route.Path != "/" {
				t.Error("Index route has wrong method or path")
			}
		case "create":
			createFound = true
			if route.Method != "POST" || route.Path != "/create" {
				t.Error("Create route has wrong method or path")
			}
		}
	}

	if !indexFound {
		t.Error("Index route not found")
	}
	if !createFound {
		t.Error("Create route not found")
	}
}