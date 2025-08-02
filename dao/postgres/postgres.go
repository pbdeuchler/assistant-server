package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Priority uint8

const (
	PriorityLow Priority = iota + 1
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

type Todo struct {
	UID            string     `json:"uid" db:"uid"`
	Title          string     `json:"title" db:"title"`
	Description    string     `json:"description" db:"description"`
	Data           string     `json:"data" db:"data"`
	Priority       Priority   `json:"priority" db:"priority"`
	DueDate        *time.Time `json:"due_date" db:"due_date"`
	RecursOn       string     `json:"recurs_on" db:"recurs_on"`
	MarkedComplete *time.Time `json:"marked_complete" db:"marked_complete"`
	ExternalURL    string     `json:"external_url" db:"external_url"`
	CreatedBy      string     `json:"created_by" db:"created_by"`
	CompletedBy    string     `json:"completed_by" db:"completed_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type Background struct {
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Preferences struct {
	Key       string    `json:"key" db:"key"`
	Specifier string    `json:"specifier" db:"specifier"`
	Data      string    `json:"data" db:"data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Notes struct {
	ID           string    `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	RelevantUser string    `json:"relevant_user" db:"relevant_user"`
	Content      string    `json:"content" db:"content"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ListOptions struct {
	Limit       int
	Offset      int
	SortBy      string
	SortDir     string
	WhereClause string
	WhereArgs   []any
}

type queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type DAO struct{ pool queryer }

func New(ctx context.Context, pool queryer) (*DAO, error) {
	return &DAO{pool}, nil
}

func (d *DAO) CreateTodo(ctx context.Context, t Todo) (Todo, error) {
	row := d.pool.QueryRow(ctx, insertTodo,
		t.UID, t.Title, t.Description, t.Data, t.Priority, t.DueDate,
		t.RecursOn, t.MarkedComplete, t.ExternalURL, t.CreatedBy, t.CompletedBy,
	)
	return scanTodo(row)
}

func (d *DAO) GetTodo(ctx context.Context, uid string) (Todo, error) {
	return scanTodo(d.pool.QueryRow(ctx, getTodo, uid))
}

func (d *DAO) ListTodos(ctx context.Context, options ListOptions) ([]Todo, error) {
	query := buildListQuery("todos", options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (d *DAO) UpdateTodo(ctx context.Context, uid string, t Todo) (Todo, error) {
	row := d.pool.QueryRow(ctx, updateTodo, uid, t.Title, t.Description, t.Data,
		t.Priority, t.DueDate, t.RecursOn, t.MarkedComplete, t.ExternalURL, t.CompletedBy,
	)
	return scanTodo(row)
}

func (d *DAO) DeleteTodo(ctx context.Context, uid string) error {
	_, err := d.pool.Exec(ctx, deleteTodo, uid)
	return err
}

func (d *DAO) CreateBackground(ctx context.Context, b Background) (Background, error) {
	row := d.pool.QueryRow(ctx, insertBackground, b.Key, b.Value)
	return scanBackground(row)
}

func (d *DAO) GetBackground(ctx context.Context, key string) (Background, error) {
	return scanBackground(d.pool.QueryRow(ctx, getBackground, key))
}

func (d *DAO) ListBackgrounds(ctx context.Context, options ListOptions) ([]Background, error) {
	query := buildListQuery("backgrounds", options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Background
	for rows.Next() {
		b, err := scanBackground(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (d *DAO) UpdateBackground(ctx context.Context, key string, b Background) (Background, error) {
	row := d.pool.QueryRow(ctx, updateBackground, key, b.Value)
	return scanBackground(row)
}

func (d *DAO) DeleteBackground(ctx context.Context, key string) error {
	_, err := d.pool.Exec(ctx, deleteBackground, key)
	return err
}

func (d *DAO) CreatePreferences(ctx context.Context, p Preferences) (Preferences, error) {
	row := d.pool.QueryRow(ctx, insertPreferences, p.Key, p.Specifier, p.Data)
	return scanPreferences(row)
}

func (d *DAO) GetPreferences(ctx context.Context, key, specifier string) (Preferences, error) {
	return scanPreferences(d.pool.QueryRow(ctx, getPreferences, key, specifier))
}

func (d *DAO) ListPreferences(ctx context.Context, options ListOptions) ([]Preferences, error) {
	query := buildListQuery("preferences", options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Preferences
	for rows.Next() {
		p, err := scanPreferences(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DAO) UpdatePreferences(ctx context.Context, key, specifier string, p Preferences) (Preferences, error) {
	row := d.pool.QueryRow(ctx, updatePreferences, key, specifier, p.Data)
	return scanPreferences(row)
}

func (d *DAO) DeletePreferences(ctx context.Context, key, specifier string) error {
	_, err := d.pool.Exec(ctx, deletePreferences, key, specifier)
	return err
}

func (d *DAO) CreateNotes(ctx context.Context, n Notes) (Notes, error) {
	row := d.pool.QueryRow(ctx, insertNotes, n.ID, n.Title, n.RelevantUser, n.Content)
	return scanNotes(row)
}

func (d *DAO) GetNotes(ctx context.Context, id string) (Notes, error) {
	return scanNotes(d.pool.QueryRow(ctx, getNotes, id))
}

func (d *DAO) ListNotes(ctx context.Context, options ListOptions) ([]Notes, error) {
	query := buildListQuery("notes", options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Notes
	for rows.Next() {
		n, err := scanNotes(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (d *DAO) UpdateNotes(ctx context.Context, id string, n Notes) (Notes, error) {
	row := d.pool.QueryRow(ctx, updateNotes, id, n.Title, n.RelevantUser, n.Content)
	return scanNotes(row)
}

func (d *DAO) DeleteNotes(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteNotes, id)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanTodo(s scannable) (Todo, error) {
	var t Todo
	err := s.Scan(&t.UID, &t.Title, &t.Description, &t.Data, &t.Priority,
		&t.DueDate, &t.RecursOn, &t.MarkedComplete, &t.ExternalURL,
		&t.CreatedBy, &t.CompletedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func scanBackground(s scannable) (Background, error) {
	var b Background
	err := s.Scan(&b.Key, &b.Value, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func scanPreferences(s scannable) (Preferences, error) {
	var p Preferences
	err := s.Scan(&p.Key, &p.Specifier, &p.Data, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func scanNotes(s scannable) (Notes, error) {
	var n Notes
	err := s.Scan(&n.ID, &n.Title, &n.RelevantUser, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func buildListQuery(tableName string, options ListOptions) string {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)

	if options.WhereClause != "" {
		query += " " + options.WhereClause
	}

	query += fmt.Sprintf(" ORDER BY %s %s", options.SortBy, options.SortDir)

	argOffset := len(options.WhereArgs)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argOffset+1, argOffset+2)

	return query
}
