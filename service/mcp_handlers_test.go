package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
)

type MockTodoDAO struct {
	mock.Mock
}

func (m *MockTodoDAO) CreateTodo(ctx context.Context, t dao.Todo) (dao.Todo, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(dao.Todo), args.Error(1)
}

func (m *MockTodoDAO) GetTodo(ctx context.Context, uid string) (dao.Todo, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(dao.Todo), args.Error(1)
}

func (m *MockTodoDAO) ListTodos(ctx context.Context, options dao.ListOptions) ([]dao.Todo, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]dao.Todo), args.Error(1)
}

func (m *MockTodoDAO) UpdateTodo(ctx context.Context, uid string, t dao.UpdateTodo) (dao.Todo, error) {
	args := m.Called(ctx, uid, t)
	return args.Get(0).(dao.Todo), args.Error(1)
}

func (m *MockTodoDAO) DeleteTodo(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

type MockNotesDAO struct {
	mock.Mock
}

func (m *MockNotesDAO) CreateNotes(ctx context.Context, n dao.Notes) (dao.Notes, error) {
	args := m.Called(ctx, n)
	return args.Get(0).(dao.Notes), args.Error(1)
}

func (m *MockNotesDAO) GetNotes(ctx context.Context, id string) (dao.Notes, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(dao.Notes), args.Error(1)
}

func (m *MockNotesDAO) ListNotes(ctx context.Context, options dao.ListOptions) ([]dao.Notes, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]dao.Notes), args.Error(1)
}

func (m *MockNotesDAO) UpdateNotes(ctx context.Context, id string, n dao.Notes) (dao.Notes, error) {
	args := m.Called(ctx, id, n)
	return args.Get(0).(dao.Notes), args.Error(1)
}

func (m *MockNotesDAO) DeleteNotes(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockPreferencesDAO struct {
	mock.Mock
}

func (m *MockPreferencesDAO) CreatePreferences(ctx context.Context, p dao.Preferences) (dao.Preferences, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(dao.Preferences), args.Error(1)
}

func (m *MockPreferencesDAO) GetPreferences(ctx context.Context, key, specifier string) (dao.Preferences, error) {
	args := m.Called(ctx, key, specifier)
	return args.Get(0).(dao.Preferences), args.Error(1)
}

func (m *MockPreferencesDAO) ListPreferences(ctx context.Context, options dao.ListOptions) ([]dao.Preferences, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]dao.Preferences), args.Error(1)
}

func (m *MockPreferencesDAO) UpdatePreferences(ctx context.Context, key, specifier string, p dao.Preferences) (dao.Preferences, error) {
	args := m.Called(ctx, key, specifier, p)
	return args.Get(0).(dao.Preferences), args.Error(1)
}

func (m *MockPreferencesDAO) DeletePreferences(ctx context.Context, key, specifier string) error {
	args := m.Called(ctx, key, specifier)
	return args.Error(0)
}

type MockRecipesDAO struct {
	mock.Mock
}

func (m *MockRecipesDAO) CreateRecipes(ctx context.Context, r dao.Recipes) (dao.Recipes, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(dao.Recipes), args.Error(1)
}

func (m *MockRecipesDAO) GetRecipes(ctx context.Context, id string) (dao.Recipes, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(dao.Recipes), args.Error(1)
}

func (m *MockRecipesDAO) ListRecipes(ctx context.Context, options dao.ListOptions) ([]dao.Recipes, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]dao.Recipes), args.Error(1)
}

func (m *MockRecipesDAO) UpdateRecipes(ctx context.Context, id string, r dao.Recipes) (dao.Recipes, error) {
	args := m.Called(ctx, id, r)
	return args.Get(0).(dao.Recipes), args.Error(1)
}

