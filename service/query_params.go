package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ListParams struct {
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
	Filters map[string]string
}

func ParseListParams(r *http.Request, allowedSortFields []string) ListParams {
	params := ListParams{
		Limit:   100,
		Offset:  0,
		SortBy:  "created_at",
		SortDir: "DESC",
		Filters: make(map[string]string),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			params.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			params.Offset = o
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		for _, allowed := range allowedSortFields {
			if sortBy == allowed {
				params.SortBy = sortBy
				break
			}
		}
	}

	if sortDir := strings.ToUpper(r.URL.Query().Get("sort_dir")); sortDir == "ASC" || sortDir == "DESC" {
		params.SortDir = sortDir
	}

	for key, values := range r.URL.Query() {
		if len(values) > 0 && !isReservedParam(key) {
			params.Filters[key] = values[0]
		}
	}

	return params
}

func isReservedParam(key string) bool {
	reserved := []string{"limit", "offset", "sort_by", "sort_dir"}
	for _, r := range reserved {
		if key == r {
			return true
		}
	}
	return false
}

func BuildWhereClause(filters map[string]string, allowedFilters []string) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	for key, value := range filters {
		// Handle tag filtering specially
		if key == "tags" {
			conditions = append(conditions, fmt.Sprintf("tags @> $%d", argIndex))
			args = append(args, []string{value})
			argIndex++
			continue
		}

		// Handle regular filters
		for _, allowed := range allowedFilters {
			if key == allowed {
				conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argIndex))
				args = append(args, value)
				argIndex++
				break
			}
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// BuildFiltersFromMCP creates a filter map from MCP tool arguments
func BuildFiltersFromMCP(arguments map[string]any, supportedFilters []string) map[string]string {
	filters := make(map[string]string)
	
	for _, filterName := range supportedFilters {
		if value, ok := arguments[filterName]; ok {
			switch v := value.(type) {
			case string:
				if v != "" {
					filters[filterName] = v
				}
			case float64:
				filters[filterName] = strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				filters[filterName] = strconv.Itoa(v)
			}
		}
	}
	
	// Handle special boolean filters
	if completedOnly, ok := arguments["completed_only"].(bool); ok && completedOnly {
		filters["completed_by"] = "NOT NULL"
	}
	if pendingOnly, ok := arguments["pending_only"].(bool); ok && pendingOnly {
		filters["completed_by"] = "IS NULL"
	}
	
	return filters
}

// Common filter configurations for each entity type
type EntityFilters struct {
	SortFields []string
	Filters    []string
}

var (
	TodoFilters = EntityFilters{
		SortFields: []string{"uid", "title", "priority", "due_date", "created_at", "updated_at", "user_id", "household_id", "completed_by"},
		Filters:    []string{"title", "priority", "user_id", "household_id", "completed_by", "tags"},
	}
	
	NotesFilters = EntityFilters{
		SortFields: []string{"id", "key", "user_id", "household_id", "created_at", "updated_at"},
		Filters:    []string{"key", "user_id", "household_id", "tags"},
	}
	
	PreferencesFilters = EntityFilters{
		SortFields: []string{"key", "specifier", "created_at", "updated_at"},
		Filters:    []string{"key", "specifier", "tags"},
	}
	
	RecipesFilters = EntityFilters{
		SortFields: []string{"id", "title", "genre", "rating", "prep_time", "cook_time", "total_time", "servings", "difficulty", "user_id", "household_id", "created_at", "updated_at"},
		Filters:    []string{"title", "genre", "rating", "difficulty", "user_id", "household_id", "tags"},
	}
)