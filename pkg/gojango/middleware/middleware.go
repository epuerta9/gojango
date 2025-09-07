package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists (from headers)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}
		
		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// Logger provides structured request logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		requestID, exists := param.Keys["request_id"]
		if !exists {
			requestID = "unknown"
		}
		
		return fmt.Sprintf("[%s] %s - %s \"%s %s %s\" %d %s \"%s\" \"%s\" %s\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			requestID,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.Request.Referer(),
			param.ErrorMessage,
		)
	})
}

// Recovery provides panic recovery with logging
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(log.Writer(), func(c *gin.Context, recovered interface{}) {
		requestID, exists := c.Get("request_id")
		if !exists {
			requestID = "unknown"
		}
		
		log.Printf("[PANIC RECOVERED] Request ID: %s, Error: %v\n", requestID, recovered)
		
		c.AbortWithStatus(500)
	})
}

// CORS provides sensible CORS defaults using gin-contrib/cors
// For custom CORS configuration, users can use cors.New(cors.Config{...}) directly
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true, // This is the correct way to allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}

// CORSPermissive provides a more permissive CORS configuration for development
func CORSPermissive() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// SecurityHeaders adds basic security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Next()
	}
}