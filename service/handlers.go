package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dao "github.com/pbdeuchler/assistant-mcp/dao/postgres"
)

type assistant interface {
	CreateTodo(ctx context.Context, t dao.Todo) (dao.Todo, error)
	GetTodo(ctx context.Context, uid string) (dao.Todo, error)
	ListTodos(ctx context.Context, options dao.ListOptions) ([]dao.Todo, error)
	UpdateTodo(ctx context.Context, uid string, t dao.Todo) (dao.Todo, error)
	DeleteTodo(ctx context.Context, uid string) error
	CreateBackground(ctx context.Context, b dao.Background) (dao.Background, error)
	GetBackground(ctx context.Context, key string) (dao.Background, error)
	ListBackgrounds(ctx context.Context, options dao.ListOptions) ([]dao.Background, error)
	UpdateBackground(ctx context.Context, key string, b dao.Background) (dao.Background, error)
	DeleteBackground(ctx context.Context, key string) error
	CreatePreferences(ctx context.Context, p dao.Preferences) (dao.Preferences, error)
	GetPreferences(ctx context.Context, key, specifier string) (dao.Preferences, error)
	ListPreferences(ctx context.Context, options dao.ListOptions) ([]dao.Preferences, error)
	UpdatePreferences(ctx context.Context, key, specifier string, p dao.Preferences) (dao.Preferences, error)
	DeletePreferences(ctx context.Context, key, specifier string) error
	CreateNotes(ctx context.Context, n dao.Notes) (dao.Notes, error)
	GetNotes(ctx context.Context, id string) (dao.Notes, error)
	ListNotes(ctx context.Context, options dao.ListOptions) ([]dao.Notes, error)
	UpdateNotes(ctx context.Context, id string, n dao.Notes) (dao.Notes, error)
	DeleteNotes(ctx context.Context, id string) error
}

type Handlers struct{ dao assistant }

func NewHandlers(dao assistant) http.Handler {
	h := &Handlers{dao}
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/{uid}", h.get)
	r.Put("/{uid}", h.update)
	r.Delete("/{uid}", h.delete)
	r.Get("/", h.list)
	return r
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	var t dao.Todo
	if json.NewDecoder(r.Body).Decode(&t) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	t.UID = uuid.NewString()
	out, err := h.dao.CreateTodo(r.Context(), t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	out, err := h.dao.GetTodo(r.Context(), chi.URLParam(r, "uid"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handlers) update(w http.ResponseWriter, r *http.Request) {
	var t dao.Todo
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

func (h *Handlers) delete(w http.ResponseWriter, r *http.Request) {
	if h.dao.DeleteTodo(r.Context(), chi.URLParam(r, "uid")) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	allowedSortFields := []string{"uid", "title", "priority", "due_date", "created_at", "updated_at", "created_by", "completed_by"}
	allowedFilters := []string{"title", "priority", "created_by", "completed_by"}

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
