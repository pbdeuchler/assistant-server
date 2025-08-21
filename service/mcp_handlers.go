package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
)

type userDAO interface {
	UpdateUser(ctx context.Context, uid string, u dao.UpdateUser) (dao.Users, error)
	GetUser(ctx context.Context, uid string) (dao.Users, error)
}

type householdDAO interface {
	UpdateHousehold(ctx context.Context, uid string, h dao.UpdateHousehold) (dao.Households, error)
	GetHousehold(ctx context.Context, uid string) (dao.Households, error)
}

type MCPHandlers struct {
	todoDAO        todoDAO
	notesDAO       notesDAO
	preferencesDAO preferencesDAO
	recipesDAO     recipesDAO
	userDAO        userDAO
	householdDAO   householdDAO
	tools          []mcp.Tool
	clientInfo     *ClientInfo
	serverInfo     ServerInfo
	capabilities   ServerCapabilities
	logger         *slog.Logger
}

func (h *MCPHandlers) log() *slog.Logger {
	if h.logger != nil {
		return h.logger
	}
	return slog.Default()
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

type ClientInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title,omitempty"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title,omitempty"`
	Version string `json:"version"`
}

type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type ClientCapabilities struct {
	Roots       *RootsCapability `json:"roots,omitempty"`
	Sampling    map[string]any   `json:"sampling,omitempty"`
	Elicitation map[string]any   `json:"elicitation,omitempty"`
}

type ServerCapabilities struct {
	Logging   map[string]any       `json:"logging,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

func NewMCP(todoDAO todoDAO, notesDAO notesDAO, preferencesDAO preferencesDAO, recipesDAO recipesDAO, userDAO userDAO, householdDAO householdDAO) *MCPHandlers {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With(
		slog.String("component", "mcp"),
		slog.String("app", "assistant-server"),
	)

	h := &MCPHandlers{
		todoDAO:        todoDAO,
		notesDAO:       notesDAO,
		preferencesDAO: preferencesDAO,
		recipesDAO:     recipesDAO,
		userDAO:        userDAO,
		householdDAO:   householdDAO,
		logger:         logger,
		serverInfo: ServerInfo{
			Name:    "assistant-server",
			Title:   "Assistant Server MCP",
			Version: "1.0.0",
		},
		capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: true,
			},
		},
	}

	h.setupTools()
	logger.Info("MCP server initialized",
		slog.Int("tools_count", len(h.tools)),
		slog.String("server_name", h.serverInfo.Name),
		slog.String("server_version", h.serverInfo.Version),
	)
	return h
}

