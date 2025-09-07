# Gojango Core Framework - Internal Architecture & Implementation

## Core Framework Architecture

### 1. App Registration System

The app registry is the heart of Gojango, managing all installed apps and their lifecycle.

```go
// pkg/gojango/registry.go
package gojango

import (
    "fmt"
    "reflect"
    "sync"
)

// Global registry - singleton pattern
var (
    registry *Registry
    once     sync.Once
)

type Registry struct {
    mu       sync.RWMutex
    apps     map[string]App
    order    []string              // Registration order matters
    models   map[string]ModelMeta  // All models across apps
    routes   map[string][]Route    // URL patterns by app
    services map[string]Service    // gRPC services
    commands map[string]Command    // CLI commands
    
    // Lifecycle hooks
    preInit  []func() error
    postInit []func() error
}

func GetRegistry() *Registry {
    once.Do(func() {
        registry = &Registry{
            apps:     make(map[string]App),
            models:   make(map[string]ModelMeta),
            routes:   make(map[string][]Route),
            services: make(map[string]Service),
            commands: make(map[string]Command),
        }
    })
    return registry
}

// Register is called in app's init() function
func Register(app App) {
    r := GetRegistry()
    r.mu.Lock()
    defer r.mu.Unlock()
    
    config := app.Config()
    
    // Validate app
    if _, exists := r.apps[config.Name]; exists {
        panic(fmt.Sprintf("App '%s' already registered", config.Name))
    }
    
    // Check dependencies
    for _, dep := range config.Dependencies {
        if _, exists := r.apps[dep]; !exists {
            panic(fmt.Sprintf("App '%s' depends on '%s' which is not registered", 
                config.Name, dep))
        }
    }
    
    r.apps[config.Name] = app
    r.order = append(r.order, config.Name)
    
    // Register models if app implements ModelProvider
    if provider, ok := app.(ModelProvider); ok {
        for _, model := range provider.Models() {
            r.registerModel(config.Name, model)
        }
    }
    
    // Register services if app implements ServiceProvider
    if provider, ok := app.(ServiceProvider); ok {
        for _, service := range provider.Services() {
            r.services[service.Name()] = service
        }
    }
}

// Initialize is called when the application starts
func (r *Registry) Initialize(ctx *Context) error {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    // Run pre-init hooks
    for _, hook := range r.preInit {
        if err := hook(); err != nil {
            return fmt.Errorf("pre-init hook failed: %w", err)
        }
    }
    
    // Initialize apps in dependency order
    sorted := r.topologicalSort()
    for _, appName := range sorted {
        app := r.apps[appName]
        
        // Create app context
        appCtx := &AppContext{
            Name:     appName,
            DB:       ctx.DB,
            Cache:    ctx.Cache,
            Settings: ctx.Settings,
            Registry: r,
        }
        
        if err := app.Initialize(appCtx); err != nil {
            return fmt.Errorf("failed to initialize app '%s': %w", appName, err)
        }
        
        // Collect routes
        if router, ok := app.(RouterProvider); ok {
            r.routes[appName] = router.Routes()
        }
    }
    
    // Run post-init hooks
    for _, hook := range r.postInit {
        if err := hook(); err != nil {
            return fmt.Errorf("post-init hook failed: %w", err)
        }
    }
    
    return nil
}

// Model registration with metadata
func (r *Registry) registerModel(appName string, model interface{}) {
    modelType := reflect.TypeOf(model)
    modelName := modelType.Name()
    
    meta := ModelMeta{
        App:       appName,
        Name:      modelName,
        FullName:  fmt.Sprintf("%s.%s", appName, modelName),
        Type:      modelType,
        TableName: toSnakeCase(fmt.Sprintf("%s_%s", appName, modelName)),
    }
    
    // Extract admin config from struct tags or annotations
    if adminConfig := extractAdminConfig(model); adminConfig != nil {
        meta.Admin = adminConfig
    }
    
    r.models[meta.FullName] = meta
}
```

### 2. App Interface & Implementation

