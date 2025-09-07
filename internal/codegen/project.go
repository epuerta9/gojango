package codegen

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// ProjectConfig contains configuration for generating a new project
type ProjectConfig struct {
	Name      string
	Module    string
	Directory string
	Frontend  string
	Database  string
	Features  []string
}

// ProjectGenerator generates new Gojango projects
type ProjectGenerator struct {
	templates *template.Template
}

// NewProjectGenerator creates a new project generator
func NewProjectGenerator() *ProjectGenerator {
	return &ProjectGenerator{}
}

// Generate creates a new project based on the configuration
func (g *ProjectGenerator) Generate(config ProjectConfig) error {
	// Create project directory
	if err := os.MkdirAll(config.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Load templates with custom functions
	tmpl := template.New("project").Funcs(template.FuncMap{
		"has": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
	})
	
	tmpl, err := tmpl.ParseFS(templateFS, 
		"templates/project/*.tmpl",
		"templates/project/cmd/server/*.tmpl",
		"templates/project/internal/settings/*.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Template data
	data := map[string]interface{}{
		"Name":      config.Name,
		"Module":    config.Module,
		"Frontend":  config.Frontend,
		"Database":  config.Database,
		"Features":  config.Features,
		"HasFeature": func(feature string) bool {
			for _, f := range config.Features {
				if f == feature {
					return true
				}
			}
			return false
		},
	}

	// Generate files
	files := []FileTemplate{
		{"main.go", "templates/project/main.go.tmpl", data},
		{"go.mod", "templates/project/go.mod.tmpl", data},
		{"Makefile", "templates/project/Makefile.tmpl", data},
		{"docker-compose.yml", "templates/project/docker-compose.yml.tmpl", data},
		{".env.example", "templates/project/env.example.tmpl", data},
		{".gitignore", "templates/project/gitignore.tmpl", data},
		{"README.md", "templates/project/README.md.tmpl", data},
		{"gojango.yaml", "templates/project/gojango.yaml.tmpl", data},
		{"cmd/server/main.go", "templates/project/cmd/server/main.go.tmpl", data},
		{"internal/settings/settings.go", "templates/project/internal/settings/settings.go.tmpl", data},
	}

	for _, file := range files {
		if err := g.generateFile(config.Directory, file, tmpl); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.Path, err)
		}
	}

	// Create directories
	directories := []string{
		"apps",
		"static",
		"templates",
		"tests",
		"docs",
	}

	for _, dir := range directories {
		dirPath := filepath.Join(config.Directory, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Add .gitkeep files to empty directories
		gitkeepPath := filepath.Join(dirPath, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %s: %w", dir, err)
		}
	}

	return nil
}

// FileTemplate represents a template file to generate
type FileTemplate struct {
	Path         string
	TemplateName string
	Data         interface{}
}

// generateFile generates a single file from template
func (g *ProjectGenerator) generateFile(projectDir string, file FileTemplate, tmpl *template.Template) error {
	// Create directory if it doesn't exist
	filePath := filepath.Join(projectDir, file.Path)
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return err
	}

	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Execute template
	return tmpl.ExecuteTemplate(f, filepath.Base(file.TemplateName), file.Data)
}