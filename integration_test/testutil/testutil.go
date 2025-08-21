package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
	"github.com/stretchr/testify/require"
)

const TestDatabaseURL = "postgres://test_user:test_password@localhost:5433/assistant_test?sslmode=disable"

type TestDatabase struct {
	Pool *pgxpool.Pool
	DAO  *dao.DAO
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	ctx := context.Background()
	
	// Connect to database
	pool, err := pgxpool.New(ctx, TestDatabaseURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Wait for database to be ready
	require.Eventually(t, func() bool {
		return pool.Ping(ctx) == nil
	}, 30*time.Second, 500*time.Millisecond, "Database did not become ready")

	// Clean up any existing schema first
	cleanupDatabase(ctx, pool)

	// Run migrations
	err = runMigrations(ctx, pool)
	require.NoError(t, err, "Failed to run migrations")

	// Create DAO instance
	testDAO, err := dao.New(ctx, pool)
	require.NoError(t, err, "Failed to create DAO")

	db := &TestDatabase{
		Pool: pool,
		DAO:  testDAO,
	}

	// Clean up after test
	t.Cleanup(func() {
		cleanupDatabase(ctx, pool)
		pool.Close()
	})

	return db
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Try to get absolute path to migrations directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// Look for migrations directory
	migrationsDir := filepath.Join(wd, "..", "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try from the project root
		migrationsDir = filepath.Join(wd, "..", "..", "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			return fmt.Errorf("migrations directory not found, tried: %s", migrationsDir)
		}
	}
	
	// Get all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}
	
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}
	
	// Sort migration files by filename (which includes timestamp)
	sort.Strings(files)
	
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}
		
		// Extract SQL from goose migration format
		sql := extractSQLFromGooseMigration(string(content))
		if sql == "" {
			fmt.Printf("Skipped migration (no SQL found): %s\n", filepath.Base(file))
			continue
		}
		
		_, err = pool.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
		
		fmt.Printf("Applied migration: %s\n", filepath.Base(file))
	}
	
	return nil
}

func extractSQLFromGooseMigration(content string) string {
	lines := strings.Split(content, "\n")
	var sqlLines []string
	inUpSection := false
	inStatementBlock := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "-- +goose Up") {
			inUpSection = true
			continue
		}
		if strings.HasPrefix(line, "-- +goose Down") {
			break // Stop at down section
		}
		if strings.HasPrefix(line, "-- +goose StatementBegin") {
			inStatementBlock = true
			continue
		}
		if strings.HasPrefix(line, "-- +goose StatementEnd") {
			inStatementBlock = false
			continue
		}
		
		if inUpSection && (inStatementBlock || !strings.HasPrefix(line, "--")) {
			if line != "" && !strings.HasPrefix(line, "--") {
				sqlLines = append(sqlLines, line)
			}
		}
	}
	
	return strings.Join(sqlLines, "\n")
}

func cleanupDatabase(ctx context.Context, pool *pgxpool.Pool) {
	// Drop all tables if they exist (in reverse dependency order)
	tables := []string{
		"recipes", "notes", "preferences", "todos", 
		"credentials", "slack_users", "users", "households",
	}
	
	for _, table := range tables {
		_, _ = pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
	}
}

// Test fixtures and helpers

func generateTestUUID(t *testing.T) string {
	return uuid.New().String()
}

func CreateTestUser(t *testing.T, db *TestDatabase) dao.Users {
	t.Helper()
	ctx := context.Background()
	
	// Generate a proper UUID for the user
	userUID := generateTestUUID(t)
	
	user := dao.Users{
		UID:         userUID,
		Name:        "Test User",
		Email:       "test@example.com",
		Description: "Test user for integration tests",
	}
	
	created, err := db.DAO.CreateUser(ctx, user)
	require.NoError(t, err)
	return created
}

func CreateTestHousehold(t *testing.T, db *TestDatabase) dao.Households {
	t.Helper()
	ctx := context.Background()
	
	// Generate a proper UUID for the household
	householdUID := generateTestUUID(t)
	
	_, err := db.Pool.Exec(ctx, 
		"INSERT INTO households (uid, name, description) VALUES ($1, $2, $3)",
		householdUID, "Test Household", "Test household for integration tests",
	)
	require.NoError(t, err)
	
	household, err := db.DAO.GetHousehold(ctx, householdUID)
	require.NoError(t, err)
	return household
}

