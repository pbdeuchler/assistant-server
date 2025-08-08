-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS credentials (
	id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id           uuid NOT NULL,
	credential_type   text NOT NULL,
	value            text NOT NULL,
	created_at       timestamptz NOT NULL DEFAULT now(),
	updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials (user_id);
CREATE INDEX IF NOT EXISTS idx_credentials_type ON credentials (credential_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_credentials_user_type ON credentials (user_id, credential_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_credentials_user_type;
DROP INDEX IF EXISTS idx_credentials_type;
DROP INDEX IF EXISTS idx_credentials_user_id;
DROP TABLE IF EXISTS credentials;
-- +goose StatementEnd