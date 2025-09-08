# Gojango Framework Makefile

.PHONY: help install build test clean dev admin-dev admin-build admin-setup example-admin

# Variables
BINARY_NAME=gojango
ADMIN_DIR=pkg/gojango/admin
EXAMPLE_ADMIN_DIR=examples/admin-example

# Default target
help:
	@echo "Gojango Framework - Make Targets"
	@echo "==============================="
	@echo ""
	@echo "🏗️  Core Framework:"
	@echo "  install      - Install all dependencies"
	@echo "  build        - Build the Gojango CLI tool"
	@echo "  test         - Run all tests"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "🎨 Admin Interface:"
	@echo "  admin-setup  - Setup admin development environment"
	@echo "  admin-dev    - Start admin frontend development server"
	@echo "  admin-build  - Build admin frontend for production"
	@echo "  admin-test   - Test admin interface"
	@echo ""
	@echo "🚀 Examples:"
	@echo "  example-admin   - Run the admin example application"
	@echo "  examples        - List all available examples"
	@echo ""
	@echo "📦 Development:"
	@echo "  dev             - Start full development environment"
	@echo "  version         - Show version information"
	@echo ""

# Core Framework
install:
	@echo "Installing Gojango dependencies..."
	go mod tidy
	go mod download

build:
	@echo "Building Gojango CLI..."
	go build -o bin/$(BINARY_NAME) ./cmd/gojango

test:
	@echo "Running all tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f $(EXAMPLE_ADMIN_DIR)/admin-example

# Admin Interface
admin-setup:
	@echo "Setting up admin development environment..."
	@echo "Installing frontend dependencies..."
	cd $(ADMIN_DIR)/frontend && npm install
	@echo "✅ Admin setup complete!"

admin-dev:
	@echo "Starting admin frontend development server..."
	cd $(ADMIN_DIR)/frontend && npm run dev

admin-build:
	@echo "Building admin frontend for production..."
	cd $(ADMIN_DIR)/frontend && npm run build
	@echo "✅ Admin built successfully!"

admin-test:
	@echo "Testing admin interface..."
	go test -v ./$(ADMIN_DIR)/...
	@echo "Testing frontend (when available)..."
	# cd $(ADMIN_DIR)/frontend && npm test

admin-generate:
	@echo "Generating TypeScript clients from protobuf..."
	cd $(ADMIN_DIR)/frontend && npm run generate

# Examples
example-admin: build-example-admin
	@echo "Starting admin example on http://localhost:8082..."
	cd $(EXAMPLE_ADMIN_DIR) && ./admin-example

build-example-admin:
	@echo "Building admin example..."
	cd $(EXAMPLE_ADMIN_DIR) && go build -o admin-example main.go

examples:
	@echo "Available Examples:"
	@echo "  make example-admin  - Django-style admin interface"
	@find examples/ -name "*.go" -exec dirname {} \; | sort | uniq | sed 's/^/  /'

# Development
dev: admin-dev
	@echo "Full development environment started!"
	@echo ""
	@echo "🎯 Admin Interface:"
	@echo "  Frontend dev: http://localhost:3000"
	@echo "  Backend proxy: Configure your app to proxy /admin to frontend"
	@echo ""

# Version information
version:
	@echo "Gojango Framework"
	@go run ./cmd/gojango version

# Deployment
deploy-admin: admin-build
	@echo "Admin interface ready for deployment!"
	@echo "Static files are in: $(ADMIN_DIR)/frontend/dist/"

# Development shortcuts
run-tests: test admin-test
	@echo "All tests completed!"

full-build: build admin-build
	@echo "Full framework build complete!"

# Help for contributors
contrib-help:
	@echo "🤝 Contributor Guide"
	@echo "==================="
	@echo ""
	@echo "Setup for development:"
	@echo "  1. make install"
	@echo "  2. make admin-setup"
	@echo "  3. make example-admin (in another terminal)"
	@echo "  4. make admin-dev (for frontend development)"
	@echo ""
	@echo "Before submitting PR:"
	@echo "  1. make test"
	@echo "  2. make admin-test"
	@echo "  3. make full-build"
	@echo ""