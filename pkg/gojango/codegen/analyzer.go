package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// SchemaAnalyzer analyzes Ent schemas and generates code
type SchemaAnalyzer struct {
	schemaDir string
	models    []*ModelInfo
}

// ModelInfo contains metadata about an Ent model
type ModelInfo struct {
	Name        string
	PackageName string
	TableName   string
	Fields      []*FieldInfo
	Edges       []*EdgeInfo
	AdminConfig *AdminConfig
}

// FieldInfo represents a model field
type FieldInfo struct {
	Name         string
	Type         string
	GoType       string
	ProtoType    string
	JSONTag      string
	Optional     bool
	Unique       bool
	Default      interface{}
	Description  string
}

// EdgeInfo represents model relationships
type EdgeInfo struct {
	Name        string
	Type        string // "O2O", "O2M", "M2O", "M2M"
	Target      string
	Field       string
	Inverse     string
	Description string
}

// AdminConfig represents admin interface configuration
type AdminConfig struct {
	ListDisplay  []string
	SearchFields []string
	Actions      []string
}

// NewSchemaAnalyzer creates a new schema analyzer
func NewSchemaAnalyzer(schemaDir string) *SchemaAnalyzer {
	return &SchemaAnalyzer{
		schemaDir: schemaDir,
		models:    make([]*ModelInfo, 0),
	}
}

// Analyze scans and analyzes all Ent schemas
func (a *SchemaAnalyzer) Analyze() error {
	// Find all .go files in schema directory
	files, err := filepath.Glob(filepath.Join(a.schemaDir, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to find schema files: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(filepath.Base(file), "_test.go") {
			continue // Skip test files
		}

		model, err := a.analyzeFile(file)
		if err != nil {
			return fmt.Errorf("failed to analyze %s: %w", file, err)
		}

		if model != nil {
			a.models = append(a.models, model)
		}
	}

	return nil
}

// analyzeFile analyzes a single schema file
func (a *SchemaAnalyzer) analyzeFile(filename string) (*ModelInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var model *ModelInfo

	// Look for struct types that implement ent.Schema
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name != nil && x.Name.IsExported() {
				// Check if this is a schema struct
				if _, ok := x.Type.(*ast.StructType); ok {
					model = &ModelInfo{
						Name:        x.Name.Name,
						PackageName: node.Name.Name,
						TableName:   toSnakeCase(x.Name.Name),
						Fields:      make([]*FieldInfo, 0),
						Edges:       make([]*EdgeInfo, 0),
					}

					// Look for Fields() and Edges() methods
					model = a.extractSchemaInfo(node, model)
				}
			}
		}
		return true
	})

	return model, nil
}

// extractSchemaInfo extracts field and edge information from schema methods
func (a *SchemaAnalyzer) extractSchemaInfo(node *ast.File, model *ModelInfo) *ModelInfo {
	// This is a simplified implementation
	// In a real implementation, you'd need to analyze the Fields() and Edges() methods
	// and extract the field definitions, types, and configuration

	// For now, we'll add some common fields that most models have
	model.Fields = append(model.Fields, []*FieldInfo{
		{
			Name:      "id",
			Type:      "int",
			GoType:    "int",
			ProtoType: "int64",
			JSONTag:   "id",
		},
		{
			Name:      "created_at",
			Type:      "time",
			GoType:    "time.Time",
			ProtoType: "google.protobuf.Timestamp",
			JSONTag:   "created_at",
		},
		{
			Name:      "updated_at",
			Type:      "time",
			GoType:    "time.Time",
			ProtoType: "google.protobuf.Timestamp",
			JSONTag:   "updated_at",
		},
	}...)

	return model
}

// GetModels returns the analyzed models
func (a *SchemaAnalyzer) GetModels() []*ModelInfo {
	return a.models
}

// getOpenAPIType converts Go types to OpenAPI types
func (f *FieldInfo) getOpenAPIType() string {
	switch f.Type {
	case "int", "int32", "int64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time":
		return "string"
	default:
		return "string"
	}
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