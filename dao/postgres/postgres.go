package postgres

import (
	"context"
	"encoding/json"
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
	UserID         string     `json:"user_id" db:"user_id"`
	HouseholdID    string     `json:"household_id" db:"household_id"`
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
	Tags      []string  `json:"tags" db:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Notes struct {
	ID          string    `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	UserID      string    `json:"user_id" db:"user_id"`
	HouseholdID string    `json:"household_id" db:"household_id"`
	Data        string    `json:"data" db:"data"`
	Tags        []string  `json:"tags" db:"tags"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Credentials struct {
	ID             string          `json:"id" db:"id"`
	UserID         string          `json:"user_id" db:"user_id"`
	CredentialType string          `json:"credential_type" db:"credential_type"`
	Value          json.RawMessage `json:"value" db:"value"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

type SlackUsers struct {
	SlackUserID string    `json:"slack_user_id" db:"slack_user_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Users struct {
	UID         string    `json:"uid" db:"uid"`
	Name        string    `json:"name" db:"name"`
	Email       string    `json:"email" db:"email"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Households struct {
	UID         string    `json:"uid" db:"uid"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Recipes struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	ExternalURL *string   `json:"external_url" db:"external_url"`
	Data        string    `json:"data" db:"data"`
	Genre       *string   `json:"genre" db:"genre"`
	GroceryList *string   `json:"grocery_list" db:"grocery_list"`
	PrepTime    *int      `json:"prep_time" db:"prep_time"`
	CookTime    *int      `json:"cook_time" db:"cook_time"`
	TotalTime   *int      `json:"total_time" db:"total_time"`
	Servings    *int      `json:"servings" db:"servings"`
	Difficulty  *string   `json:"difficulty" db:"difficulty"`
	Rating      *int      `json:"rating" db:"rating"`
	Tags        []string  `json:"tags" db:"tags"`
	UserID      string    `json:"user_id" db:"user_id"`
	HouseholdID string    `json:"household_id" db:"household_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
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
		t.RecursOn, t.MarkedComplete, t.ExternalURL, t.UserID, t.HouseholdID, t.CompletedBy,
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

type UpdateTodo struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	Data           *string `json:"data"`
	Priority       *int    `json:"priority"`
	DueDate        *string `json:"due_date"`
	RecursOn       *string `json:"recurs_on"`
	ExternalURL    *string `json:"external_url"`
	CompletedBy    *string `json:"completed_by"`
	MarkedComplete *string `json:"marked_complete"`
}

func (d *DAO) UpdateTodo(ctx context.Context, uid string, t UpdateTodo) (Todo, error) {
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
	row := d.pool.QueryRow(ctx, insertPreferences, p.Key, p.Specifier, p.Data, p.Tags)
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
	row := d.pool.QueryRow(ctx, updatePreferences, key, specifier, p.Data, p.Tags)
	return scanPreferences(row)
}

func (d *DAO) DeletePreferences(ctx context.Context, key, specifier string) error {
	_, err := d.pool.Exec(ctx, deletePreferences, key, specifier)
	return err
}

func (d *DAO) CreateNotes(ctx context.Context, n Notes) (Notes, error) {
	row := d.pool.QueryRow(ctx, insertNotes, n.ID, n.Key, n.UserID, n.HouseholdID, n.Data, n.Tags)
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
	row := d.pool.QueryRow(ctx, updateNotes, id, n.Key, n.UserID, n.HouseholdID, n.Data, n.Tags)
	return scanNotes(row)
}

func (d *DAO) DeleteNotes(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteNotes, id)
	return err
}

func (d *DAO) CreateCredentials(ctx context.Context, c Credentials) (Credentials, error) {
	row := d.pool.QueryRow(ctx, insertCredentials, c.ID, c.UserID, c.CredentialType, c.Value)
	return scanCredentials(row)
}

func (d *DAO) GetCredentials(ctx context.Context, id string) (Credentials, error) {
	return scanCredentials(d.pool.QueryRow(ctx, getCredentials, id))
}

func (d *DAO) GetCredentialsByUserAndType(ctx context.Context, userID, credentialType string) (Credentials, error) {
	return scanCredentials(d.pool.QueryRow(ctx, getCredentialsByUserAndType, userID, credentialType))
}

func (d *DAO) ListCredentials(ctx context.Context, options ListOptions) ([]Credentials, error) {
	query := buildListQuery("credentials", options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Credentials
	for rows.Next() {
		c, err := scanCredentials(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *DAO) UpdateCredentials(ctx context.Context, id string, c Credentials) (Credentials, error) {
	row := d.pool.QueryRow(ctx, updateCredentials, id, c.UserID, c.CredentialType, c.Value)
	return scanCredentials(row)
}

func (d *DAO) DeleteCredentials(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteCredentials, id)
	return err
}

func (d *DAO) GetSlackUser(ctx context.Context, slackUserID string) (SlackUsers, error) {
	return scanSlackUser(d.pool.QueryRow(ctx, getSlackUser, slackUserID))
}

func (d *DAO) GetUserBySlackUserID(ctx context.Context, slackUserID string) (Users, error) {
	return scanUser(d.pool.QueryRow(ctx, getUserBySlackUserID, slackUserID))
}

func (d *DAO) GetCredentialsByUserID(ctx context.Context, userID string) ([]Credentials, error) {
	rows, err := d.pool.Query(ctx, getCredentialsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Credentials
	for rows.Next() {
		c, err := scanCredentials(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *DAO) GetUser(ctx context.Context, uid string) (Users, error) {
	return scanUser(d.pool.QueryRow(ctx, getUser, uid))
}

func (d *DAO) GetHousehold(ctx context.Context, uid string) (Households, error) {
	return scanHousehold(d.pool.QueryRow(ctx, getHousehold, uid))
}

func (d *DAO) GetTodosByUserID(ctx context.Context, userID string) ([]Todo, error) {
	rows, err := d.pool.Query(ctx, getTodosByUserID, userID)
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

func (d *DAO) GetNotesByUserID(ctx context.Context, userID string) ([]Notes, error) {
	rows, err := d.pool.Query(ctx, getNotesByUserID, userID)
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

func (d *DAO) GetPreferencesByUserID(ctx context.Context, userID string) ([]Preferences, error) {
	rows, err := d.pool.Query(ctx, getPreferencesByUserID, userID)
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

func (d *DAO) CreateRecipes(ctx context.Context, r Recipes) (Recipes, error) {
	row := d.pool.QueryRow(ctx, insertRecipes, r.ID, r.Title, r.ExternalURL, r.Data, r.Genre, r.GroceryList, r.PrepTime, r.CookTime, r.TotalTime, r.Servings, r.Difficulty, r.Rating, r.Tags, r.UserID, r.HouseholdID)
	return scanRecipes(row)
}

func (d *DAO) GetRecipes(ctx context.Context, id string) (Recipes, error) {
	return scanRecipes(d.pool.QueryRow(ctx, getRecipes, id))
}

func (d *DAO) ListRecipes(ctx context.Context, options ListOptions) ([]Recipes, error) {
	query := buildListQuery("recipes", options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Recipes
	for rows.Next() {
		r, err := scanRecipes(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d *DAO) UpdateRecipes(ctx context.Context, id string, r Recipes) (Recipes, error) {
	row := d.pool.QueryRow(ctx, updateRecipes, id, r.Title, r.ExternalURL, r.Data, r.Genre, r.GroceryList, r.PrepTime, r.CookTime, r.TotalTime, r.Servings, r.Difficulty, r.Rating, r.Tags, r.UserID, r.HouseholdID)
	return scanRecipes(row)
}

func (d *DAO) DeleteRecipes(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteRecipes, id)
	return err
}

func (d *DAO) GetRecipesByUserID(ctx context.Context, userID string) ([]Recipes, error) {
	rows, err := d.pool.Query(ctx, getRecipesByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Recipes
	for rows.Next() {
		r, err := scanRecipes(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanTodo(s scannable) (Todo, error) {
	var t Todo
	err := s.Scan(&t.UID, &t.Title, &t.Description, &t.Data, &t.Priority,
		&t.DueDate, &t.RecursOn, &t.MarkedComplete, &t.ExternalURL,
		&t.UserID, &t.HouseholdID, &t.CompletedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func scanBackground(s scannable) (Background, error) {
	var b Background
	err := s.Scan(&b.Key, &b.Value, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func scanPreferences(s scannable) (Preferences, error) {
	var p Preferences
	err := s.Scan(&p.Key, &p.Specifier, &p.Data, &p.Tags, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func scanNotes(s scannable) (Notes, error) {
	var n Notes
	err := s.Scan(&n.ID, &n.Key, &n.UserID, &n.HouseholdID, &n.Data, &n.Tags, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func scanCredentials(s scannable) (Credentials, error) {
	var c Credentials
	err := s.Scan(&c.ID, &c.UserID, &c.CredentialType, &c.Value, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func scanSlackUser(s scannable) (SlackUsers, error) {
	var su SlackUsers
	err := s.Scan(&su.SlackUserID, &su.UserID, &su.CreatedAt, &su.UpdatedAt)
	return su, err
}

func scanUser(s scannable) (Users, error) {
	var u Users
	err := s.Scan(&u.UID, &u.Name, &u.Email, &u.Description, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func scanHousehold(s scannable) (Households, error) {
	var h Households
	err := s.Scan(&h.UID, &h.Name, &h.Description, &h.CreatedAt, &h.UpdatedAt)
	return h, err
}

func scanRecipes(s scannable) (Recipes, error) {
	var r Recipes
	err := s.Scan(&r.ID, &r.Title, &r.ExternalURL, &r.Data, &r.Genre, &r.GroceryList, &r.PrepTime, &r.CookTime, &r.TotalTime, &r.Servings, &r.Difficulty, &r.Rating, &r.Tags, &r.UserID, &r.HouseholdID, &r.CreatedAt, &r.UpdatedAt)
	return r, err
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
