-- +goose Up
-- +goose StatementBegin
ALTER TABLE todos DROP COLUMN created_by;

ALTER TABLE todos ADD COLUMN user_id uuid REFERENCES users(uid) ON DELETE CASCADE;
ALTER TABLE todos ADD COLUMN household_id uuid REFERENCES households(uid) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_todos_user ON todos (user_id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_todos_user;
ALTER TABLE todos DROP COLUMN user_id;
ALTER TABLE todos DROP COLUMN household_id;
ALTER TABLE todos ADD COLUMN created_by text; -- Revert to original state
-- +goose StatementEnd
