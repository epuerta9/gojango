package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Engine manages template discovery and rendering
type Engine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
		funcMap:   make(template.FuncMap),
	}
}

// AddFuncs adds template functions
func (e *Engine) AddFuncs(funcs template.FuncMap) {
	for name, fn := range funcs {
		e.funcMap[name] = fn
	}
}

// LoadAppTemplates loads templates for a specific app
func (e *Engine) LoadAppTemplates(appName string, templateDir string) error {
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// No templates directory - this is fine
		return nil
	}
	
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-HTML files
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		
		// Calculate template name relative to app template directory
		relPath, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		
		// Template name format: app/template.html
		templateName := fmt.Sprintf("%s/%s", appName, relPath)
		
		// Read template content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}
		
		// Parse template content with our custom name
		tmpl, err := template.New(templateName).Funcs(e.funcMap).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		
		e.templates[templateName] = tmpl
		
		return nil
	})
}

// LoadGlobalTemplates loads global templates from the templates directory
func (e *Engine) LoadGlobalTemplates(templateDir string) error {
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// No templates directory - this is fine
		return nil
	}
	
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-HTML files
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		
		// Calculate template name relative to global template directory
		relPath, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		
		// Template name is just the relative path
		templateName := relPath
		
		// Parse template
		tmpl := template.New(templateName).Funcs(e.funcMap)
		tmpl, err = tmpl.ParseFiles(path)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		
		e.templates[templateName] = tmpl
		
		return nil
	})
}

// LoadEmbeddedTemplates loads templates from an embedded filesystem
func (e *Engine) LoadEmbeddedTemplates(appName string, embedFS fs.FS, root string) error {
	return fs.WalkDir(embedFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-HTML files
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		
		// Read template content
		content, err := fs.ReadFile(embedFS, path)
		if err != nil {
			return err
		}
		
		// Calculate template name
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		
		templateName := fmt.Sprintf("%s/%s", appName, relPath)
		
		// Parse template
		tmpl := template.New(templateName).Funcs(e.funcMap)
		tmpl, err = tmpl.Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		
		e.templates[templateName] = tmpl
		
		return nil
	})
}

// Render renders a template with the given data
func (e *Engine) Render(templateName string, data interface{}) (string, error) {
	tmpl, exists := e.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template '%s' not found", templateName)
	}
	
	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template '%s': %w", templateName, err)
	}
	
	return buf.String(), nil
}

// Has checks if a template exists
func (e *Engine) Has(templateName string) bool {
	_, exists := e.templates[templateName]
	return exists
}

// List returns all available template names
func (e *Engine) List() []string {
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}