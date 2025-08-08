-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  uid          uuid PRIMARY KEY,
  name         text NOT NULL,
  email        text NOT NULL UNIQUE,
  description  text,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
