package postgres

const (
	insertTodo = `INSERT INTO todos
	(uid,title,description,data,priority,due_date,recurs_on,marked_complete,
	 external_url,created_by,completed_by,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW()) RETURNING *;`

	getTodo    = `SELECT * FROM todos WHERE uid=$1;`
	listTodos  = `SELECT * FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateTodo = `UPDATE todos SET title=$2,description=$3,data=$4,priority=$5,due_date=$6,
		recurs_on=$7,marked_complete=$8,external_url=$9,completed_by=$10,updated_at=NOW()
		WHERE uid=$1 RETURNING *;`
	deleteTodo = `DELETE FROM todos WHERE uid=$1;`

	insertBackground = `INSERT INTO backgrounds (key, value, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW()) RETURNING *;`
	getBackground    = `SELECT * FROM backgrounds WHERE key=$1;`
	listBackgrounds  = `SELECT * FROM backgrounds ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateBackground = `UPDATE backgrounds SET value=$2, updated_at=NOW()
		WHERE key=$1 RETURNING *;`
	deleteBackground = `DELETE FROM backgrounds WHERE key=$1;`

	insertPreferences = `INSERT INTO preferences (key, specifier, data, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW()) RETURNING *;`
	getPreferences    = `SELECT * FROM preferences WHERE key=$1 AND specifier=$2;`
	listPreferences   = `SELECT * FROM preferences ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updatePreferences = `UPDATE preferences SET data=$3, updated_at=NOW()
		WHERE key=$1 AND specifier=$2 RETURNING *;`
	deletePreferences = `DELETE FROM preferences WHERE key=$1 AND specifier=$2;`

	insertNotes = `INSERT INTO notes (id, title, relevant_user, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING *;`
	getNotes    = `SELECT * FROM notes WHERE id=$1;`
	listNotes   = `SELECT * FROM notes ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	updateNotes = `UPDATE notes SET title=$2, relevant_user=$3, content=$4, updated_at=NOW()
		WHERE id=$1 RETURNING *;`
	deleteNotes = `DELETE FROM notes WHERE id=$1;`
)