func (h *MCPHandlers) setupTools() {
	h.tools = []mcp.Tool{
		mcp.NewTool("create_todo",
			mcp.WithDescription("Create a new todo task"),
			mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
			mcp.WithString("description", mcp.Description("Task description")),
			mcp.WithNumber("priority", mcp.Description("Priority level 1-5 (5 is highest)")),
			mcp.WithString("due_date", mcp.Description("Due date in RFC3339 format (e.g., 2024-01-15T10:00:00Z)")),
			mcp.WithString("user_uid", mcp.Description("User ID")),
			mcp.WithString("household_uid", mcp.Description("Household ID")),
		),
		mcp.NewTool("list_todos",
			mcp.WithDescription("List todos with optional filtering"),
			mcp.WithString("user_uid", mcp.Description("Filter by user ID")),
			mcp.WithString("household_uid", mcp.Description("Filter by household ID")),
			mcp.WithNumber("priority", mcp.Description("Filter by priority level")),
			mcp.WithString("tags", mcp.Description("Filter by tags (comma-separated)")),
			mcp.WithBoolean("completed_only", mcp.Description("Show only completed todos")),
			mcp.WithBoolean("pending_only", mcp.Description("Show only pending todos")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 20)")),
		),
		mcp.NewTool("complete_todo",
			mcp.WithDescription("Mark a todo as completed"),
			mcp.WithString("todo_id", mcp.Required(), mcp.Description("Todo UID to complete")),
			mcp.WithString("completed_by", mcp.Description("User ID who completed the task")),
		),
		mcp.NewTool("save_note",
			mcp.WithDescription("Save a note with a key for later retrieval"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Unique key for the note")),
			mcp.WithString("data", mcp.Required(), mcp.Description("Structured note content")),
			mcp.WithString("user_uid", mcp.Description("User ID")),
			mcp.WithString("household_uid", mcp.Description("Household ID")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		),
		mcp.NewTool("recall_note",
			mcp.WithDescription("Retrieve a saved note by key"),
			mcp.WithString("note_id", mcp.Required(), mcp.Description("Note ID to retrieve")),
		),
		mcp.NewTool("list_notes",
			mcp.WithDescription("List notes with optional filtering"),
			mcp.WithString("key", mcp.Description("Filter by key")),
			mcp.WithString("user_uid", mcp.Description("Filter by user ID")),
			mcp.WithString("household_uid", mcp.Description("Filter by household ID")),
			mcp.WithString("tags", mcp.Description("Filter by tags (comma-separated)")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 20)")),
		),
		mcp.NewTool("set_preference",
			mcp.WithDescription("Set a user preference"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Preference key")),
			mcp.WithString("specifier", mcp.Required(), mcp.Description("Preference specifier (user-specific identifier)")),
			mcp.WithString("data", mcp.Required(), mcp.Description("Structured preference data")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		),
		mcp.NewTool("get_preference",
			mcp.WithDescription("Get a user preference"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Preference key")),
			mcp.WithString("specifier", mcp.Required(), mcp.Description("Preference specifier")),
		),
		mcp.NewTool("save_recipe",
			mcp.WithDescription("Save a recipe"),
			mcp.WithString("title", mcp.Required(), mcp.Description("Recipe title")),
			mcp.WithString("data", mcp.Required(), mcp.Description("Recipe instructions as structured data")),
			mcp.WithString("genre", mcp.Description("Recipe genre/category")),
			mcp.WithString("grocery_list", mcp.Description("Grocery list as structured data")),
			mcp.WithNumber("prep_time", mcp.Description("Prep time in minutes")),
			mcp.WithNumber("cook_time", mcp.Description("Cook time in minutes")),
			mcp.WithNumber("servings", mcp.Description("Number of servings")),
			mcp.WithNumber("difficulty", mcp.Description("Difficulty level 1-5")),
			mcp.WithNumber("rating", mcp.Description("Rating 1-5")),
			mcp.WithString("user_uid", mcp.Description("User ID")),
			mcp.WithString("household_uid", mcp.Description("Household ID")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		),
		mcp.NewTool("find_recipes",
			mcp.WithDescription("Search recipes by criteria"),
			mcp.WithString("title", mcp.Description("Filter by title (partial match)")),
			mcp.WithString("genre", mcp.Description("Filter by genre")),
			mcp.WithNumber("max_cook_time", mcp.Description("Maximum cook time in minutes")),
			mcp.WithNumber("min_rating", mcp.Description("Minimum rating")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags to filter by")),
			mcp.WithString("user_uid", mcp.Description("Filter by user ID")),
			mcp.WithString("household_uid", mcp.Description("Filter by household ID")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 20)")),
		),
		mcp.NewTool("get_recipe",
			mcp.WithDescription("Get a specific recipe by ID"),
			mcp.WithString("recipe_id", mcp.Required(), mcp.Description("Recipe ID")),
		),
		mcp.NewTool("update_user_description",
			mcp.WithDescription("Update a user's description"),
			mcp.WithString("user_uid", mcp.Required(), mcp.Description("User ID")),
			mcp.WithString("description", mcp.Required(), mcp.Description("New description for the user")),
		),
		mcp.NewTool("update_household_description",
			mcp.WithDescription("Update a household's description"),
			mcp.WithString("household_uid", mcp.Required(), mcp.Description("Household ID")),
			mcp.WithString("description", mcp.Required(), mcp.Description("New description for the household")),
		),
	}
}

func (h *MCPHandlers) handleInitialize(ctx context.Context, params InitializeParams) InitializeResult {
	h.clientInfo = &params.ClientInfo

	h.log().Info("MCP client initialized",
		slog.String("client_name", params.ClientInfo.Name),
		slog.String("client_version", params.ClientInfo.Version),
		slog.String("protocol_version", params.ProtocolVersion),
	)

	return InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    h.capabilities,
		ServerInfo:      h.serverInfo,
		Instructions:    "Assistant Server MCP provides tools for managing todos, notes, preferences, and recipes.",
	}
}

func (h *MCPHandlers) handleInitialized(ctx context.Context) {
	h.log().Info("MCP server ready to handle requests")
}

func (h *MCPHandlers) handleCreateTodo(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	h.log().Debug("Creating todo", slog.Any("arguments", arguments))

	title, ok := arguments["title"].(string)
	if !ok || title == "" {
		h.log().Warn("Create todo failed: missing title", slog.Any("arguments", arguments))
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: title is required"}},
		}
	}

	priority := 3
	if p, ok := arguments["priority"].(float64); ok {
		if p >= 1 && p <= 5 {
			priority = int(p)
		}
	}

	description, _ := arguments["description"].(string)
	userUID, _ := arguments["user_uid"].(string)
	householdUID, _ := arguments["household_uid"].(string)

	var dueDate *time.Time
	if dueDateStr, ok := arguments["due_date"].(string); ok && dueDateStr != "" {
		if parsedDate, err := time.Parse(time.RFC3339, dueDateStr); err == nil {
			dueDate = &parsedDate
		}
	}

	todo := dao.Todo{
		UID:          uuid.NewString(),
		Title:        title,
		Description:  description,
		Data:         "{}",
		Priority:     dao.Priority(priority),
		DueDate:      dueDate,
		UserUID:      &userUID,
		HouseholdUID: &householdUID,
	}

	created, err := h.todoDAO.CreateTodo(ctx, todo)
	if err != nil {
		h.log().Error("Failed to create todo",
			slog.String("error", err.Error()),
			slog.String("title", title),
			slog.String("user_uid", userUID),
			slog.String("household_uid", householdUID),
		)
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to create todo: %v", err)}},
		}
	}

	h.log().Info("Todo created successfully",
		slog.String("todo_id", created.UID),
		slog.String("title", created.Title),
	)

	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Todo created successfully with ID: %s", created.UID)}},
	}
}

func (h *MCPHandlers) handleListTodos(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	h.log().Debug("Listing todos", slog.Any("arguments", arguments))

	limit := 20
	if l, ok := arguments["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	// Use shared filtering logic
	filters := BuildFiltersFromMCP(arguments, TodoFilters.Filters)
	whereClause, whereArgs := BuildWhereClause(filters, TodoFilters.Filters)
	options := dao.ListOptions{
		Limit:       limit,
		Offset:      0,
		SortBy:      "due_date",
		SortDir:     "ASC",
		WhereClause: whereClause,
		WhereArgs:   whereArgs,
	}

	todos, err := h.todoDAO.ListTodos(ctx, options)
	if err != nil {
		h.log().Error("Failed to list todos",
			slog.String("error", err.Error()),
			slog.Any("filters", filters),
		)
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to list todos: %v", err)}},
		}
	}

	h.log().Info("Listed todos successfully",
		slog.Int("count", len(todos)),
		slog.Int("limit", limit),
	)

	result, _ := json.Marshal(todos)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(result)}},
	}
}

