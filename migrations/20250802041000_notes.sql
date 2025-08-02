-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notes (
	id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	title           text NOT NULL,
	relevant_user   text NOT NULL,
	content         text NOT NULL,
	created_at      timestamptz NOT NULL DEFAULT now(),
	updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notes_relevant_user ON notes (relevant_user);
CREATE INDEX IF NOT EXISTS idx_notes_title ON notes (title);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notes_title;
DROP INDEX IF EXISTS idx_notes_relevant_user;
DROP INDEX IF EXISTS idx_notes_created_at;
DROP TABLE IF EXISTS notes;
-- +goose StatementEnd