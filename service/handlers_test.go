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
	"github.com/pbdeuchler/assistant-mcp/dao/postgres"
)

// Mock DAO for testing
type mockDAO struct {
	createTodoFunc  func(ctx context.Context, t postgres.Todo) (postgres.Todo, error)
	getTodoFunc     func(ctx context.Context, uid string) (postgres.Todo, error)
	listTodosFunc   func(ctx context.Context, options postgres.ListOptions) ([]postgres.Todo, error)
	updateTodoFunc  func(ctx context.Context, uid string, t postgres.Todo) (postgres.Todo, error)
	deleteTodoFunc  func(ctx context.Context, uid string) error
	createBgFunc    func(ctx context.Context, b postgres.Background) (postgres.Background, error)
	getBgFunc       func(ctx context.Context, key string) (postgres.Background, error)
	listBgFunc      func(ctx context.Context, options postgres.ListOptions) ([]postgres.Background, error)
	updateBgFunc    func(ctx context.Context, key string, b postgres.Background) (postgres.Background, error)
	deleteBgFunc    func(ctx context.Context, key string) error
	createPrefFunc  func(ctx context.Context, p postgres.Preferences) (postgres.Preferences, error)
	getPrefFunc     func(ctx context.Context, key, spec string) (postgres.Preferences, error)
	listPrefFunc    func(ctx context.Context, options postgres.ListOptions) ([]postgres.Preferences, error)
	updatePrefFunc  func(ctx context.Context, key, spec string, p postgres.Preferences) (postgres.Preferences, error)
	deletePrefFunc  func(ctx context.Context, key, spec string) error
	createNotesFunc func(ctx context.Context, n postgres.Notes) (postgres.Notes, error)
	getNotesFunc    func(ctx context.Context, id string) (postgres.Notes, error)
	listNotesFunc   func(ctx context.Context, options postgres.ListOptions) ([]postgres.Notes, error)
	updateNotesFunc func(ctx context.Context, id string, n postgres.Notes) (postgres.Notes, error)
	deleteNotesFunc func(ctx context.Context, id string) error
}

func (m *mockDAO) CreateTodo(ctx context.Context, t postgres.Todo) (postgres.Todo, error) {
	if m.createTodoFunc != nil {
		return m.createTodoFunc(ctx, t)
	}
	return postgres.Todo{}, errors.New("not implemented")
}

func (m *mockDAO) GetTodo(ctx context.Context, uid string) (postgres.Todo, error) {
	if m.getTodoFunc != nil {
		return m.getTodoFunc(ctx, uid)
	}
	return postgres.Todo{}, errors.New("not implemented")
}

