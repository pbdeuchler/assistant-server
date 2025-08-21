package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
	"github.com/pbdeuchler/assistant-server/integration_test/testutil"
	"github.com/pbdeuchler/assistant-server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMCPServer(t *testing.T, db *testutil.TestDatabase) *httptest.Server {
	t.Helper()
	
	mcpRouter := service.NewMCPRouter(db.DAO, db.DAO, db.DAO, db.DAO, db.DAO, db.DAO)
	return httptest.NewServer(mcpRouter)
}

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func sendMCPRequest(t *testing.T, server *httptest.Server, req JSONRPCRequest) JSONRPCResponse {
	t.Helper()
	
	body, err := json.Marshal(req)
	require.NoError(t, err)
	
	resp, err := http.Post(server.URL+"/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	
	var mcpResp JSONRPCResponse
	err = json.NewDecoder(resp.Body).Decode(&mcpResp)
	require.NoError(t, err)
	
	return mcpResp
}

func TestMCP_Initialize_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	t.Run("Initialize MCP Server", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"roots": map[string]any{
						"listChanged": true,
					},
				},
				"clientInfo": map[string]any{
					"name":    "IntegrationTestClient",
					"version": "1.0.0",
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Equal(t, float64(1), resp.ID)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		assert.Equal(t, "2024-11-05", result["protocolVersion"])
		
		serverInfo, ok := result["serverInfo"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "assistant-server", serverInfo["name"])
		assert.Equal(t, "Assistant Server MCP", serverInfo["title"])
		assert.Equal(t, "1.0.0", serverInfo["version"])
		
		capabilities, ok := result["capabilities"].(map[string]any)
		require.True(t, ok)
		assert.NotNil(t, capabilities["tools"])
		
		instructions, ok := result["instructions"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, instructions)
	})
	
	t.Run("Initialized Notification", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "initialized",
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Equal(t, float64(2), resp.ID)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		assert.Empty(t, result) // initialized should return empty result
	})
}

func TestMCP_ToolsList_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	t.Run("List Available Tools", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/list",
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		tools, ok := result["tools"].([]any)
		require.True(t, ok)
		assert.Len(t, tools, 13) // We have 13 tools defined
		
		// Verify specific tools exist
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolMap := tool.(map[string]any)
			name := toolMap["name"].(string)
			toolNames[name] = true
		}
		
		expectedTools := []string{
			"create_todo", "list_todos", "complete_todo",
			"save_note", "recall_note", "list_notes",
			"set_preference", "get_preference",
			"save_recipe", "find_recipes", "get_recipe",
			"update_user_description", "update_household_description",
		}
		
		for _, expectedTool := range expectedTools {
			assert.True(t, toolNames[expectedTool], "Tool %s should be available", expectedTool)
		}
	})
}

func TestMCP_TodoTools_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	var todoID string
	
	t.Run("Create Todo via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "create_todo",
				"arguments": map[string]any{
					"title":         "MCP Integration Todo",
					"description":   "A todo created via MCP interface",
					"priority":      float64(4),
					"due_date":      time.Now().Add(24 * time.Hour).Format(time.RFC3339),
					"user_uid":      user.UID,
					"household_uid": household.UID,
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "text", textContent["type"])
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "Todo created successfully")
		assert.Contains(t, text, "ID:")
		
		// Extract the todo ID from the response text
		parts := []string{}
		for _, part := range []string{text} {
			if len(part) > 0 {
				parts = append(parts, part)
			}
		}
		// The response should be like "Todo created successfully with ID: <id>"
		assert.Contains(t, text, "ID:")
		start := len("Todo created successfully with ID: ")
		todoID = text[start:]
		assert.NotEmpty(t, todoID)
	})
	
	t.Run("List Todos via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "list_todos",
				"arguments": map[string]any{
					"user_uid": user.UID,
					"limit":    float64(10),
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		var todos []dao.Todo
		err := json.Unmarshal([]byte(textContent["text"].(string)), &todos)
		require.NoError(t, err)
		
		assert.GreaterOrEqual(t, len(todos), 1)
		
		found := false
		for _, todo := range todos {
			if todo.UID == todoID {
				found = true
				assert.Equal(t, "MCP Integration Todo", todo.Title)
				assert.Equal(t, user.UID, todo.UserUID)
				break
			}
		}
		assert.True(t, found, "Created todo should be in the list")
	})
	
	t.Run("Complete Todo via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "complete_todo",
				"arguments": map[string]any{
					"todo_id":      todoID,
					"completed_by": user.UID,
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "marked as completed")
		assert.Contains(t, text, todoID)
	})
}

