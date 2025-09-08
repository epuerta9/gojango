package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProtoGenerator generates protobuf definitions from analyzed schemas
type ProtoGenerator struct {
	analyzer *SchemaAnalyzer
}

// NewProtoGenerator creates a new protobuf generator
func NewProtoGenerator(analyzer *SchemaAnalyzer) *ProtoGenerator {
	return &ProtoGenerator{
		analyzer: analyzer,
	}
}

// Generate generates protobuf definitions from analyzed schemas
func (g *ProtoGenerator) Generate(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate main proto file
	protoContent := g.buildProtoContent()
	
	protoFile := filepath.Join(outputDir, "models.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		return fmt.Errorf("failed to write proto file: %w", err)
	}

	// Generate service proto file
	serviceContent := g.buildServiceProtoContent()
	
	serviceFile := filepath.Join(outputDir, "service.proto")
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service proto file: %w", err)
	}

	return nil
}

// buildProtoContent builds the main protobuf content
func (g *ProtoGenerator) buildProtoContent() string {
	var content strings.Builder
	
	content.WriteString(`syntax = "proto3";

package models;

import "google/protobuf/timestamp.proto";

option go_package = "./internal/proto/models";

`)

	models := g.analyzer.GetModels()

	// Generate message definitions for each model
	for _, model := range models {
		content.WriteString(fmt.Sprintf("// %s represents a %s model\n", model.Name, model.Name))
		content.WriteString(fmt.Sprintf("message %s {\n", model.Name))
		
		for i, field := range model.Fields {
			fieldNumber := i + 1
			content.WriteString(fmt.Sprintf("  %s %s = %d;\n", field.ProtoType, field.Name, fieldNumber))
		}
		
		content.WriteString("}\n\n")

		// Generate request/response messages
		content.WriteString(fmt.Sprintf("message Create%sRequest {\n", model.Name))
		for i, field := range model.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" {
				continue // Skip auto-generated fields
			}
			fieldNumber := i + 1
			content.WriteString(fmt.Sprintf("  %s %s = %d;\n", field.ProtoType, field.Name, fieldNumber))
		}
		content.WriteString("}\n\n")

		content.WriteString(fmt.Sprintf("message Update%sRequest {\n", model.Name))
		content.WriteString("  int64 id = 1;\n")
		for i, field := range model.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" {
				continue
			}
			fieldNumber := i + 2
			content.WriteString(fmt.Sprintf("  optional %s %s = %d;\n", field.ProtoType, field.Name, fieldNumber))
		}
		content.WriteString("}\n\n")

		content.WriteString(fmt.Sprintf("message Get%sRequest {\n", model.Name))
		content.WriteString("  int64 id = 1;\n")
		content.WriteString("}\n\n")

		content.WriteString(fmt.Sprintf("message Delete%sRequest {\n", model.Name))
		content.WriteString("  int64 id = 1;\n")
		content.WriteString("}\n\n")

		content.WriteString(fmt.Sprintf("message List%sRequest {\n", model.Name))
		content.WriteString("  int32 page = 1;\n")
		content.WriteString("  int32 page_size = 2;\n")
		content.WriteString("  string search = 3;\n")
		content.WriteString("  string sort = 4;\n")
		content.WriteString("}\n\n")

		content.WriteString(fmt.Sprintf("message List%sResponse {\n", model.Name))
		content.WriteString(fmt.Sprintf("  repeated %s items = 1;\n", model.Name))
		content.WriteString("  int32 total = 2;\n")
		content.WriteString("  int32 page = 3;\n")
		content.WriteString("  int32 page_size = 4;\n")
		content.WriteString("}\n\n")
	}

	return content.String()
}

// buildServiceProtoContent builds the service protobuf content
func (g *ProtoGenerator) buildServiceProtoContent() string {
	var content strings.Builder
	
	content.WriteString(`syntax = "proto3";

package service;

import "google/protobuf/empty.proto";
import "models.proto";

option go_package = "./internal/proto/service";

`)

	models := g.analyzer.GetModels()

	// Generate service definitions for each model
	for _, model := range models {
		serviceName := fmt.Sprintf("%sService", model.Name)
		content.WriteString(fmt.Sprintf("// %s provides CRUD operations for %s\n", serviceName, model.Name))
		content.WriteString(fmt.Sprintf("service %s {\n", serviceName))
		
		content.WriteString(fmt.Sprintf("  rpc Create%s(models.Create%sRequest) returns (models.%s);\n", model.Name, model.Name, model.Name))
		content.WriteString(fmt.Sprintf("  rpc Get%s(models.Get%sRequest) returns (models.%s);\n", model.Name, model.Name, model.Name))
		content.WriteString(fmt.Sprintf("  rpc Update%s(models.Update%sRequest) returns (models.%s);\n", model.Name, model.Name, model.Name))
		content.WriteString(fmt.Sprintf("  rpc Delete%s(models.Delete%sRequest) returns (google.protobuf.Empty);\n", model.Name, model.Name))
		content.WriteString(fmt.Sprintf("  rpc List%s(models.List%sRequest) returns (models.List%sResponse);\n", model.Name, model.Name, model.Name))
		
		content.WriteString("}\n\n")
	}

	return content.String()
}