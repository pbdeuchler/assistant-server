package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pbdeuchler/assistant-mcp/dao/postgres"
	"github.com/pbdeuchler/assistant-mcp/service"
)

// Integration tests that test the complete flow from HTTP request to response
// These tests use mock DAO implementations to avoid database dependencies

type IntegrationTestSuite struct {
	router http.Handler
	dao    *mockIntegrationDAO
}

type mockIntegrationDAO struct {
	todos       map[string]postgres.Todo
	backgrounds map[string]postgres.Background
	preferences map[string]map[string]postgres.Preferences
	notes       map[string]postgres.Notes
	nextUID     int
}

func newMockIntegrationDAO() *mockIntegrationDAO {
	return &mockIntegrationDAO{
		todos:       make(map[string]postgres.Todo),
		backgrounds: make(map[string]postgres.Background),
		preferences: make(map[string]map[string]postgres.Preferences),
		notes:       make(map[string]postgres.Notes),
		nextUID:     1,
	}
}

func (m *mockIntegrationDAO) CreateTodo(ctx context.Context, t postgres.Todo) (postgres.Todo, error) {
	uid := fmt.Sprintf("todo-%d", m.nextUID)
	m.nextUID++
	
	now := time.Now()
	t.UID = uid
	t.CreatedAt = now
	t.UpdatedAt = now
	
	m.todos[uid] = t
	return t, nil
}

func (m *mockIntegrationDAO) GetTodo(ctx context.Context, uid string) (postgres.Todo, error) {
	if todo, exists := m.todos[uid]; exists {
		return todo, nil
	}
	return postgres.Todo{}, fmt.Errorf("todo not found")
}

func (m *mockIntegrationDAO) ListTodos(ctx context.Context, options postgres.ListOptions) ([]postgres.Todo, error) {
	var result []postgres.Todo
	for _, todo := range m.todos {
		result = append(result, todo)
	}
	
	// Simple filtering for testing
	if options.WhereClause != "" && len(options.WhereArgs) > 0 {
		filtered := []postgres.Todo{}
		for _, todo := range result {
			if strings.Contains(options.WhereClause, "priority") && len(options.WhereArgs) > 0 {
				if fmt.Sprintf("%d", todo.Priority) == fmt.Sprintf("%v", options.WhereArgs[0]) {
					filtered = append(filtered, todo)
				}
			} else if strings.Contains(options.WhereClause, "title") && len(options.WhereArgs) > 0 {
				if strings.Contains(todo.Title, fmt.Sprintf("%v", options.WhereArgs[0])) {
					filtered = append(filtered, todo)
				}
			}
		}
		result = filtered
	}
	
	// Apply limit and offset
	start := options.Offset
	if start > len(result) {
		return []postgres.Todo{}, nil
	}
	
	end := start + options.Limit
	if end > len(result) {
		end = len(result)
	}
	
	return result[start:end], nil
}

func (m *mockIntegrationDAO) UpdateTodo(ctx context.Context, uid string, t postgres.Todo) (postgres.Todo, error) {
	if _, exists := m.todos[uid]; !exists {
		return postgres.Todo{}, fmt.Errorf("todo not found")
	}
	
	t.UID = uid
	t.UpdatedAt = time.Now()
	if existing := m.todos[uid]; t.CreatedAt.IsZero() {
		t.CreatedAt = existing.CreatedAt
	}
	
	m.todos[uid] = t
	return t, nil
}

func (m *mockIntegrationDAO) DeleteTodo(ctx context.Context, uid string) error {
	if _, exists := m.todos[uid]; !exists {
		return fmt.Errorf("todo not found")
	}
	delete(m.todos, uid)
	return nil
}

func (m *mockIntegrationDAO) CreateBackground(ctx context.Context, b postgres.Background) (postgres.Background, error) {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	m.backgrounds[b.Key] = b
	return b, nil
}

func (m *mockIntegrationDAO) GetBackground(ctx context.Context, key string) (postgres.Background, error) {
	if bg, exists := m.backgrounds[key]; exists {
		return bg, nil
	}
	return postgres.Background{}, fmt.Errorf("background not found")
}