func (h *MCPHandlers) handleCompleteTodo(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	h.log().Debug("Completing todo", slog.Any("arguments", arguments))

	todoID, ok := arguments["todo_id"].(string)
	if !ok || todoID == "" {
		h.log().Warn("Complete todo failed: missing todo_id", slog.Any("arguments", arguments))
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: todo_id is required"}},
		}
	}

	completedBy, _ := arguments["completed_by"].(string)

	now := time.Now()
	update := dao.UpdateTodo{
		MarkedComplete: &now,
	}
	if completedBy != "" {
		update.CompletedBy = &completedBy
	}

	_, err := h.todoDAO.UpdateTodo(ctx, todoID, update)
	if err != nil {
		h.log().Error("Failed to complete todo",
			slog.String("error", err.Error()),
			slog.String("todo_id", todoID),
			slog.String("completed_by", completedBy),
		)
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to complete todo: %v", err)}},
		}
	}

	h.log().Info("Todo completed successfully",
		slog.String("todo_id", todoID),
		slog.String("completed_by", completedBy),
	)

	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Todo %s marked as completed", todoID)}},
	}
}

func (h *MCPHandlers) handleSaveNote(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	key, ok := arguments["key"].(string)
	if !ok || key == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: key is required"}},
		}
	}

	data, ok := arguments["data"].(string)
	if !ok || data == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: data is required"}},
		}
	}

	userUID, _ := arguments["user_uid"].(string)
	householdUID, _ := arguments["household_uid"].(string)
	tagsStr, _ := arguments["tags"].(string)

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	note := dao.Notes{
		ID:           uuid.NewString(),
		Key:          key,
		UserUID:      &userUID,
		HouseholdUID: &householdUID,
		Data:         data,
		Tags:         tags,
	}

	created, err := h.notesDAO.CreateNotes(ctx, note)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to save note: %v", err)}},
		}
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Note saved successfully with ID: %s", created.ID)}},
	}
}

