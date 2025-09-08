package admin

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// EntDatabaseInterface implements DatabaseInterface for Ent models
type EntDatabaseInterface struct {
	client interface{} // Generic Ent client
}

// NewEntDatabaseInterface creates a new Ent database interface
func NewEntDatabaseInterface(client interface{}) *EntDatabaseInterface {
	return &EntDatabaseInterface{
		client: client,
	}
}

// GetAll retrieves all objects with filtering, ordering, and pagination
func (db *EntDatabaseInterface) GetAll(ctx context.Context, model interface{}, filters map[string]interface{}, ordering []string, limit, offset int) ([]interface{}, int, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use Ent's query builder
	// to create dynamic queries based on the model type and filters
	
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	// For now, return mock data
	// TODO: Implement actual Ent query generation
	objects := []interface{}{
		map[string]interface{}{
			"id":         1,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}
	
	return objects, len(objects), nil
}

// GetByID retrieves an object by its ID
func (db *EntDatabaseInterface) GetByID(ctx context.Context, model interface{}, id interface{}) (interface{}, error) {
	// Convert ID to appropriate type
	var idValue interface{}
	switch v := id.(type) {
	case string:
		if intID, err := strconv.Atoi(v); err == nil {
			idValue = intID
		} else {
			idValue = v
		}
	default:
		idValue = id
	}
	
	// TODO: Implement actual Ent query by ID
	return map[string]interface{}{
		"id":         idValue,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}, nil
}

// Create creates a new object
func (db *EntDatabaseInterface) Create(ctx context.Context, model interface{}, data map[string]interface{}) (interface{}, error) {
	// TODO: Implement actual Ent create operation
	// This would use the Ent model's Create() method
	
	// Mock implementation
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	result["id"] = 1
	result["created_at"] = time.Now()
	result["updated_at"] = time.Now()
	
	return result, nil
}

// Update updates an existing object
func (db *EntDatabaseInterface) Update(ctx context.Context, model interface{}, id interface{}, data map[string]interface{}) (interface{}, error) {
	// TODO: Implement actual Ent update operation
	
	// Mock implementation
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	result["id"] = id
	result["updated_at"] = time.Now()
	
	return result, nil
}

// Delete deletes an object
func (db *EntDatabaseInterface) Delete(ctx context.Context, model interface{}, id interface{}) error {
	// TODO: Implement actual Ent delete operation
	return nil
}

// GetSchema returns the schema for a model
func (db *EntDatabaseInterface) GetSchema(model interface{}) (*ModelSchema, error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	schema := &ModelSchema{
		Fields:    []FieldSchema{},
		Relations: []RelationSchema{},
	}
	
	// Use reflection to extract field information
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		fieldSchema := FieldSchema{
			Name: strings.ToLower(field.Name),
			Type: getFieldType(field.Type),
			Verbose: field.Name,
		}
		
		// Parse struct tags
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldSchema.Name = parts[0]
			}
			
			for _, part := range parts[1:] {
				switch part {
				case "omitempty":
					fieldSchema.Required = false
				}
			}
		}
		
		if tag := field.Tag.Get("validate"); tag != "" {
			if strings.Contains(tag, "required") {
				fieldSchema.Required = true
			}
			if strings.Contains(tag, "unique") {
				fieldSchema.Unique = true
			}
		}
		
		if tag := field.Tag.Get("ent"); tag != "" {
			// Parse ent-specific tags
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "max=") {
					if max, err := strconv.Atoi(strings.TrimPrefix(part, "max=")); err == nil {
						fieldSchema.MaxLength = &max
					}
				}
			}
		}
		
		schema.Fields = append(schema.Fields, fieldSchema)
	}
	
	return schema, nil
}

// getFieldType converts Go types to admin field types
func getFieldType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "bytes"
		}
		return "array"
	case reflect.Map:
		return "object"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "datetime"
		}
		return "object"
	default:
		return "string"
	}
}

// AutoRegisterEntModels automatically registers Ent models with the admin
func AutoRegisterEntModels(client interface{}, models ...interface{}) error {
	dbInterface := NewEntDatabaseInterface(client)
	
	for _, model := range models {
		admin := NewModelAdmin(model)
		admin.SetDatabaseInterface(dbInterface)
		
		// Auto-configure based on model reflection
		if err := configureModelAdmin(admin, model); err != nil {
			return fmt.Errorf("failed to configure model admin for %T: %w", model, err)
		}
		
		if err := Register(model, admin); err != nil {
			return fmt.Errorf("failed to register model %T: %w", model, err)
		}
	}
	
	return nil
}

// configureModelAdmin automatically configures ModelAdmin based on model structure
func configureModelAdmin(admin *ModelAdmin, model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	var listDisplay []string
	var searchFields []string
	var listFilter []string
	
	// Analyze fields for auto-configuration
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		
		if !field.IsExported() {
			continue
		}
		
		fieldName := strings.ToLower(field.Name)
		fieldType := field.Type
		
		// Parse JSON tag for field name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}
		
		// Auto-configure based on field type and name
		switch {
		case fieldName == "id":
			// ID is always in list display but not editable
		case fieldType == reflect.TypeOf(time.Time{}) && (fieldName == "created_at" || fieldName == "updated_at"):
			listFilter = append(listFilter, fieldName)
		case fieldType.Kind() == reflect.String:
			if len(searchFields) < 3 { // Limit search fields
				searchFields = append(searchFields, fieldName)
			}
			if len(listDisplay) < 5 { // Limit list display
				listDisplay = append(listDisplay, fieldName)
			}
		case fieldType.Kind() == reflect.Bool:
			listFilter = append(listFilter, fieldName)
		case isNumericType(fieldType):
			if len(listDisplay) < 5 {
				listDisplay = append(listDisplay, fieldName)
			}
		}
	}
	
	// Apply auto-configuration
	if len(listDisplay) > 0 {
		admin.SetListDisplay(listDisplay...)
	}
	if len(searchFields) > 0 {
		admin.SetSearchFields(searchFields...)
	}
	if len(listFilter) > 0 {
		admin.SetListFilter(listFilter...)
	}
	
	// Add default actions
	admin.AddAction("delete_selected", "Delete selected items", func(ctx *gin.Context, objects []interface{}) (interface{}, error) {
		count := 0
		for _, obj := range objects {
			// Extract ID from object
			objMap, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}
			id, exists := objMap["id"]
			if !exists {
				continue
			}
			
			if err := admin.DeleteObject(ctx, fmt.Sprintf("%v", id)); err != nil {
				return nil, fmt.Errorf("failed to delete object %v: %w", id, err)
			}
			count++
		}
		
		return gin.H{
			"message": fmt.Sprintf("Successfully deleted %d items", count),
			"count":   count,
		}, nil
	})
	
	return nil
}

// isNumericType checks if a type is numeric
func isNumericType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		 reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// EntModelAdmin creates a pre-configured ModelAdmin for Ent models
func EntModelAdmin(model interface{}) *ModelAdmin {
	admin := NewModelAdmin(model)
	configureModelAdmin(admin, model)
	return admin
}