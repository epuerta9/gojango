# Gojango Framework - Complete Development Handbook

## Overview

Gojango is a batteries-included Go web framework that brings Django's incredible developer experience to Go's performance and simplicity. It combines the best of Go's ecosystem with Django's conventions, providing a productive, type-safe, and scalable foundation for modern web applications.

**Core Philosophy:**
- **Convention over Configuration** - Like Django, but respecting Go idioms
- **Batteries Included** - Everything you need out of the box
- **Type Safety Everywhere** - From database to frontend
- **Code Generation over Magic** - Explicit, debuggable, fast
- **Multiple Paradigms** - HTML templates, REST, gRPC, GraphQL - your choice

## Architecture Components

### 1. Core Framework

```go
// The heart of Gojango
package gojango

type Application struct {
    // Core components
    Registry  *AppRegistry       // App registration system
    Router    *gin.Engine        // HTTP routing (Gin)
    RPC       *connect.Server    // gRPC/Connect server
    DB        *ent.Client        // Database ORM (Ent)
    Cache     *redis.Client      // Redis cache
    Queue     *asynq.Client      // Background jobs
    Signals   *SignalServer      // NATS-based signals
    
    // Configuration
    Settings  *Settings          // Starlark-based config
    
    // Admin
    Admin     *AdminSite         // Auto-generated admin
}
```

### 2. App System

Apps are self-contained modules (like Django apps):

```go
// apps/blog/app.go
package blog

import (
    "github.com/you/gojango"
    "myproject/apps/blog/schema"
)

func init() {
    // Auto-registration on import
    gojango.Register(&BlogApp{})
}

type BlogApp struct {
    ent *ent.Client
}

func (app *BlogApp) Config() gojango.AppConfig {
    return gojango.AppConfig{
        Name:         "blog",
        Label:        "Blog Application",
        Models:       []string{"Post", "Comment"},
        Dependencies: []string{"auth", "core"},
    }
}

func (app *BlogApp) Initialize(ctx *gojango.Context) error {
    app.ent = ctx.DB
    return nil
}

func (app *BlogApp) Routes(r *gin.RouterGroup) {
    r.GET("/", app.ListView)
    r.GET("/post/:id", app.DetailView)
}

func (app *BlogApp) Services() []gojango.Service {
    return []gojango.Service{
        &PostService{app: app},  // gRPC/Connect service
    }
}
```

### 3. Database Layer (Ent)

```go
// apps/blog/schema/post.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
    "github.com/you/gojango/admin"
)

type Post struct {
    ent.Schema
}

func (Post) Fields() []ent.Field {
    return []ent.Field{
        field.String("title").
            NotEmpty().
            MaxLen(200),
        field.Text("content"),
        field.String("slug").
            Unique(),
        field.Time("created_at").
            Default(time.Now),
        field.Enum("status").
            Values("draft", "published").
            Default("draft"),
    }
}

func (Post) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("author", User.Type).
            Ref("posts").
            Unique().
            Required(),
        edge.To("comments", Comment.Type),
        edge.To("tags", Tag.Type),
    }
}

func (Post) Annotations() []schema.Annotation {
    return []schema.Annotation{
        // Admin configuration
        admin.Config{
            ListDisplay:  []string{"title", "author", "status", "created_at"},
            SearchFields: []string{"title", "content"},
            ListFilter:   []string{"status", "author", "created_at"},
            Ordering:     []string{"-created_at"},
        },
        // Auto-generate gRPC service
        gojango.GenerateService(),
    }
}
```

### 4. API Layer

#### Connect/gRPC Services (Auto-generated)
```go
// Generated from Ent schema
type PostService struct {
    app *BlogApp
}

func (s *PostService) CreatePost(
    ctx context.Context,
    req *connect.Request[blogv1.CreatePostRequest],
) (*connect.Response[blogv1.Post], error) {
    // Automatic validation from proto
    post, err := s.app.ent.Post.
        Create().
        SetTitle(req.Msg.Title).
        SetContent(req.Msg.Content).
        SetAuthorID(ctx.Value("user").(*User).ID).
        Save(ctx)
        
    return connect.NewResponse(entToProto(post)), err
}
```

#### Traditional Views (Gin + Templ)
```go
// apps/blog/views.go
func (app *BlogApp) ListView(c *gin.Context) {
    posts, _ := app.ent.Post.
        Query().
        WithAuthor().
        Order(ent.Desc(post.FieldCreatedAt)).
        Limit(10).
        All(c)
    
    // Render with Templ
    html := templates.PostList(posts).Render(c.Request.Context())
    c.HTML(200, html)
}
```