func (h *MCPHandlers) handleRecallNote(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	noteID, ok := arguments["note_id"].(string)
	if !ok || noteID == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: note_id is required"}},
		}
	}

	note, err := h.notesDAO.GetNotes(ctx, noteID)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Note not found: %v", err)}},
		}
	}

	result, _ := json.Marshal(note)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(result)}},
	}
}

func (h *MCPHandlers) handleListNotes(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	limit := 20
	if l, ok := arguments["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	// Use shared filtering logic
	filters := BuildFiltersFromMCP(arguments, NotesFilters.Filters)
	whereClause, whereArgs := BuildWhereClause(filters, NotesFilters.Filters)
	options := dao.ListOptions{
		Limit:       limit,
		Offset:      0,
		SortBy:      "created_at",
		SortDir:     "DESC",
		WhereClause: whereClause,
		WhereArgs:   whereArgs,
	}

	notes, err := h.notesDAO.ListNotes(ctx, options)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to list notes: %v", err)}},
		}
	}

	result, _ := json.Marshal(notes)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(result)}},
	}
}

func (h *MCPHandlers) handleSetPreference(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	key, ok := arguments["key"].(string)
	if !ok || key == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: key is required"}},
		}
	}

	specifier, ok := arguments["specifier"].(string)
	if !ok || specifier == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: specifier is required"}},
		}
	}

	data, ok := arguments["data"].(string)
	if !ok || data == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: data is required"}},
		}
	}

	tagsStr, _ := arguments["tags"].(string)
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	pref := dao.Preferences{
		Key:       key,
		Specifier: specifier,
		Data:      data,
		Tags:      tags,
	}

	if _, err := h.preferencesDAO.GetPreferences(ctx, key, specifier); err == nil {
		_, err = h.preferencesDAO.UpdatePreferences(ctx, key, specifier, pref)
		if err != nil {
			return mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to update preference: %v", err)}},
			}
		}
		return mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Preference updated: %s/%s", key, specifier)}},
		}
	} else {
		_, err = h.preferencesDAO.CreatePreferences(ctx, pref)
		if err != nil {
			return mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to create preference: %v", err)}},
			}
		}
		return mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Preference created: %s/%s", key, specifier)}},
		}
	}
}

func (h *MCPHandlers) handleGetPreference(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	key, ok := arguments["key"].(string)
	if !ok || key == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: key is required"}},
		}
	}

	specifier, ok := arguments["specifier"].(string)
	if !ok || specifier == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: specifier is required"}},
		}
	}

	pref, err := h.preferencesDAO.GetPreferences(ctx, key, specifier)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Preference not found: %v", err)}},
		}
	}

	result, _ := json.Marshal(pref)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(result)}},
	}
}

func (h *MCPHandlers) handleSaveRecipe(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	title, ok := arguments["title"].(string)
	if !ok || title == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: title is required"}},
		}
	}

	data, ok := arguments["data"].(string)
	if !ok || data == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: data is required"}},
		}
	}

	genre, _ := arguments["genre"].(string)
	groceryList, _ := arguments["grocery_list"].(string)
	userUID, _ := arguments["user_uid"].(string)
	householdUID, _ := arguments["household_uid"].(string)
	tagsStr, _ := arguments["tags"].(string)

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	var prepTime, cookTime, servings, difficulty, rating *int
	if pt, ok := arguments["prep_time"].(float64); ok {
		prepTime = &[]int{int(pt)}[0]
	}
	if ct, ok := arguments["cook_time"].(float64); ok {
		cookTime = &[]int{int(ct)}[0]
	}
	if s, ok := arguments["servings"].(float64); ok {
		servings = &[]int{int(s)}[0]
	}
	if d, ok := arguments["difficulty"].(float64); ok && d >= 1 && d <= 5 {
		difficulty = &[]int{int(d)}[0]
	}
	if r, ok := arguments["rating"].(float64); ok && r >= 1 && r <= 5 {
		rating = &[]int{int(r)}[0]
	}

	totalTime := 0
	if prepTime != nil && cookTime != nil {
		totalTime = *prepTime + *cookTime
	}
	var totalTimePtr *int
	if totalTime > 0 {
		totalTimePtr = &totalTime
	}

	var genrePtr, groceryListPtr, difficultyPtr *string
	if genre != "" {
		genrePtr = &genre
	}
	if groceryList != "" {
		groceryListPtr = &groceryList
	}
	if difficulty != nil {
		difficultyStr := strconv.Itoa(*difficulty)
		difficultyPtr = &difficultyStr
	}

	recipe := dao.Recipes{
		ID:           uuid.NewString(),
		Title:        title,
		Data:         data,
		Genre:        genrePtr,
		GroceryList:  groceryListPtr,
		PrepTime:     prepTime,
		CookTime:     cookTime,
		TotalTime:    totalTimePtr,
		Servings:     servings,
		Difficulty:   difficultyPtr,
		Rating:       rating,
		Tags:         tags,
		UserUID:      &userUID,
		HouseholdUID: &householdUID,
	}

	created, err := h.recipesDAO.CreateRecipes(ctx, recipe)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to save recipe: %v", err)}},
		}
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Recipe saved successfully with ID: %s", created.ID)}},
	}
}

