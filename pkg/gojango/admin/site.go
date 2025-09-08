// Package admin provides Django-style admin interface for Gojango applications.
//
// The admin package automatically generates CRUD interfaces from Ent schemas,
// provides bulk actions, filtering, and a modern React-based UI similar to
// PocketBase and Django admin.
package admin

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/epuerta9/gojango/pkg/gojango/admin/proto/protoconnect"
)

// Site represents the admin site that manages all registered models
type Site struct {
	mu           sync.RWMutex
	models       map[string]*ModelAdmin
	name         string
	headerTitle  string
	indexTitle   string
	siteURL      string
	enableLogin  bool
	permissions  PermissionChecker
	entClient    interface{} // Global Ent client for database operations
}

// PermissionChecker defines interface for checking admin permissions
type PermissionChecker interface {
	HasPermission(user interface{}, perm string, obj interface{}) bool
	HasAddPermission(user interface{}, model string) bool
	HasChangePermission(user interface{}, obj interface{}) bool
	HasDeletePermission(user interface{}, obj interface{}) bool
	HasViewPermission(user interface{}, obj interface{}) bool
}

// NewSite creates a new admin site
func NewSite(name string) *Site {
	return &Site{
		models:      make(map[string]*ModelAdmin),
		name:        name,
		headerTitle: "Gojango Administration",
		indexTitle:  "Site Administration",
		siteURL:     "/",
		enableLogin: true,
	}
}

// DefaultSite is the default admin site instance
var DefaultSite = NewSite("admin")

// SetEntClient sets the Ent client for database operations
func SetEntClient(client interface{}) {
	DefaultSite.mu.Lock()
	defer DefaultSite.mu.Unlock()
	DefaultSite.entClient = client
}

// Register registers a model with its admin configuration
func (s *Site) Register(model interface{}, admin *ModelAdmin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	modelName := getModelName(model)
	if modelName == "" {
		return fmt.Errorf("unable to determine model name for %T", model)
	}

	if admin == nil {
		admin = NewModelAdmin(model)
	}
	admin.model = model
	admin.modelName = modelName

	s.models[modelName] = admin
	return nil
}

// Unregister removes a model from the admin site
func (s *Site) Unregister(model interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	modelName := getModelName(model)
	delete(s.models, modelName)
}

// GetModelAdmin returns the admin configuration for a model
func (s *Site) GetModelAdmin(modelName string) (*ModelAdmin, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	admin, exists := s.models[modelName]
	return admin, exists
}

// GetRegisteredModels returns all registered model names
func (s *Site) GetRegisteredModels() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	models := make([]string, 0, len(s.models))
	for name := range s.models {
		models = append(models, name)
	}
	return models
}

// SetupRoutes configures admin routes with the given Gin router
func (s *Site) SetupRoutes(router gin.IRouter) {
	adminGroup := router.Group("/admin")
	
	// Setup basic API routes for testing
	s.setupBasicAPIRoutes(adminGroup)
	
	// Static files for React admin (using relative path from project root)
	adminGroup.StaticFS("/static", http.Dir("../../pkg/gojango/admin/templates/static"))
	adminGroup.StaticFS("/assets", http.Dir("../../pkg/gojango/admin/templates/dist/assets"))
	
	// Serve the React app for specific admin paths (avoid conflicts with /api)
	adminGroup.GET("/", s.handleReactApp)
	adminGroup.GET("/dashboard", s.handleReactApp)
	adminGroup.GET("/dashboard/*path", s.handleReactApp)
	
	// Handle model routes - both with and without app prefix for convenience
	adminGroup.GET("/:app/:model/", s.handleModelList)
	adminGroup.GET("/:app/:model/add/", s.handleReactApp)
	adminGroup.GET("/:app/:model/:id/", s.handleReactApp)
	adminGroup.GET("/:app/:model/:id/change/", s.handleReactApp)
	
	// Handle direct model access (default to 'main' app)
	adminGroup.GET("/users", s.handleReactApp)
	adminGroup.GET("/users/*path", s.handleReactApp)
	adminGroup.GET("/posts", s.handleReactApp)  
	adminGroup.GET("/posts/*path", s.handleReactApp)
	adminGroup.GET("/categories", s.handleReactApp)
	adminGroup.GET("/categories/*path", s.handleReactApp)
	adminGroup.GET("/comments", s.handleReactApp)
	adminGroup.GET("/comments/*path", s.handleReactApp)
}