func (m *mockIntegrationDAO) ListBackgrounds(ctx context.Context, options postgres.ListOptions) ([]postgres.Background, error) {
	var result []postgres.Background
	for _, bg := range m.backgrounds {
		result = append(result, bg)
	}
	
	start := options.Offset
	if start > len(result) {
		return []postgres.Background{}, nil
	}
	
	end := start + options.Limit
	if end > len(result) {
		end = len(result)
	}
	
	return result[start:end], nil
}

func (m *mockIntegrationDAO) UpdateBackground(ctx context.Context, key string, b postgres.Background) (postgres.Background, error) {
	if _, exists := m.backgrounds[key]; !exists {
		return postgres.Background{}, fmt.Errorf("background not found")
	}
	
	b.Key = key
	b.UpdatedAt = time.Now()
	if existing := m.backgrounds[key]; b.CreatedAt.IsZero() {
		b.CreatedAt = existing.CreatedAt
	}
	
	m.backgrounds[key] = b
	return b, nil
}

func (m *mockIntegrationDAO) DeleteBackground(ctx context.Context, key string) error {
	if _, exists := m.backgrounds[key]; !exists {
		return fmt.Errorf("background not found")
	}
	delete(m.backgrounds, key)
	return nil
}

func (m *mockIntegrationDAO) CreatePreferences(ctx context.Context, p postgres.Preferences) (postgres.Preferences, error) {
	if m.preferences[p.Key] == nil {
		m.preferences[p.Key] = make(map[string]postgres.Preferences)
	}
	
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	
	m.preferences[p.Key][p.Specifier] = p
	return p, nil
}

func (m *mockIntegrationDAO) GetPreferences(ctx context.Context, key, specifier string) (postgres.Preferences, error) {
	if keyMap, exists := m.preferences[key]; exists {
		if pref, exists := keyMap[specifier]; exists {
			return pref, nil
		}
	}
	return postgres.Preferences{}, fmt.Errorf("preferences not found")
}

func (m *mockIntegrationDAO) ListPreferences(ctx context.Context, options postgres.ListOptions) ([]postgres.Preferences, error) {
	var result []postgres.Preferences
	for _, keyMap := range m.preferences {
		for _, pref := range keyMap {
			result = append(result, pref)
		}
	}
	
	start := options.Offset
	if start > len(result) {
		return []postgres.Preferences{}, nil
	}
	
	end := start + options.Limit
	if end > len(result) {
		end = len(result)
	}
	
	return result[start:end], nil
}

func (m *mockIntegrationDAO) UpdatePreferences(ctx context.Context, key, specifier string, p postgres.Preferences) (postgres.Preferences, error) {
	if keyMap, exists := m.preferences[key]; exists {
		if _, exists := keyMap[specifier]; exists {
			p.Key = key
			p.Specifier = specifier
			p.UpdatedAt = time.Now()
			if existing := keyMap[specifier]; p.CreatedAt.IsZero() {
				p.CreatedAt = existing.CreatedAt
			}
			
			keyMap[specifier] = p
			return p, nil
		}
	}
	return postgres.Preferences{}, fmt.Errorf("preferences not found")
}

func (m *mockIntegrationDAO) DeletePreferences(ctx context.Context, key, specifier string) error {
	if keyMap, exists := m.preferences[key]; exists {
		if _, exists := keyMap[specifier]; exists {
			delete(keyMap, specifier)
			return nil
		}
	}
	return fmt.Errorf("preferences not found")
}

func (m *mockIntegrationDAO) CreateNotes(ctx context.Context, n postgres.Notes) (postgres.Notes, error) {
	if n.ID == "" {
		n.ID = fmt.Sprintf("note-%d", m.nextUID)
		m.nextUID++
	}
	
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	
	m.notes[n.ID] = n
	return n, nil
}

func (m *mockIntegrationDAO) GetNotes(ctx context.Context, id string) (postgres.Notes, error) {
	if note, exists := m.notes[id]; exists {
		return note, nil
	}
	return postgres.Notes{}, fmt.Errorf("note not found")
}

func (m *mockIntegrationDAO) ListNotes(ctx context.Context, options postgres.ListOptions) ([]postgres.Notes, error) {
	var result []postgres.Notes
	for _, note := range m.notes {
		result = append(result, note)
	}
	
	start := options.Offset
	if start > len(result) {
		return []postgres.Notes{}, nil
	}
	
	end := start + options.Limit
	if end > len(result) {
		end = len(result)
	}
	
	return result[start:end], nil
}

