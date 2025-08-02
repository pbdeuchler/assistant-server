-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS backgrounds (
	key         varchar(128) PRIMARY KEY,
	value       text NOT NULL,
	created_at  timestamptz NOT NULL DEFAULT now(),
	updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS preferences (
	key         varchar(128) NOT NULL,
	specifier   varchar(256) NOT NULL,
	data        jsonb NOT NULL,
	created_at  timestamptz NOT NULL DEFAULT now(),
	updated_at  timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (key, specifier)
);

CREATE INDEX IF NOT EXISTS idx_backgrounds_created_at ON backgrounds (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_preferences_created_at ON preferences (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_preferences_key ON preferences (key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_preferences_key;
DROP INDEX IF EXISTS idx_preferences_created_at;
DROP INDEX IF EXISTS idx_backgrounds_created_at;
DROP TABLE IF EXISTS preferences;
DROP TABLE IF EXISTS backgrounds;
-- +goose StatementEnd