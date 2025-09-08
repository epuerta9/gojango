# Gojango Admin Interface Makefile

.PHONY: help install dev build clean test admin-dev admin-build admin-clean

# Default target
help:
	@echo "Gojango Admin Interface - Make Targets"
	@echo "======================================"
	@echo ""
	@echo "Development:"
	@echo "  install     - Install all dependencies (Go modules + npm packages)"
	@echo "  dev         - Start development servers (Go backend + Vite frontend)"
	@echo "  admin-dev   - Start only the React frontend dev server"
	@echo ""
	@echo "Building:"
	@echo "  build       - Build complete admin interface for production"
	@echo "  admin-build - Build only the React frontend"
	@echo "  generate    - Generate TypeScript clients from protobuf"
	@echo ""
	@echo "Maintenance:"
	@echo "  test        - Run all tests (Go + frontend)"
	@echo "  clean       - Clean all build artifacts"
	@echo "  admin-clean - Clean only frontend build"
	@echo ""

# Installation
install: install-go install-frontend

install-go:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

install-frontend:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# Development
dev: admin-dev
	@echo "Starting Gojango admin in development mode..."
	@echo "Backend will proxy to frontend dev server on http://localhost:3000"

admin-dev:
	@echo "Starting React development server..."
	cd frontend && npm run dev &

# Building
build: admin-build
	@echo "Admin interface built successfully!"
	@echo "Production files are in: frontend/dist/"

admin-build:
	@echo "Building React frontend for production..."
	cd frontend && npm run build

generate:
	@echo "Generating TypeScript clients from protobuf..."
	cd frontend && npm run generate

# Production build with protobuf generation
build-full: generate admin-build
	@echo "Full build with protobuf generation complete!"

# Testing
test: test-go test-frontend

test-go:
	@echo "Running Go tests..."
	cd ../../../.. && go test ./pkg/gojango/admin/...

test-frontend:
	@echo "Running frontend tests..."
	cd frontend && npm test

# Cleaning
clean: admin-clean
	@echo "Cleaned all build artifacts"

admin-clean:
	@echo "Cleaning frontend build..."
	cd frontend && rm -rf dist node_modules/.vite
	@echo "Frontend cleaned"

# Docker targets
docker-build:
	@echo "Building Docker image for admin..."
	docker build -t gojango-admin .

# Quick development setup for new contributors
setup: install
	@echo ""
	@echo "🎉 Gojango Admin Setup Complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev' to start development servers"
	@echo "  2. Visit http://localhost:8080/admin/ to see the admin"
	@echo "  3. Frontend dev server: http://localhost:3000"
	@echo ""

# Production deployment
deploy: build
	@echo "Admin interface ready for deployment!"
	@echo "Serve the files in frontend/dist/ as static assets"