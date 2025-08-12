-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS recipes (
	id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	title           text NOT NULL,
	external_url    text,
	data            text NOT NULL,
	genre           text,
	grocery_list    text,
	prep_time       integer,
	cook_time       integer,
	total_time      integer,
	servings        integer,
	difficulty      text,
	rating          integer CHECK (rating >= 1 AND rating <= 5),
	tags            text[],
	user_id         uuid REFERENCES users(uid) ON DELETE CASCADE,
	household_id    uuid REFERENCES households(uid) ON DELETE CASCADE,
	created_at      timestamptz NOT NULL DEFAULT now(),
	updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_recipes_created_at ON recipes (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recipes_user_id ON recipes (user_id);
CREATE INDEX IF NOT EXISTS idx_recipes_household_id ON recipes (household_id);
CREATE INDEX IF NOT EXISTS idx_recipes_title ON recipes (title);
CREATE INDEX IF NOT EXISTS idx_recipes_genre ON recipes (genre);
CREATE INDEX IF NOT EXISTS idx_recipes_rating ON recipes (rating);
CREATE INDEX IF NOT EXISTS idx_recipes_tags ON recipes USING GIN (tags);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_recipes_tags;
DROP INDEX IF EXISTS idx_recipes_rating;
DROP INDEX IF EXISTS idx_recipes_genre;
DROP INDEX IF EXISTS idx_recipes_title;
DROP INDEX IF EXISTS idx_recipes_household_id;
DROP INDEX IF EXISTS idx_recipes_user_id;
DROP INDEX IF EXISTS idx_recipes_created_at;
DROP TABLE IF EXISTS recipes;
-- +goose StatementEnd