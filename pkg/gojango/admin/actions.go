package admin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ActionRegistry manages all available admin actions
type ActionRegistry struct {
	actions map[string]Action
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	registry := &ActionRegistry{
		actions: make(map[string]Action),
	}
	
	// Register default actions
	registry.registerDefaultActions()
	
	return registry
}

// GlobalActionRegistry is the global registry for admin actions
var GlobalActionRegistry = NewActionRegistry()

// RegisterAction registers a global action
func RegisterAction(name, description string, handler func(ctx *gin.Context, objects []interface{}) (interface{}, error)) {
	GlobalActionRegistry.Register(name, description, handler)
}

// Register adds an action to the registry
func (ar *ActionRegistry) Register(name, description string, handler func(ctx *gin.Context, objects []interface{}) (interface{}, error)) {
	ar.actions[name] = Action{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

// Get retrieves an action by name
func (ar *ActionRegistry) Get(name string) (Action, bool) {
	action, exists := ar.actions[name]
	return action, exists
}

// GetAll returns all registered actions
func (ar *ActionRegistry) GetAll() map[string]Action {
	actions := make(map[string]Action)
	for name, action := range ar.actions {
		actions[name] = action
	}
	return actions
}

// registerDefaultActions registers the default admin actions
func (ar *ActionRegistry) registerDefaultActions() {
	// Delete selected action
	ar.Register("delete_selected", "Delete selected items", DeleteSelectedAction)
	
	// Export actions
	ar.Register("export_csv", "Export selected items as CSV", ExportCSVAction)
	ar.Register("export_json", "Export selected items as JSON", ExportJSONAction)
	
	// Status change actions (common for many models)
	ar.Register("mark_active", "Mark selected items as active", MarkActiveAction)
	ar.Register("mark_inactive", "Mark selected items as inactive", MarkInactiveAction)
}

// Default action implementations

// DeleteSelectedAction deletes the selected objects
func DeleteSelectedAction(ctx *gin.Context, objects []interface{}) (interface{}, error) {
	if len(objects) == 0 {
		return gin.H{"message": "No items selected", "count": 0}, nil
	}
	
	// Extract model admin from context (this would be set by the handler)
	modelAdminInterface, exists := ctx.Get("model_admin")
	if !exists {
		return nil, fmt.Errorf("model admin not found in context")
	}
	
	modelAdmin, ok := modelAdminInterface.(*ModelAdmin)
	if !ok {
		return nil, fmt.Errorf("invalid model admin type")
	}
	
	count := 0
	errors := []string{}
	
	for _, obj := range objects {
		// Extract ID from object
		id, err := extractObjectID(obj)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to extract ID from object: %v", err))
			continue
		}
		
		if err := modelAdmin.DeleteObject(ctx, id); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to delete object %s: %v", id, err))
			continue
		}
		count++
	}
	
	result := gin.H{
		"message": fmt.Sprintf("Successfully deleted %d items", count),
		"count":   count,
	}
	
	if len(errors) > 0 {
		result["errors"] = errors
	}
	
	return result, nil
}

// ExportCSVAction exports selected objects as CSV
func ExportCSVAction(ctx *gin.Context, objects []interface{}) (interface{}, error) {
	if len(objects) == 0 {
		return gin.H{"message": "No items selected for export", "count": 0}, nil
	}
	
	// Generate CSV content
	csvContent, err := generateCSV(objects)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSV: %w", err)
	}
	
	// Set headers for file download
	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", "attachment; filename=\"export.csv\"")
	
	return gin.H{
		"message": fmt.Sprintf("Exported %d items as CSV", len(objects)),
		"count":   len(objects),
		"data":    csvContent,
		"type":    "csv",
	}, nil
}

// ExportJSONAction exports selected objects as JSON
func ExportJSONAction(ctx *gin.Context, objects []interface{}) (interface{}, error) {
	if len(objects) == 0 {
		return gin.H{"message": "No items selected for export", "count": 0}, nil
	}
	
	// Set headers for file download
	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", "attachment; filename=\"export.json\"")
	
	return gin.H{
		"message": fmt.Sprintf("Exported %d items as JSON", len(objects)),
		"count":   len(objects),
		"data":    objects,
		"type":    "json",
	}, nil
}

// MarkActiveAction marks selected objects as active
func MarkActiveAction(ctx *gin.Context, objects []interface{}) (interface{}, error) {
	return updateFieldAction(ctx, objects, "active", true, "active")
}

// MarkInactiveAction marks selected objects as inactive
func MarkInactiveAction(ctx *gin.Context, objects []interface{}) (interface{}, error) {
	return updateFieldAction(ctx, objects, "active", false, "inactive")
}

// Helper functions

// extractObjectID extracts the ID from an object
func extractObjectID(obj interface{}) (string, error) {
	switch o := obj.(type) {
	case map[string]interface{}:
		if id, exists := o["id"]; exists {
			return fmt.Sprintf("%v", id), nil
		}
		return "", fmt.Errorf("no id field found")
	default:
		// Use reflection to find ID field
		// This would be more complex in a real implementation
		return "", fmt.Errorf("unsupported object type: %T", obj)
	}
}

