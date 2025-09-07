# Gojango Development Roadmap

*The idiomatic fullstack framework for Go developers*

## Vision

Create a Django-like framework for Go that combines Django's incredible developer experience with Go's performance, type safety, and simplicity. Gojango brings batteries-included web development to Go while respecting Go idioms and leveraging the best of Go's ecosystem.

## Core Principles

- **Convention over Configuration** - Like Django, but respecting Go idioms
- **Batteries Included** - Everything you need out of the box
- **Type Safety Everywhere** - From database to frontend
- **Code Generation over Magic** - Explicit, debuggable, fast
- **Multiple Paradigms** - HTML templates, REST, gRPC, GraphQL - your choice

## Phase 1: Foundation (Current) - v0.1.0
*Weeks 1-2: The absolute minimum to be useful*

### Goal
Create a working project structure generator and basic app system that developers can start building with immediately.

### Features to Build

#### ✅ Core Framework Components
- [x] **Basic app interface** - Simple App interface with Config() and Initialize() methods
- [x] **App registry system** - Global registry to manage app lifecycle and dependencies
- [x] **Minimal application struct** - Core Application struct to orchestrate everything
- [x] **Basic settings loader** - Simple environment variable loader (Starlark comes later)
- [x] **HTTP routing integration** - Gin-based routing with URL patterns and reversal

#### ✅ CLI Tool (Global Installation)
- [x] **'gojango new' command** - Generate new project with customizable options
- [x] **'gojango version' command** - Version information
- [x] **Interactive project setup** - CLI prompts for project configuration
- [x] **Template system** - Embedded templates for code generation

#### ✅ Generated Project Structure
- [x] **main.go generation** - Application entry point with proper imports
- [x] **Basic Makefile** - Developer convenience commands
- [x] **go.mod setup** - Proper Go module initialization
- [x] **Project metadata** - gojango.yaml for project configuration

### Implementation Status

#### 🎯 Currently Working On
- **Core app registry system** - The heart that manages all apps
- **Basic HTTP routing** - Gin integration with URL patterns
- **Project generation CLI** - Templates and interactive setup

#### ✅ Completed
- Project structure analysis and planning
- Development documentation and handbook
- Framework architecture design

### What Users Can Do After Phase 1

```bash
# Install Gojango globally
go install github.com/epuerta9/gojango/cmd/gojango@latest

# Create a new project
gojango new myproject
cd myproject

# The generated project structure:
myproject/
├── apps/                  # Applications (empty initially)
├── cmd/
│   └── server/
│       └── main.go       # Entry point
├── config/
│   └── settings.go       # Basic settings
├── internal/             # Generated code
├── static/               # Static files
├── templates/            # Global templates
├── Makefile             # Build automation
├── docker-compose.yml   # Local development
├── go.mod
├── gojango.yaml        # Project config
└── README.md

# Create first app manually (automated in Phase 2)
mkdir -p apps/core
# Write app.go manually for now

# Run the development server
make run    # or: go run cmd/server/main.go
```

### Tests Required

```
tests/
├── pkg/gojango/
│   ├── registry_test.go      # App registration and lifecycle
│   ├── app_test.go           # App interface implementations
│   └── application_test.go   # Core application functionality
├── cmd/gojango/
│   └── commands/
│       └── new_test.go       # Project generation
└── integration/
    └── basic_test.go         # End-to-end project creation
```

### Success Metrics for Phase 1
- [ ] Can generate a working Go project
- [ ] Project compiles and runs basic HTTP server
- [ ] Can manually create and register apps
- [ ] Community can provide feedback on structure
- [ ] Foundation ready for Phase 2 features

---

## Phase 2: Web Basics - v0.2.0
*Weeks 3-4: Make it useful for building web apps*

### Features to Build

#### ✅ Enhanced CLI
- [ ] **'gojango startapp' command** - Automated app creation
- [ ] **App template generation** - Generate boilerplate app code
- [ ] **Project health checks** - Validate project structure

#### ✅ Routing System
- [ ] **Gin integration** - Production-ready HTTP server
- [ ] **URL patterns** - Django-style URL routing
- [ ] **URL reversal** - Template functions for URL generation
- [ ] **Static file serving** - Efficient static asset handling

#### ✅ Basic Middleware
- [ ] **Request logging** - Structured request/response logging
- [ ] **Recovery middleware** - Graceful panic recovery
- [ ] **Request ID** - Unique ID per request for tracing

