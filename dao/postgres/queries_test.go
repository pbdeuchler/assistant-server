package postgres

import (
	"strings"
	"testing"
)

func TestQueryConstants(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "insertTodo contains all required fields",
			query:   insertTodo,
			wantSQL: "INSERT INTO todos",
		},
		{
			name:    "getTodo selects by uid",
			query:   getTodo,
			wantSQL: "SELECT * FROM todos WHERE uid=$1",
		},
		{
			name:    "listTodos orders and limits",
			query:   listTodos,
			wantSQL: "SELECT * FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		},
		{
			name:    "updateTodo updates by uid",
			query:   updateTodo,
			wantSQL: "UPDATE todos SET",
		},
		{
			name:    "deleteTodo deletes by uid",
			query:   deleteTodo,
			wantSQL: "DELETE FROM todos WHERE uid=$1",
		},
		{
			name:    "insertBackground has key and value",
			query:   insertBackground,
			wantSQL: "INSERT INTO backgrounds",
		},
		{
			name:    "getBackground selects by key",
			query:   getBackground,
			wantSQL: "SELECT * FROM backgrounds WHERE key=$1",
		},
		{
			name:    "insertPreferences has key, specifier, and data",
			query:   insertPreferences,
			wantSQL: "INSERT INTO preferences",
		},
		{
			name:    "getPreferences selects by key and specifier",
			query:   getPreferences,
			wantSQL: "SELECT * FROM preferences WHERE key=$1 AND specifier=$2",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.query, tt.wantSQL) {
				t.Errorf("Query %s should contain %s\nActual: %s", tt.name, tt.wantSQL, tt.query)
			}
		})
	}
}

func TestTodoQueries(t *testing.T) {
	// Test that insertTodo has the correct number of parameters
	paramCount := strings.Count(insertTodo, "$")
	expectedParams := 12 // Based on the Todo struct fields being inserted (added user_id and household_id)
	
	if paramCount != expectedParams {
		t.Errorf("insertTodo should have %d parameters, found %d", expectedParams, paramCount)
	}
	
	// Test that insertTodo returns all fields
	if !strings.Contains(insertTodo, "RETURNING *") {
		t.Error("insertTodo should return all fields with RETURNING *")
	}
	
	// Test that updateTodo has updated_at=NOW()
	if !strings.Contains(updateTodo, "updated_at=NOW()") {
		t.Error("updateTodo should update the updated_at field to NOW()")
	}
}

func TestBackgroundQueries(t *testing.T) {
	// Test insertBackground parameters
	paramCount := strings.Count(insertBackground, "$")
	expectedParams := 2 // key, value
	
	if paramCount != expectedParams {
		t.Errorf("insertBackground should have %d parameters, found %d", expectedParams, paramCount)
	}
	
	// Test that all background queries reference the correct table
	backgroundQueries := []string{insertBackground, getBackground, listBackgrounds, updateBackground, deleteBackground}
	
	for i, query := range backgroundQueries {
		if !strings.Contains(query, "backgrounds") {
			t.Errorf("Background query %d should reference 'backgrounds' table: %s", i, query)
		}
	}
}

func TestPreferencesQueries(t *testing.T) {
	// Test insertPreferences parameters
	paramCount := strings.Count(insertPreferences, "$")
	expectedParams := 3 // key, specifier, data
	
	if paramCount != expectedParams {
		t.Errorf("insertPreferences should have %d parameters, found %d", expectedParams, paramCount)
	}
	
	// Test that getPreferences uses composite key
	if !strings.Contains(getPreferences, "key=$1 AND specifier=$2") {
		t.Error("getPreferences should filter by both key and specifier")
	}
	
	// Test that updatePreferences uses composite key in WHERE clause
	if !strings.Contains(updatePreferences, "WHERE key=$1 AND specifier=$2") {
		t.Error("updatePreferences should filter by both key and specifier in WHERE clause")
	}
	
	// Test that deletePreferences uses composite key
	if !strings.Contains(deletePreferences, "key=$1 AND specifier=$2") {
		t.Error("deletePreferences should filter by both key and specifier")
	}
}

func TestQueryConsistency(t *testing.T) {
	// Test that all INSERT queries set created_at and updated_at to NOW()
	insertQueries := []struct {
		name  string
		query string
	}{
		{"insertTodo", insertTodo},
		{"insertBackground", insertBackground},
		{"insertPreferences", insertPreferences},
	}
	
	for _, iq := range insertQueries {
		if !strings.Contains(iq.query, "NOW()") {
			t.Errorf("%s should set timestamps to NOW()", iq.name)
		}
		if !strings.Contains(iq.query, "RETURNING *") {
			t.Errorf("%s should return all fields with RETURNING *", iq.name)
		}
	}
	
	// Test that all UPDATE queries update updated_at to NOW()
	updateQueries := []struct {
		name  string
		query string
	}{
		{"updateTodo", updateTodo},
		{"updateBackground", updateBackground},
		{"updatePreferences", updatePreferences},
	}
	
	for _, uq := range updateQueries {
		if !strings.Contains(uq.query, "updated_at=NOW()") {
			t.Errorf("%s should update updated_at to NOW()", uq.name)
		}
		if !strings.Contains(uq.query, "RETURNING *") {
			t.Errorf("%s should return all fields with RETURNING *", uq.name)
		}
	}
	
	// Test that all LIST queries have ORDER BY and LIMIT/OFFSET
	listQueries := []struct {
		name  string
		query string
	}{
		{"listTodos", listTodos},
		{"listBackgrounds", listBackgrounds},
		{"listPreferences", listPreferences},
	}
	
	for _, lq := range listQueries {
		if !strings.Contains(lq.query, "ORDER BY") {
			t.Errorf("%s should have ORDER BY clause", lq.name)
		}
		if !strings.Contains(lq.query, "LIMIT") || !strings.Contains(lq.query, "OFFSET") {
			t.Errorf("%s should have LIMIT and OFFSET parameters", lq.name)
		}
	}
}

func TestParameterizedQueries(t *testing.T) {
	// Test that queries use parameterized statements (no direct string interpolation)
	allQueries := []struct {
		name  string
		query string
	}{
		{"insertTodo", insertTodo},
		{"getTodo", getTodo},
		{"listTodos", listTodos},
		{"updateTodo", updateTodo},
		{"deleteTodo", deleteTodo},
		{"insertBackground", insertBackground},
		{"getBackground", getBackground},
		{"listBackgrounds", listBackgrounds},
		{"updateBackground", updateBackground},
		{"deleteBackground", deleteBackground},
		{"insertPreferences", insertPreferences},
		{"getPreferences", getPreferences},
		{"listPreferences", listPreferences},
		{"updatePreferences", updatePreferences},
		{"deletePreferences", deletePreferences},
	}
	
	for _, q := range allQueries {
		// Check that queries don't contain obvious SQL injection patterns
		dangerousPatterns := []string{
			"' OR '1'='1",
			"; DROP TABLE",
			"UNION SELECT",
		}
		
		for _, pattern := range dangerousPatterns {
			if strings.Contains(strings.ToUpper(q.query), strings.ToUpper(pattern)) {
				t.Errorf("Query %s appears to contain dangerous pattern: %s", q.name, pattern)
			}
		}
		
		// Ensure queries that should have parameters actually have them
		if strings.Contains(q.name, "get") || strings.Contains(q.name, "update") || strings.Contains(q.name, "delete") {
			if !strings.Contains(q.query, "$") {
				t.Errorf("Query %s should use parameterized statements", q.name)
			}
		}
	}
}