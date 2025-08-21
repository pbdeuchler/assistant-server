package postgres

const (
	insertTodo = `INSERT INTO todos
	(uid,title,description,data,priority,due_date,recurs_on,marked_complete,
	 external_url,user_uid,household_uid,completed_by,created_at,updated_at)
	VALUES (gen_random_uuid()::uuid,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW()) 
	RETURNING uid, title, description, data, priority, due_date, recurs_on, marked_complete, external_url, user_uid, household_uid, completed_by, created_at, updated_at;`

	getTodo    = `SELECT uid, title, description, data, priority, due_date, recurs_on, marked_complete, external_url, user_uid, household_uid, completed_by, created_at, updated_at FROM todos WHERE uid=$1;`
	listTodos  = `SELECT uid, title, description, data, priority, due_date, recurs_on, marked_complete, external_url, user_uid, household_uid, completed_by, created_at, updated_at FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateTodo = `UPDATE todos SET 
		title=COALESCE($2,title),
		description=COALESCE($3,description),
		data=COALESCE($4,data),
		priority=COALESCE($5,priority),
		due_date=COALESCE($6,due_date),
		recurs_on=COALESCE($7,recurs_on),
		marked_complete=COALESCE($8,marked_complete),
		external_url=COALESCE($9,external_url),
		completed_by=COALESCE($10,completed_by),
		updated_at=NOW()
		WHERE uid=$1 
		RETURNING uid, title, description, data, priority, due_date, recurs_on, marked_complete, external_url, user_uid, household_uid, completed_by, created_at, updated_at;`
	deleteTodo = `DELETE FROM todos WHERE uid=$1;`

	insertBackground = `INSERT INTO backgrounds (key, value, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW()) RETURNING *;`
	getBackground    = `SELECT * FROM backgrounds WHERE key=$1;`
	listBackgrounds  = `SELECT * FROM backgrounds ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateBackground = `UPDATE backgrounds SET value=$2, updated_at=NOW()
		WHERE key=$1 RETURNING *;`
	deleteBackground = `DELETE FROM backgrounds WHERE key=$1;`

	insertPreferences = `INSERT INTO preferences (key, specifier, data, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING key, specifier, data, created_at, updated_at, tags;`
	getPreferences    = `SELECT key, specifier, data, created_at, updated_at, tags FROM preferences WHERE key=$1 AND specifier=$2;`
	listPreferences   = `SELECT key, specifier, data, created_at, updated_at, tags FROM preferences ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updatePreferences = `UPDATE preferences SET data=$3, tags=$4, updated_at=NOW()
		WHERE key=$1 AND specifier=$2 RETURNING key, specifier, data, created_at, updated_at, tags;`
	deletePreferences = `DELETE FROM preferences WHERE key=$1 AND specifier=$2;`

	insertNotes = `INSERT INTO notes (key, user_uid, household_uid, data, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, key, data, created_at, updated_at, user_uid, household_uid, tags;`
	getNotes    = `SELECT id, key, data, created_at, updated_at, user_uid, household_uid, tags FROM notes WHERE id=$1;`
	listNotes   = `SELECT * FROM notes ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateNotes = `UPDATE notes SET key=$2, user_uid=$3, household_uid=$4, data=$5, tags=$6, updated_at=NOW()
		WHERE id=$1 RETURNING id, key, data, created_at, updated_at, user_uid, household_uid, tags;`
	deleteNotes = `DELETE FROM notes WHERE id=$1;`

	insertCredentials = `INSERT INTO credentials (user_uid, credential_type, value, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW()) RETURNING *;`
	getCredentials              = `SELECT * FROM credentials WHERE id=$1;`
	getCredentialsByUserAndType = `SELECT * FROM credentials WHERE user_uid=$1 AND credential_type=$2;`
	listCredentials             = `SELECT * FROM credentials ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateCredentials           = `UPDATE credentials SET user_uid=$2, credential_type=$3, value=$4, updated_at=NOW()
		WHERE id=$1 RETURNING *;`
	deleteCredentials = `DELETE FROM credentials WHERE id=$1;`

	insertRecipes = `INSERT INTO recipes (title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at)
		VALUES ($2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()) RETURNING id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at;`
	getRecipes    = `SELECT id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at FROM recipes WHERE id=$1;`
	listRecipes   = `SELECT id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at FROM recipes ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateRecipes = `UPDATE recipes SET title=$2, external_url=$3, data=$4, genre=$5, grocery_list=$6, prep_time=$7, cook_time=$8, total_time=$9, servings=$10, difficulty=$11, rating=$12, tags=$13, user_uid=$14, household_uid=$15, updated_at=NOW()
		WHERE id=$1 RETURNING id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at;`
	deleteRecipes = `DELETE FROM recipes WHERE id=$1;`

	insertUser = `INSERT INTO users (uid, name, email, description, household_uid, created_at, updated_at)
		VALUES (gen_random_uuid()::uuid, $1, $2, $3, $4, NOW(), NOW()) RETURNING uid, name, email, description, created_at, updated_at, household_uid;`
	updateUser = `UPDATE users SET name=COALESCE($2,name), email=COALESCE($3,email), description=COALESCE($4,description), household_uid=COALESCE($5,household_uid), updated_at=NOW()
		WHERE uid=$1 RETURNING uid, name, email, description, created_at, updated_at, household_uid;`

	getSlackUser            = `SELECT slack_user_uid, user_uid, created_at, updated_at FROM slack_users WHERE slack_user_uid=$1;`
	getUserBySlackUserUID   = `SELECT u.uid, u.name, u.email, u.description, u.created_at, u.updated_at, u.household_uid FROM users u JOIN slack_users su ON u.uid = su.user_uid WHERE su.slack_user_uid=$1;`
	getCredentialsByUserUID = `SELECT id, user_uid, credential_type, value, created_at, updated_at FROM credentials WHERE user_uid=$1;`
	getUser                 = `SELECT uid, name, email, description, created_at, updated_at, household_uid FROM users WHERE uid=$1;`
	getHousehold            = `SELECT * FROM households WHERE uid=$1;`
	updateHousehold         = `UPDATE households SET name=COALESCE($2,name), description=COALESCE($3,description), updated_at=NOW()
		WHERE uid=$1 RETURNING *;`
	getTodosByUserUID       = `SELECT uid, title, description, data, priority, due_date, recurs_on, marked_complete, external_url, user_uid, household_uid, completed_by, created_at, updated_at FROM todos WHERE user_uid=$1;`
	getNotesByUserUID       = `SELECT id, key, data, created_at, updated_at, user_uid, household_uid, tags FROM notes WHERE user_uid=$1;`
	getRecipesByUserUID     = `SELECT id, title, external_url, data, genre, grocery_list, prep_time, cook_time, total_time, servings, difficulty, rating, tags, user_uid, household_uid, created_at, updated_at FROM recipes WHERE user_uid=$1;`
	getPreferencesByUserUID = `SELECT key, specifier, data, created_at, updated_at, tags FROM preferences WHERE specifier=$1;`
)
