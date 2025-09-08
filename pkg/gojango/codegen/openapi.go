package codegen

import (
	"fmt"
	"os"
	"strings"
)

// OpenAPIGenerator generates OpenAPI specifications from analyzed schemas
type OpenAPIGenerator struct {
	analyzer *SchemaAnalyzer
}

// NewOpenAPIGenerator creates a new OpenAPI generator
func NewOpenAPIGenerator(analyzer *SchemaAnalyzer) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		analyzer: analyzer,
	}
}

// Generate generates OpenAPI specification from analyzed schemas
func (g *OpenAPIGenerator) Generate(outputFile string) error {
	spec := g.buildOpenAPISpec()
	
	return os.WriteFile(outputFile, []byte(spec), 0644)
}

// buildOpenAPISpec builds the OpenAPI specification
func (g *OpenAPIGenerator) buildOpenAPISpec() string {
	var content strings.Builder
	
	content.WriteString(`openapi: 3.0.3
info:
  title: Gojango API
  description: Auto-generated API from Ent schemas
  version: 1.0.0
servers:
  - url: http://localhost:8080/api/v1
    description: Development server

paths:
`)

	models := g.analyzer.GetModels()

	// Generate paths for each model
	for _, model := range models {
		modelPath := strings.ToLower(model.Name)
		content.WriteString(fmt.Sprintf("  /%s:\n", modelPath))
		content.WriteString("    get:\n")
		content.WriteString(fmt.Sprintf("      summary: List %s\n", model.Name))
		content.WriteString("      tags:\n")
		content.WriteString(fmt.Sprintf("        - %s\n", model.Name))
		content.WriteString("      parameters:\n")
		content.WriteString("        - name: page\n")
		content.WriteString("          in: query\n")
		content.WriteString("          schema:\n")
		content.WriteString("            type: integer\n")
		content.WriteString("            default: 1\n")
		content.WriteString("        - name: page_size\n")
		content.WriteString("          in: query\n")
		content.WriteString("          schema:\n")
		content.WriteString("            type: integer\n")
		content.WriteString("            default: 20\n")
		content.WriteString("        - name: search\n")
		content.WriteString("          in: query\n")
		content.WriteString("          schema:\n")
		content.WriteString("            type: string\n")
		content.WriteString("      responses:\n")
		content.WriteString("        '200':\n")
		content.WriteString("          description: Success\n")
		content.WriteString("          content:\n")
		content.WriteString("            application/json:\n")
		content.WriteString("              schema:\n")
		content.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/List%sResponse'\n", model.Name))
		content.WriteString("    post:\n")
		content.WriteString(fmt.Sprintf("      summary: Create %s\n", model.Name))
		content.WriteString("      tags:\n")
		content.WriteString(fmt.Sprintf("        - %s\n", model.Name))
		content.WriteString("      requestBody:\n")
		content.WriteString("        required: true\n")
		content.WriteString("        content:\n")
		content.WriteString("          application/json:\n")
		content.WriteString("            schema:\n")
		content.WriteString(fmt.Sprintf("              $ref: '#/components/schemas/Create%sRequest'\n", model.Name))
		content.WriteString("      responses:\n")
		content.WriteString("        '201':\n")
		content.WriteString("          description: Created\n")
		content.WriteString("          content:\n")
		content.WriteString("            application/json:\n")
		content.WriteString("              schema:\n")
		content.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", model.Name))
		
		content.WriteString(fmt.Sprintf("  /%s/{id}:\n", modelPath))
		content.WriteString("    get:\n")
		content.WriteString(fmt.Sprintf("      summary: Get %s by ID\n", model.Name))
		content.WriteString("      tags:\n")
		content.WriteString(fmt.Sprintf("        - %s\n", model.Name))
		content.WriteString("      parameters:\n")
		content.WriteString("        - name: id\n")
		content.WriteString("          in: path\n")
		content.WriteString("          required: true\n")
		content.WriteString("          schema:\n")
		content.WriteString("            type: integer\n")
		content.WriteString("            format: int64\n")
		content.WriteString("      responses:\n")
		content.WriteString("        '200':\n")
		content.WriteString("          description: Success\n")
		content.WriteString("          content:\n")
		content.WriteString("            application/json:\n")
		content.WriteString("              schema:\n")
		content.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", model.Name))
		content.WriteString("        '404':\n")
		content.WriteString("          description: Not found\n")
		content.WriteString("    put:\n")
		content.WriteString(fmt.Sprintf("      summary: Update %s\n", model.Name))
		content.WriteString("      tags:\n")
		content.WriteString(fmt.Sprintf("        - %s\n", model.Name))
		content.WriteString("      parameters:\n")
		content.WriteString("        - name: id\n")
		content.WriteString("          in: path\n")
		content.WriteString("          required: true\n")
		content.WriteString("          schema:\n")
		content.WriteString("            type: integer\n")
		content.WriteString("            format: int64\n")
		content.WriteString("      requestBody:\n")
		content.WriteString("        required: true\n")
		content.WriteString("        content:\n")
		content.WriteString("          application/json:\n")
		content.WriteString("            schema:\n")
		content.WriteString(fmt.Sprintf("              $ref: '#/components/schemas/Update%sRequest'\n", model.Name))
		content.WriteString("      responses:\n")
		content.WriteString("        '200':\n")
		content.WriteString("          description: Updated\n")
		content.WriteString("          content:\n")
		content.WriteString("            application/json:\n")
		content.WriteString("              schema:\n")
		content.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", model.Name))
		content.WriteString("        '404':\n")
		content.WriteString("          description: Not found\n")
		content.WriteString("    delete:\n")
		content.WriteString(fmt.Sprintf("      summary: Delete %s\n", model.Name))
		content.WriteString("      tags:\n")
		content.WriteString(fmt.Sprintf("        - %s\n", model.Name))
		content.WriteString("      parameters:\n")
		content.WriteString("        - name: id\n")
		content.WriteString("          in: path\n")
		content.WriteString("          required: true\n")
		content.WriteString("          schema:\n")
		content.WriteString("            type: integer\n")
		content.WriteString("            format: int64\n")
		content.WriteString("      responses:\n")
		content.WriteString("        '204':\n")
		content.WriteString("          description: Deleted\n")
		content.WriteString("        '404':\n")
		content.WriteString("          description: Not found\n\n")
	}

	content.WriteString("components:\n")
	content.WriteString("  schemas:\n")
	
	// Generate schema definitions
	for _, model := range models {
		// Main model schema
		content.WriteString(fmt.Sprintf("    %s:\n", model.Name))
		content.WriteString("      type: object\n")
		content.WriteString("      properties:\n")
		
		for _, field := range model.Fields {
			content.WriteString(fmt.Sprintf("        %s:\n", field.Name))
			content.WriteString(fmt.Sprintf("          type: %s\n", field.getOpenAPIType()))
			if field.Type == "time" {
				content.WriteString("          format: date-time\n")
			}
			if field.Name == "id" {
				content.WriteString("          format: int64\n")
			}
		}
		content.WriteString("\n")

		// Create request schema
		content.WriteString(fmt.Sprintf("    Create%sRequest:\n", model.Name))
		content.WriteString("      type: object\n")
		content.WriteString("      required:\n")
		content.WriteString("      properties:\n")
		
		for _, field := range model.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" {
				continue
			}
			content.WriteString(fmt.Sprintf("        %s:\n", field.Name))
			content.WriteString(fmt.Sprintf("          type: %s\n", field.getOpenAPIType()))
			if field.Type == "time" {
				content.WriteString("          format: date-time\n")
			}
		}
		content.WriteString("\n")

		// Update request schema
		content.WriteString(fmt.Sprintf("    Update%sRequest:\n", model.Name))
		content.WriteString("      type: object\n")
		content.WriteString("      properties:\n")
		
		for _, field := range model.Fields {
			if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" {
				continue
			}
			content.WriteString(fmt.Sprintf("        %s:\n", field.Name))
			content.WriteString(fmt.Sprintf("          type: %s\n", field.getOpenAPIType()))
			if field.Type == "time" {
				content.WriteString("          format: date-time\n")
			}
		}
		content.WriteString("\n")

		// List response schema
		content.WriteString(fmt.Sprintf("    List%sResponse:\n", model.Name))
		content.WriteString("      type: object\n")
		content.WriteString("      properties:\n")
		content.WriteString("        items:\n")
		content.WriteString("          type: array\n")
		content.WriteString("          items:\n")
		content.WriteString(fmt.Sprintf("            $ref: '#/components/schemas/%s'\n", model.Name))
		content.WriteString("        total:\n")
		content.WriteString("          type: integer\n")
		content.WriteString("        page:\n")
		content.WriteString("          type: integer\n")
		content.WriteString("        page_size:\n")
		content.WriteString("          type: integer\n")
		content.WriteString("\n")
	}

	return content.String()
}