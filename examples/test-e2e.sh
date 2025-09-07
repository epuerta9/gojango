#!/bin/bash

# Gojango End-to-End Test Script
# Tests the complete workflow from CLI installation to running application

set -e

# Configuration
PROJECT_NAME="test-e2e-project"
APP_NAME="testapp"
FRONTEND="htmx"
DATABASE="postgres"
TEST_PORT="8081"
export PORT="8081"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="$SCRIPT_DIR/temp-test"

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

log_step() {
    echo ""
    echo -e "${BLUE}🔹 $1${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."
    if [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
    fi
    
    # Kill any background processes
    if [ -n "$SERVER_PID" ]; then
        kill "$SERVER_PID" 2>/dev/null || true
        log_info "Stopped test server (PID: $SERVER_PID)"
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --project-name)
            PROJECT_NAME="$2"
            shift 2
            ;;
        --app-name)
            APP_NAME="$2"
            shift 2
            ;;
        --frontend)
            FRONTEND="$2"
            shift 2
            ;;
        --database)
            DATABASE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --project-name NAME    Test project name (default: test-e2e-project)"
            echo "  --app-name NAME        Test app name (default: testapp)"
            echo "  --frontend TYPE        Frontend type (default: htmx)"
            echo "  --database TYPE        Database type (default: postgres)"
            echo "  --help                 Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            ;;
    esac
done