### 5. Templates (Templ + HTMX)

```go
// apps/blog/templates/post_list.templ
package templates

import "myproject/ent"

templ PostList(posts []*ent.Post) {
    @Base() {
        <div class="posts" id="post-list">
            for _, post := range posts {
                @PostCard(post)
            }
        </div>
        
        <button hx-get="/blog?page=2" 
                hx-target="#post-list" 
                hx-swap="beforeend">
            Load More
        </button>
    }
}

templ PostCard(post *ent.Post) {
    <article class="post-card">
        <h2>{ post.Title }</h2>
        <p>By { post.Edges.Author.Name }</p>
        <div>{ post.Content }</div>
    </article>
}
```

### 6. Signals System (NATS-powered)

```go
// Cross-language signals via embedded NATS
package signals

// Define signals
var (
    PostPublished = gojango.Signal("blog.post_published")
    UserLoggedIn  = gojango.Signal("auth.user_logged_in")
)

// Send signal (from Go)
PostPublished.Send(map[string]interface{}{
    "post_id": post.ID,
    "title":   post.Title,
    "author":  post.AuthorID,
})

// Receive in Go
PostPublished.Connect(func(data map[string]interface{}) {
    // Clear cache, send email, etc.
    cache.Delete(fmt.Sprintf("post:%d", data["post_id"]))
})

// Receive in Python (Django app)
import nats

async def handle_post_published(msg):
    data = json.loads(msg.data)
    # Update search index, etc.
    
nc = await nats.connect("nats://localhost:4222")
await nc.subscribe("blog.post_published", cb=handle_post_published)
```

### 7. Background Jobs (Asynq)

```go
// apps/blog/tasks.go
package blog

import "github.com/you/gojango/tasks"

// Define task
var SendNewsletter = tasks.Task{
    Name: "blog:send_newsletter",
    Handler: func(ctx context.Context, t *asynq.Task) error {
        var payload struct {
            PostID int64 `json:"post_id"`
        }
        json.Unmarshal(t.Payload(), &payload)
        
        // Send newsletter
        post := app.ent.Post.Get(ctx, payload.PostID)
        subscribers := app.ent.User.Query().Where(user.Subscribed(true)).All(ctx)
        
        for _, sub := range subscribers {
            sendEmail(sub.Email, post)
        }
        return nil
    },
}

// Schedule task
SendNewsletter.Delay(map[string]interface{}{
    "post_id": post.ID,
}, tasks.In(5 * time.Minutes))
```

### 8. Settings System (Starlark)

```python
# config/settings.star
load("env", "env")
load("gojango", "secret_key")

# Environment
DEBUG = env.bool("DEBUG", True)
SECRET_KEY = env.get("SECRET_KEY", secret_key())

# Database
DATABASES = {
    "default": {
        "engine": "postgres",
        "host": env.get("DB_HOST", "localhost"),
        "port": env.int("DB_PORT", 5432),
        "name": env.get("DB_NAME", "gojango"),
        "user": env.get("DB_USER", "postgres"),
        "password": env.get("DB_PASSWORD", ""),
    }
}

# Installed apps
INSTALLED_APPS = [
    "gojango.contrib.auth",
    "gojango.contrib.admin",
    "gojango.contrib.sessions",
    
    # Your apps
    "apps.core",
    "apps.blog",
    "apps.api",
]

# Middleware
MIDDLEWARE = [
    "gojango.middleware.Security",
    "gojango.middleware.Session",
    "gojango.middleware.CSRF",
    "gojango.middleware.Auth",
    "gojango.middleware.Messages",
]

# NATS Configuration
NATS = {
    "enabled": True,
    "port": 4222,
    "jetstream": True,
    "store_dir": "./data/nats",
}

# Cache
CACHES = {
    "default": {
        "backend": "redis",
        "location": env.get("REDIS_URL", "redis://localhost:6379/0"),
    }
}

# Static files
STATIC_URL = "/static/"
STATIC_ROOT = "./staticfiles"

# Media files
MEDIA_URL = "/media/"
MEDIA_ROOT = "./media"

# Admin
ADMIN_SITE_HEADER = "Gojango Administration"

# Email
if DEBUG:
    EMAIL_BACKEND = "console"
else:
    EMAIL_BACKEND = "smtp"
    EMAIL_HOST = env.get("EMAIL_HOST")
    EMAIL_PORT = env.int("EMAIL_PORT", 587)
```

## CLI Commands

