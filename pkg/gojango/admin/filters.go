package admin

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Filter represents a list filter in the admin interface
type Filter interface {
	// Name returns the filter's name/field name
	Name() string
	
	// Title returns the human-readable title for the filter
	Title() string
	
	// Choices returns the available filter choices
	Choices() []FilterChoice
	
	// ApplyFilter applies the filter to the query parameters
	ApplyFilter(query url.Values) map[string]interface{}
	
	// IsActive checks if the filter is currently active
	IsActive(query url.Values) bool
	
	// GetActiveValue returns the currently active filter value
	GetActiveValue(query url.Values) interface{}
	
	// GetWidget returns the widget configuration for rendering
	GetWidget() FilterWidget
}

// FilterChoice represents a choice in a filter
type FilterChoice struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Count   int    `json:"count,omitempty"`
}

// FilterWidget represents the UI widget for a filter
type FilterWidget struct {
	Type        string                 `json:"type"`        // select, text, date, boolean, etc.
	Multiple    bool                   `json:"multiple"`    // Allow multiple selections
	Choices     []FilterChoice         `json:"choices"`
	Placeholder string                 `json:"placeholder"`
	Config      map[string]interface{} `json:"config"`
}

// BaseFilter provides common functionality for all filters
type BaseFilter struct {
	field       string
	title       string
	parameter   string
}

// NewBaseFilter creates a new base filter
func NewBaseFilter(field, title string) *BaseFilter {
	parameter := fmt.Sprintf("filter_%s", field)
	return &BaseFilter{
		field:     field,
		title:     title,
		parameter: parameter,
	}
}

func (f *BaseFilter) Name() string {
	return f.field
}

func (f *BaseFilter) Title() string {
	if f.title != "" {
		return f.title
	}
	// Convert field_name to Field Name
	parts := strings.Split(f.field, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func (f *BaseFilter) IsActive(query url.Values) bool {
	return query.Get(f.parameter) != ""
}

func (f *BaseFilter) GetActiveValue(query url.Values) interface{} {
	return query.Get(f.parameter)
}

// ChoiceFilter provides filtering by predefined choices
type ChoiceFilter struct {
	*BaseFilter
	choices []FilterChoice
}

// NewChoiceFilter creates a new choice filter
func NewChoiceFilter(field, title string, choices []FilterChoice) *ChoiceFilter {
	return &ChoiceFilter{
		BaseFilter: NewBaseFilter(field, title),
		choices:    choices,
	}
}

func (f *ChoiceFilter) Choices() []FilterChoice {
	allChoice := FilterChoice{
		Value:   "",
		Display: "All",
	}
	return append([]FilterChoice{allChoice}, f.choices...)
}

func (f *ChoiceFilter) ApplyFilter(query url.Values) map[string]interface{} {
	value := query.Get(f.parameter)
	if value == "" {
		return nil
	}
	
	return map[string]interface{}{
		f.field: value,
	}
}

func (f *ChoiceFilter) GetWidget() FilterWidget {
	return FilterWidget{
		Type:    "select",
		Choices: f.Choices(),
	}
}

// BooleanFilter provides filtering by boolean values
type BooleanFilter struct {
	*BaseFilter
}

// NewBooleanFilter creates a new boolean filter
func NewBooleanFilter(field, title string) *BooleanFilter {
	return &BooleanFilter{
		BaseFilter: NewBaseFilter(field, title),
	}
}

func (f *BooleanFilter) Choices() []FilterChoice {
	return []FilterChoice{
		{Value: "", Display: "All"},
		{Value: "true", Display: "Yes"},
		{Value: "false", Display: "No"},
	}
}

func (f *BooleanFilter) ApplyFilter(query url.Values) map[string]interface{} {
	value := query.Get(f.parameter)
	if value == "" {
		return nil
	}
	
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	
	return map[string]interface{}{
		f.field: boolValue,
	}
}

func (f *BooleanFilter) GetWidget() FilterWidget {
	return FilterWidget{
		Type:    "select",
		Choices: f.Choices(),
	}
}

// DateFilter provides filtering by date ranges
type DateFilter struct {
	*BaseFilter
}

// NewDateFilter creates a new date filter
func NewDateFilter(field, title string) *DateFilter {
	return &DateFilter{
		BaseFilter: NewBaseFilter(field, title),
	}
}

func (f *DateFilter) Choices() []FilterChoice {
	return []FilterChoice{
		{Value: "", Display: "Any time"},
		{Value: "today", Display: "Today"},
		{Value: "yesterday", Display: "Yesterday"},
		{Value: "week", Display: "Past 7 days"},
		{Value: "month", Display: "This month"},
		{Value: "year", Display: "This year"},
	}
}

func (f *DateFilter) ApplyFilter(query url.Values) map[string]interface{} {
	value := query.Get(f.parameter)
	if value == "" {
		return nil
	}
	
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	var startDate, endDate time.Time
	
	switch value {
	case "today":
		startDate = today
		endDate = today.Add(24 * time.Hour)
	case "yesterday":
		startDate = today.Add(-24 * time.Hour)
		endDate = today
	case "week":
		startDate = today.Add(-7 * 24 * time.Hour)
		endDate = now
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = now
	case "year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = now
	default:
		return nil
	}
	
	return map[string]interface{}{
		f.field + "__gte": startDate,
		f.field + "__lt":  endDate,
	}
}

func (f *DateFilter) GetWidget() FilterWidget {
	return FilterWidget{
		Type:    "select",
		Choices: f.Choices(),
	}
}

// TextFilter provides text-based filtering
type TextFilter struct {
	*BaseFilter
	exact bool
}

// NewTextFilter creates a new text filter
func NewTextFilter(field, title string, exact bool) *TextFilter {
	return &TextFilter{
		BaseFilter: NewBaseFilter(field, title),
		exact:      exact,
	}
}

func (f *TextFilter) Choices() []FilterChoice {
	return []FilterChoice{} // Text filters don't have predefined choices
}

func (f *TextFilter) ApplyFilter(query url.Values) map[string]interface{} {
	value := query.Get(f.parameter)
	if value == "" {
		return nil
	}
	
	if f.exact {
		return map[string]interface{}{
			f.field: value,
		}
	}
	
	return map[string]interface{}{
		f.field + "__icontains": value,
	}
}

func (f *TextFilter) GetWidget() FilterWidget {
	return FilterWidget{
		Type:        "text",
		Placeholder: fmt.Sprintf("Filter by %s", f.Title()),
	}
}

// NumericRangeFilter provides filtering by numeric ranges
type NumericRangeFilter struct {
	*BaseFilter
	min, max *float64
}

// NewNumericRangeFilter creates a new numeric range filter
func NewNumericRangeFilter(field, title string, min, max *float64) *NumericRangeFilter {
	return &NumericRangeFilter{
		BaseFilter: NewBaseFilter(field, title),
		min:        min,
		max:        max,
	}
}

func (f *NumericRangeFilter) Choices() []FilterChoice {
	return []FilterChoice{} // Range filters don't have predefined choices
}

func (f *NumericRangeFilter) ApplyFilter(query url.Values) map[string]interface{} {
	minValue := query.Get(f.parameter + "_min")
	maxValue := query.Get(f.parameter + "_max")
	
	filters := make(map[string]interface{})
	
	if minValue != "" {
		if min, err := strconv.ParseFloat(minValue, 64); err == nil {
			filters[f.field+"__gte"] = min
		}
	}
	
	if maxValue != "" {
		if max, err := strconv.ParseFloat(maxValue, 64); err == nil {
			filters[f.field+"__lte"] = max
		}
	}
	
	if len(filters) == 0 {
		return nil
	}
	
	return filters
}

func (f *NumericRangeFilter) GetWidget() FilterWidget {
	config := make(map[string]interface{})
	if f.min != nil {
		config["min"] = *f.min
	}
	if f.max != nil {
		config["max"] = *f.max
	}
	
	return FilterWidget{
		Type:   "numeric_range",
		Config: config,
	}
}

// FilterSpec represents a filter specification for auto-generation
type FilterSpec struct {
	Field       string
	FilterType  string
	Title       string
	Choices     []FilterChoice
	Config      map[string]interface{}
}

// AutoGenerateFilters automatically generates filters based on model fields
func AutoGenerateFilters(model interface{}) []Filter {
	var filters []Filter
	
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
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
		
		// Skip certain fields
		if fieldName == "id" || fieldName == "created_at" || fieldName == "updated_at" {
			continue
		}
		
		// Generate filter based on field type
		filter := generateFilterForField(fieldName, fieldType)
		if filter != nil {
			filters = append(filters, filter)
		}
	}
	
	return filters
}

// generateFilterForField generates an appropriate filter for a field type
func generateFilterForField(fieldName string, fieldType reflect.Type) Filter {
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	
	switch fieldType.Kind() {
	case reflect.Bool:
		return NewBooleanFilter(fieldName, "")
	case reflect.String:
		// Check if it looks like a status or category field
		if strings.Contains(fieldName, "status") || strings.Contains(fieldName, "type") || strings.Contains(fieldName, "category") {
			// This would typically come from database analysis or model metadata
			return NewChoiceFilter(fieldName, "", []FilterChoice{})
		}
		return NewTextFilter(fieldName, "", false)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		 reflect.Float32, reflect.Float64:
		return NewNumericRangeFilter(fieldName, "", nil, nil)
	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			return NewDateFilter(fieldName, "")
		}
	}
	
	return nil
}

