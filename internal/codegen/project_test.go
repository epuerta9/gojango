package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectGenerator(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "gojango-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project config
	config := ProjectConfig{
		Name:      "testproject",
		Module:    "github.com/test/testproject",
		Directory: tempDir,
		Frontend:  "htmx",
		Database:  "postgres",
		Features:  []string{"admin", "auth"},
	}

	// Generate project
	generator := NewProjectGenerator()
	err = generator.Generate(config)
	if err != nil {
		t.Fatalf("Project generation failed: %v", err)
	}

	// Test that expected files were created
	expectedFiles := []string{
		"main.go",
		"go.mod",
		"Makefile", 
		"docker-compose.yml",
		".env.example",
		".gitignore",
		"README.md",
		"gojango.yaml",
		"cmd/server/main.go",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(tempDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	// Test that expected directories were created
	expectedDirs := []string{
		"apps",
		"static", 
		"templates",
		"tests",
		"docs",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(tempDir, dir)
		if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
			t.Errorf("Expected directory %s was not created", dir)
		}

		// Check for .gitkeep file
		gitkeepPath := filepath.Join(dirPath, ".gitkeep")
		if _, err := os.Stat(gitkeepPath); os.IsNotExist(err) {
			t.Errorf("Expected .gitkeep in %s was not created", dir)
		}
	}
}

func TestProjectGeneratorTemplateData(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "gojango-template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project config with specific values
	config := ProjectConfig{
		Name:      "myawesomeapp",
		Module:    "github.com/user/myawesomeapp",
		Directory: tempDir,
		Frontend:  "react",
		Database:  "mysql",
		Features:  []string{"docker", "api"},
	}

	// Generate project
	generator := NewProjectGenerator()
	err = generator.Generate(config)
	if err != nil {
		t.Fatalf("Project generation failed: %v", err)
	}

	// Read generated go.mod and check module name
	goModPath := filepath.Join(tempDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	goModContent := string(content)
	if !contains(goModContent, "github.com/user/myawesomeapp") {
		t.Errorf("go.mod should contain module name, got: %s", goModContent)
	}

	// Read generated README.md and check project name
	readmePath := filepath.Join(tempDir, "README.md")
	content, err = os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	readmeContent := string(content)
	if !contains(readmeContent, "myawesomeapp") {
		t.Errorf("README.md should contain project name")
	}
	if !contains(readmeContent, "react") {
		t.Errorf("README.md should contain frontend framework")
	}
	if !contains(readmeContent, "mysql") {
		t.Errorf("README.md should contain database type")
	}

	// Check gojango.yaml
	configPath := filepath.Join(tempDir, "gojango.yaml")
	content, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read gojango.yaml: %v", err)
	}

	configContent := string(content)
	if !contains(configContent, "name: myawesomeapp") {
		t.Errorf("gojango.yaml should contain project name")
	}
	if !contains(configContent, "framework: react") {
		t.Errorf("gojango.yaml should contain frontend framework")
	}
	if !contains(configContent, "engine: mysql") {
		t.Errorf("gojango.yaml should contain database engine")
	}
}

func TestProjectGeneratorDirectoryCreation(t *testing.T) {
	// Test that generator creates the project directory if it doesn't exist
	tempDir, err := os.MkdirTemp("", "gojango-mkdir-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a subdirectory that doesn't exist yet
	projectDir := filepath.Join(tempDir, "newproject")

	config := ProjectConfig{
		Name:      "newproject", 
		Module:    "github.com/test/newproject",
		Directory: projectDir,
		Frontend:  "htmx",
		Database:  "sqlite",
		Features:  []string{},
	}

	// Generate project
	generator := NewProjectGenerator()
	err = generator.Generate(config)
	if err != nil {
		t.Fatalf("Project generation failed: %v", err)
	}

	// Check that project directory was created
	if info, err := os.Stat(projectDir); os.IsNotExist(err) || !info.IsDir() {
		t.Error("Project directory should have been created")
	}

	// Check that main.go exists in the new directory
	mainGoPath := filepath.Join(projectDir, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("main.go should exist in project directory")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 s[1:len(substr)+1] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}