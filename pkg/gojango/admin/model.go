package admin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ModelAdmin defines the admin interface for a model
type ModelAdmin struct {
	model              interface{}
	modelName          string
	verboseName        string
	verboseNamePlural  string
	
	// Display options
	listDisplay        []string
	listDisplayLinks   []string
	listFilter         []string
	searchFields       []string
	ordering           []string
	
	// Form options
	fields             []string
	exclude            []string
	readonly           []string
	
	// Permissions
	permissions        map[string]bool
	
	// Pagination
	listPerPage        int
	maxShowAll         int
	
	// Actions
	actions            map[string]Action
	actionsOnTop       bool
	actionsOnBottom    bool
	
	// Custom methods
	listMethods        map[string]func(obj interface{}) interface{}
	formMethods        map[string]func(obj interface{}) interface{}
	
	// Database interface
	dbInterface        DatabaseInterface
}

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetAll(ctx context.Context, model interface{}, filters map[string]interface{}, ordering []string, limit, offset int) ([]interface{}, int, error)
	GetByID(ctx context.Context, model interface{}, id interface{}) (interface{}, error)
	Create(ctx context.Context, model interface{}, data map[string]interface{}) (interface{}, error)
	Update(ctx context.Context, model interface{}, id interface{}, data map[string]interface{}) (interface{}, error)
	Delete(ctx context.Context, model interface{}, id interface{}) error
	GetSchema(model interface{}) (*ModelSchema, error)
}

// ModelSchema represents the database schema for a model
type ModelSchema struct {
	Fields    []FieldSchema `json:"fields"`
	Relations []RelationSchema `json:"relations"`
}

// FieldSchema represents a database field
type FieldSchema struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	Unique       bool        `json:"unique"`
	MaxLength    *int        `json:"max_length,omitempty"`
	Choices      []Choice    `json:"choices,omitempty"`
	Default      interface{} `json:"default,omitempty"`
	HelpText     string      `json:"help_text,omitempty"`
	Verbose      string      `json:"verbose_name,omitempty"`
}

// RelationSchema represents a database relation
type RelationSchema struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // ForeignKey, ManyToMany, OneToOne
	RelatedModel string `json:"related_model"`
	RelatedName  string `json:"related_name,omitempty"`
}

// Choice represents a field choice
type Choice struct {
	Value   interface{} `json:"value"`
	Display string      `json:"display"`
}

// Action represents a bulk action
type Action struct {
	Name        string
	Description string
	Handler     func(ctx *gin.Context, objects []interface{}) (interface{}, error)
}