// setupBasicAPIRoutes sets up basic API routes for testing without gRPC
func (s *Site) setupBasicAPIRoutes(adminGroup gin.IRouter) {
	apiGroup := adminGroup.Group("/api")
	
	// Models endpoint  
	apiGroup.GET("/models/", s.handleAPIModelsList)
	
	// gRPC-Web endpoints for Connect protocol  
	if routerGroup, ok := adminGroup.(*gin.RouterGroup); ok {
		s.registerConnectHandlers(routerGroup)
	}
}

// registerConnectHandlers registers the Connect-Web gRPC handlers
func (s *Site) registerConnectHandlers(group *gin.RouterGroup) {
	// Create the gRPC service handler with Ent client
	bridge := NewEntBridge(s.entClient)
	handler := NewAdminServiceHandler(s, bridge)
	handler.SetEntClient(s.entClient)
	
	// Create the Connect service
	path, connectHandler := protoconnect.NewAdminServiceHandler(handler)
	
	// Register the Connect handler with Gin
	// Connect uses POST requests for all RPCs
	group.POST(path+"*method", gin.WrapH(connectHandler))
	group.GET(path+"*method", gin.WrapH(connectHandler))  // For some Connect clients
}

// handleReactApp serves the React admin application
func (s *Site) handleReactApp(c *gin.Context) {
	// Read and serve the built index.html file
	indexPath := "../../pkg/gojango/admin/templates/dist/index.html"
	htmlContent, err := os.ReadFile(indexPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load admin interface: %v", err)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	c.Writer.Write(htmlContent)
}

func (s *Site) handleModelList(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.HTML(http.StatusNotFound, "", gin.H{})
		c.Writer.WriteString(fmt.Sprintf(`
		<html><head><title>Model Not Found</title></head><body>
		<h1>Model Not Found</h1>
		<p>The model "%s" was not found.</p>
		<a href="/admin/">← Back to Admin</a>
		</body></html>`, modelKey))
		return
	}
	
	// Modern model list view template with sidebar
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s | Gojango Admin</title>
    <style>
      * { box-sizing: border-box; }
      body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif; margin: 0; background: #f8fafc; color: #1e293b; line-height: 1.6; }
      
      /* Layout */
      .admin-layout { display: flex; min-height: 100vh; }
      
      /* Sidebar */
      .sidebar { width: 280px; background: #1e293b; color: white; position: fixed; height: 100vh; overflow-y: auto; z-index: 1000; box-shadow: 2px 0 4px rgba(0,0,0,0.1); }
      .sidebar-header { padding: 24px; border-bottom: 1px solid #334155; }
      .sidebar-header h1 { margin: 0; font-size: 20px; font-weight: 700; }
      .sidebar-header p { margin: 8px 0 0; font-size: 14px; opacity: 0.8; }
      
      .sidebar-nav { padding: 16px 0; }
      .nav-section { margin-bottom: 32px; }
      .nav-section-title { padding: 8px 24px 12px; font-size: 11px; font-weight: 600; text-transform: uppercase; opacity: 0.6; letter-spacing: 0.1em; }
      .nav-link { display: flex; align-items: center; padding: 12px 24px; color: #cbd5e1; text-decoration: none; transition: all 0.2s ease; }
      .nav-link:hover { background: #334155; color: white; }
      .nav-link.active { background: #3b82f6; color: white; }
      .nav-link-icon { width: 20px; height: 20px; margin-right: 12px; text-align: center; font-size: 16px; }
      .nav-link-text { flex: 1; }
      
      /* Main content */
      .main-content { flex: 1; margin-left: 280px; background: #f8fafc; }
      .top-bar { background: white; border-bottom: 1px solid #e2e8f0; padding: 20px 32px; display: flex; justify-content: space-between; align-items: center; box-shadow: 0 1px 2px rgba(0,0,0,0.05); }
      .page-title { margin: 0; font-size: 24px; font-weight: 700; color: #1e293b; }
      .page-subtitle { margin: 4px 0 0; font-size: 14px; color: #64748b; }
      .breadcrumb { font-size: 14px; color: #64748b; margin-bottom: 4px; }
      .breadcrumb a { color: #3b82f6; text-decoration: none; }
      .breadcrumb a:hover { text-decoration: underline; }
      .content { padding: 32px; max-width: none; }
      
      /* Content sections */
      .info-card { background: white; border-radius: 12px; padding: 24px; border: 1px solid #e2e8f0; box-shadow: 0 1px 2px rgba(0,0,0,0.05); margin-bottom: 24px; }
      .info-card-header { display: flex; align-items: center; margin-bottom: 16px; }
      .info-card-icon { width: 48px; height: 48px; background: #dbeafe; color: #3b82f6; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 20px; margin-right: 16px; }
      .info-card h3 { margin: 0; font-size: 18px; font-weight: 600; }
      .config-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 16px; margin: 20px 0; }
      .config-item { background: #f8fafc; padding: 16px; border-radius: 8px; border-left: 4px solid #3b82f6; }
      .config-label { font-size: 12px; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 4px; }
      .config-value { font-size: 14px; color: #1e293b; }
      
      /* Buttons */
      .btn { display: inline-flex; align-items: center; justify-content: center; padding: 10px 16px; background: #3b82f6; color: white; text-decoration: none; border-radius: 8px; font-size: 14px; font-weight: 500; transition: all 0.2s ease; border: none; cursor: pointer; margin-right: 12px; }
      .btn:hover { background: #2563eb; transform: translateY(-1px); box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3); }
      .btn-secondary { background: #6b7280; }
      .btn-secondary:hover { background: #4b5563; box-shadow: 0 2px 4px rgba(107, 114, 128, 0.3); }
      .btn-outline { background: transparent; color: #3b82f6; border: 1px solid #3b82f6; }
      .btn-outline:hover { background: #3b82f6; color: white; }
      
      .actions-bar { display: flex; align-items: center; margin-top: 24px; }
      
      /* Future table styling */
      .table-container { background: white; border-radius: 12px; border: 1px solid #e2e8f0; overflow: hidden; box-shadow: 0 1px 2px rgba(0,0,0,0.05); }
      .table { width: 100%%; border-collapse: collapse; }
      .table th, .table td { padding: 12px 16px; text-align: left; border-bottom: 1px solid #e2e8f0; }
      .table th { background: #f8fafc; font-weight: 600; color: #374151; }
      .table tbody tr:hover { background: #f8fafc; }
      
      /* Responsive */
      @media (max-width: 1024px) {
        .sidebar { transform: translateX(-100%%); }
        .main-content { margin-left: 0; }
        .content { padding: 20px; }
        .config-grid { grid-template-columns: 1fr; }
      }
    </style>
</head>
<body>
    <div class="admin-layout">
      <!-- Sidebar -->
      <div class="sidebar">
        <div class="sidebar-header">
          <h1>Gojango Admin</h1>
          <p>Site Administration</p>
        </div>
        <nav class="sidebar-nav">
          <div class="nav-section">
            <div class="nav-section-title">Dashboard</div>
            <a href="/admin/" class="nav-link">
              <span class="nav-link-icon">🏠</span>
              <span class="nav-link-text">Overview</span>
            </a>
          </div>
          <div class="nav-section">
            <div class="nav-section-title">Models</div>
            %s
          </div>
          <div class="nav-section">
            <div class="nav-section-title">Tools</div>
            <a href="/admin/api/models/" class="nav-link" target="_blank">
              <span class="nav-link-icon">🔗</span>
              <span class="nav-link-text">API Endpoint</span>
            </a>
          </div>
        </nav>
      </div>
      
      <!-- Main Content -->
      <div class="main-content">
        <div class="top-bar">
          <div>
            <div class="breadcrumb">
              <a href="/admin/">Dashboard</a> › %s
            </div>
            <h1 class="page-title">%s</h1>
            <p class="page-subtitle">Model: %s.%s</p>
          </div>
        </div>
        <div class="content">
          <div class="info-card">
            <div class="info-card-header">
              <div class="info-card-icon">🚧</div>
              <div>
                <h3>List View Coming Soon!</h3>
                <p style="margin: 4px 0 0; font-size: 14px; color: #64748b;">This will show a fully featured data table with pagination, search, filters, and bulk actions.</p>
              </div>
            </div>
            
            <div class="config-grid">
              <div class="config-item">
                <div class="config-label">List Display Fields</div>
                <div class="config-value">%s</div>
              </div>
              <div class="config-item">
                <div class="config-label">Search Fields</div>
                <div class="config-value">%s</div>
              </div>
              <div class="config-item">
                <div class="config-label">List Filters</div>
                <div class="config-value">%s</div>
              </div>
              <div class="config-item">
                <div class="config-label">Custom Actions</div>
                <div class="config-value">%d actions available</div>
              </div>
            </div>
            
            <p style="margin: 20px 0 0; color: #64748b; font-size: 14px;">
              This would show a paginated table with all <strong>%s</strong> objects, complete with search functionality, column sorting, advanced filters, and bulk actions for data management.
            </p>
          </div>
          
          <div class="actions-bar">
            <a href="/admin/" class="btn btn-secondary">← Back to Dashboard</a>
            <a href="/admin/api/models/" class="btn-outline" target="_blank">🔗 View API</a>
          </div>
        </div>
      </div>
    </div>
</body>
</html>`
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	// Generate complete sidebar navigation with all models
	currentModelKey := fmt.Sprintf("%s.%s", app, model)
	navLinksHTML := ""
	
	s.mu.RLock()
	for name, modelAdmin := range s.models {
		parts := strings.Split(name, ".")
		modelApp := "main"
		modelName := name
		if len(parts) == 2 {
			modelApp = parts[0]
			modelName = parts[1]
		}
		
		// Get model icon
		icon := "📊" // Default icon
		if strings.Contains(strings.ToLower(modelAdmin.verboseNamePlural), "user") {
			icon = "👥"
		} else if strings.Contains(strings.ToLower(modelAdmin.verboseNamePlural), "post") {
			icon = "📝"
		} else if strings.Contains(strings.ToLower(modelAdmin.verboseNamePlural), "category") {
			icon = "🏷️"
		}
		
		// Check if this is the active model
		activeClass := ""
		if name == currentModelKey {
			activeClass = " active"
		}
		
		navLinksHTML += fmt.Sprintf(`
		<a href="/admin/%s/%s/" class="nav-link%s">
			<span class="nav-link-icon">%s</span>
			<span class="nav-link-text">%s</span>
		</a>`, modelApp, modelName, activeClass, icon, modelAdmin.verboseNamePlural)
	}
	s.mu.RUnlock()
	
	c.Writer.WriteString(fmt.Sprintf(tmpl,
		admin.verboseNamePlural, // title
		navLinksHTML, // complete sidebar navigation
		admin.verboseNamePlural, // breadcrumb
		admin.verboseNamePlural, // page title
		app, model, // subtitle
		strings.Join(admin.listDisplay, ", "), // list display
		strings.Join(admin.searchFields, ", "), // search fields
		strings.Join(admin.listFilter, ", "), // list filters
		len(admin.actions), // actions count
		strings.ToLower(admin.verboseNamePlural), // description
	))
}

func (s *Site) handleModelAdd(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	c.HTML(http.StatusOK, "admin/change_form.html", gin.H{
		"admin": admin,
		"app":   app,
		"model": model,
		"isAdd": true,
	})
}

func (s *Site) handleModelCreate(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	// Create new instance through model admin
	obj, err := admin.CreateObject(c, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"object": obj})
}

func (s *Site) handleModelDetail(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	id := c.Param("id")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	obj, err := admin.GetObject(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}
	
	c.HTML(http.StatusOK, "admin/change_form.html", gin.H{
		"admin":  admin,
		"object": obj,
		"app":    app,
		"model":  model,
		"isAdd":  false,
	})
}

func (s *Site) handleModelEdit(c *gin.Context) {
	// Same as detail for now
	s.handleModelDetail(c)
}

func (s *Site) handleModelUpdate(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	id := c.Param("id")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	obj, err := admin.UpdateObject(c, id, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"object": obj})
}

func (s *Site) handleModelDelete(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	id := c.Param("id")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	err := admin.DeleteObject(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Site) handleBulkAction(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	result, err := admin.ExecuteBulkAction(c, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// API handlers for React frontend
func (s *Site) handleAPIModelsList(c *gin.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	models := make(map[string]interface{})
	for name, admin := range s.models {
		parts := strings.Split(name, ".")
		app := "main"
		model := name
		if len(parts) == 2 {
			app = parts[0]
			model = parts[1]
		}
		
		models[name] = gin.H{
			"name":               model,
			"app":                app,
			"verbose_name":       admin.verboseName,
			"verbose_name_plural": admin.verboseNamePlural,
			"list_display":       admin.listDisplay,
			"search_fields":      admin.searchFields,
			"list_filter":        admin.listFilter,
			"permissions":        admin.GetPermissions(c),
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"site": gin.H{
			"name":         s.name,
			"header_title": s.headerTitle,
			"index_title":  s.indexTitle,
		},
	})
}

func (s *Site) handleAPIModelData(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	data, err := admin.GetAPIData(c, c.Request.URL.Query())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, data)
}

func (s *Site) handleAPIModelSchema(c *gin.Context) {
	app := c.Param("app")
	model := c.Param("model")
	modelKey := fmt.Sprintf("%s.%s", app, model)
	
	admin, exists := s.GetModelAdmin(modelKey)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	schema := admin.GetSchema()
	c.JSON(http.StatusOK, schema)
}

func (s *Site) handleAPIModelCreate(c *gin.Context) {
	s.handleModelCreate(c)
}

func (s *Site) handleAPIModelUpdate(c *gin.Context) {
	s.handleModelUpdate(c)
}

func (s *Site) handleAPIModelDelete(c *gin.Context) {
	s.handleModelDelete(c)
}

// Helper functions
func getModelName(model interface{}) string {
	if model == nil {
		return ""
	}
	
	// Get type name
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	// Format as app.model (assuming package name is app name)
	pkg := t.PkgPath()
	typeName := t.Name()
	
	// Extract app name from package path
	parts := strings.Split(pkg, "/")
	appName := "main"
	if len(parts) > 0 {
		appName = parts[len(parts)-1]
		// Remove common suffixes
		if strings.HasSuffix(appName, "_models") || strings.HasSuffix(appName, "_ent") {
			parts := strings.Split(appName, "_")
			if len(parts) > 1 {
				appName = parts[0]
			}
		}
	}
	
	return fmt.Sprintf("%s.%s", appName, strings.ToLower(typeName))
}

// Convenience functions
func Register(model interface{}, admin *ModelAdmin) error {
	return DefaultSite.Register(model, admin)
}

func Unregister(model interface{}) {
	DefaultSite.Unregister(model)
}

func SetupRoutes(router gin.IRouter) {
	DefaultSite.SetupRoutes(router)
}