func (m *MockRecipesDAO) DeleteRecipes(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestMCPHandlers_CreateTodo(t *testing.T) {
	tests := []struct {
		name          string
		request       map[string]any
		mockTodo      dao.Todo
		mockError     error
		expectedError bool
	}{
		{
			name: "successful todo creation",
			request: map[string]any{
				"title":       "Test Todo",
				"description": "Test Description",
				"priority":    float64(4),
				"user_id":     "user123",
			},
			mockTodo: dao.Todo{
				UID:         "todo123",
				Title:       "Test Todo",
				Description: "Test Description",
				Priority:    dao.Priority(4),
				UserID:      "user123",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "missing title",
			request: map[string]any{
				"description": "Test Description",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &MockTodoDAO{}
			if !tt.expectedError {
				mockDAO.On("CreateTodo", mock.Anything, mock.AnythingOfType("postgres.Todo")).Return(tt.mockTodo, tt.mockError)
			}

			h := &MCPHandlers{todoDAO: mockDAO}
			result := h.handleCreateTodo(context.Background(), tt.request)

			if tt.expectedError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				assert.NotNil(t, result)
				if tt.mockError == nil {
					assert.Len(t, result.Content, 1)
					if textContent, ok := result.Content[0].(mcp.TextContent); ok {
						assert.Contains(t, textContent.Text, "Todo created successfully")
					}
				}
			}

			if !tt.expectedError {
				mockDAO.AssertExpectations(t)
			}
		})
	}
}

func TestMCPHandlers_ListTodos(t *testing.T) {
	tests := []struct {
		name      string
		request   map[string]any
		mockTodos []dao.Todo
		mockError error
	}{
		{
			name: "successful todo listing",
			request: map[string]any{
				"user_id": "user123",
				"limit":   float64(10),
			},
			mockTodos: []dao.Todo{
				{UID: "todo1", Title: "Todo 1", UserID: "user123"},
				{UID: "todo2", Title: "Todo 2", UserID: "user123"},
			},
			mockError: nil,
		},
		{
			name: "todo listing with tags filter",
			request: map[string]any{
				"user_id": "user123",
				"tags":    "urgent,work",
				"limit":   float64(5),
			},
			mockTodos: []dao.Todo{
				{UID: "todo1", Title: "Work Task", UserID: "user123"},
			},
			mockError: nil,
		},
		{
			name:      "empty request",
			request:   map[string]any{},
			mockTodos: []dao.Todo{},
			mockError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &MockTodoDAO{}
			mockDAO.On("ListTodos", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return(tt.mockTodos, tt.mockError)

			h := &MCPHandlers{todoDAO: mockDAO}
			result := h.handleListTodos(context.Background(), tt.request)

			assert.False(t, result.IsError)
			assert.NotNil(t, result)

			if tt.mockError == nil {
				assert.Len(t, result.Content, 1)
				if textContent, ok := result.Content[0].(mcp.TextContent); ok {
					var todos []dao.Todo
					json.Unmarshal([]byte(textContent.Text), &todos)
					assert.Equal(t, len(tt.mockTodos), len(todos))
				}
			}

			mockDAO.AssertExpectations(t)
		})
	}
}

func TestMCPHandlers_CompleteTodo(t *testing.T) {
	tests := []struct {
		name          string
		request       map[string]any
		mockTodo      dao.Todo
		mockError     error
		expectedError bool
	}{
		{
			name: "successful todo completion",
			request: map[string]any{
				"todo_id":      "todo123",
				"completed_by": "user123",
			},
			mockTodo:      dao.Todo{UID: "todo123"},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "missing todo_id",
			request: map[string]any{
				"completed_by": "user123",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &MockTodoDAO{}
			if !tt.expectedError {
				mockDAO.On("UpdateTodo", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("postgres.UpdateTodo")).Return(tt.mockTodo, tt.mockError)
			}

			h := &MCPHandlers{todoDAO: mockDAO}
			result := h.handleCompleteTodo(context.Background(), tt.request)

			if tt.expectedError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				assert.NotNil(t, result)
			}

			if !tt.expectedError {
				mockDAO.AssertExpectations(t)
			}
		})
	}
}