func (m *mockIntegrationDAO) UpdateNotes(ctx context.Context, id string, n postgres.Notes) (postgres.Notes, error) {
	if _, exists := m.notes[id]; !exists {
		return postgres.Notes{}, fmt.Errorf("note not found")
	}
	
	n.ID = id
	n.UpdatedAt = time.Now()
	if existing := m.notes[id]; n.CreatedAt.IsZero() {
		n.CreatedAt = existing.CreatedAt
	}
	
	m.notes[id] = n
	return n, nil
}

func (m *mockIntegrationDAO) DeleteNotes(ctx context.Context, id string) error {
	if _, exists := m.notes[id]; !exists {
		return fmt.Errorf("note not found")
	}
	delete(m.notes, id)
	return nil
}

func setupIntegrationTestSuite() *IntegrationTestSuite {
	dao := newMockIntegrationDAO()
	
	r := chi.NewRouter()
	r.Mount("/todos", service.NewHandlers(dao))
	r.Mount("/backgrounds", service.NewBackgroundHandlers(dao))
	r.Mount("/preferences", service.NewPreferencesHandlers(dao))
	
	return &IntegrationTestSuite{
		router: r,
		dao:    dao,
	}
}

func TestTodoCompleteWorkflow(t *testing.T) {
	suite := setupIntegrationTestSuite()
	
	// Test Create Todo
	todoJSON := `{"title":"Integration Test Todo","description":"Testing complete workflow","priority":3}`
	req := httptest.NewRequest("POST", "/todos/", strings.NewReader(todoJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Create todo failed: %d", w.Code)
	}
	
	var createdTodo postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&createdTodo); err != nil {
		t.Fatalf("Failed to decode created todo: %v", err)
	}
	
	// Test Get Todo
	req = httptest.NewRequest("GET", "/todos/"+createdTodo.UID, nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Get todo failed: %d", w.Code)
	}
	
	var fetchedTodo postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&fetchedTodo); err != nil {
		t.Fatalf("Failed to decode fetched todo: %v", err)
	}
	
	if fetchedTodo.UID != createdTodo.UID {
		t.Errorf("Expected UID %s, got %s", createdTodo.UID, fetchedTodo.UID)
	}
	
	// Test Update Todo
	updateJSON := `{"title":"Updated Todo Title","description":"Updated description","priority":2}`
	req = httptest.NewRequest("PUT", "/todos/"+createdTodo.UID, strings.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Update todo failed: %d", w.Code)
	}
	
	var updatedTodo postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&updatedTodo); err != nil {
		t.Fatalf("Failed to decode updated todo: %v", err)
	}
	
	if updatedTodo.Title != "Updated Todo Title" {
		t.Errorf("Expected title 'Updated Todo Title', got '%s'", updatedTodo.Title)
	}
	
	// Test List Todos
	req = httptest.NewRequest("GET", "/todos/", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("List todos failed: %d", w.Code)
	}
	
	var todos []postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode todos list: %v", err)
	}
	
	if len(todos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(todos))
	}
	
	// Test Delete Todo
	req = httptest.NewRequest("DELETE", "/todos/"+createdTodo.UID, nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusNoContent {
		t.Fatalf("Delete todo failed: %d", w.Code)
	}
	
	// Verify deletion
	req = httptest.NewRequest("GET", "/todos/"+createdTodo.UID, nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 after deletion, got %d", w.Code)
	}
}

