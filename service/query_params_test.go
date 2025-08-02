package service

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestParseListParams_Defaults(t *testing.T) {
	req := &http.Request{URL: &url.URL{}}
	allowedSortFields := []string{"created_at", "updated_at"}
	
	params := ParseListParams(req, allowedSortFields)
	
	expected := ListParams{
		Limit:   100,
		Offset:  0,
		SortBy:  "created_at",
		SortDir: "DESC",
		Filters: make(map[string]string),
	}
	
	if !reflect.DeepEqual(params, expected) {
		t.Errorf("Expected %+v, got %+v", expected, params)
	}
}

func TestParseListParams_CustomValues(t *testing.T) {
	u, _ := url.Parse("?limit=50&offset=25&sort_by=updated_at&sort_dir=ASC&priority=high&created_by=user123")
	req := &http.Request{URL: u}
	allowedSortFields := []string{"created_at", "updated_at", "priority"}
	
	params := ParseListParams(req, allowedSortFields)
	
	if params.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", params.Limit)
	}
	if params.Offset != 25 {
		t.Errorf("Expected offset 25, got %d", params.Offset)
	}
	if params.SortBy != "updated_at" {
		t.Errorf("Expected sort_by 'updated_at', got '%s'", params.SortBy)
	}
	if params.SortDir != "ASC" {
		t.Errorf("Expected sort_dir 'ASC', got '%s'", params.SortDir)
	}
	if params.Filters["priority"] != "high" {
		t.Errorf("Expected filter priority 'high', got '%s'", params.Filters["priority"])
	}
	if params.Filters["created_by"] != "user123" {
		t.Errorf("Expected filter created_by 'user123', got '%s'", params.Filters["created_by"])
	}
}

func TestParseListParams_InvalidSortField(t *testing.T) {
	u, _ := url.Parse("?sort_by=invalid_field")
	req := &http.Request{URL: u}
	allowedSortFields := []string{"created_at", "updated_at"}
	
	params := ParseListParams(req, allowedSortFields)
	
	if params.SortBy != "created_at" {
		t.Errorf("Expected default sort_by 'created_at' for invalid field, got '%s'", params.SortBy)
	}
}

func TestParseListParams_LimitBounds(t *testing.T) {
	tests := []struct {
		limitParam string
		expected   int
	}{
		{"0", 100},     // invalid, should use default
		{"-5", 100},    // invalid, should use default
		{"1001", 100},  // over max, should use default
		{"50", 50},     // valid
		{"1000", 1000}, // max valid
	}
	
	for _, test := range tests {
		u, _ := url.Parse("?limit=" + test.limitParam)
		req := &http.Request{URL: u}
		allowedSortFields := []string{"created_at"}
		
		params := ParseListParams(req, allowedSortFields)
		
		if params.Limit != test.expected {
			t.Errorf("For limit=%s, expected %d, got %d", test.limitParam, test.expected, params.Limit)
		}
	}
}

func TestParseListParams_InvalidSortDir(t *testing.T) {
	u, _ := url.Parse("?sort_dir=invalid")
	req := &http.Request{URL: u}
	allowedSortFields := []string{"created_at"}
	
	params := ParseListParams(req, allowedSortFields)
	
	if params.SortDir != "DESC" {
		t.Errorf("Expected default sort_dir 'DESC' for invalid value, got '%s'", params.SortDir)
	}
}

func TestBuildWhereClause_NoFilters(t *testing.T) {
	filters := map[string]string{}
	allowedFilters := []string{"name", "status"}
	
	whereClause, args := BuildWhereClause(filters, allowedFilters)
	
	if whereClause != "" {
		t.Errorf("Expected empty where clause, got '%s'", whereClause)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestBuildWhereClause_SingleFilter(t *testing.T) {
	filters := map[string]string{"status": "active"}
	allowedFilters := []string{"status", "name"}
	
	whereClause, args := BuildWhereClause(filters, allowedFilters)
	
	expected := "WHERE status = $1"
	if whereClause != expected {
		t.Errorf("Expected '%s', got '%s'", expected, whereClause)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("Expected args ['active'], got %v", args)
	}
}

func TestBuildWhereClause_MultipleFilters(t *testing.T) {
	filters := map[string]string{"status": "active", "name": "test"}
	allowedFilters := []string{"status", "name"}
	
	whereClause, args := BuildWhereClause(filters, allowedFilters)
	
	if whereClause != "WHERE status = $1 AND name = $2" && whereClause != "WHERE name = $1 AND status = $2" {
		t.Errorf("Unexpected where clause: '%s'", whereClause)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestBuildWhereClause_DisallowedFilter(t *testing.T) {
	filters := map[string]string{"status": "active", "password": "secret"}
	allowedFilters := []string{"status"}
	
	whereClause, args := BuildWhereClause(filters, allowedFilters)
	
	expected := "WHERE status = $1"
	if whereClause != expected {
		t.Errorf("Expected '%s', got '%s'", expected, whereClause)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("Expected args ['active'], got %v", args)
	}
}

func TestIsReservedParam(t *testing.T) {
	reserved := []string{"limit", "offset", "sort_by", "sort_dir"}
	notReserved := []string{"status", "name", "priority", "key"}
	
	for _, param := range reserved {
		if !isReservedParam(param) {
			t.Errorf("Expected '%s' to be reserved", param)
		}
	}
	
	for _, param := range notReserved {
		if isReservedParam(param) {
			t.Errorf("Expected '%s' to not be reserved", param)
		}
	}
}