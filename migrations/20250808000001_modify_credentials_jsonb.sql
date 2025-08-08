-- +goose Up
-- +goose StatementBegin
ALTER TABLE credentials ALTER COLUMN value TYPE jsonb USING value::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credentials ALTER COLUMN value TYPE text USING value::text;
-- +goose StatementEnd