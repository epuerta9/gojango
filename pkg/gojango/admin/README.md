# Gojango Admin Interface

A Django-style admin interface for Go web applications, built with React, TypeScript, TailwindCSS, and gRPC/Connect.

## Features

- **Modern React Frontend**: Built with Vite, TypeScript, and TailwindCSS v3
- **Type-Safe API**: gRPC with Connect for fully typed client-server communication
- **Django-Like Interface**: Familiar admin interface for Django developers
- **Auto-Generated Forms**: Automatic CRUD interfaces from Go struct definitions
- **Bulk Actions**: Select and perform actions on multiple records
- **Advanced Filtering**: Date ranges, text search, boolean filters, and more
- **Responsive Design**: Works perfectly on desktop and mobile devices
- **Extensible**: Custom actions, filters, and widgets

## Quick Start

1. **Register your models:**

```go
package main

import (
    "github.com/epuerta9/gojango/pkg/gojango"
    "github.com/epuerta9/gojango/pkg/gojango/admin"
)

type User struct {
    ID        int    `json:"id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    IsActive  bool   `json:"is_active"`
}

func main() {
    app, _ := gojango.NewApplication(&gojango.Config{
        Debug: true,
        Port:  8080,
    })

    // Register model with default admin
    admin.Register(&User{}, nil)
    
    // Setup admin routes
    app.SetupAdmin()
    
    app.Run()
}
```

2. **Visit the admin interface:**

Navigate to `http://localhost:8080/admin/` in your browser.

## Advanced Configuration

### Custom Admin Configuration

```go
// Create custom admin configuration
userAdmin := admin.NewModelAdmin(&User{}).
    SetListDisplay("id", "username", "email", "is_active").
    SetSearchFields("username", "email").
    SetListFilter("is_active", "created_at").
    SetOrdering("-created_at").
    SetListPerPage(50)

// Add custom bulk actions
userAdmin.AddAction("activate_users", "Activate selected users", 
    func(ctx *gin.Context, objects []interface{}) (interface{}, error) {
        // Your custom logic here
        return gin.H{"message": "Users activated", "count": len(objects)}, nil
    })

// Register with custom configuration
admin.Register(&User{}, userAdmin)
```

### Custom Widgets

```go
import "github.com/epuerta9/gojango/pkg/gojango/admin/widgets"

// Use specific widgets for fields
passwordWidget := widgets.NewPasswordInput().SetRenderValue(false)
emailWidget := widgets.NewTextInput()
statusWidget := widgets.NewSelect().SetChoices([]widgets.Choice{
    {Value: "active", Display: "Active"},
    {Value: "inactive", Display: "Inactive"},
})
```

### Custom Filters

```go
import "github.com/epuerta9/gojango/pkg/gojango/admin/filters"

// Create custom filters
statusFilter := filters.NewChoiceFilter("status", "Status", []filters.FilterChoice{
    {Value: "active", Display: "Active"},
    {Value: "inactive", Display: "Inactive"},
})

dateFilter := filters.NewDateFilter("created_at", "Created Date")
```

## Architecture

### Backend (Go)

- **Site Registry**: Central registry for all admin models
- **Model Admin**: Configuration for individual models  
- **Database Interface**: Abstraction for database operations (Ent integration)
- **gRPC Service**: Type-safe API using Connect
- **Bulk Actions**: System for batch operations
- **Filters & Widgets**: Extensible form components

### Frontend (React)

- **Modern Stack**: Vite + React 18 + TypeScript + TailwindCSS v3
- **Type Safety**: Full TypeScript coverage with generated Connect clients  
- **Components**: Reusable UI components following admin design patterns
- **State Management**: React Query for server state management
- **Routing**: React Router for SPA navigation

### Communication

- **gRPC/Connect**: Type-safe, efficient communication
- **Protobuf Definitions**: Shared schema between Go and TypeScript
- **Auto-Generated Clients**: TypeScript clients generated from protobuf

## Development

### Frontend Development

```bash
cd pkg/gojango/admin/frontend
npm install
npm run dev  # Starts Vite dev server on port 3000
```

### Generate Protobuf Code

```bash
cd pkg/gojango/admin/frontend
npm run generate  # Generates TypeScript clients from protobuf
```

### Build for Production

```bash
cd pkg/gojango/admin/frontend
npm run build  # Builds optimized production bundle
```

## File Structure

```
pkg/gojango/admin/
├── README.md                    # This file
├── site.go                      # Admin site registry
├── model.go                     # Model admin base class
├── generator.go                 # Auto-generation from schemas
├── actions.go                   # Bulk actions system
├── filters.go                   # List filters system
├── grpc.go                      # gRPC/Connect service
├── widgets/
│   └── widgets.go              # Form widgets
├── frontend/                    # React frontend
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── proto/                  # Protobuf definitions
│   └── src/
│       ├── components/         # React components
│       ├── pages/             # Admin pages
│       ├── services/          # API clients
│       ├── hooks/             # React hooks
│       └── types/             # TypeScript types
└── templates/
    └── index.html             # HTML template
```

## Examples

See `examples/admin-example/` for a complete working example.

## Roadmap

- [ ] Ent integration for automatic schema detection
- [ ] Advanced relationship handling (ForeignKey, ManyToMany)
- [ ] File upload widgets
- [ ] Rich text editor widgets
- [ ] Advanced permissions system
- [ ] Audit logging
- [ ] Export functionality (CSV, JSON, Excel)
- [ ] Import functionality
- [ ] Dashboard widgets
- [ ] Custom admin themes

## Contributing

The admin interface is part of the larger Gojango framework. See the main project documentation for contribution guidelines.