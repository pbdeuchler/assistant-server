package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
)

type notesDAO interface {
	CreateNotes(ctx context.Context, n dao.Notes) (dao.Notes, error)
	GetNotes(ctx context.Context, id string) (dao.Notes, error)
	ListNotes(ctx context.Context, options dao.ListOptions) ([]dao.Notes, error)
	UpdateNotes(ctx context.Context, id string, n dao.Notes) (dao.Notes, error)
	DeleteNotes(ctx context.Context, id string) error
}

type NotesHandlers struct{ dao notesDAO }

func NewNotes(dao notesDAO) http.Handler {
	h := &NotesHandlers{dao}
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Get("/", h.list)
	return r
}

func (h *NotesHandlers) create(w http.ResponseWriter, r *http.Request) {
	var n dao.Notes
	if json.NewDecoder(r.Body).Decode(&n) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	n.ID = uuid.NewString()
	out, err := h.dao.CreateNotes(r.Context(), n)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *NotesHandlers) get(w http.ResponseWriter, r *http.Request) {
	out, err := h.dao.GetNotes(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *NotesHandlers) update(w http.ResponseWriter, r *http.Request) {
	var n dao.Notes
	if json.NewDecoder(r.Body).Decode(&n) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := h.dao.UpdateNotes(r.Context(), chi.URLParam(r, "id"), n)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *NotesHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if h.dao.DeleteNotes(r.Context(), chi.URLParam(r, "id")) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotesHandlers) list(w http.ResponseWriter, r *http.Request) {
	params := ParseListParams(r, NotesFilters.SortFields)
	whereClause, whereArgs := BuildWhereClause(params.Filters, NotesFilters.Filters)

	options := dao.ListOptions{
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortDir:     params.SortDir,
		WhereClause: whereClause,
		WhereArgs:   whereArgs,
	}

	out, err := h.dao.ListNotes(r.Context(), options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}
