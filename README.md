<div align="center">
  <img src="logo.png" alt="Gojango Logo" width="200" height="200">
  
  # Gojango 🎸
  
  **Django's incredible developer experience with Go's performance and simplicity**
  
  [![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Tests](https://img.shields.io/badge/tests-passing-brightgreen.svg)](#testing)
  
</div>

## 🌟 **What is Gojango?**

Gojango brings Django's beloved developer experience to the Go ecosystem, providing a batteries-included web framework that doesn't hide the underlying packages you already know and love.

### **Framework Philosophy**
- **🔧 Code Generation over Configuration** - Your schemas drive API generation
- **🏗️ Django-Style Structure** - Familiar project layout with `apps/`, `settings.py` equivalent
- **📦 Batteries Included** - Admin interface, migrations, CLI tools, and more
- **🚀 Go Performance** - Built on proven Go packages (Gin, Ent, Cobra)
- **🎯 Developer Experience** - Simple commands, clear patterns, familiar workflows

## 🚀 **Quick Start**

### Global CLI - Project Creation
```bash
# Install Gojango CLI
go install github.com/epuerta9/gojango/cmd/gojango@latest

# Create new Django-style project
gojango new myblog --frontend htmx --database postgres

# Navigate to project
cd myblog
```

### Project CLI - Django-Style Commands
```bash
# Start development server (like Django's runserver)
go run manage.go runserver

# Run database migrations (like Django's migrate)
go run manage.go migrate

# Create new app (like Django's startapp)
go run manage.go startapp blog

# Generate APIs from schemas
go run manage.go generate proto

# Interactive shell with project context
go run manage.go shell
```

## 🏗️ **Project Structure**

Gojango creates Django-familiar project structures:

```
myblog/
├── manage.go           # Django manage.py equivalent
├── apps/              # Django-style applications
│   └── core/          # Default core app
├── config/            # Settings management
│   └── settings.star  # Starlark-based configuration
├── internal/          # Generated code
│   ├── ent/          # Ent ORM models
│   └── proto/        # Protobuf definitions
├── migrations/        # Database migrations
├── static/           # Static files
├── templates/        # HTML templates
└── docker-compose.yml # Container setup
```

## 🎯 **Key Features**

### **Django-Style CLI Separation**
- **`gojango`** (Global): Project creation, system checks
- **`manage.go`** (Project): Development server, migrations, app management

### **Multi-App Architecture**
- Django-style app system with automatic registration
- Dependency resolution and lifecycle management
- Per-app templates, static files, and routes

### **Settings Management**
- Starlark-based configuration (Python-like syntax)
- Environment variable integration
- Django patterns: `INSTALLED_APPS`, `DATABASES`, `DEBUG`

### **Code Generation**
- **Protobuf APIs** generated from Ent schemas
- **OpenAPI specs** with full REST documentation
- **Database migrations** with forward/reverse support
- **Admin interfaces** auto-generated from models

### **Admin Interface**
- Django-style admin with automatic model discovery
- Modern React frontend with TypeScript
- gRPC/Connect integration for type-safety
- Customizable list views, search, and actions

## 🧪 **Testing**

Gojango includes comprehensive end-to-end testing to ensure everything works as designed:

### **How We Test**
Our testing approach validates the complete Django-style workflow:

1. **Global CLI Testing** - Project creation, version checks, system validation
2. **Project Structure Validation** - All directories and files created correctly
3. **Django-Style Commands** - All manage.go commands work with proper context
4. **Settings System** - Starlark configuration loads and works correctly
5. **App Architecture** - Multi-app registration and routing functions
6. **Code Generation** - Schema-driven API generation produces correct output
7. **Build System** - Docker, Makefile, and dependencies work properly

### **Run End-to-End Tests**
```bash
# Run the comprehensive test suite
cd examples/e2e-test
./test-workflow.sh

# Expected output:
# 🎉 END-TO-END TEST SUMMARY
# ✅ Global CLI (gojango) - Project creation and system checks
# ✅ Project CLI (manage.go) - Django-style management interface  
# ✅ Django Structure - Apps, settings, migrations, admin
# ✅ Code Generation - Protobuf, OpenAPI, migrations
# ✅ Developer Experience - Familiar Django workflows
# ✅ Go 1.24 - Modern Go version and toolchain
```

The test creates a complete project, validates all generated files, and confirms the Django-style workflow functions correctly.

## 🛠️ **Built With**

Gojango leverages the best of the Go ecosystem:

- **[Gin](https://gin-gonic.com/)** - High-performance HTTP web framework
- **[Ent](https://entgo.io/)** - Simple, yet powerful ORM for Go
- **[Cobra](https://cobra.dev/)** - Modern CLI framework
- **[Starlark](https://github.com/google/starlark-go)** - Python-like configuration
- **[gRPC/Connect](https://connect.build/)** - Type-safe API communication
- **[React](https://reactjs.org/)** - Modern admin interface
- **[Docker](https://www.docker.com/)** - Containerization support

## 📚 **Documentation**

- [**Development Handbook**](docs/development/handbook.md) - Framework philosophy and architecture
- [**Getting Started Guide**](docs/getting-started.md) - Step-by-step tutorial
- [**API Reference**](docs/api/) - Complete API documentation
- [**Migration Guide**](docs/migration.md) - Moving from Django to Gojango

## 🤝 **Contributing**

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### **Development Setup**
```bash
git clone https://github.com/epuerta9/gojango.git
cd gojango
go mod download
go build -o bin/gojango cmd/gojango/*.go
./bin/gojango version
```

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
  
**🎸 Gojango - Where Django meets Go 🐹**

*Bringing Django's incredible developer experience to Go's performance and simplicity*

</div>
