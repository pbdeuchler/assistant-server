package service

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	dao "github.com/pbdeuchler/assistant-mcp/dao/postgres"
)

type BackgroundHandlers struct{ dao assistant }

func NewBackgroundHandlers(dao assistant) http.Handler {
	h := &BackgroundHandlers{dao}
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/{key}", h.get)
	r.Put("/{key}", h.update)
	r.Delete("/{key}", h.delete)
	r.Get("/", h.list)
	return r
}

func (h *BackgroundHandlers) create(w http.ResponseWriter, r *http.Request) {
	var b dao.Background
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := h.dao.CreateBackground(r.Context(), b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *BackgroundHandlers) get(w http.ResponseWriter, r *http.Request) {
	out, err := h.dao.GetBackground(r.Context(), chi.URLParam(r, "key"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *BackgroundHandlers) update(w http.ResponseWriter, r *http.Request) {
	var b dao.Background
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := h.dao.UpdateBackground(r.Context(), chi.URLParam(r, "key"), b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *BackgroundHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if h.dao.DeleteBackground(r.Context(), chi.URLParam(r, "key")) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BackgroundHandlers) list(w http.ResponseWriter, r *http.Request) {
	allowedSortFields := []string{"key", "created_at", "updated_at"}
	allowedFilters := []string{"key"}

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

	out, err := h.dao.ListBackgrounds(r.Context(), options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

