package gojango

import (
	"context"
	"testing"
)

// TestApp is a simple test app implementation
type TestApp struct {
	BaseApp
	name string
	deps []string
	initialized bool
}

func (app *TestApp) Config() AppConfig {
	return AppConfig{
		Name:         app.name,
		Label:        app.name + " Test App",
		Version:      "1.0.0",
		Dependencies: app.deps,
	}
}

func (app *TestApp) Initialize(ctx *AppContext) error {
	app.initialized = true
	return app.BaseApp.Initialize(ctx)
}

func TestRegistryBasics(t *testing.T) {
	// Get a fresh registry for testing
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Test empty registry
	if len(registry.GetApps()) != 0 {
		t.Error("New registry should be empty")
	}

	if registry.HasApp("nonexistent") {
		t.Error("Empty registry should not have any apps")
	}

	// Test app registration
	app1 := &TestApp{name: "app1"}
	registry.RegisterApp(app1)

	if !registry.HasApp("app1") {
		t.Error("App should be registered")
	}

	if len(registry.GetApps()) != 1 {
		t.Error("Registry should have one app")
	}

	// Test duplicate registration
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for duplicate app registration")
		}
	}()
	registry.RegisterApp(&TestApp{name: "app1"}) // Should panic
}

func TestRegistryDependencies(t *testing.T) {
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Register apps in order: core -> blog -> admin
	core := &TestApp{name: "core"}
	blog := &TestApp{name: "blog", deps: []string{"core"}}
	admin := &TestApp{name: "admin", deps: []string{"core", "blog"}}

	registry.RegisterApp(core)
	registry.RegisterApp(blog)
	registry.RegisterApp(admin)

	// Test topological sort
	sorted, err := registry.topologicalSort()
	if err != nil {
		t.Fatalf("Dependency resolution failed: %v", err)
	}

	// Core should come first
	if sorted[0] != "core" {
		t.Errorf("Expected 'core' first, got '%s'", sorted[0])
	}

	// Blog should come before admin
	blogIdx, adminIdx := -1, -1
	for i, name := range sorted {
		if name == "blog" {
			blogIdx = i
		}
		if name == "admin" {
			adminIdx = i
		}
	}

	if blogIdx == -1 || adminIdx == -1 {
		t.Error("Could not find blog or admin in sorted order")
	}

	if blogIdx > adminIdx {
		t.Error("Blog should come before admin due to dependencies")
	}
}

func TestRegistryCircularDependency(t *testing.T) {
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Create circular dependency: a -> b -> c -> a
	appA := &TestApp{name: "a", deps: []string{"c"}}
	appB := &TestApp{name: "b", deps: []string{"a"}}
	appC := &TestApp{name: "c", deps: []string{"b"}}

	registry.RegisterApp(appA)
	registry.RegisterApp(appB)
	registry.RegisterApp(appC)

	// Should detect circular dependency
	_, err := registry.topologicalSort()
	if err == nil {
		t.Error("Expected error for circular dependency")
	}
}

func TestRegistryMissingDependency(t *testing.T) {
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Register app with missing dependency
	app := &TestApp{name: "app", deps: []string{"missing"}}
	registry.RegisterApp(app)

	// Should fail during topological sort
	_, err := registry.topologicalSort()
	if err == nil {
		t.Error("Expected error for missing dependency during topological sort")
	}
}

func TestRegistryInitialization(t *testing.T) {
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Register test apps
	app1 := &TestApp{name: "app1"}
	app2 := &TestApp{name: "app2", deps: []string{"app1"}}

	registry.RegisterApp(app1)
	registry.RegisterApp(app2)

	// Create basic settings
	settings := NewBasicSettings()
	settings.Set("TEST", true)

	// Initialize registry
	ctx := context.Background()
	err := registry.Initialize(ctx, settings)
	if err != nil {
		t.Fatalf("Registry initialization failed: %v", err)
	}

	// Check that apps were initialized
	if !app1.initialized {
		t.Error("App1 should be initialized")
	}

	if !app2.initialized {
		t.Error("App2 should be initialized")
	}

	// Test reinitialization (should be allowed)
	err = registry.Initialize(ctx, settings)
	if err != nil {
		t.Errorf("Reinitialization should be allowed, got error: %v", err)
	}
}

func TestRegistryHooks(t *testing.T) {
	registry := &Registry{
		apps:     make(map[string]App),
		models:   make(map[string]ModelMeta),
		routes:   make(map[string][]Route),
		services: make(map[string]Service),
	}

	// Track hook execution
	var preInitCalled, postInitCalled bool

	registry.AddPreInitHook(func() error {
		preInitCalled = true
		return nil
	})

	registry.AddPostInitHook(func() error {
		postInitCalled = true
		return nil
	})

	// Register a test app
	app := &TestApp{name: "test"}
	registry.RegisterApp(app)

	// Initialize
	settings := NewBasicSettings()
	ctx := context.Background()
	err := registry.Initialize(ctx, settings)
	if err != nil {
		t.Fatalf("Registry initialization failed: %v", err)
	}

	// Check hooks were called
	if !preInitCalled {
		t.Error("Pre-init hook was not called")
	}

	if !postInitCalled {
		t.Error("Post-init hook was not called")
	}
}