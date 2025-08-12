package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pbdeuchler/assistant-server/dao/postgres"
	"github.com/pbdeuchler/assistant-server/mocks"
	"github.com/stretchr/testify/mock"
)

func TestTodoCreate(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	expectedTodo := postgres.Todo{
		UID:         "test-uid",
		Title:       "Test Todo",
		Description: "Test Description",
		Data:        "{}",
		Priority:    postgres.PriorityMedium,
		DueDate:     nil,
		RecursOn:    "",
		ExternalURL: "",
		UserID:      "user-123",
		HouseholdID: "household-456",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockTodoDAO.On("CreateTodo", 
		mock.Anything, 
		mock.MatchedBy(func(t postgres.Todo) bool {
			return t.Title == "Test Todo" && 
				   t.Description == "Test Description" &&
				   t.Priority == postgres.PriorityMedium &&
				   t.UserID == "user-123" &&
				   t.HouseholdID == "household-456"
		})).Return(expectedTodo, nil)

	handler := NewTodos(mockTodoDAO)

	reqBody := `{
		"title": "Test Todo",
		"description": "Test Description",
		"data": "{}",
		"priority": 2,
		"user_id": "user-123",
		"household_id": "household-456"
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Title != expectedTodo.Title {
		t.Errorf("Expected title %s, got %s", expectedTodo.Title, response.Title)
	}
}

func TestTodoCreateInvalidJSON(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	handler := NewTodos(mockTodoDAO)

	req := httptest.NewRequest("POST", "/", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestTodoGet(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	expectedTodo := postgres.Todo{
		UID:         "test-uid",
		Title:       "Test Todo",
		Description: "Test Description",
		Data:        "{}",
		Priority:    postgres.PriorityMedium,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockTodoDAO.On("GetTodo", mock.Anything, "test-uid").Return(expectedTodo, nil)

	handler := NewTodos(mockTodoDAO)
	
	req := httptest.NewRequest("GET", "/test-uid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.UID != expectedTodo.UID {
		t.Errorf("Expected UID %s, got %s", expectedTodo.UID, response.UID)
	}
}

func TestTodoGetNotFound(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	mockTodoDAO.On("GetTodo", mock.Anything, "nonexistent").Return(postgres.Todo{}, errors.New("not found"))

	handler := NewTodos(mockTodoDAO)
	
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestTodoList(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	expectedTodos := []postgres.Todo{
		{
			UID:         "test-uid-1",
			Title:       "Test Todo 1",
			Description: "Test Description 1",
			Priority:    postgres.PriorityHigh,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			UID:         "test-uid-2", 
			Title:       "Test Todo 2",
			Description: "Test Description 2",
			Priority:    postgres.PriorityLow,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockTodoDAO.On("ListTodos", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return(expectedTodos, nil)

	handler := NewTodos(mockTodoDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response []postgres.Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(expectedTodos) {
		t.Errorf("Expected %d todos, got %d", len(expectedTodos), len(response))
	}
}

func TestTodoUpdate(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	expectedTodo := postgres.Todo{
		UID:         "test-uid",
		Title:       "Updated Todo",
		Description: "Updated Description",
		Data:        "{}",
		Priority:    postgres.PriorityMedium,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockTodoDAO.On("UpdateTodo", mock.Anything, "test-uid", mock.AnythingOfType("postgres.UpdateTodo")).Return(expectedTodo, nil)

	handler := NewTodos(mockTodoDAO)

	reqBody := `{
		"title": "Updated Todo",
		"description": "Updated Description"
	}`

	req := httptest.NewRequest("PUT", "/test-uid", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Title != expectedTodo.Title {
		t.Errorf("Expected title %s, got %s", expectedTodo.Title, response.Title)
	}
}

func TestTodoDelete(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	mockTodoDAO.On("DeleteTodo", mock.Anything, "test-uid").Return(nil)

	handler := NewTodos(mockTodoDAO)
	
	req := httptest.NewRequest("DELETE", "/test-uid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rr.Code)
	}
}

func TestTodoUpdateInvalidJSON(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	handler := NewTodos(mockTodoDAO)

	req := httptest.NewRequest("PUT", "/test-uid", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestTodoUpdateError(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	mockTodoDAO.On("UpdateTodo", mock.Anything, "test-uid", mock.AnythingOfType("postgres.UpdateTodo")).Return(postgres.Todo{}, errors.New("database error"))

	handler := NewTodos(mockTodoDAO)

	reqBody := `{
		"title": "Updated Todo",
		"description": "Updated Description"
	}`

	req := httptest.NewRequest("PUT", "/test-uid", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestTodoDeleteError(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	mockTodoDAO.On("DeleteTodo", mock.Anything, "test-uid").Return(errors.New("database error"))

	handler := NewTodos(mockTodoDAO)
	
	req := httptest.NewRequest("DELETE", "/test-uid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestTodoCreateError(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	mockTodoDAO.On("CreateTodo", mock.Anything, mock.AnythingOfType("postgres.Todo")).Return(postgres.Todo{}, errors.New("database error"))

	handler := NewTodos(mockTodoDAO)

	reqBody := `{
		"title": "Test Todo",
		"description": "Test Description",
		"priority": 2,
		"user_id": "user-123",
		"household_id": "household-456"
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestTodoListError(t *testing.T) {
	mockTodoDAO := mocks.NewMocktodoDAO(t)
	
	mockTodoDAO.On("ListTodos", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return([]postgres.Todo{}, errors.New("database error"))

	handler := NewTodos(mockTodoDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}