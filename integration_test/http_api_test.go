package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
	"github.com/pbdeuchler/assistant-server/integration_test/testutil"
	"github.com/pbdeuchler/assistant-server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T, db *testutil.TestDatabase) *httptest.Server {
	t.Helper()
	
	r := chi.NewRouter()
	
	// Mount all the API routes
	r.Mount("/todos", service.NewTodos(db.DAO))
	r.Mount("/preferences", service.NewPreferences(db.DAO))
	r.Mount("/notes", service.NewNotes(db.DAO))
	r.Mount("/recipes", service.NewRecipes(db.DAO))
	r.Mount("/bootstrap", service.NewBootstrap(db.DAO))
	
	return httptest.NewServer(r)
}

func TestTodosAPI_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupTestServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	t.Run("Create Todo", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		createReq := map[string]any{
			"title":         "Integration Test Todo",
			"description":   "A todo created via HTTP API",
			"data":          `{"priority": "high"}`,
			"priority":      4,
			"due_date":      dueDate,
			"user_uid":      user.UID,
			"household_uid": household.UID,
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/todos/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var todo dao.Todo
		err = json.NewDecoder(resp.Body).Decode(&todo)
		require.NoError(t, err)
		
		assert.Equal(t, "Integration Test Todo", todo.Title)
		assert.Equal(t, "A todo created via HTTP API", todo.Description)
		assert.Equal(t, dao.Priority(4), todo.Priority)
		assert.Equal(t, user.UID, todo.UserUID)
		assert.Equal(t, household.UID, todo.HouseholdUID)
		assert.NotEmpty(t, todo.UID)
		
		// Test Get Todo
		t.Run("Get Todo", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/todos/" + todo.UID)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var fetchedTodo dao.Todo
			err = json.NewDecoder(resp.Body).Decode(&fetchedTodo)
			require.NoError(t, err)
			
			testutil.AssertTodoEqual(t, todo, fetchedTodo)
		})
		
		// Test Update Todo
		t.Run("Update Todo", func(t *testing.T) {
			updateReq := map[string]any{
				"title":           "Updated Todo Title",
				"marked_complete": time.Now().Format(time.RFC3339),
				"completed_by":    user.UID,
			}
			
			body, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest(http.MethodPut, server.URL+"/todos/"+todo.UID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var updatedTodo dao.Todo
			err = json.NewDecoder(resp.Body).Decode(&updatedTodo)
			require.NoError(t, err)
			
			assert.Equal(t, "Updated Todo Title", updatedTodo.Title)
			assert.NotNil(t, updatedTodo.MarkedComplete)
			assert.Equal(t, user.UID, updatedTodo.CompletedBy)
		})
		
		// Test List Todos
		t.Run("List Todos", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/todos/?user_uid=" + user.UID + "&limit=10")
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var todos []dao.Todo
			err = json.NewDecoder(resp.Body).Decode(&todos)
			require.NoError(t, err)
			
			assert.GreaterOrEqual(t, len(todos), 1)
			
			found := false
			for _, fetchedTodo := range todos {
				if fetchedTodo.UID == todo.UID {
					found = true
					assert.Equal(t, "Updated Todo Title", fetchedTodo.Title)
					break
				}
			}
			assert.True(t, found, "Created todo should be in the list")
		})
		
		// Test Delete Todo
		t.Run("Delete Todo", func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, server.URL+"/todos/"+todo.UID, nil)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			
			// Verify deletion
			resp, err = http.Get(server.URL + "/todos/" + todo.UID)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

func TestNotesAPI_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupTestServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	t.Run("Create and Manage Notes", func(t *testing.T) {
		createReq := map[string]any{
			"key":           "meeting-notes",
			"user_uid":      user.UID,
			"household_uid": household.UID,
			"data":          `{"meeting_date": "2024-01-15", "attendees": ["Alice", "Bob"], "summary": "Discussed project roadmap"}`,
			"tags":          []string{"meeting", "project", "important"},
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/notes/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var note dao.Notes
		err = json.NewDecoder(resp.Body).Decode(&note)
		require.NoError(t, err)
		
		assert.Equal(t, "meeting-notes", note.Key)
		assert.Equal(t, user.UID, note.UserUID)
		assert.Equal(t, household.UID, note.HouseholdUID)
		assert.Contains(t, note.Data, "project roadmap")
		assert.ElementsMatch(t, []string{"meeting", "project", "important"}, note.Tags)
		assert.NotEmpty(t, note.ID)
		
		// Test Get Note
		t.Run("Get Note", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/notes/" + note.ID)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var fetchedNote dao.Notes
			err = json.NewDecoder(resp.Body).Decode(&fetchedNote)
			require.NoError(t, err)
			
			testutil.AssertNoteEqual(t, note, fetchedNote)
		})
		
		// Test List Notes with filters
		t.Run("List Notes with Filters", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/notes/?key=meeting-notes&tags=meeting&limit=10")
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var notes []dao.Notes
			err = json.NewDecoder(resp.Body).Decode(&notes)
			require.NoError(t, err)
			
			assert.GreaterOrEqual(t, len(notes), 1)
			
			found := false
			for _, fetchedNote := range notes {
				if fetchedNote.ID == note.ID {
					found = true
					testutil.AssertNoteEqual(t, note, fetchedNote)
					break
				}
			}
			assert.True(t, found, "Created note should be in the filtered list")
		})
	})
}

