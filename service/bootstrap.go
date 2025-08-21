package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
	"golang.org/x/oauth2"
)

type bootstrapDAO interface {
	GetUserBySlackUserUID(ctx context.Context, slackUserUID string) (dao.Users, error)
	GetUser(ctx context.Context, uid string) (dao.Users, error)
	GetCredentialsByUserUID(ctx context.Context, userUID string) ([]dao.Credentials, error)
	GetTodosByUserUID(ctx context.Context, userUID string) ([]dao.Todo, error)
	GetNotesByUserUID(ctx context.Context, userUID string) ([]dao.Notes, error)
	GetPreferencesByUserUID(ctx context.Context, userUID string) ([]dao.Preferences, error)
	GetRecipesByUserUID(ctx context.Context, userUID string) ([]dao.Recipes, error)
	GetHousehold(ctx context.Context, uid string) (dao.Households, error)
	UpdateCredentials(ctx context.Context, id string, c dao.Credentials) (dao.Credentials, error)
}

type bootstrapHandlers struct{ dao bootstrapDAO }

func NewBootstrap(dao bootstrapDAO) http.Handler {
	h := &bootstrapHandlers{dao}
	r := chi.NewRouter()
	r.Use(httpLogger())
	r.Get("/", h.bootstrap)
	return r
}

type BootstrapResponse struct {
	User               dao.Users         `json:"user"`
	Household          *dao.Households   `json:"household,omitempty"`
	Todos              []dao.Todo        `json:"todos,omitempty"`
	Notes              []dao.Notes       `json:"notes,omitempty"`
	Preferences        []dao.Preferences `json:"preferences,omitempty"`
	Recipes            []dao.Recipes     `json:"recipes,omitempty"`
	Prompt             string            `json:"prompt,omitempty"`
	AppendSystemPrompt string            `json:"append_system_prompt,omitempty"`
	AllowedTools       []string          `json:"allowed_tools,omitempty"`
	DisallowedTools    []string          `json:"disallowed_tools,omitempty"`
	Env                map[string]string `json:"env"`
}