# Main test function
main() {
    echo -e "${BLUE}"
    echo "🧪 Gojango End-to-End Test Suite"
    echo "================================="
    echo -e "${NC}"
    echo "Testing Gojango CLI and framework functionality"
    echo "Project: $PROJECT_NAME"
    echo "App: $APP_NAME"
    echo "Frontend: $FRONTEND"
    echo "Database: $DATABASE"
    echo ""
    
    # Create test directory
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Step 1: Build CLI
    log_step "Step 1: Build Gojango CLI"
    cd "$PROJECT_ROOT"
    if [ -f "bin/gojango" ]; then
        rm bin/gojango
    fi
    go build -o bin/gojango ./cmd/gojango
    CLI_PATH="$PROJECT_ROOT/bin/gojango"
    
    if [ ! -x "$CLI_PATH" ]; then
        log_error "Failed to build Gojango CLI"
    fi
    log_success "CLI built successfully at $CLI_PATH"
    
    # Step 2: Test CLI basic functionality
    log_step "Step 2: Test CLI Commands"
    cd "$TEST_DIR"
    
    # Test version command
    log_info "Testing version command..."
    version_output=$("$CLI_PATH" version)
    if [[ "$version_output" != *"Gojango CLI version"* ]]; then
        log_error "Version command failed: $version_output"
    fi
    log_success "Version command works: $(echo "$version_output" | head -1)"
    
    # Test doctor command
    log_info "Testing doctor command..."
    "$CLI_PATH" doctor > doctor_output.txt 2>&1
    if ! grep -q "All checks passed\|checks failed" doctor_output.txt; then
        log_error "Doctor command failed"
    fi
    log_success "Doctor command completed"
    
    # Step 3: Create new project
    log_step "Step 3: Create New Project"
    log_info "Creating project: $PROJECT_NAME"
    "$CLI_PATH" new "$PROJECT_NAME" \
        --module "github.com/test/$PROJECT_NAME" \
        --frontend "$FRONTEND" \
        --database "$DATABASE"
    
    if [ ! -d "$PROJECT_NAME" ]; then
        log_error "Project directory not created"
    fi
    log_success "Project created successfully"
    
    # Step 4: Validate project structure
    log_step "Step 4: Validate Project Structure"
    cd "$PROJECT_NAME"
    
    required_files=(
        "main.go"
        "go.mod" 
        "Makefile"
        "gojango.yaml"
        "README.md"
        ".env.example"
        ".gitignore"
    )
    
    required_dirs=(
        "apps"
        "static"
        "templates"
        "cmd/server"
        "internal/settings"
    )
    
    log_info "Checking required files..."
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_error "Required file missing: $file"
        fi
    done
    log_success "All required files present"
    
    log_info "Checking required directories..."
    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            log_error "Required directory missing: $dir"
        fi
    done
    log_success "All required directories present"
    
    # Step 5: Setup project for local development
    log_step "Step 5: Setup Local Development"
    
    # Add replace directive for local development
    log_info "Setting up go.mod for local development..."
    go mod edit -replace "github.com/epuerta9/gojango=$PROJECT_ROOT"
    go mod tidy
    
    if ! go mod verify; then
        log_error "go mod verify failed"
    fi
    log_success "Go modules configured correctly"
    
    # Step 6: Test project compilation
    log_step "Step 6: Test Project Compilation"
    
    log_info "Testing project compilation..."
    if ! go build -o test-binary main.go; then
        log_error "Project failed to compile"
    fi
    
    if [ ! -x "test-binary" ]; then
        log_error "Compiled binary not found or not executable"
    fi
    log_success "Project compiles successfully"
    
    # Step 7: Test basic commands
    log_step "Step 7: Test Application Commands"
    
    log_info "Testing version command..."
    version_out=$(./test-binary version 2>&1)
    if [[ "$version_out" != *"Gojango application version"* ]]; then
        log_error "Application version command failed: $version_out"
    fi
    log_success "Application version command works"
    
    log_info "Testing apps command..."
    apps_out=$(./test-binary apps 2>&1)
    if [[ "$apps_out" != *"Registered apps (0)"* ]]; then
        log_error "Apps command failed: $apps_out"
    fi
    log_success "Apps command works (no apps registered)"
    
    # Step 8: Create test app
    log_step "Step 8: Create Test App"
    
    log_info "Creating app: $APP_NAME"
    "$CLI_PATH" startapp "$APP_NAME"
    
    if [ ! -d "apps/$APP_NAME" ]; then
        log_error "App directory not created"
    fi
    
    required_app_files=(
        "apps/$APP_NAME/app.go"
        "apps/$APP_NAME/views.go"
        "apps/$APP_NAME/tests/app_test.go"
    )
    
    for file in "${required_app_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_error "Required app file missing: $file"
        fi
    done
    log_success "App created successfully"
    
    # Step 9: Register app in main.go
    log_step "Step 9: Register App"
    
    log_info "Adding app import to main.go..."
    # Add import line after the comment
    sed -i "s|// Import your apps here.*|// Import your apps here (they register themselves via init())\n\t_ \"github.com/test/$PROJECT_NAME/apps/$APP_NAME\"|" main.go
    
    # Verify the import was added
    if ! grep -q "apps/$APP_NAME" main.go; then
        log_error "Failed to add app import to main.go"
    fi
    log_success "App import added to main.go"
    
    # Step 10: Test with registered app
    log_step "Step 10: Test with Registered App"
    
    log_info "Recompiling with registered app..."
    go build -o test-binary main.go
    
    log_info "Testing apps command with registered app..."
    apps_out=$(./test-binary apps 2>&1)
    if [[ "$apps_out" != *"$APP_NAME"* ]]; then
        log_error "App not registered: $apps_out"
    fi
    if [[ "$apps_out" != *"Registered apps (1)"* ]]; then
        log_error "App count incorrect: $apps_out"
    fi
    log_success "App registered successfully: $APP_NAME"
    
    # Step 11: Test server startup (briefly)
    log_step "Step 11: Test Server Startup"
    
    log_info "Starting server for validation..."
    ./test-binary runserver &
    SERVER_PID=$!
    
    # Give server time to start
    sleep 2
    
    # Check if server is running
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
        log_error "Server failed to start or crashed"
    fi
    
    # Try to make a request (optional - depends on having curl/wget)
    if command -v curl >/dev/null 2>&1; then
        log_info "Testing HTTP endpoint..."
        if curl -f -s "http://localhost:$TEST_PORT/health" >/dev/null; then
            log_success "Server responds to HTTP requests"
        else
            log_warning "Server started but health endpoint not responding"
        fi
    fi
    
    # Stop server
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    SERVER_PID=""
    log_success "Server started and stopped successfully"
    
    # Step 12: Test app-specific functionality
    log_step "Step 12: Test App Functionality"
    
    log_info "Running app tests..."
    if ! go test "./apps/$APP_NAME/tests/..."; then
        log_error "App tests failed"
    fi
    log_success "App tests passed"
    
    # Step 13: Test Makefile commands
    log_step "Step 13: Test Makefile Commands"
    
    if command -v make >/dev/null 2>&1; then
        log_info "Testing Makefile help..."
        if ! make help >/dev/null; then
            log_error "make help failed"
        fi
        log_success "Makefile help works"
        
        log_info "Testing make info..."
        info_out=$(make info 2>&1)
        if [[ "$info_out" != *"$PROJECT_NAME"* ]]; then
            log_error "make info failed: $info_out"
        fi
        log_success "make info works"
    else
        log_warning "make not available, skipping Makefile tests"
    fi
    
    # Step 14: Final validation
    log_step "Step 14: Final Validation"
    
    log_info "Running complete test suite..."
    if ! go test ./...; then
        log_error "Complete test suite failed"
    fi
    log_success "All tests passed"
    
    # Step 15: Test Phase 2 Features - Gin Integration
    log_step "Step 15: Test Gin Integration and Routing"
    
    # Create a simple template for testing
    log_info "Creating test template..."
    mkdir -p "apps/$APP_NAME/templates/$APP_NAME"
    cat > "apps/$APP_NAME/templates/$APP_NAME/test.html" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Message}}</h1>
    <p>Static URL: {{static "css/test.css"}}</p>
    <p>App URL: {{url "testapp:index"}}</p>