func TestRecipesAPI_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupTestServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	t.Run("Create and Search Recipes", func(t *testing.T) {
		createReq := map[string]any{
			"title":         "Spaghetti Carbonara",
			"data":          `{"ingredients": ["spaghetti", "eggs", "bacon", "cheese"], "instructions": ["Boil pasta", "Cook bacon", "Mix eggs", "Combine all"]}`,
			"genre":         "italian",
			"grocery_list":  `["spaghetti", "eggs", "bacon", "parmesan cheese"]`,
			"prep_time":     15,
			"cook_time":     20,
			"servings":      4,
			"difficulty":    "medium",
			"rating":        5,
			"tags":          []string{"pasta", "italian", "dinner", "comfort-food"},
			"user_uid":      user.UID,
			"household_uid": household.UID,
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/recipes/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var recipe dao.Recipes
		err = json.NewDecoder(resp.Body).Decode(&recipe)
		require.NoError(t, err)
		
		assert.Equal(t, "Spaghetti Carbonara", recipe.Title)
		assert.Equal(t, user.UID, recipe.UserUID)
		assert.Equal(t, household.UID, recipe.HouseholdUID)
		assert.Equal(t, "italian", *recipe.Genre)
		assert.Equal(t, 5, *recipe.Rating)
		assert.ElementsMatch(t, []string{"pasta", "italian", "dinner", "comfort-food"}, recipe.Tags)
		assert.NotEmpty(t, recipe.ID)
		
		// Test Get Recipe
		t.Run("Get Recipe", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/recipes/" + recipe.ID)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var fetchedRecipe dao.Recipes
			err = json.NewDecoder(resp.Body).Decode(&fetchedRecipe)
			require.NoError(t, err)
			
			testutil.AssertRecipeEqual(t, recipe, fetchedRecipe)
		})
		
		// Test Search Recipes
		t.Run("Search Recipes", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/recipes/?title=carbonara&genre=italian&min_rating=4&limit=10")
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var recipes []dao.Recipes
			err = json.NewDecoder(resp.Body).Decode(&recipes)
			require.NoError(t, err)
			
			assert.GreaterOrEqual(t, len(recipes), 1)
			
			found := false
			for _, fetchedRecipe := range recipes {
				if fetchedRecipe.ID == recipe.ID {
					found = true
					testutil.AssertRecipeEqual(t, recipe, fetchedRecipe)
					break
				}
			}
			assert.True(t, found, "Created recipe should be in the search results")
		})
		
		// Test Search by Tags
		t.Run("Search Recipes by Tags", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/recipes/?tags=pasta,italian&limit=10")
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var recipes []dao.Recipes
			err = json.NewDecoder(resp.Body).Decode(&recipes)
			require.NoError(t, err)
			
			assert.GreaterOrEqual(t, len(recipes), 1)
			
			found := false
			for _, fetchedRecipe := range recipes {
				if fetchedRecipe.ID == recipe.ID {
					found = true
					// Verify the recipe has the searched tags
					assert.Contains(t, fetchedRecipe.Tags, "pasta")
					assert.Contains(t, fetchedRecipe.Tags, "italian")
					break
				}
			}
			assert.True(t, found, "Recipe with matching tags should be found")
		})
	})
}

