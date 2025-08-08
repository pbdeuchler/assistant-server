package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
)

type todoDAO interface {
	CreateTodo(ctx context.Context, t dao.Todo) (dao.Todo, error)
	GetTodo(ctx context.Context, uid string) (dao.Todo, error)
	ListTodos(ctx context.Context, options dao.ListOptions) ([]dao.Todo, error)
	UpdateTodo(ctx context.Context, uid string, t dao.UpdateTodo) (dao.Todo, error)
	DeleteTodo(ctx context.Context, uid string) error
}

type todoHandlers struct{ dao todoDAO }

func NewTodos(dao todoDAO) http.Handler {
	h := &todoHandlers{dao}
	r := chi.NewRouter()
	r.Use(httpLogger())
	r.Post("/", h.create)
	r.Get("/{uid}", h.get)
	r.Put("/{uid}", h.update)
	r.Delete("/{uid}", h.delete)
	r.Get("/", h.list)
	return r
}

type createTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Data        string `json:"data"`
	Priority    int    `json:"priority"`
	DueDate     string `json:"due_date"`
	RecursOn    string `json:"recurs_on"`
	ExternalURL string `json:"external_url"`
	UserID      string `json:"user_id"`
	HouseholdID string `json:"household_id"`
}

func (h *todoHandlers) create(w http.ResponseWriter, r *http.Request) {
	var todoReq createTodoRequest
	if json.NewDecoder(r.Body).Decode(&todoReq) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var dueDate *time.Time
	if todoReq.DueDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, todoReq.DueDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid due date: " + err.Error()})
			return
		} else {
			dueDate = &parsedDate
		}
	} else {
		dueDate = nil
	}
	if todoReq.Priority < 1 || todoReq.Priority > 5 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "priority must be between 1 and 5"})
		return
	}

	if todoReq.Data == "" {
		todoReq.Data = "{}" // Default to empty JSON object if no data is provided
	} else {
		// Validate that Data is a valid JSON string
		var js map[string]any
		if err := json.Unmarshal([]byte(todoReq.Data), &js); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json submitted for data: " + err.Error()})
			return
		}
	}

	priority := dao.Priority(todoReq.Priority)
	t := dao.Todo{
		Title:       todoReq.Title,
		Description: todoReq.Description,
		Data:        todoReq.Data,
		Priority:    priority,
		DueDate:     dueDate,
		RecursOn:    todoReq.RecursOn,
		ExternalURL: todoReq.ExternalURL,
		UserID:      todoReq.UserID,
		HouseholdID: todoReq.HouseholdID,
		UID:         uuid.NewString(),
	}
	out, err := h.dao.CreateTodo(r.Context(), t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to create todo", "error", err)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *todoHandlers) get(w http.ResponseWriter, r *http.Request) {
	out, err := h.dao.GetTodo(r.Context(), chi.URLParam(r, "uid"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *todoHandlers) update(w http.ResponseWriter, r *http.Request) {
	var t dao.UpdateTodo
	if json.NewDecoder(r.Body).Decode(&t) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := h.dao.UpdateTodo(r.Context(), chi.URLParam(r, "uid"), t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *todoHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if h.dao.DeleteTodo(r.Context(), chi.URLParam(r, "uid")) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *todoHandlers) list(w http.ResponseWriter, r *http.Request) {
	allowedSortFields := []string{"uid", "title", "priority", "due_date", "created_at", "updated_at", "user_id", "household_id", "completed_by"}
	allowedFilters := []string{"title", "priority", "user_id", "household_id", "completed_by"}

	params := ParseListParams(r, allowedSortFields)
	whereClause, whereArgs := BuildWhereClause(params.Filters, allowedFilters)

	options := dao.ListOptions{
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortDir:     params.SortDir,
		WhereClause: whereClause,
		WhereArgs:   whereArgs,
	}

	out, err := h.dao.ListTodos(r.Context(), options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}
