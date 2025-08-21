-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN household_id uuid REFERENCES households(uid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS household_id;
-- +goose StatementEnd