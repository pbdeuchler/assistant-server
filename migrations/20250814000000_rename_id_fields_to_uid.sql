-- +goose Up
-- +goose StatementBegin

-- Rename user_id to user_uid in todos table
ALTER TABLE todos RENAME COLUMN user_id TO user_uid;

-- Rename household_id to household_uid in todos table  
ALTER TABLE todos RENAME COLUMN household_id TO household_uid;

-- Rename user_id to user_uid in notes table
ALTER TABLE notes RENAME COLUMN user_id TO user_uid;

-- Rename household_id to household_uid in notes table
ALTER TABLE notes RENAME COLUMN household_id TO household_uid;

-- Rename user_id to user_uid in credentials table
ALTER TABLE credentials RENAME COLUMN user_id TO user_uid;

-- Rename user_id to user_uid in recipes table
ALTER TABLE recipes RENAME COLUMN user_id TO user_uid;

-- Rename household_id to household_uid in recipes table
ALTER TABLE recipes RENAME COLUMN household_id TO household_uid;

-- Rename slack_user_id to slack_user_uid in slack_users table
ALTER TABLE slack_users RENAME COLUMN slack_user_id TO slack_user_uid;

-- Rename user_id to user_uid in slack_users table
ALTER TABLE slack_users RENAME COLUMN user_id TO user_uid;

-- Rename household_id to household_uid in users table
ALTER TABLE users RENAME COLUMN household_id TO household_uid;

-- Update indexes
DROP INDEX IF EXISTS idx_todos_user;
CREATE INDEX IF NOT EXISTS idx_todos_user_uid ON todos (user_uid DESC);

DROP INDEX IF EXISTS idx_notes_user;
CREATE INDEX IF NOT EXISTS idx_notes_user_uid ON notes (user_uid DESC);

DROP INDEX IF EXISTS idx_credentials_user_id;
CREATE INDEX IF NOT EXISTS idx_credentials_user_uid ON credentials (user_uid);

DROP INDEX IF EXISTS idx_credentials_user_type;
CREATE UNIQUE INDEX IF NOT EXISTS idx_credentials_user_type ON credentials (user_uid, credential_type);

DROP INDEX IF EXISTS idx_recipes_user_id;
CREATE INDEX IF NOT EXISTS idx_recipes_user_uid ON recipes (user_uid);

DROP INDEX IF EXISTS idx_recipes_household_id;
CREATE INDEX IF NOT EXISTS idx_recipes_household_uid ON recipes (household_uid);

DROP INDEX IF EXISTS idx_slack_users_user_id;
CREATE INDEX IF NOT EXISTS idx_slack_users_user_uid ON slack_users (user_uid);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert user_uid to user_id in todos table
ALTER TABLE todos RENAME COLUMN user_uid TO user_id;

-- Revert household_uid to household_id in todos table  
ALTER TABLE todos RENAME COLUMN household_uid TO household_id;

-- Revert user_uid to user_id in notes table
ALTER TABLE notes RENAME COLUMN user_uid TO user_id;

-- Revert household_uid to household_id in notes table
ALTER TABLE notes RENAME COLUMN household_uid TO household_id;

-- Revert user_uid to user_id in credentials table
ALTER TABLE credentials RENAME COLUMN user_uid TO user_id;

-- Revert user_uid to user_id in recipes table
ALTER TABLE recipes RENAME COLUMN user_uid TO user_id;

-- Revert household_uid to household_id in recipes table
ALTER TABLE recipes RENAME COLUMN household_uid TO household_id;

-- Revert slack_user_uid to slack_user_id in slack_users table
ALTER TABLE slack_users RENAME COLUMN slack_user_uid TO slack_user_id;

-- Revert user_uid to user_id in slack_users table
ALTER TABLE slack_users RENAME COLUMN user_uid TO user_id;

-- Revert household_uid to household_id in users table
ALTER TABLE users RENAME COLUMN household_uid TO household_id;

-- Revert indexes
DROP INDEX IF EXISTS idx_todos_user_uid;
CREATE INDEX IF NOT EXISTS idx_todos_user ON todos (user_id DESC);

DROP INDEX IF EXISTS idx_notes_user_uid;
CREATE INDEX IF NOT EXISTS idx_notes_user ON notes (user_id DESC);

DROP INDEX IF EXISTS idx_credentials_user_uid;
CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials (user_id);

DROP INDEX IF EXISTS idx_credentials_user_type;
CREATE UNIQUE INDEX IF NOT EXISTS idx_credentials_user_type ON credentials (user_id, credential_type);

DROP INDEX IF EXISTS idx_recipes_user_uid;
CREATE INDEX IF NOT EXISTS idx_recipes_user_id ON recipes (user_id);

DROP INDEX IF EXISTS idx_recipes_household_uid;
CREATE INDEX IF NOT EXISTS idx_recipes_household_id ON recipes (household_id);

DROP INDEX IF EXISTS idx_slack_users_user_uid;
CREATE INDEX IF NOT EXISTS idx_slack_users_user_id ON slack_users (user_id);

-- +goose StatementEnd