-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes ADD COLUMN tags text[] DEFAULT '{}';
ALTER TABLE preferences ADD COLUMN tags text[] DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_notes_tags ON notes USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_preferences_tags ON preferences USING GIN (tags);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_preferences_tags;
DROP INDEX IF EXISTS idx_notes_tags;
ALTER TABLE preferences DROP COLUMN tags;
ALTER TABLE notes DROP COLUMN tags;
-- +goose StatementEnd