func TestPreferencesAPI_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupTestServer(t, db)
	defer server.Close()
	
	t.Run("Create and Manage Preferences", func(t *testing.T) {
		createReq := map[string]any{
			"key":       "ui-theme",
			"specifier": "user-123",
			"data":      `{"theme": "dark", "font_size": 14, "sidebar_collapsed": false}`,
			"tags":      []string{"ui", "personalization"},
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/preferences/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var pref dao.Preferences
		err = json.NewDecoder(resp.Body).Decode(&pref)
		require.NoError(t, err)
		
		assert.Equal(t, "ui-theme", pref.Key)
		assert.Equal(t, "user-123", pref.Specifier)
		assert.Contains(t, pref.Data, "dark")
		assert.ElementsMatch(t, []string{"ui", "personalization"}, pref.Tags)
		
		// Test Get Preference
		t.Run("Get Preference", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/preferences/ui-theme/user-123")
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var fetchedPref dao.Preferences
			err = json.NewDecoder(resp.Body).Decode(&fetchedPref)
			require.NoError(t, err)
			
			assert.Equal(t, pref.Key, fetchedPref.Key)
			assert.Equal(t, pref.Specifier, fetchedPref.Specifier)
			assert.Equal(t, pref.Data, fetchedPref.Data)
			assert.ElementsMatch(t, pref.Tags, fetchedPref.Tags)
		})
		
		// Test Update Preference
		t.Run("Update Preference", func(t *testing.T) {
			updateReq := map[string]any{
				"data": `{"theme": "light", "font_size": 16, "sidebar_collapsed": true}`,
				"tags": []string{"ui", "personalization", "updated"},
			}
			
			body, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest(http.MethodPut, server.URL+"/preferences/ui-theme/user-123", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			
			var updatedPref dao.Preferences
			err = json.NewDecoder(resp.Body).Decode(&updatedPref)
			require.NoError(t, err)
			
			assert.Contains(t, updatedPref.Data, "light")
			assert.Contains(t, updatedPref.Data, "16")
			assert.ElementsMatch(t, []string{"ui", "personalization", "updated"}, updatedPref.Tags)
		})
	})
}

func TestBootstrapAPI_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupTestServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	// Create some test data for the user
	testutil.CreateTestTodo(t, db, user.UID, household.UID)
	testutil.CreateTestNote(t, db, user.UID, household.UID)
	testutil.CreateTestRecipe(t, db, user.UID, household.UID)
	
	t.Run("Bootstrap User Data", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/bootstrap/" + user.UID)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var bootstrap map[string]any
		err = json.NewDecoder(resp.Body).Decode(&bootstrap)
		require.NoError(t, err)
		
		// Verify user data is included
		require.Contains(t, bootstrap, "user")
		userData := bootstrap["user"].(map[string]any)
		assert.Equal(t, user.UID, userData["uid"])
		assert.Equal(t, user.Name, userData["name"])
		
		// Verify todos are included
		require.Contains(t, bootstrap, "todos")
		todos := bootstrap["todos"].([]any)
		assert.GreaterOrEqual(t, len(todos), 1)
		
		// Verify notes are included
		require.Contains(t, bootstrap, "notes")
		notes := bootstrap["notes"].([]any)
		assert.GreaterOrEqual(t, len(notes), 1)
		
		// Verify recipes are included
		require.Contains(t, bootstrap, "recipes")
		recipes := bootstrap["recipes"].([]any)
		assert.GreaterOrEqual(t, len(recipes), 1)
		
		// Verify preferences are included (may be empty)
		require.Contains(t, bootstrap, "preferences")
	})
}

func TestAPIErrors_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupTestServer(t, db)
	defer server.Close()
	
	t.Run("Invalid Todo Creation", func(t *testing.T) {
		// Missing required fields
		createReq := map[string]any{
			"description": "Missing title",
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/todos/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
	
	t.Run("Invalid Priority", func(t *testing.T) {
		createReq := map[string]any{
			"title":    "Test Todo",
			"priority": 10, // Invalid priority (should be 1-5)
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/todos/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Contains(t, strings.ToLower(errorResp["error"]), "priority")
	})
	
	t.Run("Invalid JSON Data", func(t *testing.T) {
		createReq := map[string]any{
			"title": "Test Todo",
			"data":  "invalid json{", // Invalid JSON
		}
		
		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/todos/", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Contains(t, strings.ToLower(errorResp["error"]), "json")
	})
	
	t.Run("Not Found", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/todos/non-existent-id")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}