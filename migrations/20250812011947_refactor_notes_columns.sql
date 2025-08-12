-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes RENAME COLUMN title TO key;
ALTER TABLE notes RENAME COLUMN content TO data;

-- Update indexes to use the new column names
DROP INDEX IF EXISTS idx_notes_title;
CREATE INDEX IF NOT EXISTS idx_notes_key ON notes (key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert the column renames
ALTER TABLE notes RENAME COLUMN key TO title;
ALTER TABLE notes RENAME COLUMN data TO content;

-- Revert the indexes
DROP INDEX IF EXISTS idx_notes_key;
CREATE INDEX IF NOT EXISTS idx_notes_title ON notes (title);
-- +goose StatementEnd