```go
// pkg/gojango/app.go
package gojango

// Core App interface - minimal requirements
type App interface {
    Config() AppConfig
    Initialize(*AppContext) error
}

// Optional interfaces apps can implement
type RouterProvider interface {
    Routes() []Route
}

type ModelProvider interface {
    Models() []interface{}
}

type ServiceProvider interface {
    Services() []Service
}

type CommandProvider interface {
    Commands() []*cobra.Command
}

type AdminProvider interface {
    AdminSite() *AdminSite
}

type SignalProvider interface {
    Signals() []Signal
}

// AppConfig defines app metadata
type AppConfig struct {
    Name         string
    Label        string
    Version      string
    Dependencies []string
    Settings     map[string]interface{}
}

// AppContext passed to app during initialization
type AppContext struct {
    Name     string
    DB       *ent.Client
    Cache    cache.Cache
    Queue    *asynq.Client
    Settings *Settings
    Registry *Registry
}

// Example app implementation
// apps/blog/app.go
package blog

import (
    "embed"
    "github.com/you/gojango"
    "myproject/apps/blog/schema"
    "myproject/ent"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed migrations/*.sql
var migrations embed.FS

func init() {
    gojango.Register(&BlogApp{})
}

type BlogApp struct {
    ent      *ent.Client
    settings *gojango.Settings
}

func (app *BlogApp) Config() gojango.AppConfig {
    return gojango.AppConfig{
        Name:    "blog",
        Label:   "Blog Application",
        Version: "1.0.0",
        Dependencies: []string{"auth", "core"},
        Settings: map[string]interface{}{
            "posts_per_page": 10,
            "enable_comments": true,
        },
    }
}

func (app *BlogApp) Initialize(ctx *gojango.AppContext) error {
    app.ent = ctx.DB
    app.settings = ctx.Settings
    
    // Register templates
    ctx.Registry.RegisterTemplates("blog", templateFS)
    
    // Register static files
    ctx.Registry.RegisterStatic("blog", staticFS)
    
    // Setup signals
    app.setupSignals(ctx)
    
    return nil
}

func (app *BlogApp) Routes() []gojango.Route {
    return []gojango.Route{
        {Method: "GET", Path: "/", Handler: app.ListView, Name: "blog:list"},
        {Method: "GET", Path: "/post/:slug", Handler: app.DetailView, Name: "blog:detail"},
        {Method: "GET", Path: "/tag/:tag", Handler: app.TagView, Name: "blog:tag"},
    }
}

func (app *BlogApp) Models() []interface{} {
    return []interface{}{
        &schema.Post{},
        &schema.Comment{},
        &schema.Tag{},
    }
}

func (app *BlogApp) Services() []gojango.Service {
    return []gojango.Service{
        &PostService{app: app},
        &CommentService{app: app},
    }
}
```

### 3. Multi-App Database Schema System

```go
// pkg/gojango/schema/manager.go
package schema

import (
    "fmt"
    "entgo.io/ent"
    "entgo.io/ent/schema"
)

// SchemaManager handles multi-app schemas
type SchemaManager struct {
    apps     map[string]*AppSchema
    prefixes map[string]string  // Table prefixes per app
}

type AppSchema struct {
    Name    string
    Models  []ent.Interface
    Prefix  string  // Table prefix (e.g., "blog_")
}

// GenerateSchema creates Ent schema with app prefixes
func (m *SchemaManager) GenerateSchema() error {
    for appName, appSchema := range m.apps {
        for _, model := range appSchema.Models {
            // Inject table prefix annotation
            annotations := model.Annotations()
            annotations = append(annotations, 
                entsql.Annotation{
                    Table: fmt.Sprintf("%s%s", 
                        appSchema.Prefix, 
                        toSnakeCase(modelName(model))),
                },
            )
        }
    }
    return nil
}

// Migration tracking per app
type MigrationTracker struct {
    db *sql.DB
}

func (mt *MigrationTracker) CreateMigrationTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS gojango_migrations (
            id SERIAL PRIMARY KEY,
            app VARCHAR(100) NOT NULL,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP DEFAULT NOW(),
            UNIQUE(app, name)
        )
    `
    _, err := mt.db.Exec(query)
    return err
}

func (mt *MigrationTracker) HasMigration(app, name string) (bool, error) {
    var exists bool
    err := mt.db.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM gojango_migrations WHERE app = $1 AND name = $2)",
        app, name,
    ).Scan(&exists)
    return exists, err
}

func (mt *MigrationTracker) RecordMigration(app, name string) error {
    _, err := mt.db.Exec(
        "INSERT INTO gojango_migrations (app, name) VALUES ($1, $2)",
        app, name,
    )
    return err
}
```

### 4. Route Registration & URL Reversal

```go
// pkg/gojango/routing/router.go
package routing

import (
    "fmt"
    "strings"
    "github.com/gin-gonic/gin"
)