func TestMCP_NotesTools_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	var noteID string
	
	t.Run("Save Note via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "save_note",
				"arguments": map[string]any{
					"key":           "mcp-test-note",
					"data":          `{"type": "meeting", "attendees": ["Alice", "Bob"], "summary": "Discussed MCP integration"}`,
					"user_uid":      user.UID,
					"household_uid": household.UID,
					"tags":          "meeting,mcp,integration",
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "Note saved successfully")
		assert.Contains(t, text, "ID:")
		
		// Extract the note ID
		start := len("Note saved successfully with ID: ")
		noteID = text[start:]
		assert.NotEmpty(t, noteID)
	})
	
	t.Run("Recall Note via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "recall_note",
				"arguments": map[string]any{
					"note_id": noteID,
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		var note dao.Notes
		err := json.Unmarshal([]byte(textContent["text"].(string)), &note)
		require.NoError(t, err)
		
		assert.Equal(t, noteID, note.ID)
		assert.Equal(t, "mcp-test-note", note.Key)
		assert.Equal(t, user.UID, note.UserUID)
		assert.Contains(t, note.Data, "MCP integration")
		assert.Contains(t, note.Tags, "meeting")
		assert.Contains(t, note.Tags, "mcp")
	})
	
	t.Run("List Notes via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "list_notes",
				"arguments": map[string]any{
					"user_uid": user.UID,
					"tags":     "mcp",
					"limit":    float64(10),
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		var notes []dao.Notes
		err := json.Unmarshal([]byte(textContent["text"].(string)), &notes)
		require.NoError(t, err)
		
		assert.GreaterOrEqual(t, len(notes), 1)
		
		found := false
		for _, note := range notes {
			if note.ID == noteID {
				found = true
				assert.Equal(t, "mcp-test-note", note.Key)
				assert.Contains(t, note.Tags, "mcp")
				break
			}
		}
		assert.True(t, found, "Created note should be in the filtered list")
	})
}

func TestMCP_RecipeTools_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	var recipeID string
	
	t.Run("Save Recipe via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "save_recipe",
				"arguments": map[string]any{
					"title":        "MCP Pasta Recipe",
					"data":         `{"ingredients": ["pasta", "tomato sauce", "cheese"], "steps": ["Boil pasta", "Add sauce", "Top with cheese"]}`,
					"genre":        "italian",
					"grocery_list": `["spaghetti", "marinara sauce", "mozzarella cheese"]`,
					"prep_time":    float64(10),
					"cook_time":    float64(15),
					"servings":     float64(2),
					"difficulty":   float64(2),
					"rating":       float64(4),
					"tags":         "pasta,italian,quick,mcp",
					"user_uid":     user.UID,
					"household_uid": household.UID,
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "Recipe saved successfully")
		assert.Contains(t, text, "ID:")
		
		// Extract the recipe ID
		start := len("Recipe saved successfully with ID: ")
		recipeID = text[start:]
		assert.NotEmpty(t, recipeID)
	})
	
	t.Run("Get Recipe via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "get_recipe",
				"arguments": map[string]any{
					"recipe_id": recipeID,
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		var recipe dao.Recipes
		err := json.Unmarshal([]byte(textContent["text"].(string)), &recipe)
		require.NoError(t, err)
		
		assert.Equal(t, recipeID, recipe.ID)
		assert.Equal(t, "MCP Pasta Recipe", recipe.Title)
		assert.Equal(t, user.UID, recipe.UserUID)
		assert.Equal(t, "italian", *recipe.Genre)
		assert.Equal(t, 4, *recipe.Rating)
		assert.Contains(t, recipe.Tags, "pasta")
		assert.Contains(t, recipe.Tags, "mcp")
	})
	
	t.Run("Find Recipes via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "find_recipes",
				"arguments": map[string]any{
					"genre":      "italian",
					"tags":       "pasta",
					"min_rating": float64(3),
					"user_uid":   user.UID,
					"limit":      float64(10),
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		var recipes []dao.Recipes
		err := json.Unmarshal([]byte(textContent["text"].(string)), &recipes)
		require.NoError(t, err)
		
		assert.GreaterOrEqual(t, len(recipes), 1)
		
		found := false
		for _, recipe := range recipes {
			if recipe.ID == recipeID {
				found = true
				assert.Equal(t, "MCP Pasta Recipe", recipe.Title)
				assert.Equal(t, "italian", *recipe.Genre)
				assert.Contains(t, recipe.Tags, "pasta")
				break
			}
		}
		assert.True(t, found, "Created recipe should be in the search results")
	})
}

