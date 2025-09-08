package admin

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test model for admin tests
type TestUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type TestPost struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID int    `json:"author_id"`
}

// Mock database interface for testing
type mockDBInterface struct {
	objects map[string][]interface{}
}

func newMockDBInterface() *mockDBInterface {
	return &mockDBInterface{
		objects: make(map[string][]interface{}),
	}
}

func (m *mockDBInterface) GetAll(ctx context.Context, model interface{}, filters map[string]interface{}, ordering []string, limit, offset int) ([]interface{}, int, error) {
	modelName := getModelName(model)
	objects := m.objects[modelName]
	
	total := len(objects)
	start := offset
	end := offset + limit
	
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	
	return objects[start:end], total, nil
}

func (m *mockDBInterface) GetByID(ctx context.Context, model interface{}, id interface{}) (interface{}, error) {
	modelName := getModelName(model)
	objects := m.objects[modelName]
	
	for _, obj := range objects {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if objMap["id"] == id {
				return obj, nil
			}
		}
	}
	
	return nil, nil
}

func (m *mockDBInterface) Create(ctx context.Context, model interface{}, data map[string]interface{}) (interface{}, error) {
	modelName := getModelName(model)
	
	// Add ID and timestamps
	data["id"] = len(m.objects[modelName]) + 1
	data["created_at"] = time.Now()
	data["updated_at"] = time.Now()
	
	m.objects[modelName] = append(m.objects[modelName], data)
	return data, nil
}

func (m *mockDBInterface) Update(ctx context.Context, model interface{}, id interface{}, data map[string]interface{}) (interface{}, error) {
	modelName := getModelName(model)
	objects := m.objects[modelName]
	
	for i, obj := range objects {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if objMap["id"] == id {
				// Update fields
				for key, value := range data {
					objMap[key] = value
				}
				objMap["updated_at"] = time.Now()
				m.objects[modelName][i] = objMap
				return objMap, nil
			}
		}
	}
	
	return nil, nil
}

func (m *mockDBInterface) Delete(ctx context.Context, model interface{}, id interface{}) error {
	modelName := getModelName(model)
	objects := m.objects[modelName]
	
	for i, obj := range objects {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if objMap["id"] == id {
				// Remove from slice
				m.objects[modelName] = append(objects[:i], objects[i+1:]...)
				return nil
			}
		}
	}
	
	return nil
}

func (m *mockDBInterface) GetSchema(model interface{}) (*ModelSchema, error) {
	return &ModelSchema{
		Fields: []FieldSchema{
			{Name: "id", Type: "integer", Required: true, Unique: true},
			{Name: "username", Type: "string", Required: true},
			{Name: "email", Type: "string", Required: true},
			{Name: "is_active", Type: "boolean"},
			{Name: "created_at", Type: "datetime"},
		},
		Relations: []RelationSchema{},
	}, nil
}

func TestSiteCreation(t *testing.T) {
	site := NewSite("test")
	
	assert.Equal(t, "test", site.name)
	assert.Equal(t, "Gojango Administration", site.headerTitle)
	assert.Equal(t, "Site Administration", site.indexTitle)
	assert.Empty(t, site.models)
}

func TestModelRegistration(t *testing.T) {
	site := NewSite("test")
	
	// Register model without admin config
	err := site.Register(&TestUser{}, nil)
	require.NoError(t, err)
	
	models := site.GetRegisteredModels()
	assert.Contains(t, models, "main.testuser")
	
	// Get model admin
	admin, exists := site.GetModelAdmin("main.testuser")
	assert.True(t, exists)
	assert.NotNil(t, admin)
	assert.Equal(t, "TestUser", admin.verboseName)
}

func TestModelAdminConfiguration(t *testing.T) {
	admin := NewModelAdmin(&TestUser{}).
		SetListDisplay("id", "username", "email").
		SetSearchFields("username", "email").
		SetListFilter("is_active").
		SetOrdering("-created_at").
		SetListPerPage(25)
	
	assert.Equal(t, []string{"id", "username", "email"}, admin.listDisplay)
	assert.Equal(t, []string{"username", "email"}, admin.searchFields)
	assert.Equal(t, []string{"is_active"}, admin.listFilter)
	assert.Equal(t, []string{"-created_at"}, admin.ordering)
	assert.Equal(t, 25, admin.listPerPage)
}

