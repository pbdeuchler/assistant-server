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

func TestNotesCreate(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	expectedNote := postgres.Notes{
		ID:          "generated-id",
		Key:         "Test Note",
		UserUID:      "user-123",
		HouseholdUID: "household-456",
		Data:        "This is the content",
		Tags:        []string{"tag1", "tag2"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockNotesDAO.On("CreateNotes", 
		mock.Anything, 
		mock.MatchedBy(func(n postgres.Notes) bool {
			return n.Key == "Test Note" && 
				   n.UserUID == "user-123" &&
				   n.HouseholdUID == "household-456" &&
				   n.Data == "This is the content" &&
				   len(n.Tags) == 2
		})).Return(expectedNote, nil)

	handler := NewNotes(mockNotesDAO)

	reqBody := `{
		"key": "Test Note",
		"user_uid": "user-123",
		"household_uid": "household-456",
		"data": "This is the content",
		"tags": ["tag1", "tag2"]
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Notes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Key != expectedNote.Key {
		t.Errorf("Expected key %s, got %s", expectedNote.Key, response.Key)
	}
}

func TestNotesCreateInvalidJSON(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	handler := NewNotes(mockNotesDAO)

	req := httptest.NewRequest("POST", "/", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestNotesCreateDAOError(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	mockNotesDAO.On("CreateNotes", mock.Anything, mock.AnythingOfType("postgres.Notes")).Return(postgres.Notes{}, errors.New("database error"))

	handler := NewNotes(mockNotesDAO)

	reqBody := `{
		"key": "Test Note",
		"user_uid": "user-123",
		"household_uid": "household-456",
		"data": "This is the content"
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestNotesGet(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	expectedNote := postgres.Notes{
		ID:          "test-id",
		Key:         "Test Note",
		UserUID:      "user-123",
		HouseholdUID: "household-456",
		Data:        "This is the content",
		Tags:        []string{"tag1"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockNotesDAO.On("GetNotes", mock.Anything, "test-id").Return(expectedNote, nil)

	handler := NewNotes(mockNotesDAO)
	
	req := httptest.NewRequest("GET", "/test-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Notes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.ID != expectedNote.ID {
		t.Errorf("Expected ID %s, got %s", expectedNote.ID, response.ID)
	}
}

func TestNotesGetNotFound(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	mockNotesDAO.On("GetNotes", mock.Anything, "nonexistent").Return(postgres.Notes{}, errors.New("not found"))

	handler := NewNotes(mockNotesDAO)
	
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestNotesUpdate(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	expectedNote := postgres.Notes{
		ID:          "test-id",
		Key:         "Updated Note",
		UserUID:      "user-123",
		HouseholdUID: "household-456",
		Data:        "Updated content",
		Tags:        []string{"updated"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockNotesDAO.On("UpdateNotes", mock.Anything, "test-id", mock.AnythingOfType("postgres.Notes")).Return(expectedNote, nil)

	handler := NewNotes(mockNotesDAO)

	reqBody := `{
		"key": "Updated Note",
		"user_uid": "user-123",
		"household_uid": "household-456",
		"data": "Updated content",
		"tags": ["updated"]
	}`

	req := httptest.NewRequest("PUT", "/test-id", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Notes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Key != expectedNote.Key {
		t.Errorf("Expected key %s, got %s", expectedNote.Key, response.Key)
	}
}

func TestNotesUpdateInvalidJSON(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	handler := NewNotes(mockNotesDAO)

	req := httptest.NewRequest("PUT", "/test-id", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestNotesUpdateDAOError(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	mockNotesDAO.On("UpdateNotes", mock.Anything, "test-id", mock.AnythingOfType("postgres.Notes")).Return(postgres.Notes{}, errors.New("database error"))

	handler := NewNotes(mockNotesDAO)

	reqBody := `{
		"key": "Updated Note",
		"data": "Updated content"
	}`

	req := httptest.NewRequest("PUT", "/test-id", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestNotesDelete(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	mockNotesDAO.On("DeleteNotes", mock.Anything, "test-id").Return(nil)

	handler := NewNotes(mockNotesDAO)
	
	req := httptest.NewRequest("DELETE", "/test-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rr.Code)
	}
}

func TestNotesDeleteError(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	mockNotesDAO.On("DeleteNotes", mock.Anything, "test-id").Return(errors.New("database error"))

	handler := NewNotes(mockNotesDAO)
	
	req := httptest.NewRequest("DELETE", "/test-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestNotesList(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	expectedNotes := []postgres.Notes{
		{
			ID:          "test-id-1",
			Key:         "Test Note 1",
			UserUID:      "user-123",
			HouseholdUID: "household-456",
			Data:        "Content 1",
			Tags:        []string{"tag1"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "test-id-2",
			Key:         "Test Note 2",
			UserUID:      "user-123",
			HouseholdUID: "household-456",
			Data:        "Content 2",
			Tags:        []string{"tag2"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockNotesDAO.On("ListNotes", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return(expectedNotes, nil)

	handler := NewNotes(mockNotesDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response []postgres.Notes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(expectedNotes) {
		t.Errorf("Expected %d notes, got %d", len(expectedNotes), len(response))
	}
}

func TestNotesListError(t *testing.T) {
	mockNotesDAO := mocks.NewMocknotesDAO(t)
	
	mockNotesDAO.On("ListNotes", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return([]postgres.Notes{}, errors.New("database error"))

	handler := NewNotes(mockNotesDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}