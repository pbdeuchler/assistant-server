-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS backgrounds;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS backgrounds (
	key         varchar(128) PRIMARY KEY,
	value       text NOT NULL,
	created_at  timestamptz NOT NULL DEFAULT now(),
	updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_backgrounds_created_at ON backgrounds (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_preferences_created_at ON preferences (created_at DESC);
-- +goose StatementEnd