func (h *MCPHandlers) handleFindRecipes(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	limit := 20
	if l, ok := arguments["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	// Use shared filtering logic
	filters := BuildFiltersFromMCP(arguments, RecipesFilters.Filters)

	// Handle special min_rating filter
	if minRating, ok := arguments["min_rating"].(float64); ok {
		filters["rating"] = ">=" + strconv.Itoa(int(minRating))
	}

	whereClause, whereArgs := BuildWhereClause(filters, RecipesFilters.Filters)
	options := dao.ListOptions{
		Limit:       limit,
		Offset:      0,
		SortBy:      "rating",
		SortDir:     "DESC",
		WhereClause: whereClause,
		WhereArgs:   whereArgs,
	}

	recipes, err := h.recipesDAO.ListRecipes(ctx, options)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to find recipes: %v", err)}},
		}
	}

	result, _ := json.Marshal(recipes)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(result)}},
	}
}

func (h *MCPHandlers) handleGetRecipe(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	recipeID, ok := arguments["recipe_id"].(string)
	if !ok || recipeID == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: recipe_id is required"}},
		}
	}

	recipe, err := h.recipesDAO.GetRecipes(ctx, recipeID)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Recipe not found: %v", err)}},
		}
	}

	result, _ := json.Marshal(recipe)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(result)}},
	}
}

func (h *MCPHandlers) handleUpdateUserDescription(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	userUID, ok := arguments["user_uid"].(string)
	if !ok || userUID == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: user_uid is required"}},
		}
	}

	description, ok := arguments["description"].(string)
	if !ok {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: description is required"}},
		}
	}

	update := dao.UpdateUser{
		Description: &description,
	}

	updatedUser, err := h.userDAO.UpdateUser(ctx, userUID, update)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to update user description: %v", err)}},
		}
	}

	result, _ := json.Marshal(updatedUser)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("User description updated successfully: %s", string(result))}},
	}
}

func (h *MCPHandlers) handleUpdateHouseholdDescription(ctx context.Context, arguments map[string]any) mcp.CallToolResult {
	householdUID, ok := arguments["household_uid"].(string)
	if !ok || householdUID == "" {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: household_uid is required"}},
		}
	}

	description, ok := arguments["description"].(string)
	if !ok {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: description is required"}},
		}
	}

	update := dao.UpdateHousehold{
		Description: &description,
	}

	updatedHousehold, err := h.householdDAO.UpdateHousehold(ctx, householdUID, update)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Failed to update household description: %v", err)}},
		}
	}

	result, _ := json.Marshal(updatedHousehold)
	return mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Household description updated successfully: %s", string(result))}},
	}
}

func (h *MCPHandlers) callTool(ctx context.Context, name string, arguments map[string]any) mcp.CallToolResult {
	h.log().Info("Calling MCP tool",
		slog.String("tool_name", name),
		slog.Any("arguments", arguments),
	)

	start := time.Now()
	defer func() {
		h.log().Debug("Tool execution completed",
			slog.String("tool_name", name),
			slog.Duration("duration", time.Since(start)),
		)
	}()

	switch name {
	case "create_todo":
		return h.handleCreateTodo(ctx, arguments)
	case "list_todos":
		return h.handleListTodos(ctx, arguments)
	case "complete_todo":
		return h.handleCompleteTodo(ctx, arguments)
	case "save_note":
		return h.handleSaveNote(ctx, arguments)
	case "recall_note":
		return h.handleRecallNote(ctx, arguments)
	case "list_notes":
		return h.handleListNotes(ctx, arguments)
	case "set_preference":
		return h.handleSetPreference(ctx, arguments)
	case "get_preference":
		return h.handleGetPreference(ctx, arguments)
	case "save_recipe":
		return h.handleSaveRecipe(ctx, arguments)
	case "find_recipes":
		return h.handleFindRecipes(ctx, arguments)
	case "get_recipe":
		return h.handleGetRecipe(ctx, arguments)
	case "update_user_description":
		return h.handleUpdateUserDescription(ctx, arguments)
	case "update_household_description":
		return h.handleUpdateHouseholdDescription(ctx, arguments)
	default:
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: Unknown tool: %s", name)}},
		}
	}
}