```bash
# Project management
gojango new myproject [--frontend react|htmx] [--api grpc|rest|graphql]
gojango run                    # Run development server
gojango build                   # Build for production

# App management
gojango startapp blog           # Create new app
gojango generate model Post     # Generate Ent model
gojango generate service Post   # Generate gRPC service
gojango generate admin Post     # Generate admin interface

# Database
gojango makemigrations [app]    # Create migrations
gojango migrate                  # Apply migrations
gojango dbshell                 # Database shell
gojango seed                    # Load fixtures

# Code generation
gojango generate all            # Run all generators
gojango generate ent            # Generate Ent code
gojango generate proto          # Generate protobuf
gojango generate typescript     # Generate TS client
gojango generate admin          # Generate admin

# Development tools
gojango shell                   # Interactive shell
gojango test [app]             # Run tests
gojango lint                   # Run linters
gojango format                 # Format code

# Admin
gojango createsuperuser        # Create admin user
gojango collectstatic          # Collect static files

# Background jobs
gojango worker                 # Start background worker
gojango scheduler              # Start task scheduler
gojango tasks list            # List scheduled tasks

# Signals/NATS
gojango signals list          # List all signals
gojango signals monitor       # Monitor signal traffic
gojango nats stats           # NATS server stats
```

## Project Structure

```
myproject/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── apps/                     # Your applications
│   ├── core/
│   │   ├── app.go           # App configuration
│   │   ├── schema/          # Ent schemas
│   │   │   └── user.go
│   │   ├── services/        # gRPC services
│   │   ├── views.go         # HTTP handlers
│   │   ├── templates/       # Templ templates
│   │   ├── static/          # App static files
│   │   └── migrations/      # SQL migrations
│   ├── blog/
│   └── admin/
├── config/
│   ├── settings.star        # Main settings
│   ├── settings_dev.star    # Dev overrides
│   └── settings_prod.star   # Prod settings
├── internal/
│   ├── ent/                # Generated Ent code
│   │   ├── schema/
│   │   └── migrate/
│   └── proto/              # Generated proto code
├── web/                    # Frontend (if React)
│   ├── src/
│   │   ├── api/           # Generated TS client
│   │   ├── components/
│   │   └── pages/
│   └── package.json
├── static/                 # Global static files
├── media/                  # User uploads
├── templates/              # Global templates
│   └── base.templ
├── proto/                  # Proto definitions
│   └── blog/
│       └── v1/
│           └── blog.proto
├── migrations/             # Database migrations
├── fixtures/               # Test data
├── tests/                  # Tests
├── docker-compose.yml      # Local development
├── Dockerfile             # Production build
├── Makefile               # Build commands
├── go.mod
├── go.sum
└── gojango.yaml           # Project config
```

## Full Example: Blog App

### 1. Create the project
```bash
gojango new myblog --frontend htmx --api grpc
cd myblog
```

### 2. Create blog app
```bash
gojango startapp blog
```

### 3. Define models
```go
// apps/blog/schema/post.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "github.com/you/gojango/admin"
)

type Post struct {
    ent.Schema
}

func (Post) Fields() []ent.Field {
    return []ent.Field{
        field.String("title").NotEmpty(),
        field.Text("content"),
        field.String("slug").Unique(),
        field.Time("published_at").Optional(),
        field.Enum("status").
            Values("draft", "published").
            Default("draft"),
    }
}

func (Post) Annotations() []schema.Annotation {
    return []schema.Annotation{
        admin.Config{
            ListDisplay: []string{"title", "status", "published_at"},
            SearchFields: []string{"title", "content"},
        },
        gojango.GenerateService(),
    }
}
```

### 4. Generate everything
```bash
gojango generate all
gojango migrate
```

### 5. Create views
```go
// apps/blog/views.go
package blog

func (app *BlogApp) ListView(c *gin.Context) {
    posts := app.ent.Post.Query().
        Where(post.Status(post.StatusPublished)).
        Order(ent.Desc(post.FieldPublishedAt)).
        All(c)
    
    c.HTML(200, templates.PostList(posts))
}
```

### 6. Run server
```bash
gojango run
# Visit http://localhost:8080
# Admin at http://localhost:8080/admin
```

## Development Roadmap

### Phase 1: Core Foundation (Weeks 1-4)
- [x] Project structure
- [ ] CLI tool (Cobra)
- [ ] App registry system
- [ ] Settings loader (Starlark)
- [ ] Basic routing (Gin)
- [ ] Ent integration
- [ ] Migration system