func TestMCP_PreferenceTools_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	t.Run("Set and Get Preference via MCP", func(t *testing.T) {
		// Set preference
		setReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "set_preference",
				"arguments": map[string]any{
					"key":       "mcp-settings",
					"specifier": "test-user-123",
					"data":      `{"notifications": true, "theme": "auto", "language": "en"}`,
					"tags":      "ui,settings,mcp",
				},
			},
		}
		
		resp := sendMCPRequest(t, server, setReq)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "Preference created: mcp-settings/test-user-123")
		
		// Get preference
		getReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "get_preference",
				"arguments": map[string]any{
					"key":       "mcp-settings",
					"specifier": "test-user-123",
				},
			},
		}
		
		resp = sendMCPRequest(t, server, getReq)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok = resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok = result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok = content[0].(map[string]any)
		require.True(t, ok)
		
		var pref dao.Preferences
		err := json.Unmarshal([]byte(textContent["text"].(string)), &pref)
		require.NoError(t, err)
		
		assert.Equal(t, "mcp-settings", pref.Key)
		assert.Equal(t, "test-user-123", pref.Specifier)
		assert.Contains(t, pref.Data, "notifications")
		assert.Contains(t, pref.Tags, "mcp")
	})
}

func TestMCP_UserAndHouseholdTools_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	user := testutil.CreateTestUser(t, db)
	household := testutil.CreateTestHousehold(t, db)
	
	t.Run("Update User Description via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "update_user_description",
				"arguments": map[string]any{
					"user_uid":    user.UID,
					"description": "Updated via MCP: User with integration testing experience",
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "User description updated successfully")
		
		// Parse the returned user JSON to verify the update
		start := len("User description updated successfully: ")
		userJSON := text[start:]
		
		var updatedUser dao.Users
		err := json.Unmarshal([]byte(userJSON), &updatedUser)
		require.NoError(t, err)
		
		assert.Equal(t, user.UID, updatedUser.UID)
		assert.Contains(t, updatedUser.Description, "Updated via MCP")
		assert.Contains(t, updatedUser.Description, "integration testing")
	})
	
	t.Run("Update Household Description via MCP", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "update_household_description",
				"arguments": map[string]any{
					"household_uid": household.UID,
					"description":   "Updated via MCP: Household with comprehensive testing setup",
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "Household description updated successfully")
		
		// Parse the returned household JSON to verify the update
		start := len("Household description updated successfully: ")
		householdJSON := text[start:]
		
		var updatedHousehold dao.Households
		err := json.Unmarshal([]byte(householdJSON), &updatedHousehold)
		require.NoError(t, err)
		
		assert.Equal(t, household.UID, updatedHousehold.UID)
		assert.Contains(t, updatedHousehold.Description, "Updated via MCP")
		assert.Contains(t, updatedHousehold.Description, "comprehensive testing")
	})
}

func TestMCP_ErrorHandling_Integration(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	server := setupMCPServer(t, db)
	defer server.Close()
	
	t.Run("Invalid Method", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "nonexistent/method",
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.NotNil(t, resp.Error)
		
		errorMap, ok := resp.Error.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(-32601), errorMap["code"])
		assert.Contains(t, errorMap["message"].(string), "Method not found")
	})
	
	t.Run("Invalid Tool Name", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name":      "nonexistent_tool",
				"arguments": map[string]any{},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error) // The method call succeeds, but the tool result contains an error
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		// Check if this is an error result
		isError, exists := result["isError"]
		if exists && isError.(bool) {
			content, ok := result["content"].([]any)
			require.True(t, ok)
			assert.Len(t, content, 1)
			
			textContent, ok := content[0].(map[string]any)
			require.True(t, ok)
			
			text := textContent["text"].(string)
			assert.Contains(t, text, "Unknown tool")
		}
	})
	
	t.Run("Missing Required Tool Arguments", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "create_todo",
				"arguments": map[string]any{
					"description": "Missing title",
				},
			},
		}
		
		resp := sendMCPRequest(t, server, req)
		
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.Error)
		
		result, ok := resp.Result.(map[string]any)
		require.True(t, ok)
		
		isError, exists := result["isError"]
		require.True(t, exists)
		assert.True(t, isError.(bool))
		
		content, ok := result["content"].([]any)
		require.True(t, ok)
		assert.Len(t, content, 1)
		
		textContent, ok := content[0].(map[string]any)
		require.True(t, ok)
		
		text := textContent["text"].(string)
		assert.Contains(t, text, "title is required")
	})
}