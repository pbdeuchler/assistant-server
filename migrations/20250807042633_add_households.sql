-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS households (
  uid          uuid PRIMARY KEY,
  name         text NOT NULL,
  description  text,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS households;
-- +goose StatementEnd
