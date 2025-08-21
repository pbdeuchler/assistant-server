package integration_test

import (
	"context"
	"testing"

	"github.com/pbdeuchler/assistant-server/integration_test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	ctx := context.Background()
	
	// Check that all expected tables exist
	expectedTables := []string{
		"todos", "users", "households", "notes", "preferences", "recipes", "credentials", "slack_users",
	}
	
	for _, tableName := range expectedTables {
		t.Run("Table_"+tableName+"_exists", func(t *testing.T) {
			var exists bool
			err := db.Pool.QueryRow(ctx, 
				"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
				tableName).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "Table %s should exist", tableName)
		})
	}
	
	// Check that we can describe the todos table
	t.Run("Todos_table_schema", func(t *testing.T) {
		rows, err := db.Pool.Query(ctx, `
			SELECT column_name, data_type, is_nullable 
			FROM information_schema.columns 
			WHERE table_name = 'todos' 
			ORDER BY ordinal_position
		`)
		require.NoError(t, err)
		defer rows.Close()
		
		columns := make(map[string]string)
		for rows.Next() {
			var columnName, dataType, isNullable string
			err := rows.Scan(&columnName, &dataType, &isNullable)
			require.NoError(t, err)
			columns[columnName] = dataType
		}
		
		// Check expected columns exist
		expectedColumns := []string{
			"uid", "title", "description", "data", "priority", 
			"due_date", "recurs_on", "marked_complete", "external_url",
			"user_uid", "household_uid", "completed_by", "created_at", "updated_at",
		}
		
		for _, col := range expectedColumns {
			assert.Contains(t, columns, col, "Column %s should exist in todos table", col)
		}
		
		// Print the schema for debugging
		t.Logf("Todos table schema: %+v", columns)
	})
}