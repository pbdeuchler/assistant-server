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

func TestPreferencesCreate(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	expectedPreference := postgres.Preferences{
		Key:       "theme",
		Specifier: "user-123",
		Data:      "{\"color\": \"dark\"}",
		Tags:      []string{"ui", "appearance"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockPreferencesDAO.On("CreatePreferences", 
		mock.Anything, 
		mock.MatchedBy(func(p postgres.Preferences) bool {
			return p.Key == "theme" && 
				   p.Specifier == "user-123" &&
				   p.Data == "{\"color\": \"dark\"}" &&
				   len(p.Tags) == 2
		})).Return(expectedPreference, nil)

	handler := NewPreferences(mockPreferencesDAO)

	reqBody := `{
		"key": "theme",
		"specifier": "user-123",
		"data": "{\"color\": \"dark\"}",
		"tags": ["ui", "appearance"]
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Preferences
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Key != expectedPreference.Key {
		t.Errorf("Expected key %s, got %s", expectedPreference.Key, response.Key)
	}
}

func TestPreferencesCreateInvalidJSON(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	handler := NewPreferences(mockPreferencesDAO)

	req := httptest.NewRequest("POST", "/", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestPreferencesCreateDAOError(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	mockPreferencesDAO.On("CreatePreferences", mock.Anything, mock.AnythingOfType("postgres.Preferences")).Return(postgres.Preferences{}, errors.New("database error"))

	handler := NewPreferences(mockPreferencesDAO)

	reqBody := `{
		"key": "theme",
		"specifier": "user-123",
		"data": "{\"color\": \"dark\"}"
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestPreferencesGet(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	expectedPreference := postgres.Preferences{
		Key:       "theme",
		Specifier: "user-123",
		Data:      "{\"color\": \"dark\"}",
		Tags:      []string{"ui"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockPreferencesDAO.On("GetPreferences", mock.Anything, "theme", "user-123").Return(expectedPreference, nil)

	handler := NewPreferences(mockPreferencesDAO)
	
	req := httptest.NewRequest("GET", "/theme/user-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Preferences
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Key != expectedPreference.Key {
		t.Errorf("Expected key %s, got %s", expectedPreference.Key, response.Key)
	}
	if response.Specifier != expectedPreference.Specifier {
		t.Errorf("Expected specifier %s, got %s", expectedPreference.Specifier, response.Specifier)
	}
}

func TestPreferencesGetNotFound(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	mockPreferencesDAO.On("GetPreferences", mock.Anything, "nonexistent", "user-123").Return(postgres.Preferences{}, errors.New("not found"))

	handler := NewPreferences(mockPreferencesDAO)
	
	req := httptest.NewRequest("GET", "/nonexistent/user-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "nonexistent")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestPreferencesUpdate(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	expectedPreference := postgres.Preferences{
		Key:       "theme",
		Specifier: "user-123",
		Data:      "{\"color\": \"light\"}",
		Tags:      []string{"ui", "updated"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockPreferencesDAO.On("UpdatePreferences", mock.Anything, "theme", "user-123", mock.AnythingOfType("postgres.Preferences")).Return(expectedPreference, nil)

	handler := NewPreferences(mockPreferencesDAO)

	reqBody := `{
		"key": "theme",
		"specifier": "user-123",
		"data": "{\"color\": \"light\"}",
		"tags": ["ui", "updated"]
	}`

	req := httptest.NewRequest("PUT", "/theme/user-123", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Preferences
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Data != expectedPreference.Data {
		t.Errorf("Expected data %s, got %s", expectedPreference.Data, response.Data)
	}
}

func TestPreferencesUpdateInvalidJSON(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	handler := NewPreferences(mockPreferencesDAO)

	req := httptest.NewRequest("PUT", "/theme/user-123", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestPreferencesUpdateDAOError(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	mockPreferencesDAO.On("UpdatePreferences", mock.Anything, "theme", "user-123", mock.AnythingOfType("postgres.Preferences")).Return(postgres.Preferences{}, errors.New("database error"))

	handler := NewPreferences(mockPreferencesDAO)

	reqBody := `{
		"data": "{\"color\": \"light\"}"
	}`

	req := httptest.NewRequest("PUT", "/theme/user-123", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestPreferencesDelete(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	mockPreferencesDAO.On("DeletePreferences", mock.Anything, "theme", "user-123").Return(nil)

	handler := NewPreferences(mockPreferencesDAO)
	
	req := httptest.NewRequest("DELETE", "/theme/user-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rr.Code)
	}
}

func TestPreferencesDeleteError(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	mockPreferencesDAO.On("DeletePreferences", mock.Anything, "theme", "user-123").Return(errors.New("database error"))

	handler := NewPreferences(mockPreferencesDAO)
	
	req := httptest.NewRequest("DELETE", "/theme/user-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "theme")
	rctx.URLParams.Add("specifier", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestPreferencesList(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	expectedPreferences := []postgres.Preferences{
		{
			Key:       "theme",
			Specifier: "user-123",
			Data:      "{\"color\": \"dark\"}",
			Tags:      []string{"ui"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Key:       "language",
			Specifier: "user-123",
			Data:      "{\"lang\": \"en\"}",
			Tags:      []string{"locale"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockPreferencesDAO.On("ListPreferences", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return(expectedPreferences, nil)

	handler := NewPreferences(mockPreferencesDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response []postgres.Preferences
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(expectedPreferences) {
		t.Errorf("Expected %d preferences, got %d", len(expectedPreferences), len(response))
	}
}

func TestPreferencesListError(t *testing.T) {
	mockPreferencesDAO := mocks.NewMockpreferencesDAO(t)
	
	mockPreferencesDAO.On("ListPreferences", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return([]postgres.Preferences{}, errors.New("database error"))

	handler := NewPreferences(mockPreferencesDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}