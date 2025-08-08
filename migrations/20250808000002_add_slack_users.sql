-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS slack_users (
	slack_user_id    text PRIMARY KEY,
	user_id          uuid NOT NULL REFERENCES users(uid),
	created_at       timestamptz NOT NULL DEFAULT now(),
	updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_slack_users_user_id ON slack_users (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_slack_users_user_id;
DROP TABLE IF EXISTS slack_users;
-- +goose StatementEnd