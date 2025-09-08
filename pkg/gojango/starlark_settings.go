package gojango

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.starlark.net/starlark"
)

// StarlarkSettings provides Django-style configuration using Starlark
type StarlarkSettings struct {
	data   map[string]interface{}
	thread *starlark.Thread
	globals starlark.StringDict
}

// NewStarlarkSettings creates a new StarlarkSettings instance
func NewStarlarkSettings() *StarlarkSettings {
	s := &StarlarkSettings{
		data:   make(map[string]interface{}),
		thread: &starlark.Thread{Name: "settings"},
	}
	
	// Setup built-in functions
	s.setupBuiltins()
	
	return s
}

// LoadFromFile loads settings from a Starlark file (like Django's settings.py)
func (s *StarlarkSettings) LoadFromFile(filename string) error {
	// Read the settings file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read settings file %s: %w", filename, err)
	}
	
	// Execute the Starlark code
	globals, err := starlark.ExecFile(s.thread, filename, content, s.globals)
	if err != nil {
		return fmt.Errorf("failed to execute settings file %s: %w", filename, err)
	}
	
	// Convert Starlark values to Go values and store in data
	for name, value := range globals {
		if !strings.HasPrefix(name, "_") { // Skip private variables
			goValue, err := s.starlarkToGo(value)
			if err != nil {
				return fmt.Errorf("failed to convert setting %s: %w", name, err)
			}
			s.data[name] = goValue
		}
	}
	
	return nil
}

// setupBuiltins sets up Django-style built-in functions for Starlark
func (s *StarlarkSettings) setupBuiltins() {
	s.globals = starlark.StringDict{
		"env": s.makeEnvFunction(),
		"load": starlark.NewBuiltin("load", s.loadBuiltin),
	}
}

// makeEnvFunction creates the Django-style env() function
func (s *StarlarkSettings) makeEnvFunction() *starlark.Builtin {
	return starlark.NewBuiltin("env", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// env.get(key, default) or env(key, default)
		var key string
		var defaultVal starlark.Value = starlark.None
		
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "key", &key, "default?", &defaultVal); err != nil {
			return nil, err
		}
		
		// Get environment variable
		if val := os.Getenv(key); val != "" {
			return starlark.String(val), nil
		}
		
		return defaultVal, nil
	}).BindReceiver(&envModule{})
}

// envModule provides Django-style environment variable access
type envModule struct{}

func (e *envModule) String() string        { return "env" }
func (e *envModule) Type() string          { return "env" }
func (e *envModule) Freeze()               {}
func (e *envModule) Truth() starlark.Bool  { return starlark.True }
func (e *envModule) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: env") }

func (e *envModule) Attr(name string) (starlark.Value, error) {
	switch name {
	case "get":
		return starlark.NewBuiltin("env.get", e.get), nil
	case "bool":
		return starlark.NewBuiltin("env.bool", e.getBool), nil
	case "int":
		return starlark.NewBuiltin("env.int", e.getInt), nil
	case "list":
		return starlark.NewBuiltin("env.list", e.getList), nil
	default:
		return nil, nil
	}
}

func (e *envModule) AttrNames() []string {
	return []string{"get", "bool", "int", "list"}
}

// env.get(key, default)
func (e *envModule) get(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key string
	var defaultVal starlark.Value = starlark.String("")
	
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "key", &key, "default?", &defaultVal); err != nil {
		return nil, err
	}
	
	if val := os.Getenv(key); val != "" {
		return starlark.String(val), nil
	}
	
	return defaultVal, nil
}

// env.bool(key, default)
func (e *envModule) getBool(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key string
	var defaultVal starlark.Bool = starlark.False
	
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "key", &key, "default?", &defaultVal); err != nil {
		return nil, err
	}
	
	if val := os.Getenv(key); val != "" {
		switch strings.ToLower(val) {
		case "true", "1", "yes", "on":
			return starlark.True, nil
		case "false", "0", "no", "off":
			return starlark.False, nil
		}
	}
	
	return defaultVal, nil
}

// env.int(key, default)
func (e *envModule) getInt(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key string
	var defaultVal starlark.Int = starlark.MakeInt(0)
	
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "key", &key, "default?", &defaultVal); err != nil {
		return nil, err
	}
	
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return starlark.MakeInt(i), nil
		}
	}
	
	return defaultVal, nil
}