func (m *mockDAO) ListTodos(ctx context.Context, options postgres.ListOptions) ([]postgres.Todo, error) {
	if m.listTodosFunc != nil {
		return m.listTodosFunc(ctx, options)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDAO) UpdateTodo(ctx context.Context, uid string, t postgres.Todo) (postgres.Todo, error) {
	if m.updateTodoFunc != nil {
		return m.updateTodoFunc(ctx, uid, t)
	}
	return postgres.Todo{}, errors.New("not implemented")
}

func (m *mockDAO) DeleteTodo(ctx context.Context, uid string) error {
	if m.deleteTodoFunc != nil {
		return m.deleteTodoFunc(ctx, uid)
	}
	return errors.New("not implemented")
}

func (m *mockDAO) CreateBackground(ctx context.Context, b postgres.Background) (postgres.Background, error) {
	if m.createBgFunc != nil {
		return m.createBgFunc(ctx, b)
	}
	return postgres.Background{}, errors.New("not implemented")
}

func (m *mockDAO) GetBackground(ctx context.Context, key string) (postgres.Background, error) {
	if m.getBgFunc != nil {
		return m.getBgFunc(ctx, key)
	}
	return postgres.Background{}, errors.New("not implemented")
}

func (m *mockDAO) ListBackgrounds(ctx context.Context, options postgres.ListOptions) ([]postgres.Background, error) {
	if m.listBgFunc != nil {
		return m.listBgFunc(ctx, options)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDAO) UpdateBackground(ctx context.Context, key string, b postgres.Background) (postgres.Background, error) {
	if m.updateBgFunc != nil {
		return m.updateBgFunc(ctx, key, b)
	}
	return postgres.Background{}, errors.New("not implemented")
}

func (m *mockDAO) DeleteBackground(ctx context.Context, key string) error {
	if m.deleteBgFunc != nil {
		return m.deleteBgFunc(ctx, key)
	}
	return errors.New("not implemented")
}

func (m *mockDAO) CreatePreferences(ctx context.Context, p postgres.Preferences) (postgres.Preferences, error) {
	if m.createPrefFunc != nil {
		return m.createPrefFunc(ctx, p)
	}
	return postgres.Preferences{}, errors.New("not implemented")
}

func (m *mockDAO) GetPreferences(ctx context.Context, key, spec string) (postgres.Preferences, error) {
	if m.getPrefFunc != nil {
		return m.getPrefFunc(ctx, key, spec)
	}
	return postgres.Preferences{}, errors.New("not implemented")
}

func (m *mockDAO) ListPreferences(ctx context.Context, options postgres.ListOptions) ([]postgres.Preferences, error) {
	if m.listPrefFunc != nil {
		return m.listPrefFunc(ctx, options)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDAO) UpdatePreferences(ctx context.Context, key, spec string, p postgres.Preferences) (postgres.Preferences, error) {
	if m.updatePrefFunc != nil {
		return m.updatePrefFunc(ctx, key, spec, p)
	}
	return postgres.Preferences{}, errors.New("not implemented")
}

func (m *mockDAO) DeletePreferences(ctx context.Context, key, spec string) error {
	if m.deletePrefFunc != nil {
		return m.deletePrefFunc(ctx, key, spec)
	}
	return errors.New("not implemented")
}

func (m *mockDAO) CreateNotes(ctx context.Context, n postgres.Notes) (postgres.Notes, error) {
	if m.createNotesFunc != nil {
		return m.createNotesFunc(ctx, n)
	}
	return postgres.Notes{}, errors.New("not implemented")
}

func (m *mockDAO) GetNotes(ctx context.Context, id string) (postgres.Notes, error) {
	if m.getNotesFunc != nil {
		return m.getNotesFunc(ctx, id)
	}
	return postgres.Notes{}, errors.New("not implemented")
}

func (m *mockDAO) ListNotes(ctx context.Context, options postgres.ListOptions) ([]postgres.Notes, error) {
	if m.listNotesFunc != nil {
		return m.listNotesFunc(ctx, options)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDAO) UpdateNotes(ctx context.Context, id string, n postgres.Notes) (postgres.Notes, error) {
	if m.updateNotesFunc != nil {
		return m.updateNotesFunc(ctx, id, n)
	}
	return postgres.Notes{}, errors.New("not implemented")
}

func (m *mockDAO) DeleteNotes(ctx context.Context, id string) error {
	if m.deleteNotesFunc != nil {
		return m.deleteNotesFunc(ctx, id)
	}
	return errors.New("not implemented")
}

// Todo Handlers Tests
func TestTodoCreate(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		createTodoFunc: func(ctx context.Context, todo postgres.Todo) (postgres.Todo, error) {
			todo.UID = "generated-uid"
			todo.CreatedAt = now
			todo.UpdatedAt = now
			return todo, nil
		},
	}

	handlers := &Handlers{dao: mockDAO}

	todoJSON := `{"title":"Test Todo","description":"Test Description","priority":3}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(todoJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.create(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.UID != "generated-uid" {
		t.Errorf("Expected UID 'generated-uid', got '%s'", result.UID)
	}
	if result.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got '%s'", result.Title)
	}
}

func TestTodoCreateInvalidJSON(t *testing.T) {
	mockDAO := &mockDAO{}
	handlers := &Handlers{dao: mockDAO}

	req := httptest.NewRequest("POST", "/", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	handlers.create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTodoGet(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		getTodoFunc: func(ctx context.Context, uid string) (postgres.Todo, error) {
			if uid == "test-uid" {
				return postgres.Todo{
					UID:       "test-uid",
					Title:     "Test Todo",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			}
			return postgres.Todo{}, errors.New("not found")
		},
	}

	handlers := &Handlers{dao: mockDAO}

	req := httptest.NewRequest("GET", "/test-uid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handlers.get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.UID != "test-uid" {
		t.Errorf("Expected UID 'test-uid', got '%s'", result.UID)
	}
}

func TestTodoGetNotFound(t *testing.T) {
	mockDAO := &mockDAO{
		getTodoFunc: func(ctx context.Context, uid string) (postgres.Todo, error) {
			return postgres.Todo{}, errors.New("not found")
		},
	}

	handlers := &Handlers{dao: mockDAO}

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handlers.get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestTodoList(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		listTodosFunc: func(ctx context.Context, options postgres.ListOptions) ([]postgres.Todo, error) {
			return []postgres.Todo{
				{UID: "uid1", Title: "Todo 1", CreatedAt: now, UpdatedAt: now},
				{UID: "uid2", Title: "Todo 2", CreatedAt: now, UpdatedAt: now},
			}, nil
		},
	}

	handlers := &Handlers{dao: mockDAO}

	req := httptest.NewRequest("GET", "/?limit=10&sort_by=title&sort_dir=ASC&priority=high", nil)
	w := httptest.NewRecorder()

	handlers.list(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(result))
	}
}

func TestTodoUpdate(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		updateTodoFunc: func(ctx context.Context, uid string, todo postgres.Todo) (postgres.Todo, error) {
			if uid == "test-uid" {
				todo.UID = uid
				todo.UpdatedAt = now
				return todo, nil
			}
			return postgres.Todo{}, errors.New("not found")
		},
	}

	handlers := &Handlers{dao: mockDAO}

	todoJSON := `{"title":"Updated Todo","description":"Updated Description"}`
	req := httptest.NewRequest("PUT", "/test-uid", strings.NewReader(todoJSON))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handlers.update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Todo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.Title != "Updated Todo" {
		t.Errorf("Expected title 'Updated Todo', got '%s'", result.Title)
	}
}

func TestTodoDelete(t *testing.T) {
	mockDAO := &mockDAO{
		deleteTodoFunc: func(ctx context.Context, uid string) error {
			if uid == "test-uid" {
				return nil
			}
			return errors.New("not found")
		},
	}

	handlers := &Handlers{dao: mockDAO}

	req := httptest.NewRequest("DELETE", "/test-uid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", "test-uid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handlers.delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

// Background Handlers Tests
func TestBackgroundCreate(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		createBgFunc: func(ctx context.Context, bg postgres.Background) (postgres.Background, error) {
			bg.CreatedAt = now
			bg.UpdatedAt = now
			return bg, nil
		},
	}

	handlers := &BackgroundHandlers{dao: mockDAO}

	bgJSON := `{"key":"theme","value":"dark"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(bgJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.create(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Background
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.Key != "theme" {
		t.Errorf("Expected key 'theme', got '%s'", result.Key)
	}
	if result.Value != "dark" {
		t.Errorf("Expected value 'dark', got '%s'", result.Value)
	}
}

func TestBackgroundGet(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		getBgFunc: func(ctx context.Context, key string) (postgres.Background, error) {
			if key == "theme" {
				return postgres.Background{
					Key:       "theme",
					Value:     "dark",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			}
			return postgres.Background{}, errors.New("not found")
		},
	}

	handlers := &BackgroundHandlers{dao: mockDAO}

	req := httptest.NewRequest("GET", "/theme", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handlers.get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Background
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.Key != "theme" {
		t.Errorf("Expected key 'theme', got '%s'", result.Key)
	}
}

// Preferences Handlers Tests
func TestPreferencesCreate(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		createPrefFunc: func(ctx context.Context, pref postgres.Preferences) (postgres.Preferences, error) {
			pref.CreatedAt = now
			pref.UpdatedAt = now
			return pref, nil
		},
	}

	handlers := &PreferencesHandlers{dao: mockDAO}

	prefJSON := `{"key":"ui","specifier":"theme","data":"{\"mode\":\"dark\"}"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(prefJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.create(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Preferences
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.Key != "ui" {
		t.Errorf("Expected key 'ui', got '%s'", result.Key)
	}
	if result.Specifier != "theme" {
		t.Errorf("Expected specifier 'theme', got '%s'", result.Specifier)
	}
}

func TestPreferencesGet(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		getPrefFunc: func(ctx context.Context, key, spec string) (postgres.Preferences, error) {
			if key == "ui" && spec == "theme" {
				return postgres.Preferences{
					Key:       "ui",
					Specifier: "theme",
					Data:      "{\"mode\":\"dark\"}",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			}
			return postgres.Preferences{}, errors.New("not found")
		},
	}

	handlers := &PreferencesHandlers{dao: mockDAO}

	req := httptest.NewRequest("GET", "/ui/theme", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "ui")
	rctx.URLParams.Add("specifier", "theme")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handlers.get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Preferences
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.Key != "ui" {
		t.Errorf("Expected key 'ui', got '%s'", result.Key)
	}
	if result.Specifier != "theme" {
		t.Errorf("Expected specifier 'theme', got '%s'", result.Specifier)
	}
}

// Notes Handlers Tests
func TestNotesCreate(t *testing.T) {
	now := time.Now()
	mockDAO := &mockDAO{
		createNotesFunc: func(ctx context.Context, note postgres.Notes) (postgres.Notes, error) {
			note.ID = "generated-id"
			note.CreatedAt = now
			note.UpdatedAt = now
			return note, nil
		},
	}

	handlers := &NotesHandlers{dao: mockDAO}

	noteJSON := `{"title":"Test Note","relevant_user":"user123","content":"This is a test note"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(noteJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.create(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result postgres.Notes
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.ID != "generated-id" {
		t.Errorf("Expected ID 'generated-id', got '%s'", result.ID)
	}
	if result.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got '%s'", result.Title)
	}
	if result.RelevantUser != "user123" {
		t.Errorf("Expected relevant_user 'user123', got '%s'", result.RelevantUser)
	}
}
