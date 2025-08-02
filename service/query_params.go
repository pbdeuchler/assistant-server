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