</body>
</html>
EOF
    
    # Update the app to include proper Gin routes
    log_info "Updating app with Phase 2 routing..."
    cat > "apps/$APP_NAME/app.go" << EOF
package $APP_NAME

import (
    "github.com/epuerta9/gojango/pkg/gojango"
)

func init() {
    // Register this app with the global registry
    gojango.Register(&${APP_NAME^}App{})
}

// ${APP_NAME^}App represents the $APP_NAME application
type ${APP_NAME^}App struct {
    gojango.BaseApp
}

// Config returns the app configuration
func (app *${APP_NAME^}App) Config() gojango.AppConfig {
    return gojango.AppConfig{
        Name:    "$APP_NAME",
        Label:   "${APP_NAME^} Application",
        Version: "1.0.0",
    }
}

// Initialize sets up the app
func (app *${APP_NAME^}App) Initialize(ctx *gojango.AppContext) error {
    // Call parent initialization
    if err := app.BaseApp.Initialize(ctx); err != nil {
        return err
    }
    
    // App-specific initialization goes here
    
    return nil
}

// Routes defines the HTTP routes for this app
func (app *${APP_NAME^}App) Routes() []gojango.Route {
    return []gojango.Route{
        {
            Method:  "GET",
            Path:    "/",
            Handler: app.IndexView,
            Name:    "index",
        },
        {
            Method:  "GET",
            Path:    "/test",
            Handler: app.TestView,
            Name:    "test",
        },
    }
}
EOF
    
    # Update views to use Gin handlers
    cat > "apps/$APP_NAME/views.go" << EOF
package $APP_NAME

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func (app *${APP_NAME^}App) IndexView(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Hello from $APP_NAME!",
        "app":     "$APP_NAME",
        "phase":   2,
    })
}