func TestModelAdminActions(t *testing.T) {
	admin := NewModelAdmin(&TestUser{})
	
	// Add custom action
	actionCalled := false
	admin.AddAction("test_action", "Test Action", func(ctx *gin.Context, objects []interface{}) (interface{}, error) {
		actionCalled = true
		return map[string]interface{}{
			"message": "Action executed",
			"count":   len(objects),
		}, nil
	})
	
	// Check action was added
	assert.Contains(t, admin.actions, "test_action")
	assert.Equal(t, "Test Action", admin.actions["test_action"].Description)
	
	// Execute action
	ctx := &gin.Context{}
	objects := []interface{}{
		map[string]interface{}{"id": 1},
		map[string]interface{}{"id": 2},
	}
	
	result, err := admin.actions["test_action"].Handler(ctx, objects)
	require.NoError(t, err)
	assert.True(t, actionCalled)
	
	resultMap := result.(map[string]interface{})
	assert.Equal(t, "Action executed", resultMap["message"])
	assert.Equal(t, 2, resultMap["count"])
}

func TestModelAdminWithDatabase(t *testing.T) {
	admin := NewModelAdmin(&TestUser{})
	mockDB := newMockDBInterface()
	admin.SetDatabaseInterface(mockDB)
	
	// Add test data
	testUsers := []interface{}{
		map[string]interface{}{"id": 1, "username": "john", "email": "john@example.com", "is_active": true},
		map[string]interface{}{"id": 2, "username": "jane", "email": "jane@example.com", "is_active": false},
	}
	mockDB.objects["main.testuser"] = testUsers
	
	ctx := context.Background()
	
	// Test GetAll
	objects, total, err := mockDB.GetAll(ctx, &TestUser{}, nil, nil, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, objects, 2)
	
	// Test GetByID
	obj, err := mockDB.GetByID(ctx, &TestUser{}, 1)
	require.NoError(t, err)
	assert.NotNil(t, obj)
	objMap := obj.(map[string]interface{})
	assert.Equal(t, "john", objMap["username"])
	
	// Test Create
	newUser := map[string]interface{}{
		"username": "bob",
		"email":    "bob@example.com",
		"is_active": true,
	}
	created, err := mockDB.Create(ctx, &TestUser{}, newUser)
	require.NoError(t, err)
	createdMap := created.(map[string]interface{})
	assert.Equal(t, 3, createdMap["id"])
	assert.Equal(t, "bob", createdMap["username"])
	
	// Test Update
	updateData := map[string]interface{}{"username": "bob_updated"}
	updated, err := mockDB.Update(ctx, &TestUser{}, 3, updateData)
	require.NoError(t, err)
	updatedMap := updated.(map[string]interface{})
	assert.Equal(t, "bob_updated", updatedMap["username"])
	
	// Test Delete
	err = mockDB.Delete(ctx, &TestUser{}, 3)
	require.NoError(t, err)
	
	// Verify delete
	obj, err = mockDB.GetByID(ctx, &TestUser{}, 3)
	require.NoError(t, err)
	assert.Nil(t, obj)
}

func TestAutoGenerateFilters(t *testing.T) {
	filters := AutoGenerateFilters(&TestUser{})
	
	assert.NotEmpty(t, filters)
	
	// Check for expected filters
	var boolFilter, textFilter Filter
	for _, filter := range filters {
		switch filter.Name() {
		case "is_active":
			boolFilter = filter
		case "username":
			textFilter = filter
		}
	}
	
	assert.NotNil(t, boolFilter, "Should have boolean filter for is_active")
	assert.NotNil(t, textFilter, "Should have text filter for username")
	
	// Test boolean filter choices
	if boolFilter != nil {
		choices := boolFilter.Choices()
		assert.Len(t, choices, 3) // All, Yes, No
		assert.Equal(t, "", choices[0].Value)
		assert.Equal(t, "All", choices[0].Display)
	}
}

func TestDefaultActions(t *testing.T) {
	registry := NewActionRegistry()
	
	// Check default actions are registered
	assert.Contains(t, registry.actions, "delete_selected")
	assert.Contains(t, registry.actions, "export_csv")
	assert.Contains(t, registry.actions, "export_json")
	
	deleteAction, exists := registry.Get("delete_selected")
	assert.True(t, exists)
	assert.Equal(t, "Delete selected items", deleteAction.Description)
}

func TestModelNameExtraction(t *testing.T) {
	testCases := []struct {
		model    interface{}
		expected string
	}{
		{&TestUser{}, "main.testuser"},
		{&TestPost{}, "main.testpost"},
		{TestUser{}, "main.testuser"},
	}
	
	for _, tc := range testCases {
		name := getModelName(tc.model)
		assert.Equal(t, tc.expected, name, "Model name extraction failed for %T", tc.model)
	}
}

func BenchmarkModelRegistration(b *testing.B) {
	site := NewSite("benchmark")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		site.Register(&TestUser{}, nil)
		site.Unregister(&TestUser{})
	}
}

func BenchmarkGetModelAdmin(b *testing.B) {
	site := NewSite("benchmark")
	site.Register(&TestUser{}, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = site.GetModelAdmin("main.testuser")
	}
}