package service

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
)

type PreferencesHandlers struct{ dao assistant }

func NewPreferencesHandlers(dao assistant) http.Handler {
	h := &PreferencesHandlers{dao}
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/{key}/{specifier}", h.get)
	r.Put("/{key}/{specifier}", h.update)
	r.Delete("/{key}/{specifier}", h.delete)
	r.Get("/", h.list)
	return r
}

func (h *PreferencesHandlers) create(w http.ResponseWriter, r *http.Request) {
	var p dao.Preferences
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := h.dao.CreatePreferences(r.Context(), p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *PreferencesHandlers) get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	specifier := chi.URLParam(r, "specifier")
	out, err := h.dao.GetPreferences(r.Context(), key, specifier)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *PreferencesHandlers) update(w http.ResponseWriter, r *http.Request) {
	var p dao.Preferences
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := chi.URLParam(r, "key")
	specifier := chi.URLParam(r, "specifier")
	out, err := h.dao.UpdatePreferences(r.Context(), key, specifier, p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *PreferencesHandlers) delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	specifier := chi.URLParam(r, "specifier")
	if h.dao.DeletePreferences(r.Context(), key, specifier) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PreferencesHandlers) list(w http.ResponseWriter, r *http.Request) {
	allowedSortFields := []string{"key", "specifier", "created_at", "updated_at"}
	allowedFilters := []string{"key", "specifier"}

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

	out, err := h.dao.ListPreferences(r.Context(), options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}
