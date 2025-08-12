package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Mock queryer for testing
type mockQueryer struct {
	queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	execFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockQueryer) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, errors.New("query not implemented")
}

func (m *mockQueryer) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &mockRow{err: errors.New("queryRow not implemented")}
}

func (m *mockQueryer) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, errors.New("exec not implemented")
}

// Mock row for testing
type mockRow struct {
	scanFunc func(dest ...any) error
	err      error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	return m.err
}

// Simplified mock for basic testing - avoid complex pgx.Rows interface

func TestNew(t *testing.T) {
	mockPool := &mockQueryer{}
	dao, err := New(context.Background(), mockPool)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dao == nil {
		t.Error("Expected DAO instance, got nil")
	}
	if dao.pool != mockPool {
		t.Error("Expected DAO to use provided pool")
	}
}

func TestCreateTodo(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == insertTodo {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						// Simulate scanning a complete Todo
						if len(dest) >= 14 {
							*dest[0].(*string) = "test-uid"         // UID
							*dest[1].(*string) = "Test Title"       // Title  
							*dest[2].(*string) = "Test Description" // Description
							*dest[3].(*string) = "{}"               // Data
							*dest[4].(*Priority) = PriorityHigh     // Priority
							// dest[5] is DueDate (*time.Time) - leave nil
							*dest[6].(*string) = ""                 // RecursOn
							// dest[7] is MarkedComplete (*time.Time) - leave nil
							*dest[8].(*string) = ""                 // ExternalURL
							*dest[9].(*string) = "user-123"        // UserID
							*dest[10].(*string) = "household-456"  // HouseholdID
							*dest[11].(*string) = ""               // CompletedBy
							*dest[12].(*time.Time) = now           // CreatedAt
							*dest[13].(*time.Time) = now           // UpdatedAt
						}
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("unexpected query")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	todo := Todo{
		UID:         "test-uid",
		Title:       "Test Title",
		Description: "Test Description",
		Priority:    PriorityHigh,
	}
	
	result, err := dao.CreateTodo(context.Background(), todo)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.UID != "test-uid" {
		t.Errorf("Expected UID 'test-uid', got '%s'", result.UID)
	}
	if result.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", result.Title)
	}
}

func TestGetTodo(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == getTodo && len(args) == 1 && args[0] == "test-uid" {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-uid"
						*dest[1].(*string) = "Test Title"
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("todo not found")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	result, err := dao.GetTodo(context.Background(), "test-uid")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.UID != "test-uid" {
		t.Errorf("Expected UID 'test-uid', got '%s'", result.UID)
	}
}

func TestListTodosQueryBuilding(t *testing.T) {
	// Test the query building functionality without complex mocking
	options := ListOptions{
		Limit:   10,
		Offset:  0,
		SortBy:  "created_at",
		SortDir: "DESC",
	}
	
	query := buildListQuery("todos", options)
	expectedQuery := "SELECT * FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	
	if query != expectedQuery {
		t.Errorf("Expected query: %s\nGot: %s", expectedQuery, query)
	}
}

func TestCreateBackground(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == insertBackground {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-key"
						*dest[1].(*string) = "test-value"
						*dest[2].(*time.Time) = now
						*dest[3].(*time.Time) = now
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("unexpected query")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	bg := Background{
		Key:   "test-key",
		Value: "test-value",
	}
	
	result, err := dao.CreateBackground(context.Background(), bg)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", result.Key)
	}
	if result.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", result.Value)
	}
}

func TestCreatePreferences(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == insertPreferences {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-key"
						*dest[1].(*string) = "test-specifier"
						*dest[2].(*string) = "{\"theme\": \"dark\"}"
						*dest[3].(*[]string) = []string{"theme", "ui"}
						*dest[4].(*time.Time) = now
						*dest[5].(*time.Time) = now
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("unexpected query")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	pref := Preferences{
		Key:       "test-key",
		Specifier: "test-specifier",
		Data:      "{\"theme\": \"dark\"}",
		Tags:      []string{"theme", "ui"},
	}
	
	result, err := dao.CreatePreferences(context.Background(), pref)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", result.Key)
	}
	if result.Specifier != "test-specifier" {
		t.Errorf("Expected specifier 'test-specifier', got '%s'", result.Specifier)
	}
}