type Router struct {
    engine     *gin.Engine
    routes     map[string]*Route  // name -> route
    middleware []gin.HandlerFunc
}

type Route struct {
    Method  string
    Path    string
    Handler gin.HandlerFunc
    Name    string
    App     string
}

func (r *Router) RegisterApp(appName string, routes []Route) {
    group := r.engine.Group("/" + appName)
    
    for _, route := range routes {
        route.App = appName
        fullName := fmt.Sprintf("%s:%s", appName, route.Name)
        r.routes[fullName] = &route
        
        // Register with Gin
        switch route.Method {
        case "GET":
            group.GET(route.Path, route.Handler)
        case "POST":
            group.POST(route.Path, route.Handler)
        // ... other methods
        }
    }
}

// URL reversal like Django's reverse()
func (r *Router) Reverse(name string, params ...interface{}) string {
    route, exists := r.routes[name]
    if !exists {
        panic(fmt.Sprintf("Route '%s' not found", name))
    }
    
    path := route.Path
    
    // Replace parameters
    for i, param := range params {
        placeholder := fmt.Sprintf(":param%d", i)
        path = strings.Replace(path, placeholder, fmt.Sprint(param), 1)
    }
    
    return path
}

// Template function for URL reversal
func (r *Router) TemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "url": r.Reverse,
        "static": func(path string) string {
            return "/static/" + path
        },
        "media": func(path string) string {
            return "/media/" + path
        },
    }
}
```

### 5. Admin Model Generation

```go
// pkg/gojango/admin/generator.go
package admin

import (
    "reflect"
    "entgo.io/ent"
)

// AdminConfig extracted from model annotations
type AdminConfig struct {
    ListDisplay   []string
    ListFilter    []string
    SearchFields  []string
    Ordering      []string
    ReadonlyFields []string
    Fields        []string
    Fieldsets     []Fieldset
    InlineModels  []string
    Actions       []Action
}

type ModelAdmin struct {
    Model       interface{}
    Config      AdminConfig
    ent         *ent.Client
    modelType   reflect.Type
}

// GenerateAdmin creates admin from Ent schema
func GenerateAdmin(model interface{}) *ModelAdmin {
    modelType := reflect.TypeOf(model)
    config := extractAdminConfig(model)
    
    return &ModelAdmin{
        Model:     model,
        Config:    config,
        modelType: modelType,
    }
}

func extractAdminConfig(model interface{}) AdminConfig {
    config := AdminConfig{}
    
    // Use reflection to find admin method or struct tags
    modelType := reflect.TypeOf(model)
    
    // Check for Admin() method
    if method, exists := modelType.MethodByName("Admin"); exists {
        result := method.Func.Call([]reflect.Value{reflect.ValueOf(model)})
        if len(result) > 0 {
            if adminConfig, ok := result[0].Interface().(AdminConfig); ok {
                return adminConfig
            }
        }
    }
    
    // Parse struct tags
    for i := 0; i < modelType.NumField(); i++ {
        field := modelType.Field(i)
        
        if adminTag := field.Tag.Get("admin"); adminTag != "" {
            // Parse admin tag
            parseAdminTag(&config, field.Name, adminTag)
        }
    }
    
    return config
}

// Auto-generate list view
func (ma *ModelAdmin) ListView(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    search := c.Query("search")
    
    // Build query
    query := ma.ent.Client.Query()
    
    // Apply search
    if search != "" && len(ma.Config.SearchFields) > 0 {
        // Build OR conditions for search fields
        predicates := []predicate.Predicate{}
        for _, field := range ma.Config.SearchFields {
            predicates = append(predicates, 
                predicate.Contains(field, search))
        }
        query = query.Where(predicate.Or(predicates...))
    }
    
    // Apply ordering
    for _, order := range ma.Config.Ordering {
        if strings.HasPrefix(order, "-") {
            query = query.Order(ent.Desc(order[1:]))
        } else {
            query = query.Order(ent.Asc(order))
        }
    }
    
    // Paginate
    items, _ := query.
        Limit(50).
        Offset((page - 1) * 50).
        All(c)
    
    // Render admin template
    c.HTML(200, "admin/list.html", gin.H{
        "model":       ma.modelType.Name(),
        "items":       items,
        "config":      ma.Config,
        "search":      search,
        "page":        page,
    })
}