// generateCSV generates CSV content from objects
func generateCSV(objects []interface{}) (string, error) {
	if len(objects) == 0 {
		return "", nil
	}
	
	// Get field names from first object
	var fieldNames []string
	if firstObj, ok := objects[0].(map[string]interface{}); ok {
		for key := range firstObj {
			fieldNames = append(fieldNames, key)
		}
	}
	
	// Build CSV content
	csv := ""
	
	// Header row
	for i, field := range fieldNames {
		if i > 0 {
			csv += ","
		}
		csv += fmt.Sprintf("\"%s\"", field)
	}
	csv += "\n"
	
	// Data rows
	for _, obj := range objects {
		if objMap, ok := obj.(map[string]interface{}); ok {
			for i, field := range fieldNames {
				if i > 0 {
					csv += ","
				}
				value := objMap[field]
				csv += fmt.Sprintf("\"%v\"", value)
			}
			csv += "\n"
		}
	}
	
	return csv, nil
}

// updateFieldAction is a helper for actions that update a field
func updateFieldAction(ctx *gin.Context, objects []interface{}, fieldName string, fieldValue interface{}, statusName string) (interface{}, error) {
	if len(objects) == 0 {
		return gin.H{"message": fmt.Sprintf("No items selected to mark as %s", statusName), "count": 0}, nil
	}
	
	// Extract model admin from context
	modelAdminInterface, exists := ctx.Get("model_admin")
	if !exists {
		return nil, fmt.Errorf("model admin not found in context")
	}
	
	modelAdmin, ok := modelAdminInterface.(*ModelAdmin)
	if !ok {
		return nil, fmt.Errorf("invalid model admin type")
	}
	
	count := 0
	errors := []string{}
	
	for _, obj := range objects {
		// Extract ID from object
		id, err := extractObjectID(obj)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to extract ID from object: %v", err))
			continue
		}
		
		// Create mock request with form data
		request := &http.Request{
			Method: "POST",
			Header: make(http.Header),
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		// Update object
		_, err = modelAdmin.UpdateObject(ctx, id, request)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to update object %s: %v", id, err))
			continue
		}
		count++
	}
	
	result := gin.H{
		"message": fmt.Sprintf("Successfully marked %d items as %s", count, statusName),
		"count":   count,
	}
	
	if len(errors) > 0 {
		result["errors"] = errors
	}
	
	return result, nil
}

// ActionContext provides additional context for actions
type ActionContext struct {
	Request     *http.Request
	User        interface{}
	ModelAdmin  *ModelAdmin
	Site        *Site
}

// ExtendedAction represents an action with additional context
type ExtendedAction struct {
	Action
	RequiresConfirmation bool
	ConfirmationMessage  string
	Permissions          []string
	Icon                 string
	CssClass            string
}

// ActionResult represents the result of an action execution
type ActionResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	Count        int           `json:"count"`
	Errors       []string      `json:"errors,omitempty"`
	Data         interface{}   `json:"data,omitempty"`
	Type         string        `json:"type,omitempty"`
	RedirectURL  string        `json:"redirect_url,omitempty"`
}

// ExecuteActionWithContext executes an action with additional context
func ExecuteActionWithContext(ctx *gin.Context, action Action, objects []interface{}, actionCtx *ActionContext) (*ActionResult, error) {
	// Set action context in gin context
	if actionCtx != nil {
		ctx.Set("action_context", actionCtx)
		if actionCtx.ModelAdmin != nil {
			ctx.Set("model_admin", actionCtx.ModelAdmin)
		}
	}
	
	result, err := action.Handler(ctx, objects)
	if err != nil {
		return &ActionResult{
			Success: false,
			Message: err.Error(),
			Count:   0,
		}, err
	}
	
	// Convert result to ActionResult
	if resultMap, ok := result.(gin.H); ok {
		actionResult := &ActionResult{
			Success: true,
		}
		
		if msg, exists := resultMap["message"]; exists {
			if msgStr, ok := msg.(string); ok {
				actionResult.Message = msgStr
			}
		}
		
		if count, exists := resultMap["count"]; exists {
			if countInt, ok := count.(int); ok {
				actionResult.Count = countInt
			}
		}
		
		if errors, exists := resultMap["errors"]; exists {
			if errorSlice, ok := errors.([]string); ok {
				actionResult.Errors = errorSlice
			}
		}
		
		if data, exists := resultMap["data"]; exists {
			actionResult.Data = data
		}
		
		if typ, exists := resultMap["type"]; exists {
			if typeStr, ok := typ.(string); ok {
				actionResult.Type = typeStr
			}
		}
		
		return actionResult, nil
	}
	
	return &ActionResult{
		Success: true,
		Message: "Action completed successfully",
		Count:   len(objects),
		Data:    result,
	}, nil
}

// ValidateActionPermissions checks if the user has permission to execute an action
func ValidateActionPermissions(ctx *gin.Context, action ExtendedAction, user interface{}) bool {
	// TODO: Implement permission checking based on action.Permissions
	// This would integrate with the authentication system
	return true
}

// FormatActionCount formats the count message for actions
func FormatActionCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}