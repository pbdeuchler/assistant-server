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
	UserUID        *string    `json:"user_uid" db:"user_uid"`
	HouseholdUID   *string    `json:"household_uid" db:"household_uid"`
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
	ID           string    `json:"id" db:"id"`
	Key          string    `json:"key" db:"key"`
	UserUID      *string   `json:"user_uid" db:"user_uid"`
	HouseholdUID *string   `json:"household_uid" db:"household_uid"`
	Data         string    `json:"data" db:"data"`
	Tags         []string  `json:"tags" db:"tags"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Credentials struct {
	ID             string          `json:"id" db:"id"`
	UserUID        string          `json:"user_uid" db:"user_uid"`
	CredentialType string          `json:"credential_type" db:"credential_type"`
	Value          json.RawMessage `json:"value" db:"value"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

type SlackUsers struct {
	SlackUserUID string    `json:"slack_user_uid" db:"slack_user_uid"`
	UserUID      string    `json:"user_uid" db:"user_uid"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Users struct {
	UID          string    `json:"uid" db:"uid"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	Description  string    `json:"description" db:"description"`
	HouseholdUID *string   `json:"household_uid" db:"household_uid"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateUser struct {
	Name         *string `json:"name"`
	Email        *string `json:"email"`
	Description  *string `json:"description"`
	HouseholdUID *string `json:"household_uid"`
}

type UpdateHousehold struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type Households struct {
	UID         string    `json:"uid" db:"uid"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Recipes struct {
	ID           string    `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	ExternalURL  *string   `json:"external_url" db:"external_url"`
	Data         string    `json:"data" db:"data"`
	Genre        *string   `json:"genre" db:"genre"`
	GroceryList  *string   `json:"grocery_list" db:"grocery_list"`
	PrepTime     *int      `json:"prep_time" db:"prep_time"`
	CookTime     *int      `json:"cook_time" db:"cook_time"`
	TotalTime    *int      `json:"total_time" db:"total_time"`
	Servings     *int      `json:"servings" db:"servings"`
	Difficulty   *string   `json:"difficulty" db:"difficulty"`
	Rating       *int      `json:"rating" db:"rating"`
	Tags         []string  `json:"tags" db:"tags"`
	UserUID      *string   `json:"user_uid" db:"user_uid"`
	HouseholdUID *string   `json:"household_uid" db:"household_uid"`
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

func handleUIDRefs(userUID, householdUID *string) (*string, *string) {
	var userUIDPtr *string
	if userUID != nil && *userUID != "" {
		userUIDPtr = userUID
	}

	var householdUIDPtr *string
	if householdUID != nil && *householdUID != "" {
		householdUIDPtr = householdUID
	}

	return userUIDPtr, householdUIDPtr
}

func (d *DAO) CreateTodo(ctx context.Context, t Todo) (Todo, error) {
	userUID, householdUID := handleUIDRefs(t.UserUID, t.HouseholdUID)

	row := d.pool.QueryRow(ctx, insertTodo,
		t.Title, t.Description, t.Data, t.Priority, t.DueDate,
		t.RecursOn, t.MarkedComplete, t.ExternalURL, userUID, householdUID, t.CompletedBy,
	)
	return scanTodo(row)
}

func (d *DAO) GetTodo(ctx context.Context, uid string) (Todo, error) {
	return scanTodo(d.pool.QueryRow(ctx, getTodo, uid))
}

func (d *DAO) ListTodos(ctx context.Context, options ListOptions) ([]Todo, error) {
	todoColumns := "uid, title, description, data, priority, due_date, recurs_on, marked_complete, external_url, user_uid, household_uid, completed_by, created_at, updated_at"
	query := buildListQuery("todos", todoColumns, options)
	args := append(options.WhereArgs, options.Limit, options.Offset)
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Todo{}
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
	Title          *string    `json:"title"`
	Description    *string    `json:"description"`
	Data           *string    `json:"data"`
	Priority       *int       `json:"priority"`
	DueDate        *time.Time `json:"due_date"`
	RecursOn       *string    `json:"recurs_on"`
	ExternalURL    *string    `json:"external_url"`
	CompletedBy    *string    `json:"completed_by"`
	MarkedComplete *time.Time `json:"marked_complete"`
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
	backgroundColumns := "*"
	query := buildListQuery("backgrounds", backgroundColumns, options)
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
	preferencesColumns := "key, specifier, data, created_at, updated_at, tags"
	query := buildListQuery("preferences", preferencesColumns, options)
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
	userUID, householdUID := handleUIDRefs(n.UserUID, n.HouseholdUID)
	row := d.pool.QueryRow(ctx, insertNotes, n.Key, userUID, householdUID, n.Data, n.Tags)
	return scanNotes(row)
}

func (d *DAO) GetNotes(ctx context.Context, id string) (Notes, error) {
	return scanNotes(d.pool.QueryRow(ctx, getNotes, id))
}

func (d *DAO) ListNotes(ctx context.Context, options ListOptions) ([]Notes, error) {
	notesColumns := "id, key, data, created_at, updated_at, user_uid, household_uid, tags"
	query := buildListQuery("notes", notesColumns, options)
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
	row := d.pool.QueryRow(ctx, updateNotes, id, n.Key, n.UserUID, n.HouseholdUID, n.Data, n.Tags)
	return scanNotes(row)
}

func (d *DAO) DeleteNotes(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteNotes, id)
	return err
}

func (d *DAO) CreateCredentials(ctx context.Context, c Credentials) (Credentials, error) {
	row := d.pool.QueryRow(ctx, insertCredentials, c.UserUID, c.CredentialType, c.Value)
	return scanCredentials(row)
}

func (d *DAO) GetCredentials(ctx context.Context, id string) (Credentials, error) {
	return scanCredentials(d.pool.QueryRow(ctx, getCredentials, id))
}

func (d *DAO) GetCredentialsByUserAndType(ctx context.Context, userID, credentialType string) (Credentials, error) {
	return scanCredentials(d.pool.QueryRow(ctx, getCredentialsByUserAndType, userID, credentialType))
}

func (d *DAO) ListCredentials(ctx context.Context, options ListOptions) ([]Credentials, error) {
	credentialsColumns := "*"
	query := buildListQuery("credentials", credentialsColumns, options)
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
	row := d.pool.QueryRow(ctx, updateCredentials, id, c.UserUID, c.CredentialType, c.Value)
	return scanCredentials(row)
}

func (d *DAO) DeleteCredentials(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteCredentials, id)
	return err
}

func (d *DAO) GetSlackUser(ctx context.Context, slackUserUID string) (SlackUsers, error) {
	return scanSlackUser(d.pool.QueryRow(ctx, getSlackUser, slackUserUID))
}

func (d *DAO) GetUserBySlackUserUID(ctx context.Context, slackUserUID string) (Users, error) {
	return scanUser(d.pool.QueryRow(ctx, getUserBySlackUserUID, slackUserUID))
}

func (d *DAO) GetCredentialsByUserUID(ctx context.Context, userUID string) ([]Credentials, error) {
	rows, err := d.pool.Query(ctx, getCredentialsByUserUID, userUID)
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

func (d *DAO) CreateUser(ctx context.Context, u Users) (Users, error) {
	row := d.pool.QueryRow(ctx, insertUser, u.Name, u.Email, u.Description, u.HouseholdUID)
	return scanUser(row)
}

func (d *DAO) UpdateUser(ctx context.Context, uid string, u UpdateUser) (Users, error) {
	row := d.pool.QueryRow(ctx, updateUser, uid, u.Name, u.Email, u.Description, u.HouseholdUID)
	return scanUser(row)
}

func (d *DAO) GetUser(ctx context.Context, uid string) (Users, error) {
	return scanUser(d.pool.QueryRow(ctx, getUser, uid))
}

func (d *DAO) GetHousehold(ctx context.Context, uid string) (Households, error) {
	return scanHousehold(d.pool.QueryRow(ctx, getHousehold, uid))
}

func (d *DAO) UpdateHousehold(ctx context.Context, uid string, h UpdateHousehold) (Households, error) {
	row := d.pool.QueryRow(ctx, updateHousehold, uid, h.Name, h.Description)
	return scanHousehold(row)
}

func (d *DAO) GetTodosByUserUID(ctx context.Context, userUID string) ([]Todo, error) {
	rows, err := d.pool.Query(ctx, getTodosByUserUID, userUID)
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

func (d *DAO) GetNotesByUserUID(ctx context.Context, userUID string) ([]Notes, error) {
	rows, err := d.pool.Query(ctx, getNotesByUserUID, userUID)
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

func (d *DAO) GetPreferencesByUserUID(ctx context.Context, userUID string) ([]Preferences, error) {
	rows, err := d.pool.Query(ctx, getPreferencesByUserUID, userUID)
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
	userUID, householdUID := handleUIDRefs(r.UserUID, r.HouseholdUID)
	row := d.pool.QueryRow(ctx, insertRecipes, r.Title, r.ExternalURL, r.Data, r.Genre, r.GroceryList, r.PrepTime, r.CookTime, r.TotalTime, r.Servings, r.Difficulty, r.Rating, r.Tags, userUID, householdUID)
	return scanRecipes(row)
}

func (d *DAO) GetRecipes(ctx context.Context, id string) (Recipes, error) {
	return scanRecipes(d.pool.QueryRow(ctx, getRecipes, id))
}

func (d *DAO) ListRecipes(ctx context.Context, options ListOptions) ([]Recipes, error) {
	recipesColumns := "id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at"
	query := buildListQuery("recipes", recipesColumns, options)
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
	row := d.pool.QueryRow(ctx, updateRecipes, id, r.Title, r.ExternalURL, r.Data, r.Genre, r.GroceryList, r.PrepTime, r.CookTime, r.TotalTime, r.Servings, r.Difficulty, r.Rating, r.Tags, r.UserUID, r.HouseholdUID)
	return scanRecipes(row)
}

func (d *DAO) DeleteRecipes(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, deleteRecipes, id)
	return err
}

func (d *DAO) GetRecipesByUserUID(ctx context.Context, userUID string) ([]Recipes, error) {
	rows, err := d.pool.Query(ctx, getRecipesByUserUID, userUID)
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
		&t.UserUID, &t.HouseholdUID, &t.CompletedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func scanBackground(s scannable) (Background, error) {
	var b Background
	err := s.Scan(&b.Key, &b.Value, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func scanPreferences(s scannable) (Preferences, error) {
	var p Preferences
	err := s.Scan(&p.Key, &p.Specifier, &p.Data, &p.CreatedAt, &p.UpdatedAt, &p.Tags)
	return p, err
}

func scanNotes(s scannable) (Notes, error) {
	var n Notes
	err := s.Scan(&n.ID, &n.Key, &n.Data, &n.CreatedAt, &n.UpdatedAt, &n.UserUID, &n.HouseholdUID, &n.Tags)
	return n, err
}

func scanCredentials(s scannable) (Credentials, error) {
	var c Credentials
	err := s.Scan(&c.ID, &c.UserUID, &c.CredentialType, &c.Value, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func scanSlackUser(s scannable) (SlackUsers, error) {
	var su SlackUsers
	err := s.Scan(&su.SlackUserUID, &su.UserUID, &su.CreatedAt, &su.UpdatedAt)
	return su, err
}

func scanUser(s scannable) (Users, error) {
	var u Users
	err := s.Scan(&u.UID, &u.Name, &u.Email, &u.Description, &u.CreatedAt, &u.UpdatedAt, &u.HouseholdUID)
	return u, err
}

func scanHousehold(s scannable) (Households, error) {
	var h Households
	err := s.Scan(&h.UID, &h.Name, &h.Description, &h.CreatedAt, &h.UpdatedAt)
	return h, err
}

func scanRecipes(s scannable) (Recipes, error) {
	var r Recipes
	err := s.Scan(&r.ID, &r.Title, &r.ExternalURL, &r.Data, &r.Genre, &r.GroceryList, &r.PrepTime, &r.CookTime, &r.TotalTime, &r.Servings, &r.Difficulty, &r.Rating, &r.Tags, &r.UserUID, &r.HouseholdUID, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func buildListQuery(tableName string, columns string, options ListOptions) string {
	query := fmt.Sprintf("SELECT %s FROM %s", columns, tableName)

	if options.WhereClause != "" {
		query += " " + options.WhereClause
	}

	query += fmt.Sprintf(" ORDER BY %s %s", options.SortBy, options.SortDir)

	argOffset := len(options.WhereArgs)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argOffset+1, argOffset+2)

	return query
}