// Auto-generate form
func (ma *ModelAdmin) FormView(c *gin.Context) {
    id := c.Param("id")
    
    var instance interface{}
    if id != "" {
        // Edit existing
        instance, _ = ma.ent.Query().Where(predicate.ID(id)).First(c)
    } else {
        // Create new
        instance = reflect.New(ma.modelType).Interface()
    }
    
    // Generate form fields
    form := ma.generateForm(instance)
    
    c.HTML(200, "admin/form.html", gin.H{
        "model":    ma.modelType.Name(),
        "instance": instance,
        "form":     form,
        "config":   ma.Config,
    })
}
```

### 6. Settings System with Starlark

```go
// pkg/gojango/settings/loader.go
package settings

import (
    "fmt"
    "os"
    "go.starlark.net/starlark"
    "go.starlark.net/starlarkstruct"
)

type Settings struct {
    data     starlark.StringDict
    globals  starlark.StringDict
    filename string
}

func LoadSettings(filename string) (*Settings, error) {
    // Define built-in functions available in settings
    predeclared := starlark.StringDict{
        "env":        starlark.NewBuiltin("env", envFunc),
        "secret_key": starlark.NewBuiltin("secret_key", secretKeyFunc),
        "load_file":  starlark.NewBuiltin("load_file", loadFileFunc),
    }
    
    // Load and execute the Starlark file
    thread := &starlark.Thread{Name: "settings"}
    globals, err := starlark.ExecFile(thread, filename, nil, predeclared)
    if err != nil {
        return nil, fmt.Errorf("failed to load settings: %w", err)
    }
    
    s := &Settings{
        data:     globals,
        globals:  predeclared,
        filename: filename,
    }
    
    // Process INSTALLED_APPS
    if err := s.processInstalledApps(); err != nil {
        return nil, err
    }
    
    return s, nil
}

// Get retrieves a setting value
func (s *Settings) Get(key string, defaultValue ...interface{}) interface{} {
    if val, exists := s.data[key]; exists {
        return starlarkToGo(val)
    }
    if len(defaultValue) > 0 {
        return defaultValue[0]
    }
    return nil
}

// Process INSTALLED_APPS to register apps
func (s *Settings) processInstalledApps() error {
    appsVal, exists := s.data["INSTALLED_APPS"]
    if !exists {
        return fmt.Errorf("INSTALLED_APPS not defined in settings")
    }
    
    appsList, ok := appsVal.(*starlark.List)
    if !ok {
        return fmt.Errorf("INSTALLED_APPS must be a list")
    }
    
    // Extract app names
    var apps []string
    iter := appsList.Iterate()
    defer iter.Done()
    
    var val starlark.Value
    for iter.Next(&val) {
        if appStr, ok := val.(starlark.String); ok {
            apps = append(apps, string(appStr))
        }
    }
    
    // Validate all apps are registered
    registry := GetRegistry()
    for _, appName := range apps {
        if !registry.HasApp(appName) {
            return fmt.Errorf("app '%s' in INSTALLED_APPS is not registered", appName)
        }
    }
    
    return nil
}

// Built-in functions for Starlark
func envFunc(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
    var key string
    var defaultVal starlark.Value
    
    if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "default?", &defaultVal); err != nil {
        return nil, err
    }
    
    if val := os.Getenv(key); val != "" {
        return starlark.String(val), nil
    }
    
    if defaultVal != nil {
        return defaultVal, nil
    }
    
    return starlark.None, nil
}
```

### 7. Plugin System

```go
// pkg/gojango/plugin/plugin.go
package plugin

import (
    "github.com/you/gojango"
)

// Plugin interface
type Plugin interface {
    Name() string
    Version() string
    Initialize(*gojango.Context) error
    Hooks() Hooks
}

type Hooks struct {
    // Request lifecycle
    PreRequest   func(*gin.Context) error
    PostRequest  func(*gin.Context) error
    
    // Model lifecycle
    PreSave      func(model interface{}) error
    PostSave     func(model interface{}) error
    PreDelete    func(model interface{}) error
    PostDelete   func(model interface{}) error
    
    // Admin customization
    AdminMenu    func() []MenuItem
    AdminWidgets func() []Widget
}

// PluginManager manages all plugins
type PluginManager struct {
    plugins map[string]Plugin
    hooks   map[string][]func(interface{}) error
}

func (pm *PluginManager) Register(plugin Plugin) {
    pm.plugins[plugin.Name()] = plugin
    
    // Register hooks
    hooks := plugin.Hooks()
    if hooks.PreSave != nil {
        pm.addHook("pre_save", hooks.PreSave)
    }
    if hooks.PostSave != nil {
        pm.addHook("post_save", hooks.PostSave)
    }
    // ... register other hooks
}

