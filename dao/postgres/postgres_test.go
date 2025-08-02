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
						if len(dest) >= 13 {
							*dest[0].(*string) = "test-uid"
							*dest[1].(*string) = "Test Title"
							*dest[2].(*string) = "Test Description"
							*dest[3].(*string) = "{}"
							*dest[4].(*Priority) = PriorityHigh
							*dest[11].(*time.Time) = now
							*dest[12].(*time.Time) = now
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
						*dest[3].(*time.Time) = now
						*dest[4].(*time.Time) = now
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
						*dest[0].(*string) = "test-id"
						*dest[1].(*string) = "Test Note"
						*dest[2].(*string) = "user123"
						*dest[3].(*string) = "This is the content of the note"
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
	
	note := Notes{
		ID:           "test-id",
		Title:        "Test Note",
		RelevantUser: "user123",
		Content:      "This is the content of the note",
	}
	
	result, err := dao.CreateNotes(context.Background(), note)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", result.ID)
	}
	if result.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got '%s'", result.Title)
	}
	if result.RelevantUser != "user123" {
		t.Errorf("Expected relevant_user 'user123', got '%s'", result.RelevantUser)
	}
}

func TestGetNotes(t *testing.T) {
	now := time.Now()
	mockPool := &mockQueryer{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if sql == getNotes && len(args) == 1 && args[0] == "test-id" {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*string) = "test-id"
						*dest[1].(*string) = "Test Note"
						*dest[2].(*string) = "user123"
						*dest[3].(*string) = "This is the content"
						*dest[4].(*time.Time) = now
						*dest[5].(*time.Time) = now
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
	if result.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got '%s'", result.Title)
	}
}