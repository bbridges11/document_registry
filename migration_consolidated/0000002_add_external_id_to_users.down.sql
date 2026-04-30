-- +goose Down
-- Description: Remove external_id column from users table

-- Drop the unique index
DROP INDEX IF EXISTS idx_users_external_id;

-- Drop the external_id column
ALTER TABLE users DROP COLUMN IF EXISTS external_id;