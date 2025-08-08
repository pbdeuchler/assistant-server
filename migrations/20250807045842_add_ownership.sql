-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes DROP COLUMN relevant_user;

ALTER TABLE notes ADD COLUMN user_id uuid REFERENCES users(uid) ON DELETE CASCADE;
ALTER TABLE notes ADD COLUMN household_id uuid REFERENCES households(uid) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_notes_user ON notes (user_id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notes_user;
ALTER TABLE notes DROP COLUMN user_id;
ALTER TABLE notes DROP COLUMN household_id;
ALTER TABLE notes ADD COLUMN relevant_user text; -- Revert to original state
-- +goose StatementEnd