// ListData represents the data for admin list view
type ListData struct {
	Objects    []interface{} `json:"objects"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	HasNext    bool         `json:"has_next"`
	HasPrev    bool         `json:"has_prev"`
	NumPages   int          `json:"num_pages"`
	Filters    interface{}  `json:"filters"`
	Query      string       `json:"query"`
}

// NewModelAdmin creates a new ModelAdmin with default settings
func NewModelAdmin(model interface{}) *ModelAdmin {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	return &ModelAdmin{
		model:              model,
		verboseName:        modelType.Name(),
		verboseNamePlural:  modelType.Name() + "s",
		listDisplay:        []string{"__str__"},
		listDisplayLinks:   []string{},
		listFilter:         []string{},
		searchFields:       []string{},
		ordering:           []string{},
		fields:             []string{},
		exclude:            []string{},
		readonly:           []string{},
		permissions:        make(map[string]bool),
		listPerPage:        100,
		maxShowAll:         200,
		actions:            make(map[string]Action),
		actionsOnTop:       false,
		actionsOnBottom:    true,
		listMethods:        make(map[string]func(obj interface{}) interface{}),
		formMethods:        make(map[string]func(obj interface{}) interface{}),
	}
}

// SetDatabaseInterface sets the database interface for the model admin
func (ma *ModelAdmin) SetDatabaseInterface(db DatabaseInterface) {
	ma.dbInterface = db
}

// GetListData retrieves data for the admin list view
func (ma *ModelAdmin) GetListData(ctx *gin.Context, query url.Values) (*ListData, error) {
	if ma.dbInterface == nil {
		return nil, fmt.Errorf("database interface not set")
	}
	
	// Parse query parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	
	perPage := ma.listPerPage
	if p := query.Get("per_page"); p != "" {
		if pp, err := strconv.Atoi(p); err == nil && pp > 0 && pp <= ma.maxShowAll {
			perPage = pp
		}
	}
	
	searchQuery := query.Get("q")
	
	// Build filters
	filters := make(map[string]interface{})
	for key, values := range query {
		if strings.HasPrefix(key, "filter_") && len(values) > 0 {
			fieldName := strings.TrimPrefix(key, "filter_")
			filters[fieldName] = values[0]
		}
	}
	
	// Add search filters
	if searchQuery != "" && len(ma.searchFields) > 0 {
		searchFilters := make(map[string]interface{})
		for _, field := range ma.searchFields {
			searchFilters[field+"__icontains"] = searchQuery
		}
		filters["__search"] = searchFilters
	}
	
	offset := (page - 1) * perPage
	objects, total, err := ma.dbInterface.GetAll(ctx, ma.model, filters, ma.ordering, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get objects: %w", err)
	}
	
	numPages := (total + perPage - 1) / perPage
	
	return &ListData{
		Objects:  objects,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
		HasNext:  page < numPages,
		HasPrev:  page > 1,
		NumPages: numPages,
		Query:    searchQuery,
		Filters:  ma.getFilterData(ctx),
	}, nil
}

// GetAPIData retrieves data for API endpoints
func (ma *ModelAdmin) GetAPIData(ctx *gin.Context, query url.Values) (interface{}, error) {
	listData, err := ma.GetListData(ctx, query)
	if err != nil {
		return nil, err
	}
	
	return gin.H{
		"results":     listData.Objects,
		"count":       listData.Total,
		"page":        listData.Page,
		"per_page":    listData.PerPage,
		"num_pages":   listData.NumPages,
		"has_next":    listData.HasNext,
		"has_prev":    listData.HasPrev,
		"query":       listData.Query,
		"filters":     listData.Filters,
		"list_display": ma.listDisplay,
		"search_fields": ma.searchFields,
		"list_filter":  ma.listFilter,
		"actions":      ma.getActionsList(),
	}, nil
}

// GetObject retrieves a single object by ID
func (ma *ModelAdmin) GetObject(ctx *gin.Context, id string) (interface{}, error) {
	if ma.dbInterface == nil {
		return nil, fmt.Errorf("database interface not set")
	}
	
	return ma.dbInterface.GetByID(ctx, ma.model, id)
}

// CreateObject creates a new object
func (ma *ModelAdmin) CreateObject(ctx *gin.Context, request *http.Request) (interface{}, error) {
	if ma.dbInterface == nil {
		return nil, fmt.Errorf("database interface not set")
	}
	
	data, err := ma.extractFormData(request)
	if err != nil {
		return nil, fmt.Errorf("failed to extract form data: %w", err)
	}
	
	// Validate data
	if err := ma.validateData(data, true); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	return ma.dbInterface.Create(ctx, ma.model, data)
}

// UpdateObject updates an existing object
func (ma *ModelAdmin) UpdateObject(ctx *gin.Context, id string, request *http.Request) (interface{}, error) {
	if ma.dbInterface == nil {
		return nil, fmt.Errorf("database interface not set")
	}
	
	data, err := ma.extractFormData(request)
	if err != nil {
		return nil, fmt.Errorf("failed to extract form data: %w", err)
	}
	
	// Validate data
	if err := ma.validateData(data, false); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	return ma.dbInterface.Update(ctx, ma.model, id, data)
}

// DeleteObject deletes an object
func (ma *ModelAdmin) DeleteObject(ctx *gin.Context, id string) error {
	if ma.dbInterface == nil {
		return fmt.Errorf("database interface not set")
	}
	
	return ma.dbInterface.Delete(ctx, ma.model, id)
}

// ExecuteBulkAction executes a bulk action on selected objects
func (ma *ModelAdmin) ExecuteBulkAction(ctx *gin.Context, request *http.Request) (interface{}, error) {
	actionName := request.FormValue("action")
	if actionName == "" {
		return nil, fmt.Errorf("no action specified")
	}
	
	action, exists := ma.actions[actionName]
	if !exists {
		return nil, fmt.Errorf("unknown action: %s", actionName)
	}
	
	// Get selected object IDs
	selectedIDs := request.Form["_selected_action"]
	if len(selectedIDs) == 0 {
		return nil, fmt.Errorf("no objects selected")
	}
	
	// Get selected objects
	var objects []interface{}
	for _, id := range selectedIDs {
		obj, err := ma.GetObject(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", id, err)
		}
		objects = append(objects, obj)
	}
	
	return action.Handler(ctx, objects)
}

// GetSchema returns the model schema
func (ma *ModelAdmin) GetSchema() *ModelSchema {
	if ma.dbInterface == nil {
		return &ModelSchema{}
	}
	
	schema, _ := ma.dbInterface.GetSchema(ma.model)
	return schema
}

// GetPermissions returns the permissions for the current user
func (ma *ModelAdmin) GetPermissions(ctx *gin.Context) map[string]bool {
	// TODO: Implement actual permission checking
	return map[string]bool{
		"add":    true,
		"change": true,
		"delete": true,
		"view":   true,
	}
}

// Configuration methods
func (ma *ModelAdmin) SetListDisplay(fields ...string) *ModelAdmin {
	ma.listDisplay = fields
	return ma
}

func (ma *ModelAdmin) SetListFilter(fields ...string) *ModelAdmin {
	ma.listFilter = fields
	return ma
}

func (ma *ModelAdmin) SetSearchFields(fields ...string) *ModelAdmin {
	ma.searchFields = fields
	return ma
}

func (ma *ModelAdmin) SetOrdering(fields ...string) *ModelAdmin {
	ma.ordering = fields
	return ma
}

func (ma *ModelAdmin) SetListPerPage(count int) *ModelAdmin {
	ma.listPerPage = count
	return ma
}

func (ma *ModelAdmin) AddAction(name, description string, handler func(ctx *gin.Context, objects []interface{}) (interface{}, error)) *ModelAdmin {
	ma.actions[name] = Action{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
	return ma
}

// Helper methods
func (ma *ModelAdmin) extractFormData(request *http.Request) (map[string]interface{}, error) {
	if err := request.ParseForm(); err != nil {
		return nil, err
	}
	
	data := make(map[string]interface{})
	for key, values := range request.Form {
		if len(values) > 0 {
			data[key] = values[0]
		}
	}
	
	return data, nil
}

func (ma *ModelAdmin) validateData(data map[string]interface{}, isCreate bool) error {
	// TODO: Implement field validation based on model schema
	return nil
}

func (ma *ModelAdmin) getFilterData(ctx *gin.Context) interface{} {
	// TODO: Generate filter widget data based on list_filter
	filters := make(map[string]interface{})
	for _, field := range ma.listFilter {
		filters[field] = map[string]interface{}{
			"type": "text",
			"choices": []Choice{},
		}
	}
	return filters
}

func (ma *ModelAdmin) getActionsList() []map[string]interface{} {
	var actions []map[string]interface{}
	for name, action := range ma.actions {
		actions = append(actions, map[string]interface{}{
			"name":        name,
			"description": action.Description,
		})
	}
	return actions
}

// Default actions
func init() {
	// These will be added to every ModelAdmin by default
	// TODO: Implement default actions like delete_selected
}