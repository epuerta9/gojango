package templates

import (
	"html/template"
	"os"
	"path/filepath"
	"testing"
)

func TestEngineCreation(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.templates == nil {
		t.Fatal("Expected templates map to be initialized")
	}

	if engine.funcMap == nil {
		t.Fatal("Expected funcMap to be initialized")
	}
}

func TestAddFuncs(t *testing.T) {
	engine := NewEngine()

	funcs := template.FuncMap{
		"upper": func(s string) string { return s + "_UPPER" },
		"test":  func() string { return "test_func" },
	}

	engine.AddFuncs(funcs)

	// Verify functions are added
	if len(engine.funcMap) != 2 {
		t.Errorf("Expected 2 functions, got: %d", len(engine.funcMap))
	}

	if _, exists := engine.funcMap["upper"]; !exists {
		t.Error("Expected 'upper' function to be added")
	}

	if _, exists := engine.funcMap["test"]; !exists {
		t.Error("Expected 'test' function to be added")
	}
}

func TestLoadAppTemplatesNotExists(t *testing.T) {
	engine := NewEngine()

	// Loading from non-existent directory should not error
	err := engine.LoadAppTemplates("testapp", "/nonexistent/path")
	if err != nil {
		t.Errorf("Expected no error for non-existent directory, got: %v", err)
	}
}

func TestLoadAppTemplatesWithFiles(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "gojango-templates-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	appTemplateDir := filepath.Join(tempDir, "templates")
	err = os.MkdirAll(appTemplateDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create app template dir: %v", err)
	}

	// Create test template files
	testTemplates := map[string]string{
		"index.html":      "<h1>{{.Title}}</h1>",
		"detail.html":     "<div>{{.Content}}</div>",
		"subdir/sub.html": "<p>{{.Message}}</p>",
	}

	for path, content := range testTemplates {
		fullPath := filepath.Join(appTemplateDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", path, err)
		}
	}

	// Load templates
	engine := NewEngine()
	err = engine.LoadAppTemplates("testapp", appTemplateDir)
	if err != nil {
		t.Fatalf("Expected no error loading templates, got: %v", err)
	}

	// Verify templates are loaded
	expectedTemplates := []string{
		"testapp/index.html",
		"testapp/detail.html",
		"testapp/subdir/sub.html",
	}

	for _, templateName := range expectedTemplates {
		if !engine.Has(templateName) {
			t.Errorf("Expected template '%s' to be loaded", templateName)
		}
	}

	// Test template rendering
	html, err := engine.Render("testapp/index.html", map[string]string{"Title": "Test Title"})
	if err != nil {
		t.Errorf("Expected no error rendering template, got: %v", err)
	}

	expected := "<h1>Test Title</h1>"
	if html != expected {
		t.Errorf("Expected '%s', got: '%s'", expected, html)
	}
}

func TestRenderNonExistentTemplate(t *testing.T) {
	engine := NewEngine()

	_, err := engine.Render("nonexistent.html", nil)
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	expectedError := "template 'nonexistent.html' not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}
}

func TestHasTemplate(t *testing.T) {
	engine := NewEngine()

	// Should not have any templates initially
	if engine.Has("test.html") {
		t.Error("Expected template 'test.html' to not exist")
	}

	// Create a simple template manually
	tmpl := template.New("test.html")
	engine.templates["test.html"] = tmpl

	// Should now have the template
	if !engine.Has("test.html") {
		t.Error("Expected template 'test.html' to exist")
	}
}

func TestListTemplates(t *testing.T) {
	engine := NewEngine()

	// Should be empty initially
	templates := engine.List()
	if len(templates) != 0 {
		t.Errorf("Expected 0 templates, got: %d", len(templates))
	}

	// Add some templates manually
	engine.templates["template1.html"] = template.New("template1.html")
	engine.templates["template2.html"] = template.New("template2.html")

	templates = engine.List()
	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got: %d", len(templates))
	}

	// Verify template names
	templateMap := make(map[string]bool)
	for _, name := range templates {
		templateMap[name] = true
	}

	if !templateMap["template1.html"] {
		t.Error("Expected 'template1.html' in template list")
	}

	if !templateMap["template2.html"] {
		t.Error("Expected 'template2.html' in template list")
	}
}

func TestLoadGlobalTemplates(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "gojango-global-templates-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test template file
	templateContent := "<base>{{.Content}}</base>"
	templatePath := filepath.Join(tempDir, "base.html")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Load global templates
	engine := NewEngine()
	err = engine.LoadGlobalTemplates(tempDir)
	if err != nil {
		t.Fatalf("Expected no error loading global templates, got: %v", err)
	}

	// Verify template is loaded
	if !engine.Has("base.html") {
		t.Error("Expected global template 'base.html' to be loaded")
	}

	// Test rendering
	html, err := engine.Render("base.html", map[string]string{"Content": "Test Content"})
	if err != nil {
		t.Errorf("Expected no error rendering global template, got: %v", err)
	}

	expected := "<base>Test Content</base>"
	if html != expected {
		t.Errorf("Expected '%s', got: '%s'", expected, html)
	}
}

func TestTemplateFunctions(t *testing.T) {
	engine := NewEngine()

	// Add custom template functions
	engine.AddFuncs(template.FuncMap{
		"upper": func(s string) string { return "UPPER_" + s },
		"add":   func(a, b int) int { return a + b },
	})

	// Create temporary directory and template with functions
	tempDir, err := os.MkdirTemp("", "gojango-template-funcs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templateContent := `<div>{{upper .Name}} - {{add .A .B}}</div>`
	templatePath := filepath.Join(tempDir, "test.html")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Load template
	err = engine.LoadGlobalTemplates(tempDir)
	if err != nil {
		t.Fatalf("Expected no error loading templates, got: %v", err)
	}

	// Render with data
	data := map[string]interface{}{
		"Name": "test",
		"A":    5,
		"B":    3,
	}

	html, err := engine.Render("test.html", data)
	if err != nil {
		t.Errorf("Expected no error rendering template with functions, got: %v", err)
	}

	expected := "<div>UPPER_test - 8</div>"
	if html != expected {
		t.Errorf("Expected '%s', got: '%s'", expected, html)
	}
}