func TestBackgroundCompleteWorkflow(t *testing.T) {
	suite := setupIntegrationTestSuite()
	
	// Test Create Background
	bgJSON := `{"key":"test-theme","value":"dark-mode-config"}`
	req := httptest.NewRequest("POST", "/backgrounds/", strings.NewReader(bgJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Create background failed: %d", w.Code)
	}
	
	var createdBg postgres.Background
	if err := json.NewDecoder(w.Body).Decode(&createdBg); err != nil {
		t.Fatalf("Failed to decode created background: %v", err)
	}
	
	// Test Get Background
	req = httptest.NewRequest("GET", "/backgrounds/test-theme", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Get background failed: %d", w.Code)
	}
	
	// Test List Backgrounds
	req = httptest.NewRequest("GET", "/backgrounds/?sort_by=key&sort_dir=ASC", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("List backgrounds failed: %d", w.Code)
	}
	
	var backgrounds []postgres.Background
	if err := json.NewDecoder(w.Body).Decode(&backgrounds); err != nil {
		t.Fatalf("Failed to decode backgrounds list: %v", err)
	}
	
	if len(backgrounds) != 1 {
		t.Errorf("Expected 1 background, got %d", len(backgrounds))
	}
}

func TestPreferencesCompleteWorkflow(t *testing.T) {
	suite := setupIntegrationTestSuite()
	
	// Test Create Preferences
	prefJSON := `{"key":"ui","specifier":"theme","data":"{\"mode\":\"dark\",\"accent\":\"blue\"}"}`
	req := httptest.NewRequest("POST", "/preferences/", strings.NewReader(prefJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Create preferences failed: %d", w.Code)
	}
	
	var createdPref postgres.Preferences
	if err := json.NewDecoder(w.Body).Decode(&createdPref); err != nil {
		t.Fatalf("Failed to decode created preferences: %v", err)
	}
	
	// Test Get Preferences
	req = httptest.NewRequest("GET", "/preferences/ui/theme", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Get preferences failed: %d", w.Code)
	}
	
	var fetchedPref postgres.Preferences
	if err := json.NewDecoder(w.Body).Decode(&fetchedPref); err != nil {
		t.Fatalf("Failed to decode fetched preferences: %v", err)
	}
	
	if fetchedPref.Key != "ui" || fetchedPref.Specifier != "theme" {
		t.Errorf("Expected key 'ui' and specifier 'theme', got '%s' and '%s'", fetchedPref.Key, fetchedPref.Specifier)
	}
	
	// Test Update Preferences
	updateJSON := `{"key":"ui","specifier":"theme","data":"{\"mode\":\"light\",\"accent\":\"green\"}"}`
	req = httptest.NewRequest("PUT", "/preferences/ui/theme", strings.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Update preferences failed: %d", w.Code)
	}
	
	// Test List Preferences
	req = httptest.NewRequest("GET", "/preferences/?limit=50", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("List preferences failed: %d", w.Code)
	}
	
	var preferences []postgres.Preferences
	if err := json.NewDecoder(w.Body).Decode(&preferences); err != nil {
		t.Fatalf("Failed to decode preferences list: %v", err)
	}
	
	if len(preferences) != 1 {
		t.Errorf("Expected 1 preference, got %d", len(preferences))
	}
}

func TestListEndpointsWithFiltering(t *testing.T) {
	suite := setupIntegrationTestSuite()
	
	// Create test data
	todo1JSON := `{"title":"High Priority Task","description":"Important task","priority":3}`
	todo2JSON := `{"title":"Low Priority Task","description":"Less important","priority":1}`
	
	// Create todos
	for _, todoJSON := range []string{todo1JSON, todo2JSON} {
		req := httptest.NewRequest("POST", "/todos/", strings.NewReader(todoJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Fatalf("Failed to create test todo: %d", w.Code)
		}
	}
	
	// Test filtering by priority
	req := httptest.NewRequest("GET", "/todos/?priority=3", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("List todos with filter failed: %d", w.Code)
	}
	
	var filteredTodos []postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&filteredTodos); err != nil {
		t.Fatalf("Failed to decode filtered todos: %v", err)
	}
	
	if len(filteredTodos) != 1 {
		t.Errorf("Expected 1 filtered todo, got %d", len(filteredTodos))
	}
	
	if len(filteredTodos) > 0 && filteredTodos[0].Priority != postgres.PriorityHigh {
		t.Errorf("Expected high priority todo, got priority %d", filteredTodos[0].Priority)
	}
}

func TestErrorHandling(t *testing.T) {
	suite := setupIntegrationTestSuite()
	
	// Test 404 for non-existent todo
	req := httptest.NewRequest("GET", "/todos/nonexistent-id", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent todo, got %d", w.Code)
	}
	
	// Test 400 for invalid JSON
	req = httptest.NewRequest("POST", "/todos/", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
	}
}