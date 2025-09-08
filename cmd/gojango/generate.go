package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/epuerta9/gojango/pkg/gojango/codegen"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [type]",
		Short: "Generate code from schemas",
		Long: `Generate code from Ent schemas and other definitions.

Available generators:
  ent     - Generate Ent ORM code
  proto   - Generate protobuf files from schemas
  openapi - Generate OpenAPI spec from schemas
  all     - Generate all code`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			genType := args[0]

			switch genType {
			case "ent":
				return generateEnt()
			case "proto":
				return generateProto()
			case "openapi":
				return generateOpenAPI()
			case "all":
				if err := generateEnt(); err != nil {
					return err
				}
				if err := generateProto(); err != nil {
					return err
				}
				return generateOpenAPI()
			default:
				return fmt.Errorf("unknown generation type: %s", genType)
			}
		},
	}

	return cmd
}

func generateEnt() error {
	fmt.Println("🔧 Generating Ent code...")
	
	// Find schema directories
	schemaDirs := []string{"schema", "internal/ent/schema"}
	var schemaDir string
	
	for _, dir := range schemaDirs {
		if _, err := os.Stat(dir); err == nil {
			schemaDir = dir
			break
		}
	}
	
	if schemaDir == "" {
		return fmt.Errorf("no schema directory found (tried: %v)", schemaDirs)
	}

	entCmd := exec.Command("go", "run", "-mod=mod", "entgo.io/ent/cmd/ent", "generate", "./"+schemaDir)
	entCmd.Stdout = os.Stdout
	entCmd.Stderr = os.Stderr
	
	if err := entCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate Ent code: %w", err)
	}

	fmt.Println("✅ Ent code generated")
	return nil
}

func generateProto() error {
	fmt.Println("🔧 Generating protobuf files...")
	
	// Find schema directories
	schemaDirs := []string{"schema", "apps/*/schema", "internal/ent/schema"}
	var schemaDir string
	
	for _, pattern := range schemaDirs {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if stat, err := os.Stat(match); err == nil && stat.IsDir() {
				schemaDir = match
				break
			}
		}
		if schemaDir != "" {
			break
		}
	}
	
	if schemaDir == "" {
		return fmt.Errorf("no schema directory found (tried: %v)", schemaDirs)
	}

	// Analyze schemas
	analyzer := codegen.NewSchemaAnalyzer(schemaDir)
	if err := analyzer.Analyze(); err != nil {
		return fmt.Errorf("failed to analyze schemas: %w", err)
	}

	models := analyzer.GetModels()
	if len(models) == 0 {
		fmt.Println("⚠️  No models found in schema directory")
		return nil
	}

	// Generate protobuf files
	protoGenerator := codegen.NewProtoGenerator(analyzer)
	outputDir := "internal/proto"
	
	if err := protoGenerator.Generate(outputDir); err != nil {
		return fmt.Errorf("failed to generate protobuf files: %w", err)
	}

	fmt.Printf("✅ Generated protobuf files for %d models in %s\n", len(models), outputDir)
	return nil
}

func generateOpenAPI() error {
	fmt.Println("🔧 Generating OpenAPI specification...")
	
	// Find schema directories
	schemaDirs := []string{"schema", "apps/*/schema", "internal/ent/schema"}
	var schemaDir string
	
	for _, pattern := range schemaDirs {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if stat, err := os.Stat(match); err == nil && stat.IsDir() {
				schemaDir = match
				break
			}
		}
		if schemaDir != "" {
			break
		}
	}
	
	if schemaDir == "" {
		return fmt.Errorf("no schema directory found (tried: %v)", schemaDirs)
	}

	// Analyze schemas
	analyzer := codegen.NewSchemaAnalyzer(schemaDir)
	if err := analyzer.Analyze(); err != nil {
		return fmt.Errorf("failed to analyze schemas: %w", err)
	}

	models := analyzer.GetModels()
	if len(models) == 0 {
		fmt.Println("⚠️  No models found in schema directory")
		return nil
	}

	// Generate OpenAPI specification
	openAPIGenerator := codegen.NewOpenAPIGenerator(analyzer)
	outputFile := "openapi.yaml"
	
	if err := openAPIGenerator.Generate(outputFile); err != nil {
		return fmt.Errorf("failed to generate OpenAPI spec: %w", err)
	}

	fmt.Printf("✅ Generated OpenAPI specification for %d models in %s\n", len(models), outputFile)
	return nil
}