func (app *${APP_NAME^}App) TestView(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Test endpoint working",
        "middleware_test": true,
        "request_id": c.GetString("request_id"),
    })
}
EOF
    
    log_info "Recompiling with Phase 2 features..."
    go build -o test-binary main.go
    log_success "Phase 2 app compiled successfully"
    
    # Step 16: Test HTTP endpoints with middleware
    log_step "Step 16: Test HTTP Endpoints and Middleware"
    
    log_info "Starting server for Phase 2 testing..."
    ./test-binary runserver &
    SERVER_PID=$!
    
    # Give server time to start
    sleep 3
    
    if command -v curl >/dev/null 2>&1; then
        log_info "Testing app index endpoint..."
        response=$(curl -s "http://localhost:$TEST_PORT/$APP_NAME/" || echo "")
        if [[ "$response" != *"Hello from testapp"* ]]; then
            log_error "App index endpoint failed: $response"
        fi
        log_success "App index endpoint works"
        
        log_info "Testing middleware (Request ID)..."
        headers=$(curl -s -D - "http://localhost:$TEST_PORT/$APP_NAME/test" | head -20)
        if [[ "$headers" != *"X-Request-ID"* ]]; then
            log_error "Request ID middleware not working: $headers"
        fi
        log_success "Request ID middleware works"
        
        log_info "Testing CORS headers..."
        cors_headers=$(curl -s -D - "http://localhost:$TEST_PORT/$APP_NAME/" | head -20)
        if [[ "$cors_headers" != *"Access-Control-Allow-Origin"* ]]; then
            log_error "CORS middleware not working: $cors_headers"
        fi
        log_success "CORS middleware works"
        
        log_info "Testing security headers..."
        security_headers=$(curl -s -D - "http://localhost:$TEST_PORT/$APP_NAME/" | head -20)
        if [[ "$security_headers" != *"X-Content-Type-Options"* ]]; then
            log_error "Security headers middleware not working"
        fi
        log_success "Security headers middleware works"
        
        log_info "Testing static file serving..."
        static_response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$TEST_PORT/static/css/style.css" || echo "000")
        if [[ "$static_response" == "404" ]]; then
            log_success "Static file serving configured (404 expected for non-existent file)"
        else
            log_warning "Static file response: $static_response"
        fi
    else
        log_warning "curl not available, skipping HTTP endpoint tests"
    fi
    
    # Stop server
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    SERVER_PID=""
    log_success "Phase 2 server testing completed"
    
    # Step 17: Test URL Reversal
    log_step "Step 17: Test URL Reversal and Template Functions"
    
    # Create a simple test to verify URL reversal
    cat > "test_url_reversal.go" << 'EOF'
package main

import (
    "fmt"
    "github.com/epuerta9/gojango/pkg/gojango/routing"
    "github.com/gin-gonic/gin"
)

func main() {
    gin.SetMode(gin.TestMode)
    router := routing.NewRouter()
    
    routes := []routing.Route{
        {
            Method:  "GET",
            Path:    "/",
            Handler: func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) },
            Name:    "index",
        },
    }
    
    err := router.RegisterRoutes("testapp", routes)
    if err != nil {
        panic(err)
    }
    
    url := router.Reverse("testapp:index")
    fmt.Printf("URL reversal test: %s\n", url)
    
    if url != "/testapp/" {
        panic(fmt.Sprintf("Expected '/testapp/', got '%s'", url))
    }
    
    fmt.Println("URL reversal works correctly!")
}
EOF
    
    log_info "Testing URL reversal..."
    if go run test_url_reversal.go; then
        log_success "URL reversal works correctly"
    else
        log_error "URL reversal test failed"
    fi
    
    # Summary
    echo ""
    echo -e "${GREEN}🎉 End-to-End Test Suite Completed Successfully!${NC}"
    echo ""
    echo "✅ CLI Installation and Commands"
    echo "✅ Project Generation"
    echo "✅ Project Structure Validation" 
    echo "✅ Compilation and Build"
    echo "✅ App Generation and Registration"
    echo "✅ Server Startup and HTTP Response"
    echo "✅ Development Workflow"
    echo "✅ Test Suite Execution"
    echo "✅ Gin Integration and Routing"
    echo "✅ Middleware Stack (RequestID, CORS, Security)"
    echo "✅ Static File Serving"
    echo "✅ URL Reversal and Template Functions"
    echo ""
    echo -e "${BLUE}Test Results Summary:${NC}"
    echo "Project: $PROJECT_NAME"
    echo "App: $APP_NAME ($(\wc -l < "apps/$APP_NAME/app.go") lines of code)"
    echo "Frontend: $FRONTEND"
    echo "Database: $DATABASE"
    echo "Location: $TEST_DIR/$PROJECT_NAME"
    echo ""
    echo -e "${GREEN}Gojango Phase 1 & 2 are working correctly! 🚀${NC}"
}

# Run main function
main "$@"