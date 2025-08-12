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

func TestRecipesCreate(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	prepTime := 15
	cookTime := 30
	totalTime := 45
	servings := 4
	rating := 5
	difficulty := "medium"
	externalURL := "https://example.com/recipe"
	genre := "Italian"
	groceryList := "tomatoes, pasta, cheese"

	expectedRecipe := postgres.Recipes{
		ID:          "generated-id",
		Title:       "Test Recipe",
		ExternalURL: &externalURL,
		Data:        "Recipe instructions here",
		Genre:       &genre,
		GroceryList: &groceryList,
		PrepTime:    &prepTime,
		CookTime:    &cookTime,
		TotalTime:   &totalTime,
		Servings:    &servings,
		Difficulty:  &difficulty,
		Rating:      &rating,
		Tags:        []string{"pasta", "dinner"},
		UserID:      "user-123",
		HouseholdID: "household-456",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRecipesDAO.On("CreateRecipes", 
		mock.Anything, 
		mock.MatchedBy(func(r postgres.Recipes) bool {
			return r.Title == "Test Recipe" && 
				   r.UserID == "user-123" &&
				   r.HouseholdID == "household-456" &&
				   r.Data == "Recipe instructions here" &&
				   len(r.Tags) == 2
		})).Return(expectedRecipe, nil)

	handler := NewRecipes(mockRecipesDAO)

	reqBody := `{
		"title": "Test Recipe",
		"external_url": "https://example.com/recipe",
		"data": "Recipe instructions here",
		"genre": "Italian",
		"grocery_list": "tomatoes, pasta, cheese",
		"prep_time": 15,
		"cook_time": 30,
		"total_time": 45,
		"servings": 4,
		"difficulty": "medium",
		"rating": 5,
		"tags": ["pasta", "dinner"],
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

	var response postgres.Recipes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Title != expectedRecipe.Title {
		t.Errorf("Expected title %s, got %s", expectedRecipe.Title, response.Title)
	}
}

func TestRecipesCreateInvalidJSON(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	handler := NewRecipes(mockRecipesDAO)

	req := httptest.NewRequest("POST", "/", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestRecipesCreateDAOError(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	mockRecipesDAO.On("CreateRecipes", mock.Anything, mock.AnythingOfType("postgres.Recipes")).Return(postgres.Recipes{}, errors.New("database error"))

	handler := NewRecipes(mockRecipesDAO)

	reqBody := `{
		"title": "Test Recipe",
		"data": "Recipe instructions",
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

func TestRecipesGet(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	rating := 4
	servings := 6
	expectedRecipe := postgres.Recipes{
		ID:          "test-id",
		Title:       "Test Recipe",
		Data:        "Recipe instructions",
		Rating:      &rating,
		Servings:    &servings,
		Tags:        []string{"dessert"},
		UserID:      "user-123",
		HouseholdID: "household-456",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRecipesDAO.On("GetRecipes", mock.Anything, "test-id").Return(expectedRecipe, nil)

	handler := NewRecipes(mockRecipesDAO)
	
	req := httptest.NewRequest("GET", "/test-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response postgres.Recipes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.ID != expectedRecipe.ID {
		t.Errorf("Expected ID %s, got %s", expectedRecipe.ID, response.ID)
	}
}

func TestRecipesGetNotFound(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	mockRecipesDAO.On("GetRecipes", mock.Anything, "nonexistent").Return(postgres.Recipes{}, errors.New("not found"))

	handler := NewRecipes(mockRecipesDAO)
	
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

func TestRecipesUpdate(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	rating := 5
	expectedRecipe := postgres.Recipes{
		ID:          "test-id",
		Title:       "Updated Recipe",
		Data:        "Updated instructions",
		Rating:      &rating,
		Tags:        []string{"updated"},
		UserID:      "user-123",
		HouseholdID: "household-456",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRecipesDAO.On("UpdateRecipes", mock.Anything, "test-id", mock.AnythingOfType("postgres.Recipes")).Return(expectedRecipe, nil)

	handler := NewRecipes(mockRecipesDAO)

	reqBody := `{
		"title": "Updated Recipe",
		"data": "Updated instructions",
		"rating": 5,
		"tags": ["updated"],
		"user_id": "user-123",
		"household_id": "household-456"
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

	var response postgres.Recipes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Title != expectedRecipe.Title {
		t.Errorf("Expected title %s, got %s", expectedRecipe.Title, response.Title)
	}
}

func TestRecipesUpdateInvalidJSON(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	handler := NewRecipes(mockRecipesDAO)

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

func TestRecipesUpdateDAOError(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	mockRecipesDAO.On("UpdateRecipes", mock.Anything, "test-id", mock.AnythingOfType("postgres.Recipes")).Return(postgres.Recipes{}, errors.New("database error"))

	handler := NewRecipes(mockRecipesDAO)

	reqBody := `{
		"title": "Updated Recipe",
		"data": "Updated instructions"
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

func TestRecipesDelete(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	mockRecipesDAO.On("DeleteRecipes", mock.Anything, "test-id").Return(nil)

	handler := NewRecipes(mockRecipesDAO)
	
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

func TestRecipesDeleteError(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	mockRecipesDAO.On("DeleteRecipes", mock.Anything, "test-id").Return(errors.New("database error"))

	handler := NewRecipes(mockRecipesDAO)
	
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

func TestRecipesList(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	rating1 := 4
	rating2 := 5
	expectedRecipes := []postgres.Recipes{
		{
			ID:          "test-id-1",
			Title:       "Recipe 1",
			Data:        "Instructions 1",
			Rating:      &rating1,
			Tags:        []string{"breakfast"},
			UserID:      "user-123",
			HouseholdID: "household-456",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "test-id-2",
			Title:       "Recipe 2",
			Data:        "Instructions 2",
			Rating:      &rating2,
			Tags:        []string{"dinner"},
			UserID:      "user-123",
			HouseholdID: "household-456",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockRecipesDAO.On("ListRecipes", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return(expectedRecipes, nil)

	handler := NewRecipes(mockRecipesDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response []postgres.Recipes
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(expectedRecipes) {
		t.Errorf("Expected %d recipes, got %d", len(expectedRecipes), len(response))
	}
}

func TestRecipesListError(t *testing.T) {
	mockRecipesDAO := mocks.NewMockrecipesDAO(t)
	
	mockRecipesDAO.On("ListRecipes", mock.Anything, mock.AnythingOfType("postgres.ListOptions")).Return([]postgres.Recipes{}, errors.New("database error"))

	handler := NewRecipes(mockRecipesDAO)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}