func TestBuildListQuery(t *testing.T) {
	tests := []struct {
		name        string
		tableName   string
		options     ListOptions
		expectedSQL string
	}{
		{
			name:      "basic query",
			tableName: "todos",
			options: ListOptions{
				Limit:   10,
				Offset:  0,
				SortBy:  "created_at",
				SortDir: "DESC",
			},
			expectedSQL: "SELECT * FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		},
		{
			name:      "with where clause",
			tableName: "todos",
			options: ListOptions{
				Limit:       10,
				Offset:      0,
				SortBy:      "created_at",
				SortDir:     "ASC",
				WhereClause: "WHERE priority = $1",
				WhereArgs:   []any{"high"},
			},
			expectedSQL: "SELECT * FROM todos WHERE priority = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3",
		},
		{
			name:      "backgrounds table",
			tableName: "backgrounds",
			options: ListOptions{
				Limit:   50,
				Offset:  25,
				SortBy:  "key",
				SortDir: "ASC",
			},
			expectedSQL: "SELECT * FROM backgrounds ORDER BY key ASC LIMIT $1 OFFSET $2",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildListQuery(test.tableName, test.options)
			if result != test.expectedSQL {
				t.Errorf("Expected SQL: %s\nGot: %s", test.expectedSQL, result)
			}
		})
	}
}

func TestBuildListQueryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		tableName   string
		options     ListOptions
		expectedSQL string
	}{
		{
			name:      "empty table name",
			tableName: "",
			options: ListOptions{
				Limit:   10,
				Offset:  0,
				SortBy:  "id",
				SortDir: "ASC",
			},
			expectedSQL: "SELECT * FROM  ORDER BY id ASC LIMIT $1 OFFSET $2",
		},
		{
			name:      "empty sort field defaults to provided",
			tableName: "test_table",
			options: ListOptions{
				Limit:   5,
				Offset:  10,
				SortBy:  "",
				SortDir: "DESC",
			},
			expectedSQL: "SELECT * FROM test_table ORDER BY  DESC LIMIT $1 OFFSET $2",
		},
		{
			name:      "complex where clause with multiple args",
			tableName: "todos",
			options: ListOptions{
				Limit:       20,
				Offset:      5,
				SortBy:      "priority",
				SortDir:     "ASC",
				WhereClause: "WHERE user_id = $1 AND priority > $2 AND created_at > $3",
				WhereArgs:   []any{"user-123", 1, "2024-01-01"},
			},
			expectedSQL: "SELECT * FROM todos WHERE user_id = $1 AND priority > $2 AND created_at > $3 ORDER BY priority ASC LIMIT $4 OFFSET $5",
		},
		{
			name:      "zero limit and offset",
			tableName: "notes",
			options: ListOptions{
				Limit:   0,
				Offset:  0,
				SortBy:  "created_at",
				SortDir: "DESC",
			},
			expectedSQL: "SELECT * FROM notes ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		},
		{
			name:      "large offset values",
			tableName: "preferences",
			options: ListOptions{
				Limit:   100,
				Offset:  1000,
				SortBy:  "key",
				SortDir: "ASC",
			},
			expectedSQL: "SELECT * FROM preferences ORDER BY key ASC LIMIT $1 OFFSET $2",
		},
		{
			name:      "no where args but has where clause",
			tableName: "backgrounds",
			options: ListOptions{
				Limit:       10,
				Offset:      0,
				SortBy:      "updated_at",
				SortDir:     "DESC",
				WhereClause: "WHERE value IS NOT NULL",
				WhereArgs:   []any{},
			},
			expectedSQL: "SELECT * FROM backgrounds WHERE value IS NOT NULL ORDER BY updated_at DESC LIMIT $1 OFFSET $2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildListQuery(test.tableName, test.options)
			if result != test.expectedSQL {
				t.Errorf("Expected SQL: %s\nGot: %s", test.expectedSQL, result)
			}
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	mockPool := &mockQueryer{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if sql == deleteTodo && len(args) == 1 && args[0] == "test-uid" {
				return pgconn.CommandTag{}, nil
			}
			return pgconn.CommandTag{}, errors.New("todo not found")
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	err := dao.DeleteTodo(context.Background(), "test-uid")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUpdateBackground(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == updateBackground {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-key"
						*dest[1].(*string) = "updated-value"
						*dest[2].(*time.Time) = now
						*dest[3].(*time.Time) = now
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("unexpected query")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	bg := Background{
		Key:   "test-key",
		Value: "updated-value",
	}
	
	result, err := dao.UpdateBackground(context.Background(), "test-key", bg)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Value != "updated-value" {
		t.Errorf("Expected value 'updated-value', got '%s'", result.Value)
	}
}

func TestCreateNotes(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == insertNotes {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-id"                      // ID
						*dest[1].(*string) = "Test Note"                    // Key
						*dest[2].(*string) = "user123"                      // UserID
						*dest[3].(*string) = "household456"                 // HouseholdID
						*dest[4].(*string) = "This is the content of the note" // Data
						*dest[5].(*[]string) = []string{"tag1", "tag2"}    // Tags
						*dest[6].(*time.Time) = now                         // CreatedAt
						*dest[7].(*time.Time) = now                         // UpdatedAt
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("unexpected query")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	note := Notes{
		ID:          "test-id",
		Key:         "Test Note",
		UserID:      "user123",
		HouseholdID: "household456",
		Data:        "This is the content of the note",
		Tags:        []string{"tag1", "tag2"},
	}
	
	result, err := dao.CreateNotes(context.Background(), note)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", result.ID)
	}
	if result.Key != "Test Note" {
		t.Errorf("Expected key 'Test Note', got '%s'", result.Key)
	}
	if result.UserID != "user123" {
		t.Errorf("Expected user_id 'user123', got '%s'", result.UserID)
	}
	if result.HouseholdID != "household456" {
		t.Errorf("Expected household_id 'household456', got '%s'", result.HouseholdID)
	}
}

func TestGetNotes(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == getNotes && len(args) == 1 && args[0] == "test-id" {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-id"           // ID
						*dest[1].(*string) = "Test Note"         // Key
						*dest[2].(*string) = "user123"           // UserID
						*dest[3].(*string) = "household456"      // HouseholdID
						*dest[4].(*string) = "This is the content" // Data
						*dest[5].(*[]string) = []string{"tag1"}   // Tags
						*dest[6].(*time.Time) = now              // CreatedAt
						*dest[7].(*time.Time) = now              // UpdatedAt
						return nil
					},
				}
			}
			return &mockRow{err: errors.New("note not found")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	result, err := dao.GetNotes(context.Background(), "test-id")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", result.ID)
	}
	if result.Key != "Test Note" {
		t.Errorf("Expected key 'Test Note', got '%s'", result.Key)
	}
}

func TestGetNotesError(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{err: errors.New("database error")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	_, err := dao.GetNotes(context.Background(), "nonexistent")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateTodoError(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{err: errors.New("insert failed")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	todo := Todo{
		UID:         "test-uid",
		Title:       "Test Title",
		Description: "Test Description",
		Priority:    PriorityHigh,
	}
	
	_, err := dao.CreateTodo(context.Background(), todo)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateBackgroundError(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{err: errors.New("insert failed")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	bg := Background{
		Key:   "test-key",
		Value: "test-value",
	}
	
	_, err := dao.CreateBackground(context.Background(), bg)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreatePreferencesError(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{err: errors.New("insert failed")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	pref := Preferences{
		Key:       "test-key",
		Specifier: "test-specifier",
		Data:      "{\"theme\": \"dark\"}",
		Tags:      []string{"theme", "ui"},
	}
	
	_, err := dao.CreatePreferences(context.Background(), pref)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateNotesError(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{err: errors.New("insert failed")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	note := Notes{
		ID:          "test-id",
		Key:         "Test Note",
		UserID:      "user123",
		HouseholdID: "household456",
		Data:        "This is the content of the note",
		Tags:        []string{"tag1", "tag2"},
	}
	
	_, err := dao.CreateNotes(context.Background(), note)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDeleteTodoError(t *testing.T) {
	mockPool := &mockQueryer{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, errors.New("delete failed")
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	err := dao.DeleteTodo(context.Background(), "test-uid")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateBackgroundError(t *testing.T) {
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{err: errors.New("update failed")}
		},
	}
	
	dao, _ := New(context.Background(), mockPool)
	
	bg := Background{
		Key:   "test-key",
		Value: "updated-value",
	}
	
	_, err := dao.UpdateBackground(context.Background(), "test-key", bg)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestScanTodo(t *testing.T) {
	now := time.Now()
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "test-uid"         // UID
			*dest[1].(*string) = "Test Title"       // Title  
			*dest[2].(*string) = "Test Description" // Description
			*dest[3].(*string) = "{}"               // Data
			*dest[4].(*Priority) = PriorityHigh     // Priority
			// dest[5] is DueDate (*time.Time) - leave nil
			*dest[6].(*string) = ""                 // RecursOn
			// dest[7] is MarkedComplete (*time.Time) - leave nil
			*dest[8].(*string) = ""                 // ExternalURL
			*dest[9].(*string) = "user-123"        // UserID
			*dest[10].(*string) = "household-456"  // HouseholdID
			*dest[11].(*string) = ""               // CompletedBy
			*dest[12].(*time.Time) = now           // CreatedAt
			*dest[13].(*time.Time) = now           // UpdatedAt
			return nil
		},
	}

	todo, err := scanTodo(mockRow)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if todo.UID != "test-uid" {
		t.Errorf("Expected UID 'test-uid', got '%s'", todo.UID)
	}
	if todo.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", todo.Title)
	}
	if todo.Priority != PriorityHigh {
		t.Errorf("Expected priority %d, got %d", PriorityHigh, todo.Priority)
	}
}

func TestScanTodoError(t *testing.T) {
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanTodo(mockRow)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestScanBackground(t *testing.T) {
	now := time.Now()
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "test-key"
			*dest[1].(*string) = "test-value"
			*dest[2].(*time.Time) = now
			*dest[3].(*time.Time) = now
			return nil
		},
	}

	bg, err := scanBackground(mockRow)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if bg.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", bg.Key)
	}
	if bg.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", bg.Value)
	}
}

func TestScanBackgroundError(t *testing.T) {
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanBackground(mockRow)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestScanPreferences(t *testing.T) {
	now := time.Now()
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "test-key"
			*dest[1].(*string) = "test-specifier"
			*dest[2].(*string) = "{\"theme\": \"dark\"}"
			*dest[3].(*[]string) = []string{"theme", "ui"}
			*dest[4].(*time.Time) = now
			*dest[5].(*time.Time) = now
			return nil
		},
	}

	pref, err := scanPreferences(mockRow)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if pref.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", pref.Key)
	}
	if len(pref.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(pref.Tags))
	}
}

func TestScanPreferencesError(t *testing.T) {
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanPreferences(mockRow)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestScanNotes(t *testing.T) {
	now := time.Now()
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "test-id"
			*dest[1].(*string) = "Test Note"
			*dest[2].(*string) = "user123"
			*dest[3].(*string) = "household456"
			*dest[4].(*string) = "This is the content"
			*dest[5].(*[]string) = []string{"tag1", "tag2"}
			*dest[6].(*time.Time) = now
			*dest[7].(*time.Time) = now
			return nil
		},
	}

	note, err := scanNotes(mockRow)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if note.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", note.ID)
	}
	if note.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", note.UserID)
	}
	if len(note.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(note.Tags))
	}
}

func TestScanNotesError(t *testing.T) {
	mockRow := &mockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanNotes(mockRow)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}