// Example plugin
type AuditLogPlugin struct{}

func (p *AuditLogPlugin) Name() string { return "audit_log" }
func (p *AuditLogPlugin) Version() string { return "1.0.0" }

func (p *AuditLogPlugin) Initialize(ctx *gojango.Context) error {
    // Create audit log table
    ctx.DB.Schema.Create(
        ctx.Context,
        migrate.Table("audit_logs"),
        migrate.Columns(
            &migrate.Column{Name: "id", Type: field.TypeInt},
            &migrate.Column{Name: "user_id", Type: field.TypeInt},
            &migrate.Column{Name: "action", Type: field.TypeString},
            &migrate.Column{Name: "model", Type: field.TypeString},
            &migrate.Column{Name: "object_id", Type: field.TypeInt},
            &migrate.Column{Name: "changes", Type: field.TypeJSON},
            &migrate.Column{Name: "created_at", Type: field.TypeTime},
        ),
    )
    return nil
}

func (p *AuditLogPlugin) Hooks() Hooks {
    return Hooks{
        PostSave: func(model interface{}) error {
            // Log the save operation
            return logAudit("save", model)
        },
        PostDelete: func(model interface{}) error {
            // Log the delete operation
            return logAudit("delete", model)
        },
    }
}
```

### 8. How Users Use The Framework

```go
// main.go - User's application entry point
package main

import (
    "log"
    
    "github.com/you/gojango"
    
    // Import apps - their init() functions register them
    _ "myproject/apps/core"
    _ "myproject/apps/auth"
    _ "myproject/apps/blog"
    _ "myproject/apps/shop"
    
    // Import plugins
    _ "github.com/you/gojango-debug-toolbar"
    _ "github.com/you/gojango-celery"
)

func main() {
    // Create application
    app := gojango.New()
    
    // Load settings
    if err := app.LoadSettings("config/settings.star"); err != nil {
        log.Fatal(err)
    }
    
    // The app now has everything from init() registrations
    // Execute CLI - this handles runserver, migrate, etc
    if err := app.Execute(); err != nil {
        log.Fatal(err)
    }
}

// apps/blog/models.go
package blog

//go:generate gojango generate ent

// This generates:
// - Ent schema
// - Admin interface
// - gRPC service
// - GraphQL resolvers

// apps/blog/views.go
package blog

import (
    "github.com/gin-gonic/gin"
    "github.com/you/gojango"
)

// Traditional view
func (app *BlogApp) ListView(c *gin.Context) {
    // Get pagination
    page := gojango.Paginate(c, 10)
    
    // Query with automatic user filtering
    posts := app.ent.Post.
        Query().
        Where(post.StatusEQ(post.StatusPublished)).
        WithAuthor().
        Order(ent.Desc(post.FieldCreatedAt)).
        Limit(page.Limit).
        Offset(page.Offset).
        All(c)
    
    // Render with context
    gojango.Render(c, "blog/list.html", gojango.H{
        "posts": posts,
        "page":  page,
        "user":  gojango.GetUser(c),  // Automatic from middleware
    })
}

// API view (automatically generates OpenAPI spec)
func (app *BlogApp) APIListView(c *gin.Context) {
    var req ListPostsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        gojango.BadRequest(c, err)
        return
    }
    
    posts := app.queryPosts(req)
    
    gojango.JSON(c, posts)  // Automatic serialization
}
```

### 9. Database Schema Declaration

```go
// apps/blog/schema/post.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/index"
    "github.com/you/gojango/admin"
    "github.com/you/gojango/api"
)

type Post struct {
    ent.Schema
}

// Mixin for common fields
func (Post) Mixin() []ent.Mixin {
    return []ent.Mixin{
        gojango.TimestampMixin{},  // created_at, updated_at
        gojango.SoftDeleteMixin{},  // deleted_at
    }
}

func (Post) Fields() []ent.Field {
    return []ent.Field{
        field.String("title").
            NotEmpty().
            MaxLen(200).
            Comment("Post title"),
            
        field.Text("content").
            Comment("Post content in Markdown"),
            
        field.String("slug").
            Unique().
            Immutable().
            Comment("URL slug"),
            
        field.Enum("status").
            Values("draft", "published", "archived").
            Default("draft"),
            
        field.Time("published_at").
            Optional().
            Nillable(),
    }
}