func TestMCPHandlers_HTTPIntegration(t *testing.T) {
	mockTodoDAO := &MockTodoDAO{}
	mockNotesDAO := &MockNotesDAO{}
	mockPrefsDAO := &MockPreferencesDAO{}
	mockRecipesDAO := &MockRecipesDAO{}

	mockTodo := dao.Todo{
		UID:         "test-todo-id",
		Title:       "Test Todo",
		Description: "Test Description",
		Priority:    dao.Priority(3),
		Data:        "{}",
	}

	mockTodoDAO.On("CreateTodo", mock.Anything, mock.AnythingOfType("postgres.Todo")).Return(mockTodo, nil)

	router := NewMCPRouter(mockTodoDAO, mockNotesDAO, mockPrefsDAO, mockRecipesDAO)

	mcpRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_todo",
			"arguments": map[string]any{
				"title":       "Test Todo",
				"description": "Test Description",
				"priority":    float64(3),
			},
		},
	}

	reqBody, _ := json.Marshal(mcpRequest)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "2.0", response["jsonrpc"])
	assert.Equal(t, float64(1), response["id"])

	mockTodoDAO.AssertExpectations(t)
}

func TestMCPHandlers_FindRecipes(t *testing.T) {
	tests := []struct {
		name        string
		request     map[string]any
		mockRecipes []dao.Recipes
		mockError   error
	}{
		{
			name: "successful recipe search",
			request: map[string]any{
				"user_id": "user123",
				"title":   "pasta",
				"limit":   float64(10),
			},
			mockRecipes: []dao.Recipes{
				{ID: "recipe1", Title: "Pasta Carbonara", UserID: "user123"},
				{ID: "recipe2", Title: "Pasta Bolognese", UserID: "user123"},
			},
			mockError: nil,
		},
		{
			name: "recipe search with tags filter",
			request: map[string]any{
				"user_id": "user123",
				"tags":    "italian,dinner",
				"limit":   float64(5),
			},
			mockRecipes: []dao.Recipes{
				{ID: "recipe1", Title: "Pasta Carbonara", UserID: "user123"},
			},
			mockError: nil,
		},
		{
			name:        "empty request",
			request:     map[string]any{},
			mockRecipes: []dao.Recipes{},
			mockError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &MockRecipesDAO{}
			mockDAO.On("ListRecipes", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return(tt.mockRecipes, tt.mockError)

			h := &MCPHandlers{recipesDAO: mockDAO}
			result := h.handleFindRecipes(context.Background(), tt.request)

			assert.False(t, result.IsError)
			assert.NotNil(t, result)

			if tt.mockError == nil {
				assert.Len(t, result.Content, 1)
				if textContent, ok := result.Content[0].(mcp.TextContent); ok {
					var recipes []dao.Recipes
					json.Unmarshal([]byte(textContent.Text), &recipes)
					assert.Equal(t, len(tt.mockRecipes), len(recipes))
				}
			}

			mockDAO.AssertExpectations(t)
		})
	}
}

func TestMCPHandlers_ToolsList(t *testing.T) {
	mockTodoDAO := &MockTodoDAO{}
	mockNotesDAO := &MockNotesDAO{}
	mockPrefsDAO := &MockPreferencesDAO{}
	mockRecipesDAO := &MockRecipesDAO{}

	router := NewMCPRouter(mockTodoDAO, mockNotesDAO, mockPrefsDAO, mockRecipesDAO)

	mcpRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}

	reqBody, _ := json.Marshal(mcpRequest)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "2.0", response["jsonrpc"])

	result, ok := response["result"].(map[string]any)
	assert.True(t, ok)

	tools, ok := result["tools"].([]any)
	assert.True(t, ok)
	assert.Len(t, tools, 11) // We have 11 tools defined
}