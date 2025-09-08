package gojango

import (
	"github.com/epuerta9/gojango/pkg/gojango/admin"
)

// SetupAdmin sets up the admin interface for the application
func (app *Application) SetupAdmin() {
	// Setup admin routes with the Gin router
	admin.DefaultSite.SetupRoutes(app.GetRouter())
}

// RegisterAdminModel registers a model with the admin interface
func (app *Application) RegisterAdminModel(model interface{}, adminConfig *admin.ModelAdmin) error {
	return admin.Register(model, adminConfig)
}

// RegisterAdminModels registers multiple models with auto-generated admin configuration
func (app *Application) RegisterAdminModels(models ...interface{}) error {
	// This would integrate with Ent client when available
	return admin.AutoRegisterEntModels(nil, models...)
}

// GetAdminSite returns the default admin site
func (app *Application) GetAdminSite() *admin.Site {
	return admin.DefaultSite
}