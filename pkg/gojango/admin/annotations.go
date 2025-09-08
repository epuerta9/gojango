package admin

import (
	"entgo.io/ent/schema"
)

// Config annotation for Ent schema admin configuration
type Config struct {
	ListDisplay     []string
	SearchFields    []string
	ListFilter      []string
	Ordering        []string
	ListPerPage     int
	ReadonlyFields  []string
	DateHierarchy   string
	Fieldsets       []Fieldset
	Actions         []AdminAction
}

// Name returns the annotation name for Ent
func (Config) Name() string {
	return "AdminConfig"
}

// Fieldset represents a grouped set of fields in admin forms
type Fieldset struct {
	Name    string
	Fields  []string
	Classes []string
}

// AdminAction represents a bulk action in admin
type AdminAction struct {
	Name        string
	Description string
}

// Ensure Config implements schema.Annotation
var _ schema.Annotation = Config{}