func (h *bootstrapHandlers) bootstrap(w http.ResponseWriter, r *http.Request) {
	slackID := r.URL.Query().Get("slack_id")
	if slackID == "" {
		http.Error(w, "slack_id query parameter is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Look up the user by slack ID
	user, err := h.dao.GetUserBySlackUserUID(ctx, slackID)
	if err != nil {
		http.Error(w, "User not found for slack ID: "+err.Error(), http.StatusNotFound)
		return
	}

	// Get user credentials
	credentials, err := h.dao.GetCredentialsByUserUID(ctx, user.UID)
	if err != nil {
		slog.Error("Failed to get credentials", "user_id", user.UID, "error", err)
		credentials = []dao.Credentials{} // Continue with empty credentials
	}

	// Validate and refresh credentials, collect environment variables
	env := make(map[string]string)
	for _, cred := range credentials {
		if credEnv, err := h.validateAndRefreshCredential(ctx, cred); err == nil {
			for key, value := range credEnv {
				env[key] = value
			}
		} else {
			slog.Error("Failed to validate credential", "credential_id", cred.ID, "error", err)
			// return the oauth url and an error message to the user
			if cred.CredentialType == "GOOGLE_CALENDAR" {
				oauthURL := fmt.Sprintf("/oauth/google?user_id=%s", user.UID) // Scope for Google Calendar
				http.Error(w, fmt.Sprintf("Please authorize your Google Calendar account: %s", oauthURL), http.StatusUnauthorized)
				return
			} else {
				slog.Warn("Unsupported credential type, skipping", "credential_type", cred.CredentialType)
			}
		}
	}

	if os.Getenv("DEVELOPMENT") != "" {
		env["ASSISTANTSERVER_HOST"] = "http://127.0.0.1:8080/mcp"
		env["FS_SHIM"] = "1"
	}

	// Get todos
	todos, err := h.dao.GetTodosByUserUID(ctx, user.UID)
	if err != nil {
		slog.Error("Failed to get todos", "user_id", user.UID, "error", err)
		todos = []dao.Todo{}
	}

	// Get notes
	notes, err := h.dao.GetNotesByUserUID(ctx, user.UID)
	if err != nil {
		slog.Error("Failed to get notes", "user_id", user.UID, "error", err)
		notes = []dao.Notes{}
	}

	// Get preferences
	preferences, err := h.dao.GetPreferencesByUserUID(ctx, user.UID)
	if err != nil {
		slog.Error("Failed to get preferences", "user_id", user.UID, "error", err)
		preferences = []dao.Preferences{}
	}

	// Get recipes
	// recipes, err := h.dao.GetRecipesByUserUID(ctx, user.UID)
	// if err != nil {
	// 	slog.Error("Failed to get recipes", "user_id", user.UID, "error", err)
	// 	recipes = []dao.Recipes{}
	// }

	// Try to get household if user is associated with one
	var household *dao.Households
	if user.HouseholdUID != nil && *user.HouseholdUID != "" {
		if h, err := h.dao.GetHousehold(ctx, *user.HouseholdUID); err == nil {
			household = &h
		} else {
			slog.Error("Failed to get household", "household_uid", *user.HouseholdUID, "error", err)
		}
	}

	// Compile structured prompt for LLM
	prompt := h.compileLLMPrompt(user, household, todos, notes, preferences)

	response := BootstrapResponse{
		User:               user,
		Todos:              todos,
		Notes:              notes,
		Preferences:        preferences,
		AppendSystemPrompt: prompt,
		AllowedTools:       []string{"mcp__assistant-mcp"},
		DisallowedTools:    []string{"TodoWrite"},
		Env:                env,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *bootstrapHandlers) validateAndRefreshCredential(ctx context.Context, cred dao.Credentials) (map[string]string, error) {
	env := make(map[string]string)

	if cred.CredentialType != "GOOGLE_CALENDAR" {
		// For now, only handle Google Calendar credentials
		return env, nil
	}

	// Parse the OAuth token from JSON
	var token oauth2.Token
	if err := json.Unmarshal(cred.Value, &token); err != nil {
		slog.Error("Failed to unmarshal OAuth token", "credential_id", cred.ID, "error", err)
		return nil, fmt.Errorf("failed to unmarshal OAuth token: %w", err)
	}

	// Check if token is expired and has refresh token
	if token.Expiry.Before(time.Now()) && token.RefreshToken != "" {
		slog.Info("Token expired, attempting refresh", "credential_id", cred.ID)

		// Create OAuth2 config for token refresh
		oauth2Config := &oauth2.Config{
			// We need these from environment or config - for now use placeholders
			ClientID:     os.Getenv("GCLOUD_CLIENT_ID"),
			ClientSecret: os.Getenv("GCLOUD_CLIENT_SECRET"),
		}

		// Attempt to refresh the token
		newToken, err := oauth2Config.TokenSource(ctx, &token).Token()
		if err != nil {
			slog.Error("Failed to refresh token", "credential_id", cred.ID, "error", err)
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Update the credential with the new token
		newTokenJSON, err := json.Marshal(newToken)
		if err != nil {
			slog.Error("Failed to marshal refreshed token", "credential_id", cred.ID, "error", err)
			return nil, fmt.Errorf("failed to marshal refreshed token: %w", err)
		}

		cred.Value = newTokenJSON
		_, err = h.dao.UpdateCredentials(ctx, cred.ID, cred)
		if err != nil {
			slog.Error("Failed to update credential", "credential_id", cred.ID, "error", err)
			return nil, fmt.Errorf("failed to update credential: %w", err)
		}

		slog.Info("Successfully refreshed and updated token", "credential_id", cred.ID)
		env["GOOGLE_API_ACCESS_TOKEN"] = newToken.AccessToken
		return env, nil
	}

	// Token is still valid
	env["GOOGLE_API_ACCESS_TOKEN"] = token.AccessToken
	return env, nil
}

func (h *bootstrapHandlers) compileLLMPrompt(user dao.Users, household *dao.Households, todos []dao.Todo, notes []dao.Notes, preferences []dao.Preferences) string {
	var prompt strings.Builder

	prompt.WriteString("# User Context\n\n")
	prompt.WriteString(fmt.Sprintf("**User:** \n %s | %s | user_uid=%s\n", user.Name, user.Email, user.UID))
	if user.Description != "" {
		prompt.WriteString(fmt.Sprintf("**Description:** %s\n", user.Description))
	}
	prompt.WriteString("\n")

	if household != nil {
		prompt.WriteString("# Household Context\n\n")
		prompt.WriteString(fmt.Sprintf("**Household:** %s (uid=%s)\n", household.Name, household.UID))
		if household.Description != "" {
			prompt.WriteString(fmt.Sprintf("**Description:** %s\n", household.Description))
		}
		prompt.WriteString("\n")
	}

	if len(todos) > 0 {
		prompt.WriteString("# Todos\n\n")
		for _, todo := range todos {
			prompt.WriteString(fmt.Sprintf("- **%s**", todo.Title))
			if todo.Description != "" {
				prompt.WriteString(fmt.Sprintf(" - %s", todo.Description))
			}
			if todo.DueDate != nil {
				prompt.WriteString(fmt.Sprintf(" (Due: %s)", todo.DueDate.Format("2006-01-02")))
			}
			prompt.WriteString("\n")
		}
		prompt.WriteString("\n")
	}

	if len(notes) > 0 {
		prompt.WriteString("# Notes\n\n")
		for _, note := range notes {
			prompt.WriteString(fmt.Sprintf("- **%s**: %s\n", note.Key, note.Data))
		}
		prompt.WriteString("\n")
	}

	if len(preferences) > 0 {
		prompt.WriteString("# Preferences\n\n")
		for _, pref := range preferences {
			prompt.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", pref.Key, pref.Specifier, pref.Data))
		}
		prompt.WriteString("\n")
	}

	return prompt.String()
}