### Phase 2: Web Layer (Weeks 5-8)
- [ ] Template system (Templ)
- [ ] HTMX integration
- [ ] Static file handling
- [ ] Session management
- [ ] CSRF protection
- [ ] Middleware pipeline
- [ ] Form handling

### Phase 3: API Layer (Weeks 9-12)
- [ ] Connect/gRPC setup
- [ ] Proto generation from Ent
- [ ] TypeScript client generation
- [ ] GraphQL optional support
- [ ] OpenAPI generation
- [ ] API authentication

### Phase 4: Admin Interface (Weeks 13-16)
- [ ] Admin site structure
- [ ] Auto-registration from Ent
- [ ] CRUD interfaces
- [ ] List views with filters
- [ ] Form generation
- [ ] File uploads
- [ ] Permissions system

### Phase 5: Advanced Features (Weeks 17-20)
- [ ] NATS signals system
- [ ] Background jobs (Asynq)
- [ ] Cache framework
- [ ] Email system
- [ ] Testing framework
- [ ] Debugging toolbar
- [ ] Logging system

### Phase 6: Frontend Options (Weeks 21-24)
- [ ] React integration
- [ ] Vite setup
- [ ] Hot module reload
- [ ] TypeScript generation
- [ ] React component library
- [ ] Mobile app support (React Native)

### Phase 7: Production Ready (Weeks 25-28)
- [ ] Performance optimizations
- [ ] Security audit
- [ ] Documentation
- [ ] Example projects
- [ ] CI/CD templates
- [ ] Deployment guides
- [ ] Plugin system

## Getting Started (Once Built)

### Installation
```bash
# Install Gojango CLI
go install github.com/you/gojango/cmd/gojango@latest

# Create new project
gojango new myapp
cd myapp

# Install dependencies
go mod download
npm install  # If using React frontend

# Setup database
docker-compose up -d  # PostgreSQL + Redis + NATS
gojango migrate

# Create superuser
gojango createsuperuser

# Run development server
gojango run
```

### Your First App
```go
// apps/hello/app.go
package hello

import "github.com/you/gojango"

func init() {
    gojango.Register(&HelloApp{})
}

type HelloApp struct{}

func (app *HelloApp) Config() gojango.AppConfig {
    return gojango.AppConfig{Name: "hello"}
}

func (app *HelloApp) Routes(r *gin.RouterGroup) {
    r.GET("/", func(c *gin.Context) {
        c.HTML(200, "<h1>Hello, Gojango!</h1>")
    })
}
```

## Key Design Decisions

1. **Ent over GORM/sqlc**: Better code generation, admin metadata support
2. **Connect over pure gRPC**: Browser compatibility, same port as HTTP
3. **Templ over html/template**: Type-safe, better DX
4. **Starlark over YAML/TOML**: Logic in config when needed
5. **NATS for signals**: Cross-language, persistent, scalable
6. **Embedded everything**: Single binary deployment

## Comparison with Django

| Feature | Django | Gojango |
|---------|---------|---------|
| ORM | Django ORM (dynamic) | Ent (code-gen) |
| Templates | Django Templates | Templ (type-safe) |
| Admin | Automatic | Generated from schema |
| API | DRF (separate) | Built-in (gRPC/REST) |
| Migrations | Automatic | Semi-automatic |
| Forms | Class-based | Struct-based + generation |
| Signals | Python-only | Cross-language (NATS) |
| Background | Celery | Asynq (built-in) |
| Type Safety | Runtime | Compile-time |
| Performance | Good | Excellent |
| Deployment | Multiple services | Single binary |

## Contributing

The framework will be open source (MIT license). Key areas needing help:

1. **Core Framework**: App system, routing, middleware
2. **Admin Interface**: UI components, form generation
3. **Code Generation**: Ent → Proto → TypeScript pipeline
4. **Documentation**: Tutorials, examples, guides
5. **Ecosystem**: Plugins, integrations, tools

## Summary

Gojango brings Django's incredible developer experience to Go, providing:
- **Batteries included**: Everything you need out of the box
- **Type safety**: From database to frontend
- **Multiple paradigms**: HTML, REST, gRPC, GraphQL
- **Cross-language**: Signals work with Python, Node.js, etc.
- **Single binary**: Deploy one file with everything embedded
- **Amazing DX**: Hot reload, code generation, admin interface

The goal is to make Go web development as productive as Django while maintaining Go's simplicity, performance, and deployment story. By combining the best libraries in the Go ecosystem with Django's proven conventions, Gojango will be the framework Go developers have been waiting for.