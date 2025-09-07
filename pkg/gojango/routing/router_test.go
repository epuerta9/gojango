package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRouterCreation(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("Expected router to be created")
	}

	engine := router.GetEngine()
	if engine == nil {
		t.Fatal("Expected Gin engine to be available")
	}
}

func TestRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter()

	// Test handler
	testHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}

	routes := []Route{
		{
			Method:  "GET",
			Path:    "/",
			Handler: testHandler,
			Name:    "index",
		},
		{
			Method:  "POST",
			Path:    "/create",
			Handler: testHandler,
			Name:    "create",
		},
	}

	err := router.RegisterRoutes("testapp", routes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify routes are registered
	registeredRoutes := router.GetRoutes()
	if len(registeredRoutes) != 2 {
		t.Fatalf("Expected 2 routes, got: %d", len(registeredRoutes))
	}

	// Test route names
	if _, exists := registeredRoutes["testapp:index"]; !exists {
		t.Error("Expected 'testapp:index' route to be registered")
	}

	if _, exists := registeredRoutes["testapp:create"]; !exists {
		t.Error("Expected 'testapp:create' route to be registered")
	}
}

func TestRouteHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	routes := make([]Route, len(methods))

	for i, method := range methods {
		routes[i] = Route{
			Method:  method,
			Path:    "/" + method,
			Handler: func(c *gin.Context) { c.String(200, method) },
			Name:    method,
		}
	}

	err := router.RegisterRoutes("testapp", routes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Test each method works
	for _, method := range methods {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, "/testapp/"+method, nil)
		router.GetEngine().ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for %s method, got: %d", method, w.Code)
		}
	}
}

func TestURLReversal(t *testing.T) {
	router := NewRouter()

	routes := []Route{
		{
			Method:  "GET",
			Path:    "/",
			Handler: func(c *gin.Context) {},
			Name:    "index",
		},
		{
			Method:  "GET",
			Path:    "/detail",
			Handler: func(c *gin.Context) {},
			Name:    "detail",
		},
	}

	router.RegisterRoutes("blog", routes)

	// Test URL reversal
	indexURL := router.Reverse("blog:index")
	if indexURL != "/blog/" {
		t.Errorf("Expected '/blog/', got: %s", indexURL)
	}

	detailURL := router.Reverse("blog:detail")
	if detailURL != "/blog/detail" {
		t.Errorf("Expected '/blog/detail', got: %s", detailURL)
	}

	// Test nonexistent route
	missingURL := router.Reverse("blog:missing")
	if missingURL != "" {
		t.Errorf("Expected empty string for missing route, got: %s", missingURL)
	}
}

func TestDuplicateRouteNames(t *testing.T) {
	router := NewRouter()

	routes := []Route{
		{
			Method:  "GET",
			Path:    "/",
			Handler: func(c *gin.Context) {},
			Name:    "index",
		},
	}

	// Register first time
	err := router.RegisterRoutes("app1", routes)
	if err != nil {
		t.Fatalf("Expected no error on first registration, got: %v", err)
	}

	// Register second time with different app - should work
	err = router.RegisterRoutes("app2", routes)
	if err != nil {
		t.Fatalf("Expected no error on second registration with different app, got: %v", err)
	}

	// Register same app again - should fail
	err = router.RegisterRoutes("app1", routes)
	if err == nil {
		t.Fatal("Expected error on duplicate registration")
	}
}

func TestTemplateFunctions(t *testing.T) {
	router := NewRouter()

	funcs := router.TemplateFuncs()

	// Test URL function exists
	if _, exists := funcs["url"]; !exists {
		t.Error("Expected 'url' template function")
	}

	// Test static function exists
	if _, exists := funcs["static"]; !exists {
		t.Error("Expected 'static' template function")
	}

	// Test static function behavior
	staticFunc := funcs["static"].(func(string) string)
	result := staticFunc("css/style.css")
	expected := "/static/css/style.css"
	if result != expected {
		t.Errorf("Expected '%s', got: '%s'", expected, result)
	}
}

func TestUnsupportedHTTPMethod(t *testing.T) {
	router := NewRouter()

	routes := []Route{
		{
			Method:  "INVALID",
			Path:    "/",
			Handler: func(c *gin.Context) {},
			Name:    "test",
		},
	}

	err := router.RegisterRoutes("testapp", routes)
	if err == nil {
		t.Fatal("Expected error for unsupported HTTP method")
	}

	if err.Error() != "unsupported HTTP method: INVALID" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}