func (h *MCPHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log().Error("Invalid JSON-RPC request",
			slog.String("error", err.Error()),
			slog.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	h.log().Debug("Received JSON-RPC request",
		slog.String("method", req.Method),
		slog.Any("id", req.ID),
		slog.String("remote_addr", r.RemoteAddr),
	)

	var response JSONRPCResponse
	response.JSONRPC = "2.0"
	response.ID = req.ID

	switch req.Method {
	case "initialize":
		if params, ok := req.Params.(map[string]any); ok {
			var initParams InitializeParams

			if protocolVersion, ok := params["protocolVersion"].(string); ok {
				initParams.ProtocolVersion = protocolVersion
			}

			if clientInfo, ok := params["clientInfo"].(map[string]any); ok {
				if name, ok := clientInfo["name"].(string); ok {
					initParams.ClientInfo.Name = name
				}
				if title, ok := clientInfo["title"].(string); ok {
					initParams.ClientInfo.Title = title
				}
				if version, ok := clientInfo["version"].(string); ok {
					initParams.ClientInfo.Version = version
				}
			}

			if capabilities, ok := params["capabilities"].(map[string]any); ok {
				if roots, ok := capabilities["roots"].(map[string]any); ok {
					if listChanged, ok := roots["listChanged"].(bool); ok {
						initParams.Capabilities.Roots = &RootsCapability{ListChanged: listChanged}
					}
				}
				if sampling, ok := capabilities["sampling"].(map[string]any); ok {
					initParams.Capabilities.Sampling = sampling
				}
				if elicitation, ok := capabilities["elicitation"].(map[string]any); ok {
					initParams.Capabilities.Elicitation = elicitation
				}
			}

			result := h.handleInitialize(r.Context(), initParams)
			response.Result = result
		} else {
			response.Error = map[string]any{"code": -32602, "message": "Invalid params"}
		}
	case "initialized":
		h.handleInitialized(r.Context())
		response.Result = map[string]any{}
	case "tools/list":
		response.Result = mcp.ListToolsResult{Tools: h.tools}
	case "tools/call":
		params, ok := req.Params.(map[string]any)
		if !ok {
			response.Error = map[string]any{"code": -32602, "message": "Invalid params"}
		} else {
			toolName, ok := params["name"].(string)
			if !ok {
				response.Error = map[string]any{"code": -32602, "message": "Tool name is required"}
			} else {
				arguments, _ := params["arguments"].(map[string]any)
				result := h.callTool(r.Context(), toolName, arguments)
				response.Result = result
			}
		}
	default:
		h.log().Warn("Unknown JSON-RPC method",
			slog.String("method", req.Method),
			slog.Any("id", req.ID),
		)
		response.Error = map[string]any{"code": -32601, "message": "Method not found"}
	}

	if response.Error != nil {
		h.log().Error("JSON-RPC request failed",
			slog.String("method", req.Method),
			slog.Any("id", req.ID),
			slog.Any("error", response.Error),
		)
	} else {
		h.log().Debug("JSON-RPC request completed successfully",
			slog.String("method", req.Method),
			slog.Any("id", req.ID),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log().Error("Failed to encode JSON-RPC response",
			slog.String("error", err.Error()),
		)
	}
}

func NewMCPRouter(todoDAO todoDAO, notesDAO notesDAO, preferencesDAO preferencesDAO, recipesDAO recipesDAO, userDAO userDAO, householdDAO householdDAO) http.Handler {
	h := NewMCP(todoDAO, notesDAO, preferencesDAO, recipesDAO, userDAO, householdDAO)

	r := chi.NewRouter()
	r.Post("/", h.ServeHTTP)
	return r
}
