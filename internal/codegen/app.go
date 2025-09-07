package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// AppConfig contains configuration for generating a new app
type AppConfig struct {
	Name      string
	Module    string
	Features  []string
}

// AppGenerator generates new Gojango apps
type AppGenerator struct{}

// NewAppGenerator creates a new app generator
func NewAppGenerator() *AppGenerator {
	return &AppGenerator{}
}

// Generate creates a new app based on the configuration
func (g *AppGenerator) Generate(config AppConfig) error {
	appDir := filepath.Join("apps", config.Name)

	// Create app directory
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Load templates
	tmpl := template.New("app").Funcs(template.FuncMap{
		"Title": strings.Title,
		"HasFeature": func(feature string) bool {
			for _, f := range config.Features {
				if f == feature {
					return true
				}
			}
			return false
		},
	})

	// Template data
	data := map[string]interface{}{
		"Name":        config.Name,
		"NameTitle":   toTitleCase(config.Name),
		"Module":      config.Module,
		"Package":     config.Name,
		"Features":    config.Features,
	}

	// Generate files
	files := []AppFileTemplate{
		{"app.go", appGoTemplate, data},
		{"views.go", viewsGoTemplate, data},
		{"schema/README.md", schemaReadmeTemplate, data},
		{"templates/README.md", templatesReadmeTemplate, data},
		{"static/README.md", staticReadmeTemplate, data},
		{"tests/app_test.go", appTestGoTemplate, data},
	}

	for _, file := range files {
		if err := g.generateAppFile(appDir, file, tmpl); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.Path, err)
		}
	}

	// Create additional directories
	directories := []string{"schema", "templates", "static", "tests"}
	for _, dir := range directories {
		dirPath := filepath.Join(appDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// AppFileTemplate represents an app template file
type AppFileTemplate struct {
	Path         string
	TemplateText string
	Data         interface{}
}

// generateAppFile generates a single app file from template
func (g *AppGenerator) generateAppFile(appDir string, file AppFileTemplate, tmpl *template.Template) error {
	// Parse template
	t, err := tmpl.Parse(file.TemplateText)
	if err != nil {
		return fmt.Errorf("failed to parse template for %s: %w", file.Path, err)
	}

	// Create directory if needed
	filePath := filepath.Join(appDir, file.Path)
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return err
	}

	// Create file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Execute template
	return t.Execute(f, file.Data)
}

// Template strings for app generation
const appGoTemplate = `package {{.Package}}

import (
	"github.com/epuerta9/gojango/pkg/gojango"
)

func init() {
	// Register this app with the global registry
	gojango.Register(&{{.NameTitle}}App{})
}

// {{.NameTitle}}App represents the {{.Name}} application
type {{.NameTitle}}App struct {
	gojango.BaseApp
}

// Config returns the app configuration
func (app *{{.NameTitle}}App) Config() gojango.AppConfig {
	return gojango.AppConfig{
		Name:    "{{.Name}}",
		Label:   "{{.NameTitle}} Application",
		Version: "1.0.0",
	}
}

// Initialize sets up the app
func (app *{{.NameTitle}}App) Initialize(ctx *gojango.AppContext) error {
	// Call parent initialization
	if err := app.BaseApp.Initialize(ctx); err != nil {
		return err
	}

	// App-specific initialization goes here
	
	return nil
}

// Routes defines the HTTP routes for this app
func (app *{{.NameTitle}}App) Routes() []gojango.Route {
	return []gojango.Route{
		{
			Method:  "GET",
			Path:    "/",
			Handler: app.IndexView,
			Name:    "index",
		},
		// Add more routes here
	}
}
`

const viewsGoTemplate = `package {{.Package}}

import (
	"fmt"
	"net/http"
)

// IndexView handles the main page for this app
func (app *{{.NameTitle}}App) IndexView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Welcome to {{.NameTitle}}</h1>")
	fmt.Fprintf(w, "<p>This is the {{.Name}} app index page.</p>")
	fmt.Fprintf(w, "<p>TODO: Implement your views here!</p>")
}

// Add more view functions here
`

const schemaReadmeTemplate = `# {{.NameTitle}} Schema

This directory contains the database schema definitions for the {{.Name}} app.

## Usage

In Phase 3, you'll define your Ent schemas here:

` + "```go" + `
// post.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type Post struct {
    ent.Schema
}

func (Post) Fields() []ent.Field {
    return []ent.Field{
        field.String("title").NotEmpty(),
        field.Text("content"),
        field.Time("created_at").Default(time.Now),
    }
}
` + "```" + `

For now, this is a placeholder until database functionality is implemented.
`

const templatesReadmeTemplate = `# {{.NameTitle}} Templates

This directory contains HTML templates for the {{.Name}} app.

## Usage

In Phase 2, you'll create HTML templates here:

` + "```html" + `
<!-- index.html -->
<h1>{{.NameTitle}}</h1>
<p>Welcome to the {{.Name}} app!</p>
` + "```" + `

For now, this is a placeholder until template functionality is implemented.
`

const staticReadmeTemplate = `# {{.NameTitle}} Static Files

This directory contains static assets for the {{.Name}} app:

- CSS files
- JavaScript files  
- Images
- Other static assets

## Structure

` + "```" + `
static/
├── css/
│   └── {{.Name}}.css
├── js/
│   └── {{.Name}}.js
└── img/
    └── (images)
` + "```" + `

For now, this is a placeholder until static file handling is implemented.
`

const appTestGoTemplate = `package {{.Package}}_test

import (
	"testing"
	"{{.Module}}/apps/{{.Name}}"
)

func TestConfig(t *testing.T) {
	app := &{{.Package}}.{{.NameTitle}}App{}
	config := app.Config()
	
	if config.Name != "{{.Name}}" {
		t.Errorf("Expected app name '{{.Name}}', got '%s'", config.Name)
	}
	
	if config.Label != "{{.NameTitle}} Application" {
		t.Errorf("Expected app label '{{.NameTitle}} Application', got '%s'", config.Label)
	}
}

func TestRoutes(t *testing.T) {
	app := &{{.Package}}.{{.NameTitle}}App{}
	routes := app.Routes()
	
	if len(routes) == 0 {
		t.Error("Expected at least one route")
	}
	
	// Check that index route exists
	found := false
	for _, route := range routes {
		if route.Name == "index" && route.Method == "GET" && route.Path == "/" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Index route not found")
	}
}
`

// toTitleCase converts a string to TitleCase (proper PascalCase)
// For app names ending with "app", we want TestApp not Testapp
// e.g., "testapp" -> "TestApp", "blog_post" -> "BlogPost", "user" -> "User"
func toTitleCase(s string) string {
	if s == "" {
		return s
	}
	
	// Special handling for names ending with "app"
	if strings.HasSuffix(s, "app") && len(s) > 3 {
		base := s[:len(s)-3]
		return strings.ToUpper(string(base[0])) + base[1:] + "App"
	}
	
	// Handle underscores and dashes
	if strings.ContainsAny(s, "_-") {
		result := ""
		parts := strings.FieldsFunc(s, func(c rune) bool {
			return c == '_' || c == '-'
		})
		for _, part := range parts {
			if len(part) > 0 {
				result += strings.ToUpper(string(part[0])) + part[1:]
			}
		}
		return result
	}
	
	// Simple case: just capitalize first letter
	return strings.ToUpper(string(s[0])) + s[1:]
}