// FilterSet manages a collection of filters for a model
type FilterSet struct {
	filters   []Filter
	filterMap map[string]Filter
}

// NewFilterSet creates a new filter set
func NewFilterSet() *FilterSet {
	return &FilterSet{
		filters:   []Filter{},
		filterMap: make(map[string]Filter),
	}
}

// AddFilter adds a filter to the set
func (fs *FilterSet) AddFilter(filter Filter) {
	fs.filters = append(fs.filters, filter)
	fs.filterMap[filter.Name()] = filter
}

// GetFilter gets a filter by name
func (fs *FilterSet) GetFilter(name string) (Filter, bool) {
	filter, exists := fs.filterMap[name]
	return filter, exists
}

// GetAllFilters returns all filters
func (fs *FilterSet) GetAllFilters() []Filter {
	return fs.filters
}

// ApplyFilters applies all active filters to generate query conditions
func (fs *FilterSet) ApplyFilters(query url.Values) map[string]interface{} {
	allFilters := make(map[string]interface{})
	
	for _, filter := range fs.filters {
		if filter.IsActive(query) {
			filterConditions := filter.ApplyFilter(query)
			for key, value := range filterConditions {
				allFilters[key] = value
			}
		}
	}
	
	return allFilters
}

// GetFilterData returns filter data for the frontend
func (fs *FilterSet) GetFilterData(query url.Values) []map[string]interface{} {
	var filterData []map[string]interface{}
	
	for _, filter := range fs.filters {
		data := map[string]interface{}{
			"name":        filter.Name(),
			"title":       filter.Title(),
			"widget":      filter.GetWidget(),
			"is_active":   filter.IsActive(query),
			"active_value": filter.GetActiveValue(query),
		}
		filterData = append(filterData, data)
	}
	
	return filterData
}