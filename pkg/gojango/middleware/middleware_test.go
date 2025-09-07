package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		if !exists {
			t.Error("Expected request_id to be set in context")
		}
		
		if requestID == "" {
			t.Error("Expected request_id to be non-empty")
		}
		
		c.String(200, "OK")
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	// Check that X-Request-ID header is set in response
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestRequestIDFromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	
	expectedID := "test-request-id"
	
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("request_id")
		if requestID != expectedID {
			t.Errorf("Expected request_id to be '%s', got: %s", expectedID, requestID)
		}
		c.String(200, "OK")
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", expectedID)
	router.ServeHTTP(w, req)
	
	// Check that the same ID is returned
	responseID := w.Header().Get("X-Request-ID")
	if responseID != expectedID {
		t.Errorf("Expected response X-Request-ID to be '%s', got: %s", expectedID, responseID)
	}
}

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID()) // Need this for logger
	router.Use(Logger())
	
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}
}

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(Recovery())
	
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)
	
	// Recovery should return 500
	if w.Code != 500 {
		t.Errorf("Expected status 500 after panic recovery, got: %d", w.Code)
	}
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})
	
	// Test preflight OPTIONS request with proper headers
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)
	
	// Check CORS preflight response
	if w.Code != 204 {
		t.Errorf("Expected status 204 for OPTIONS preflight, got: %d", w.Code)
	}
	
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	
	// Test actual CORS request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	router.ServeHTTP(w, req)
	
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: * on actual request, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SecurityHeaders())
	
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	// Check security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":      "nosniff",
		"X-Frame-Options":             "DENY",
		"X-XSS-Protection":           "1; mode=block",
		"Referrer-Policy":            "strict-origin-when-cross-origin",
	}
	
	for header, expectedValue := range expectedHeaders {
		if w.Header().Get(header) != expectedValue {
			t.Errorf("Expected header %s to be '%s', got: '%s'", 
				header, expectedValue, w.Header().Get(header))
		}
	}
}

func TestMiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add all middleware in typical order
	router.Use(RequestID())
	router.Use(Logger())
	router.Use(Recovery())
	router.Use(SecurityHeaders())
	router.Use(CORS())
	
	router.GET("/test", func(c *gin.Context) {
		// Verify request ID is available
		requestID, exists := c.Get("request_id")
		if !exists || requestID == "" {
			t.Error("Expected request_id to be set by RequestID middleware")
		}
		
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com") // Add Origin header to trigger CORS
	router.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}
	
	// Verify various headers are set
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header from RequestID middleware")
	}
	
	if w.Header().Get("X-Content-Type-Options") == "" {
		t.Error("Expected X-Content-Type-Options header from SecurityHeaders middleware")
	}
	
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header from CORS middleware")
	}
	
	// Verify JSON response
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Error("Expected JSON response with status ok")
	}
}