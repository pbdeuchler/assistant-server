package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
)

type recipesDAO interface {
	CreateRecipes(ctx context.Context, r dao.Recipes) (dao.Recipes, error)
	GetRecipes(ctx context.Context, id string) (dao.Recipes, error)
	ListRecipes(ctx context.Context, options dao.ListOptions) ([]dao.Recipes, error)
	UpdateRecipes(ctx context.Context, id string, r dao.Recipes) (dao.Recipes, error)
	DeleteRecipes(ctx context.Context, id string) error
}

type RecipesHandlers struct{ dao recipesDAO }

func NewRecipes(dao recipesDAO) http.Handler {
	h := &RecipesHandlers{dao}
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Get("/", h.list)
	return r
}

func (h *RecipesHandlers) create(w http.ResponseWriter, r *http.Request) {
	var recipe dao.Recipes
	if json.NewDecoder(r.Body).Decode(&recipe) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recipe.ID = uuid.NewString()
	out, err := h.dao.CreateRecipes(r.Context(), recipe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *RecipesHandlers) get(w http.ResponseWriter, r *http.Request) {
	out, err := h.dao.GetRecipes(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *RecipesHandlers) update(w http.ResponseWriter, r *http.Request) {
	var recipe dao.Recipes
	if json.NewDecoder(r.Body).Decode(&recipe) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := h.dao.UpdateRecipes(r.Context(), chi.URLParam(r, "id"), recipe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *RecipesHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if h.dao.DeleteRecipes(r.Context(), chi.URLParam(r, "id")) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RecipesHandlers) list(w http.ResponseWriter, r *http.Request) {
	params := ParseListParams(r, RecipesFilters.SortFields)
	
	// Handle special recipe filters
	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		params.Filters["rating"] = ">=" + minRating
	}
	if maxCookTime := r.URL.Query().Get("max_cook_time"); maxCookTime != "" {
		params.Filters["cook_time"] = "<=" + maxCookTime
	}
	
	whereClause, whereArgs := BuildWhereClause(params.Filters, RecipesFilters.Filters)

	options := dao.ListOptions{
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortDir:     params.SortDir,
		WhereClause: whereClause,
		WhereArgs:   whereArgs,
	}

	out, err := h.dao.ListRecipes(r.Context(), options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}