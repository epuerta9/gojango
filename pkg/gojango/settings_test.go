package gojango

import (
	"os"
	"testing"
)

func TestBasicSettingsCreation(t *testing.T) {
	settings := NewBasicSettings()
	if settings == nil {
		t.Fatal("NewBasicSettings() should return a valid settings instance")
	}

	if settings.data == nil {
		t.Error("Settings data map should be initialized")
	}
}

func TestBasicSettingsSetAndGet(t *testing.T) {
	settings := NewBasicSettings()

	// Test string setting
	settings.Set("TEST_STRING", "hello")
	value := settings.Get("TEST_STRING")
	if value != "hello" {
		t.Errorf("Expected 'hello', got %v", value)
	}

	// Test with default value
	value = settings.Get("NONEXISTENT", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got %v", value)
	}

	// Test without default value
	value = settings.Get("NONEXISTENT")
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}
}

func TestBasicSettingsGetString(t *testing.T) {
	settings := NewBasicSettings()

	// Test string value
	settings.Set("STRING_VAL", "test")
	value := settings.GetString("STRING_VAL")
	if value != "test" {
		t.Errorf("Expected 'test', got '%s'", value)
	}

	// Test with default
	value = settings.GetString("NONEXISTENT", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}

	// Test without default
	value = settings.GetString("NONEXISTENT")
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}
}

func TestBasicSettingsGetInt(t *testing.T) {
	settings := NewBasicSettings()

	// Test integer value
	settings.Set("INT_VAL", 42)
	value := settings.GetInt("INT_VAL")
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Test string integer
	settings.Set("STRING_INT", "123")
	value = settings.GetInt("STRING_INT")
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}

	// Test with default
	value = settings.GetInt("NONEXISTENT", 99)
	if value != 99 {
		t.Errorf("Expected 99, got %d", value)
	}

	// Test invalid string
	settings.Set("INVALID_INT", "not-a-number")
	value = settings.GetInt("INVALID_INT", 50)
	if value != 50 {
		t.Errorf("Expected 50 (default), got %d", value)
	}
}

func TestBasicSettingsGetBool(t *testing.T) {
	settings := NewBasicSettings()

	// Test boolean value
	settings.Set("BOOL_VAL", true)
	value := settings.GetBool("BOOL_VAL")
	if value != true {
		t.Errorf("Expected true, got %v", value)
	}

	// Test string boolean values
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"on", true},
		{"ON", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"NO", false},
		{"off", false},
		{"OFF", false},
		{"random", false},
	}

	for _, tc := range testCases {
		settings.Set("TEST_BOOL", tc.input)
		value := settings.GetBool("TEST_BOOL")
		if value != tc.expected {
			t.Errorf("For input '%s', expected %v, got %v", tc.input, tc.expected, value)
		}
	}

	// Test with default
	value = settings.GetBool("NONEXISTENT", true)
	if value != true {
		t.Errorf("Expected true (default), got %v", value)
	}
}

func TestBasicSettingsHas(t *testing.T) {
	settings := NewBasicSettings()

	// Test non-existent key
	if settings.Has("NONEXISTENT") {
		t.Error("Should return false for non-existent key")
	}

	// Test existing key
	settings.Set("EXISTS", "value")
	if !settings.Has("EXISTS") {
		t.Error("Should return true for existing key")
	}
}

func TestBasicSettingsGetAll(t *testing.T) {
	settings := NewBasicSettings()

	// Set some values
	settings.Set("KEY1", "value1")
	settings.Set("KEY2", 42)
	settings.Set("KEY3", true)

	all := settings.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 settings, got %d", len(all))
	}

	if all["KEY1"] != "value1" {
		t.Errorf("Expected 'value1', got %v", all["KEY1"])
	}

	if all["KEY2"] != 42 {
		t.Errorf("Expected 42, got %v", all["KEY2"])
	}

	if all["KEY3"] != true {
		t.Errorf("Expected true, got %v", all["KEY3"])
	}
}

func TestBasicSettingsLoadFromEnv(t *testing.T) {
	settings := NewBasicSettings()

	// Set environment variables
	os.Setenv("DEBUG", "true")
	os.Setenv("SECRET_KEY", "test-secret")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("PORT", "9000")
	os.Setenv("GOJANGO_CUSTOM", "custom-value")
	
	// Clean up after test
	defer func() {
		os.Unsetenv("DEBUG")
		os.Unsetenv("SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("PORT")
		os.Unsetenv("GOJANGO_CUSTOM")
	}()

	// Load from environment
	settings.LoadFromEnv()

	// Test boolean conversion
	if !settings.GetBool("DEBUG") {
		t.Error("DEBUG should be true")
	}

	// Test string values
	if settings.GetString("SECRET_KEY") != "test-secret" {
		t.Errorf("Expected 'test-secret', got '%s'", settings.GetString("SECRET_KEY"))
	}

	if settings.GetString("DATABASE_URL") != "postgres://localhost/test" {
		t.Errorf("Unexpected DATABASE_URL: %s", settings.GetString("DATABASE_URL"))
	}

	// Test integer conversion
	if settings.GetInt("PORT") != 9000 {
		t.Errorf("Expected 9000, got %d", settings.GetInt("PORT"))
	}

	// Test GOJANGO_ prefix handling
	if settings.GetString("CUSTOM") != "custom-value" {
		t.Errorf("Expected 'custom-value', got '%s'", settings.GetString("CUSTOM"))
	}
}

func TestBasicSettingsInterface(t *testing.T) {
	// Test that BasicSettings implements Settings interface
	var settings Settings = NewBasicSettings()
	
	// Test interface methods
	settings.Get("test")
	settings.GetString("test")
	settings.GetInt("test")
	settings.GetBool("test")

	// If this compiles, the interface is implemented correctly
}