func (Post) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("author", User.Type).
            Ref("posts").
            Required().
            Unique().
            Comment("Post author"),
            
        edge.To("comments", Comment.Type).
            Annotations(
                admin.Inline(),  // Show inline in admin
            ),
            
        edge.To("tags", Tag.Type),
    }
}

func (Post) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("slug"),
        index.Fields("status", "published_at"),
        index.Fields("created_at").Desc(),
    }
}

func (Post) Annotations() []schema.Annotation {
    return []schema.Annotation{
        // Table name (app prefix added automatically)
        entsql.Annotation{Table: "posts"},
        
        // Admin configuration
        admin.Config{
            ListDisplay:   []string{"title", "author", "status", "published_at"},
            ListFilter:    []string{"status", "author", "tags"},
            SearchFields:  []string{"title", "content"},
            Ordering:      []string{"-published_at"},
            DateHierarchy: "published_at",
            Actions:       []string{"publish", "archive"},
            
            Fieldsets: []admin.Fieldset{
                {
                    Name: "Basic Information",
                    Fields: []string{"title", "slug", "author"},
                },
                {
                    Name: "Content",
                    Fields: []string{"content", "tags"},
                },
                {
                    Name: "Publishing",
                    Fields: []string{"status", "published_at"},
                },
            },
        },
        
        // API configuration
        api.Config{
            Methods: []string{"GET", "POST", "PUT", "DELETE"},
            Fields:  []string{"id", "title", "content", "author", "status"},
            Authentication: api.AuthRequired,
            Permissions: []string{"blog.view_post", "blog.add_post"},
            RateLimit: "100/hour",
        },
        
        // GraphQL configuration  
        gojango.GraphQL{
            GenerateQuery:    true,
            GenerateMutation: true,
            GenerateSubscription: true,
        },
    }
}

// Custom methods
func (Post) Hooks() []ent.Hook {
    return []ent.Hook{
        // Auto-generate slug
        hook.On(
            func(next ent.Mutator) ent.Mutator {
                return hook.PostFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
                    if title, exists := m.Field("title"); exists {
                        m.SetField("slug", slugify(title.(string)))
                    }
                    return next.Mutate(ctx, m)
                })
            },
            ent.OpCreate,
        ),
        
        // Clear cache on save
        hook.On(
            func(next ent.Mutator) ent.Mutator {
                return hook.PostFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
                    v, err := next.Mutate(ctx, m)
                    if err == nil {
                        gojango.Signal("post_saved").Send(v)
                    }
                    return v, err
                })
            },
            ent.OpUpdate|ent.OpUpdateOne,
        ),
    }
}
```

### 10. Multi-App Schema Management

```go
// pkg/gojango/schema/multi_app.go
package schema

// Each app's tables are prefixed
// blog app tables: blog_posts, blog_comments
// shop app tables: shop_products, shop_orders

type SchemaBuilder struct {
    apps map[string]*AppSchemaBuilder
}

type AppSchemaBuilder struct {
    name   string
    prefix string
    models []ent.Interface
}

func (sb *SchemaBuilder) Build() (*ent.Schema, error) {
    var allModels []ent.Interface
    
    for appName, app := range sb.apps {
        for _, model := range app.models {
            // Inject app prefix into model
            wrapModelWithPrefix(model, app.prefix)
            allModels = append(allModels, model)
        }
    }
    
    // Generate migrations per app
    for appName, app := range sb.apps {
        if err := sb.generateAppMigrations(appName, app); err != nil {
            return nil, err
        }
    }
    
    return ent.NewSchema(allModels...), nil
}

// Foreign keys between apps
type CrossAppRelation struct {
    From    string // e.g., "blog.Post"
    To      string // e.g., "auth.User"
    Field   string // e.g., "author"
    OnDelete string // CASCADE, SET_NULL, etc.
}

func (sb *SchemaBuilder) AddCrossAppRelation(rel CrossAppRelation) {
    // Generate appropriate foreign key
    fromApp, fromModel := parseModelPath(rel.From)
    toApp, toModel := parseModelPath(rel.To)
    
    fromTable := fmt.Sprintf("%s_%s", fromApp, toSnakeCase(fromModel))
    toTable := fmt.Sprintf("%s_%s", toApp, toSnakeCase(toModel))
    
    // Add foreign key constraint
    sb.constraints = append(sb.constraints, 
        fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY (%s_id) REFERENCES %s(id) ON DELETE %s",
            fromTable, rel.Field, toTable, rel.OnDelete))
}
```