#### ✅ Template System (Basic)
- [ ] **html/template integration** - Standard Go templates (Templ comes later)
- [ ] **Template discovery** - Auto-load from apps/*/templates/
- [ ] **Template functions** - URL reversal, static files, etc.

### What Users Can Do After Phase 2

```bash
# Create an app with CLI
gojango startapp blog

# Generated app structure:
apps/blog/
├── app.go              # App configuration
├── views.go           # HTTP handlers
├── templates/         # App templates
│   ├── list.html
│   └── detail.html
├── static/           # App static files
│   ├── css/
│   └── js/
└── tests/           # App tests

# Define routes in apps/blog/app.go
# Create templates with URL reversal
# Build a basic website!

make run
# Visit http://localhost:8080
# Blog at http://localhost:8080/blog/
```

---

## Phase 3: Database Layer - v0.3.0
*Weeks 5-6: Add data persistence*

### Features to Build

#### ✅ Ent Integration
- [ ] **Basic Ent setup** - Schema generation and client
- [ ] **Multi-app schema** - App-prefixed table names
- [ ] **Database connection** - Connection pooling and management
- [ ] **Common mixins** - Timestamps, soft delete, UUID

#### ✅ Migration System
- [ ] **Migration tracking** - Per-app migration history
- [ ] **Auto-generation** - Create migrations from schema changes
- [ ] **Migration runner** - Apply/rollback migrations

#### ✅ CLI Commands
- [ ] **'make migrate' command** - Apply pending migrations
- [ ] **'make dbshell' command** - Database shell access
- [ ] **'gojango generate ent'** - Generate Ent code

### What Users Can Do After Phase 3

```go
// apps/blog/schema/post.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "github.com/epuerta9/gojango/pkg/gojango"
)

type Post struct {
    ent.Schema
}

func (Post) Mixin() []ent.Mixin {
    return []ent.Mixin{
        gojango.TimestampMixin{},  // created_at, updated_at
    }
}

func (Post) Fields() []ent.Field {
    return []ent.Field{
        field.String("title").NotEmpty(),
        field.Text("content"),
        field.String("slug").Unique(),
        field.Enum("status").Values("draft", "published").Default("draft"),
    }
}
```

```bash
# Generate Ent code
gojango generate ent

# Create and run migrations
make migrate

# Now build database-backed applications!
```

---

## Remaining Phases Overview

### Phase 4: Settings & Configuration - v0.4.0
- Starlark-based settings system
- Environment management
- Configuration validation

### Phase 5: Admin Interface - v0.5.0
- **THE KILLER FEATURE**
- Auto-generated admin from Ent schemas
- CRUD interfaces with list/detail views
- Bulk actions and filtering

### Phase 6: Authentication & Authorization - v0.6.0
- User management system
- Session handling
- Permission framework

### Phase 7: Templates & HTMX - v0.7.0
- Templ integration (type-safe templates)
- HTMX support for modern UX
- Component system

### Phase 8: API Layer - v0.8.0
- Connect/gRPC services
- Auto-generated from Ent schemas
- TypeScript client generation

### Phase 9: Background Tasks - v0.9.0
- Asynq integration
- Task scheduling and monitoring
- Cron jobs

### Phase 10: Signals & NATS - v1.0.0
- Cross-language event system
- Embedded NATS server
- Real-time features

---

## Development Approach

### Ship Early, Ship Often
- **2-week release cycles** during early phases
- **Working software** at each milestone
- **Community feedback** drives priorities
- **Backward compatibility** once we hit v0.5.0

### Quality Assurance
- **Tests for everything** - Unit, integration, and example apps
- **Documentation** - Updated with each release
- **Example projects** - Real applications built with each version
- **Performance monitoring** - Benchmark critical paths

### Community Building
- **Open source** from day one (MIT license)
- **Weekly progress** updates on GitHub
- **Discord server** for real-time discussions
- **Blog posts** documenting design decisions

---

## Current Status

🎯 **Phase 1 in progress** - Building the foundation
📅 **Target completion** - January 2025
🏗️ **Next major milestone** - Phase 2 (Web Basics)

The goal is to have something useful by Phase 2, amazing by Phase 5 (admin interface), and production-ready by Phase 10.

---

## Get Involved

This is an ambitious project that will benefit from community involvement:

1. **Try early releases** and provide feedback
2. **Report bugs** and suggest improvements  
3. **Contribute code** - especially if you have Django experience
4. **Build example apps** to test real-world usage
5. **Spread the word** if you like what you see

Together, we'll build the framework Go developers have been waiting for! 🚀