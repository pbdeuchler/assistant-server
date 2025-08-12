package postgres

const (
	insertTodo = `INSERT INTO todos
	(uid,title,description,data,priority,due_date,recurs_on,marked_complete,
	 external_url,user_id,household_id,completed_by,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW(),NOW()) RETURNING *;`

	getTodo    = `SELECT * FROM todos WHERE uid=$1;`
	listTodos  = `SELECT * FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateTodo = `UPDATE todos SET title=coalesce(title,$2),description=coalesce(description,$3),data=coalesce(data,$4),priority=coalesce(priority,$5),due_date=coalesce(due_date,$6),
		recurs_on=coalesce(recurs_on, $7),marked_complete=coalesce(marked_complete, $8),external_url=coalesce(external_url, $9),completed_by=coalesce(completed_by, $10),updated_at=NOW()
		WHERE uid=$1 RETURNING *;`
	deleteTodo = `DELETE FROM todos WHERE uid=$1;`

	insertBackground = `INSERT INTO backgrounds (key, value, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW()) RETURNING *;`
	getBackground    = `SELECT * FROM backgrounds WHERE key=$1;`
	listBackgrounds  = `SELECT * FROM backgrounds ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateBackground = `UPDATE backgrounds SET value=$2, updated_at=NOW()
		WHERE key=$1 RETURNING *;`
	deleteBackground = `DELETE FROM backgrounds WHERE key=$1;`

	insertPreferences = `INSERT INTO preferences (key, specifier, data, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING *;`
	getPreferences    = `SELECT * FROM preferences WHERE key=$1 AND specifier=$2;`
	listPreferences   = `SELECT * FROM preferences ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updatePreferences = `UPDATE preferences SET data=$3, tags=$4, updated_at=NOW()
		WHERE key=$1 AND specifier=$2 RETURNING *;`
	deletePreferences = `DELETE FROM preferences WHERE key=$1 AND specifier=$2;`

	insertNotes = `INSERT INTO notes (id, key, user_id, household_id, data, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING *;`
	getNotes    = `SELECT * FROM notes WHERE id=$1;`
	listNotes   = `SELECT * FROM notes ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateNotes = `UPDATE notes SET key=$2, user_id=$3, household_id=$4, data=$5, tags=$6, updated_at=NOW()
		WHERE id=$1 RETURNING *;`
	deleteNotes = `DELETE FROM notes WHERE id=$1;`

	insertCredentials = `INSERT INTO credentials (id, user_id, credential_type, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING *;`
	getCredentials = `SELECT * FROM credentials WHERE id=$1;`
	getCredentialsByUserAndType = `SELECT * FROM credentials WHERE user_id=$1 AND credential_type=$2;`
	listCredentials = `SELECT * FROM credentials ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateCredentials = `UPDATE credentials SET user_id=$2, credential_type=$3, value=$4, updated_at=NOW()
		WHERE id=$1 RETURNING *;`
	deleteCredentials = `DELETE FROM credentials WHERE id=$1;`

	insertRecipes = `INSERT INTO recipes (id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_id, household_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()) RETURNING *;`
	getRecipes    = `SELECT * FROM recipes WHERE id=$1;`
	listRecipes   = `SELECT * FROM recipes ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateRecipes = `UPDATE recipes SET title=$2, external_url=$3, data=$4, genre=$5, grocery_list=$6, prep_time=$7, cook_time=$8, total_time=$9, servings=$10, difficulty=$11, rating=$12, tags=$13, user_id=$14, household_id=$15, updated_at=NOW()
		WHERE id=$1 RETURNING *;`
	deleteRecipes = `DELETE FROM recipes WHERE id=$1;`

	getSlackUser = `SELECT * FROM slack_users WHERE slack_user_id=$1;`
	getUserBySlackUserID = `SELECT u.* FROM users u JOIN slack_users su ON u.uid = su.user_id WHERE su.slack_user_id=$1;`
	getCredentialsByUserID = `SELECT * FROM credentials WHERE user_id=$1;`
	getUser = `SELECT * FROM users WHERE uid=$1;`
	getHousehold = `SELECT * FROM households WHERE uid=$1;`
	getTodosByUserID = `SELECT * FROM todos WHERE user_id=$1;`
	getNotesByUserID = `SELECT * FROM notes WHERE user_id=$1;`
	getRecipesByUserID = `SELECT * FROM recipes WHERE user_id=$1;`
	getPreferencesByUserID = `SELECT * FROM preferences WHERE specifier=$1;`
)
