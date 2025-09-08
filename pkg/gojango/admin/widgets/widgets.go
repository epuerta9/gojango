// Package widgets provides form widgets for the Gojango admin interface.
//
// Widgets are responsible for rendering form fields in the admin interface
// and handling their input/output. They work similarly to Django's form widgets
// but are designed for React-based frontends.
package widgets

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Widget represents a form widget
type Widget interface {
	// Render returns the widget configuration for frontend rendering
	Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig

	// GetMediaFiles returns CSS/JS files needed by the widget
	GetMediaFiles() MediaFiles

	// FormatValue formats a value for display
	FormatValue(value interface{}) interface{}

	// ValueFromForm extracts value from form data
	ValueFromForm(formData map[string]interface{}, name string) (interface{}, error)
}

// WidgetConfig represents the configuration sent to the frontend
type WidgetConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Value      interface{}            `json:"value"`
	Attributes map[string]interface{} `json:"attrs"`
	Choices    []Choice               `json:"choices,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Required   bool                   `json:"required"`
	HelpText   string                 `json:"help_text,omitempty"`
	Label      string                 `json:"label,omitempty"`
}

// Choice represents an option in a select widget
type Choice struct {
	Value   interface{} `json:"value"`
	Display string      `json:"display"`
	Group   string      `json:"group,omitempty"`
}

// MediaFiles represents CSS and JavaScript files
type MediaFiles struct {
	CSS []string `json:"css"`
	JS  []string `json:"js"`
}

// BaseWidget provides common functionality for all widgets
type BaseWidget struct {
	attrs      map[string]interface{}
	mediaFiles MediaFiles
}

// NewBaseWidget creates a new base widget
func NewBaseWidget() *BaseWidget {
	return &BaseWidget{
		attrs: make(map[string]interface{}),
		mediaFiles: MediaFiles{
			CSS: []string{},
			JS:  []string{},
		},
	}
}

func (w *BaseWidget) GetMediaFiles() MediaFiles {
	return w.mediaFiles
}

func (w *BaseWidget) FormatValue(value interface{}) interface{} {
	if value == nil {
		return ""
	}
	return value
}

func (w *BaseWidget) ValueFromForm(formData map[string]interface{}, name string) (interface{}, error) {
	return formData[name], nil
}

// TextInput widget
type TextInput struct {
	*BaseWidget
}

// NewTextInput creates a new text input widget
func NewTextInput() *TextInput {
	return &TextInput{
		BaseWidget: NewBaseWidget(),
	}
}

func (w *TextInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	// Merge default attrs with provided attrs
	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	return WidgetConfig{
		Type:       "text",
		Name:       name,
		Value:      w.FormatValue(value),
		Attributes: mergedAttrs,
	}
}

// NumberInput widget
type NumberInput struct {
	*BaseWidget
	min, max *float64
	step     *float64
}

// NewNumberInput creates a new number input widget
func NewNumberInput() *NumberInput {
	return &NumberInput{
		BaseWidget: NewBaseWidget(),
	}
}

func (w *NumberInput) SetRange(min, max float64) *NumberInput {
	w.min = &min
	w.max = &max
	return w
}

func (w *NumberInput) SetStep(step float64) *NumberInput {
	w.step = &step
	return w
}

func (w *NumberInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	config := make(map[string]interface{})
	if w.min != nil {
		config["min"] = *w.min
	}
	if w.max != nil {
		config["max"] = *w.max
	}
	if w.step != nil {
		config["step"] = *w.step
	}

	return WidgetConfig{
		Type:       "number",
		Name:       name,
		Value:      w.FormatValue(value),
		Attributes: mergedAttrs,
		Config:     config,
	}
}

func (w *NumberInput) ValueFromForm(formData map[string]interface{}, name string) (interface{}, error) {
	value, exists := formData[name]
	if !exists {
		return nil, nil
	}

	strValue, ok := value.(string)
	if !ok {
		return value, nil
	}

	if strValue == "" {
		return nil, nil
	}

	// Try to parse as float first, then int
	if floatValue, err := strconv.ParseFloat(strValue, 64); err == nil {
		if intValue, err := strconv.Atoi(strValue); err == nil && float64(intValue) == floatValue {
			return intValue, nil
		}
		return floatValue, nil
	}

	return nil, fmt.Errorf("invalid number format: %s", strValue)
}

// Textarea widget
type Textarea struct {
	*BaseWidget
	rows, cols int
}

// NewTextarea creates a new textarea widget
func NewTextarea() *Textarea {
	return &Textarea{
		BaseWidget: NewBaseWidget(),
		rows:       4,
		cols:       50,
	}
}

func (w *Textarea) SetSize(rows, cols int) *Textarea {
	w.rows = rows
	w.cols = cols
	return w
}

func (w *Textarea) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	mergedAttrs["rows"] = w.rows
	mergedAttrs["cols"] = w.cols

	return WidgetConfig{
		Type:       "textarea",
		Name:       name,
		Value:      w.FormatValue(value),
		Attributes: mergedAttrs,
	}
}

// CheckboxInput widget
type CheckboxInput struct {
	*BaseWidget
}

// NewCheckboxInput creates a new checkbox input widget
func NewCheckboxInput() *CheckboxInput {
	return &CheckboxInput{
		BaseWidget: NewBaseWidget(),
	}
}

func (w *CheckboxInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	// Convert value to boolean
	boolValue := false
	if value != nil {
		switch v := value.(type) {
		case bool:
			boolValue = v
		case string:
			boolValue = v == "true" || v == "1" || v == "on"
		case int:
			boolValue = v != 0
		}
	}

	return WidgetConfig{
		Type:       "checkbox",
		Name:       name,
		Value:      boolValue,
		Attributes: mergedAttrs,
	}
}

func (w *CheckboxInput) ValueFromForm(formData map[string]interface{}, name string) (interface{}, error) {
	value, exists := formData[name]
	if !exists {
		return false, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true" || v == "1" || v == "on", nil
	default:
		return false, nil
	}
}

// Select widget
type Select struct {
	*BaseWidget
	choices []Choice
}

// NewSelect creates a new select widget
func NewSelect() *Select {
	return &Select{
		BaseWidget: NewBaseWidget(),
		choices:    []Choice{},
	}
}

func (w *Select) SetChoices(choices []Choice) *Select {
	w.choices = choices
	return w
}

func (w *Select) AddChoice(value interface{}, display string) *Select {
	w.choices = append(w.choices, Choice{
		Value:   value,
		Display: display,
	})
	return w
}

func (w *Select) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	return WidgetConfig{
		Type:       "select",
		Name:       name,
		Value:      w.FormatValue(value),
		Attributes: mergedAttrs,
		Choices:    w.choices,
	}
}

// SelectMultiple widget
type SelectMultiple struct {
	*Select
}

// NewSelectMultiple creates a new multi-select widget
func NewSelectMultiple() *SelectMultiple {
	return &SelectMultiple{
		Select: NewSelect(),
	}
}

func (w *SelectMultiple) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	config := w.Select.Render(name, value, attrs)
	config.Type = "select_multiple"
	config.Attributes["multiple"] = true
	return config
}

func (w *SelectMultiple) ValueFromForm(formData map[string]interface{}, name string) (interface{}, error) {
	value, exists := formData[name]
	if !exists {
		return []interface{}{}, nil
	}

	// Handle multiple values
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	default:
		return []interface{}{v}, nil
	}
}

// DateInput widget
type DateInput struct {
	*BaseWidget
	format string
}

// NewDateInput creates a new date input widget
func NewDateInput() *DateInput {
	return &DateInput{
		BaseWidget: NewBaseWidget(),
		format:     "2006-01-02",
	}
}

func (w *DateInput) SetFormat(format string) *DateInput {
	w.format = format
	return w
}

func (w *DateInput) FormatValue(value interface{}) interface{} {
	if value == nil {
		return ""
	}

	if t, ok := value.(time.Time); ok {
		return t.Format(w.format)
	}

	return value
}

func (w *DateInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	return WidgetConfig{
		Type:       "date",
		Name:       name,
		Value:      w.FormatValue(value),
		Attributes: mergedAttrs,
		Config: map[string]interface{}{
			"format": w.format,
		},
	}
}

func (w *DateInput) ValueFromForm(formData map[string]interface{}, name string) (interface{}, error) {
	value, exists := formData[name]
	if !exists {
		return nil, nil
	}

	strValue, ok := value.(string)
	if !ok || strValue == "" {
		return nil, nil
	}

	return time.Parse(w.format, strValue)
}

// DateTimeInput widget
type DateTimeInput struct {
	*DateInput
}

// NewDateTimeInput creates a new datetime input widget
func NewDateTimeInput() *DateTimeInput {
	return &DateTimeInput{
		DateInput: &DateInput{
			BaseWidget: NewBaseWidget(),
			format:     "2006-01-02T15:04:05",
		},
	}
}

func (w *DateTimeInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	config := w.DateInput.Render(name, value, attrs)
	config.Type = "datetime-local"
	return config
}

// FileInput widget
type FileInput struct {
	*BaseWidget
	accept string
}

// NewFileInput creates a new file input widget
func NewFileInput() *FileInput {
	return &FileInput{
		BaseWidget: NewBaseWidget(),
	}
}

func (w *FileInput) SetAccept(accept string) *FileInput {
	w.accept = accept
	return w
}

func (w *FileInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	if w.accept != "" {
		mergedAttrs["accept"] = w.accept
	}

	return WidgetConfig{
		Type:       "file",
		Name:       name,
		Value:      nil, // File inputs don't have values for security reasons
		Attributes: mergedAttrs,
	}
}

// HiddenInput widget
type HiddenInput struct {
	*BaseWidget
}

// NewHiddenInput creates a new hidden input widget
func NewHiddenInput() *HiddenInput {
	return &HiddenInput{
		BaseWidget: NewBaseWidget(),
	}
}

func (w *HiddenInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	return WidgetConfig{
		Type:       "hidden",
		Name:       name,
		Value:      w.FormatValue(value),
		Attributes: mergedAttrs,
	}
}

// PasswordInput widget
type PasswordInput struct {
	*BaseWidget
	renderValue bool
}

// NewPasswordInput creates a new password input widget
func NewPasswordInput() *PasswordInput {
	return &PasswordInput{
		BaseWidget:  NewBaseWidget(),
		renderValue: false,
	}
}

func (w *PasswordInput) SetRenderValue(render bool) *PasswordInput {
	w.renderValue = render
	return w
}

func (w *PasswordInput) Render(name string, value interface{}, attrs map[string]interface{}) WidgetConfig {
	mergedAttrs := make(map[string]interface{})

	for k, v := range w.attrs {
		mergedAttrs[k] = v
	}
	for k, v := range attrs {
		mergedAttrs[k] = v
	}

	displayValue := ""
	if w.renderValue {
		displayValue = fmt.Sprintf("%v", w.FormatValue(value))
	}

	return WidgetConfig{
		Type:       "password",
		Name:       name,
		Value:      displayValue,
		Attributes: mergedAttrs,
	}
}

// Widget registry for auto-selection based on field types
var WidgetRegistry = map[string]func() Widget{
	"string":   func() Widget { return NewTextInput() },
	"text":     func() Widget { return NewTextarea() },
	"integer":  func() Widget { return NewNumberInput() },
	"float":    func() Widget { return NewNumberInput() },
	"boolean":  func() Widget { return NewCheckboxInput() },
	"date":     func() Widget { return NewDateInput() },
	"datetime": func() Widget { return NewDateTimeInput() },
	"email":    func() Widget { return NewTextInput() },
	"url":      func() Widget { return NewTextInput() },
	"password": func() Widget { return NewPasswordInput() },
	"file":     func() Widget { return NewFileInput() },
	"hidden":   func() Widget { return NewHiddenInput() },
	"select":   func() Widget { return NewSelect() },
	"multiple": func() Widget { return NewSelectMultiple() },
}

// GetWidgetForType returns an appropriate widget for a field type
func GetWidgetForType(fieldType string) Widget {
	if factory, exists := WidgetRegistry[fieldType]; exists {
		return factory()
	}

	// Default to text input
	return NewTextInput()
}

// AutoDetectWidget detects the appropriate widget based on Go type
func AutoDetectWidget(value reflect.Type) Widget {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		return NewTextInput()
	case reflect.Bool:
		return NewCheckboxInput()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewNumberInput()
	case reflect.Float32, reflect.Float64:
		return NewNumberInput()
	case reflect.Struct:
		if value == reflect.TypeOf(time.Time{}) {
			return NewDateTimeInput()
		}
		return NewTextInput()
	case reflect.Slice:
		return NewSelectMultiple()
	default:
		return NewTextInput()
	}
}
