-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS todos (
	uid             uuid PRIMARY KEY,
	title           text NOT NULL,
	description     text,
	data            jsonb,
	priority        smallint,
	due_date        timestamptz,
	recurs_on       text,
	marked_complete timestamptz,
	external_url    text,
	created_by      text,
	completed_by    text,
	created_at      timestamptz NOT NULL DEFAULT now(),
	updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_todos_created_at;
DROP TABLE IF EXISTS todos;
-- +goose StatementEnd
