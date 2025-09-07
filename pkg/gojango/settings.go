package gojango

import (
	"os"
	"strconv"
	"strings"
)

// BasicSettings provides a simple environment-variable based settings implementation
// This is a minimal implementation for Phase 1. Phase 4 will add Starlark support.
type BasicSettings struct {
	data map[string]interface{}
}

// NewBasicSettings creates a new BasicSettings instance
func NewBasicSettings() *BasicSettings {
	return &BasicSettings{
		data: make(map[string]interface{}),
	}
}

// LoadFromEnv loads settings from environment variables
func (s *BasicSettings) LoadFromEnv() {
	// Common Django-style settings from environment
	if val := os.Getenv("DEBUG"); val != "" {
		s.data["DEBUG"] = strings.ToLower(val) == "true" || val == "1"
	}
	
	if val := os.Getenv("SECRET_KEY"); val != "" {
		s.data["SECRET_KEY"] = val
	}
	
	if val := os.Getenv("DATABASE_URL"); val != "" {
		s.data["DATABASE_URL"] = val
	}
	
	if val := os.Getenv("REDIS_URL"); val != "" {
		s.data["REDIS_URL"] = val
	}
	
	if val := os.Getenv("PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			s.data["PORT"] = port
		}
	}
	
	if val := os.Getenv("HOST"); val != "" {
		s.data["HOST"] = val
	}
	
	// Load any GOJANGO_* prefixed environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], "GOJANGO_") {
			key := strings.TrimPrefix(parts[0], "GOJANGO_")
			s.data[key] = parts[1]
		}
	}
}

// Set adds or updates a setting
func (s *BasicSettings) Set(key string, value interface{}) {
	s.data[key] = value
}

// Get retrieves a setting value with optional default
func (s *BasicSettings) Get(key string, defaultValue ...interface{}) interface{} {
	if val, exists := s.data[key]; exists {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// GetString retrieves a string setting with optional default
func (s *BasicSettings) GetString(key string, defaultValue ...string) string {
	val := s.Get(key)
	if val == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	
	if str, ok := val.(string); ok {
		return str
	}
	
	// Try to convert to string
	return strings.TrimSpace(val.(string))
}

// GetInt retrieves an integer setting with optional default
func (s *BasicSettings) GetInt(key string, defaultValue ...int) int {
	val := s.Get(key)
	if val == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	
	if i, ok := val.(int); ok {
		return i
	}
	
	// Try to convert string to int
	if str, ok := val.(string); ok {
		if i, err := strconv.Atoi(str); err == nil {
			return i
		}
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetBool retrieves a boolean setting with optional default
func (s *BasicSettings) GetBool(key string, defaultValue ...bool) bool {
	val := s.Get(key)
	if val == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}
	
	if b, ok := val.(bool); ok {
		return b
	}
	
	// Try to convert string to bool
	if str, ok := val.(string); ok {
		str = strings.ToLower(strings.TrimSpace(str))
		return str == "true" || str == "1" || str == "yes" || str == "on"
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// Has checks if a setting exists
func (s *BasicSettings) Has(key string) bool {
	_, exists := s.data[key]
	return exists
}

// GetAll returns all settings
func (s *BasicSettings) GetAll() map[string]interface{} {
	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}