// env.list(key, separator, default)
func (e *envModule) getList(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key string
	var separator string = ","
	var defaultVal *starlark.List = starlark.NewList(nil)
	
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "key", &key, "separator?", &separator, "default?", &defaultVal); err != nil {
		return nil, err
	}
	
	if val := os.Getenv(key); val != "" {
		parts := strings.Split(val, separator)
		items := make([]starlark.Value, len(parts))
		for i, part := range parts {
			items[i] = starlark.String(strings.TrimSpace(part))
		}
		return starlark.NewList(items), nil
	}
	
	return defaultVal, nil
}

// loadBuiltin handles the load() function for importing modules
func (s *StarlarkSettings) loadBuiltin(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var module string
	var symbols []string
	
	if args.Len() < 1 {
		return nil, fmt.Errorf("load: missing module argument")
	}
	
	if moduleVal, ok := args.Index(0).(starlark.String); ok {
		module = string(moduleVal)
	} else {
		return nil, fmt.Errorf("load: module must be a string")
	}
	
	// Extract symbol names
	for i := 1; i < args.Len(); i++ {
		if sym, ok := args.Index(i).(starlark.String); ok {
			symbols = append(symbols, string(sym))
		}
	}
	
	// Handle built-in modules
	switch module {
	case "env":
		// env module is already available as global
		return starlark.None, nil
	default:
		return nil, fmt.Errorf("load: unknown module %s", module)
	}
}

// starlarkToGo converts Starlark values to Go values
func (s *StarlarkSettings) starlarkToGo(val starlark.Value) (interface{}, error) {
	switch v := val.(type) {
	case starlark.String:
		return string(v), nil
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("integer too large")
		}
		return int(i), nil
	case starlark.Bool:
		return bool(v), nil
	case starlark.Float:
		return float64(v), nil
	case *starlark.List:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			item, err := s.starlarkToGo(v.Index(i))
			if err != nil {
				return nil, err
			}
			result[i] = item
		}
		return result, nil
	case *starlark.Dict:
		result := make(map[string]interface{})
		for _, item := range v.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("dict key must be string")
			}
			value, err := s.starlarkToGo(item[1])
			if err != nil {
				return nil, err
			}
			result[string(key)] = value
		}
		return result, nil
	case starlark.NoneType:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported Starlark type: %T", v)
	}
}

// Settings interface implementation

// Get retrieves a setting value with optional default
func (s *StarlarkSettings) Get(key string, defaultValue ...interface{}) interface{} {
	if val, exists := s.data[key]; exists {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// GetString retrieves a string setting with optional default
func (s *StarlarkSettings) GetString(key string, defaultValue ...string) string {
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
	return fmt.Sprintf("%v", val)
}

// GetInt retrieves an integer setting with optional default
func (s *StarlarkSettings) GetInt(key string, defaultValue ...int) int {
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
func (s *StarlarkSettings) GetBool(key string, defaultValue ...bool) bool {
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

// GetStringSlice retrieves a string slice setting
func (s *StarlarkSettings) GetStringSlice(key string, defaultValue ...[]string) []string {
	val := s.Get(key)
	if val == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return []string{}
	}
	
	if slice, ok := val.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			result[i] = fmt.Sprintf("%v", v)
		}
		return result
	}
	
	if slice, ok := val.([]string); ok {
		return slice
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return []string{}
}

// GetMap retrieves a map setting
func (s *StarlarkSettings) GetMap(key string, defaultValue ...map[string]interface{}) map[string]interface{} {
	val := s.Get(key)
	if val == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return make(map[string]interface{})
	}
	
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return make(map[string]interface{})
}

// Has checks if a setting exists
func (s *StarlarkSettings) Has(key string) bool {
	_, exists := s.data[key]
	return exists
}

// GetAll returns all settings
func (s *StarlarkSettings) GetAll() map[string]interface{} {
	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// Helper method to add LoadSettingsFromFile to Application
func (app *Application) LoadSettingsFromFile(filename string) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("settings file not found: %s", filename)
	}
	
	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(filename))
	
	var settings Settings
	
	switch ext {
	case ".star", ".bzl":
		// Starlark settings file
		starlarkSettings := NewStarlarkSettings()
		if err := starlarkSettings.LoadFromFile(filename); err != nil {
			return fmt.Errorf("failed to load Starlark settings: %w", err)
		}
		settings = starlarkSettings
	default:
		// Fall back to basic settings with environment variables
		basicSettings := NewBasicSettings()
		basicSettings.LoadFromEnv()
		settings = basicSettings
	}
	
	return app.LoadSettings(settings)
}