func CreateTestTodo(t *testing.T, db *TestDatabase, userUID, householdUID string) dao.Todo {
	t.Helper()
	ctx := context.Background()
	
	dueDate := time.Now().Add(24 * time.Hour)
	todo := dao.Todo{
		UID:          generateTestUUID(t),
		Title:        "Test Todo",
		Description:  "Test todo for integration tests",
		Data:         `{"test": true}`,
		Priority:     dao.Priority(3),
		DueDate:      &dueDate,
		UserUID:      userUID,
		HouseholdUID: householdUID,
	}
	
	created, err := db.DAO.CreateTodo(ctx, todo)
	require.NoError(t, err)
	return created
}

func CreateTestNote(t *testing.T, db *TestDatabase, userUID, householdUID string) dao.Notes {
	t.Helper()
	ctx := context.Background()
	
	note := dao.Notes{
		ID:           generateTestUUID(t),
		Key:          "test-key",
		UserUID:      userUID,
		HouseholdUID: householdUID,
		Data:         `{"content": "Test note content", "test": true}`,
		Tags:         []string{"test", "integration"},
	}
	
	created, err := db.DAO.CreateNotes(ctx, note)
	require.NoError(t, err)
	return created
}

func CreateTestRecipe(t *testing.T, db *TestDatabase, userUID, householdUID string) dao.Recipes {
	t.Helper()
	ctx := context.Background()
	
	prepTime := 15
	cookTime := 30
	totalTime := 45
	servings := 4
	rating := 5
	genre := "italian"
	difficulty := "medium"
	groceryList := `["pasta", "tomatoes", "cheese"]`
	
	recipe := dao.Recipes{
		ID:           generateTestUUID(t),
		Title:        "Test Recipe",
		Data:         `{"instructions": ["Step 1", "Step 2"], "test": true}`,
		Genre:        &genre,
		GroceryList:  &groceryList,
		PrepTime:     &prepTime,
		CookTime:     &cookTime,
		TotalTime:    &totalTime,
		Servings:     &servings,
		Difficulty:   &difficulty,
		Rating:       &rating,
		Tags:         []string{"test", "pasta", "italian"},
		UserUID:      userUID,
		HouseholdUID: householdUID,
	}
	
	created, err := db.DAO.CreateRecipes(ctx, recipe)
	require.NoError(t, err)
	return created
}

func CreateTestPreference(t *testing.T, db *TestDatabase) dao.Preferences {
	t.Helper()
	ctx := context.Background()
	
	pref := dao.Preferences{
		Key:       "test-key",
		Specifier: generateTestUUID(t),
		Data:      `{"theme": "dark", "test": true}`,
		Tags:      []string{"test", "ui"},
	}
	
	created, err := db.DAO.CreatePreferences(ctx, pref)
	require.NoError(t, err)
	return created
}

// Assertion helpers

func AssertTodoEqual(t *testing.T, expected, actual dao.Todo) {
	t.Helper()
	require.Equal(t, expected.UID, actual.UID)
	require.Equal(t, expected.Title, actual.Title)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Priority, actual.Priority)
	require.Equal(t, expected.UserUID, actual.UserUID)
	require.Equal(t, expected.HouseholdUID, actual.HouseholdUID)
}

func AssertNoteEqual(t *testing.T, expected, actual dao.Notes) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Key, actual.Key)
	require.Equal(t, expected.Data, actual.Data)
	require.Equal(t, expected.UserUID, actual.UserUID)
	require.Equal(t, expected.HouseholdUID, actual.HouseholdUID)
	require.ElementsMatch(t, expected.Tags, actual.Tags)
}

func AssertRecipeEqual(t *testing.T, expected, actual dao.Recipes) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Title, actual.Title)
	require.Equal(t, expected.Data, actual.Data)
	require.Equal(t, expected.UserUID, actual.UserUID)
	require.Equal(t, expected.HouseholdUID, actual.HouseholdUID)
	if expected.Genre != nil && actual.Genre != nil {
		require.Equal(t, *expected.Genre, *actual.Genre)
	}
	if expected.Rating != nil && actual.Rating != nil {
		require.Equal(t, *expected.Rating, *actual.Rating)
	}
	require.ElementsMatch(t, expected.Tags, actual.Tags)
}