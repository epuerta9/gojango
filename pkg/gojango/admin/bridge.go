// Package admin provides Django-style admin bridge for Ent models
package admin

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	adminpb "github.com/epuerta9/gojango/pkg/gojango/admin/proto"
)

// EntBridge connects the admin system to Ent models
type EntBridge struct {
	client    interface{}          // Generic client interface to avoid Ent import dependency
	modelMap  map[string]*ModelAdmin
}

// NewEntBridge creates a new bridge to Ent models
func NewEntBridge(client interface{}) *EntBridge {
	return &EntBridge{
		client:   client,
		modelMap: make(map[string]*ModelAdmin),
	}
}

// ModelQueryer interface for Ent model queries
type ModelQueryer interface {
	Count(ctx context.Context) (int, error)
	Limit(int) interface{}
	Offset(int) interface{}
	Order(interface{}) interface{}
	Where(interface{}) interface{}
	All(ctx context.Context) (interface{}, error)
	First(ctx context.Context) (interface{}, error)
}

// ModelMutator interface for Ent model mutations
type ModelMutator interface {
	Create() interface{}
	Update() interface{}
	Delete() interface{}
	Save(ctx context.Context) (interface{}, error)
}

// FieldInfo represents field metadata for admin forms and display
type FieldInfo struct {
	Name         string
	FieldType    string
	VerboseName  string
	HelpText     string
	Required     bool
	Editable     bool
	Blank        bool
	Null         bool
	MaxLength    int
	Unique       bool
	RelatedModel string
	WidgetType   string
}

// EntModelReflector provides model introspection capabilities
type EntModelReflector struct {
	modelType reflect.Type
}

// GetFieldsFromModel extracts field information from an Ent model
func (r *EntModelReflector) GetFields() []FieldInfo {
	var fields []FieldInfo
	
	if r.modelType == nil {
		return fields
	}

	for i := 0; i < r.modelType.NumField(); i++ {
		field := r.modelType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldInfo := FieldInfo{
			Name:       field.Name,
			FieldType:  r.getFieldType(field.Type),
			Required:   r.isFieldRequired(field),
			Editable:   r.isFieldEditable(field),
		}

		// Extract JSON tag for field name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldInfo.Name = parts[0]
			}
		}

		fields = append(fields, fieldInfo)
	}

	return fields
}

func (r *EntModelReflector) getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Bool:
		return "boolean"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "datetime"
		}
		return "object"
	case reflect.Ptr:
		return r.getFieldType(t.Elem())
	case reflect.Slice:
		return "array"
	default:
		return "unknown"
	}
}

func (r *EntModelReflector) isFieldRequired(field reflect.StructField) bool {
	// Check for required tag or analyze the field type
	tag := field.Tag.Get("validate")
	return strings.Contains(tag, "required")
}

func (r *EntModelReflector) isFieldEditable(field reflect.StructField) bool {
	// Fields like ID, CreatedAt, UpdatedAt are typically not editable
	name := strings.ToLower(field.Name)
	return name != "id" && name != "createdat" && name != "updatedat"
}

// ConvertEntObjectToObjectData converts an Ent model instance to protobuf ObjectData
func ConvertEntObjectToObjectData(obj interface{}) (*adminpb.ObjectData, error) {
	if obj == nil {
		return nil, fmt.Errorf("object is nil")
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object is not a struct")
	}

	fields := make(map[string]*structpb.Value)
	var id string
	var createdAt, updatedAt *timestamppb.Timestamp

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !field.IsExported() {
			continue
		}

		// Get field name from JSON tag
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		// Handle special fields
		switch strings.ToLower(field.Name) {
		case "id":
			if fieldValue.IsValid() && !fieldValue.IsZero() {
				id = fmt.Sprintf("%v", fieldValue.Interface())
			}
		case "createdat":
			if t, ok := fieldValue.Interface().(time.Time); ok && !t.IsZero() {
				createdAt = timestamppb.New(t)
			}
		case "updatedat":
			if t, ok := fieldValue.Interface().(time.Time); ok && !t.IsZero() {
				updatedAt = timestamppb.New(t)
			}
		}

		// Convert field value to protobuf Value
		pbValue, err := convertToProtobufValue(fieldValue.Interface())
		if err != nil {
			// Skip fields that can't be converted
			continue
		}
		fields[fieldName] = pbValue
	}

	return &adminpb.ObjectData{
		Id:               id,
		Fields:           fields,
		StrRepresentation: fmt.Sprintf("%v", obj),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

// convertToProtobufValue converts a Go value to protobuf Value
func convertToProtobufValue(val interface{}) (*structpb.Value, error) {
	if val == nil {
		return structpb.NewNullValue(), nil
	}

	switch v := val.(type) {
	case string:
		return structpb.NewStringValue(v), nil
	case int, int8, int16, int32, int64:
		return structpb.NewNumberValue(float64(reflect.ValueOf(v).Int())), nil
	case uint, uint8, uint16, uint32, uint64:
		return structpb.NewNumberValue(float64(reflect.ValueOf(v).Uint())), nil
	case float32, float64:
		return structpb.NewNumberValue(reflect.ValueOf(v).Float()), nil
	case bool:
		return structpb.NewBoolValue(v), nil
	case time.Time:
		return structpb.NewStringValue(v.Format(time.RFC3339)), nil
	default:
		// For complex types, convert to string representation
		return structpb.NewStringValue(fmt.Sprintf("%v", val)), nil
	}
}

// ParseFilterParams parses filter parameters for database queries
func ParseFilterParams(filters map[string]string) map[string]interface{} {
	parsed := make(map[string]interface{})
	
	for key, value := range filters {
		// Handle different filter types
		if strings.HasSuffix(key, "__exact") {
			parsed[strings.TrimSuffix(key, "__exact")] = value
		} else if strings.HasSuffix(key, "__icontains") {
			parsed[strings.TrimSuffix(key, "__icontains")] = "%" + value + "%"
		} else if strings.HasSuffix(key, "__gt") {
			if num, err := strconv.Atoi(value); err == nil {
				parsed[strings.TrimSuffix(key, "__gt")] = map[string]int{"gt": num}
			}
		} else if strings.HasSuffix(key, "__lt") {
			if num, err := strconv.Atoi(value); err == nil {
				parsed[strings.TrimSuffix(key, "__lt")] = map[string]int{"lt": num}
			}
		} else {
			parsed[key] = value
		}
	}
	
	return parsed
}

// BuildOrderClause builds ordering clause for queries
func BuildOrderClause(ordering string) interface{} {
	if ordering == "" {
		return nil
	}

	desc := false
	if strings.HasPrefix(ordering, "-") {
		desc = true
		ordering = strings.TrimPrefix(ordering, "-")
	}

	// This would need to be implemented based on your Ent schema
	// For now, return the ordering string
	if desc {
		return "-" + ordering
	}
	return ordering
}