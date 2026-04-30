-- +goose Up
-- Description: Add external_id column to users table for client-provided user identifiers

-- Add external_id column (7-character alphanumeric identifier)
ALTER TABLE users ADD COLUMN external_id VARCHAR(7);

-- Create unique index for fast lookups
CREATE UNIQUE INDEX idx_users_external_id ON users(external_id);

-- Add comment for documentation
COMMENT ON COLUMN users.external_id IS 'Client-provided 7-character alphanumeric identifier (e.g., a123456). Used in X-User-ID header for API requests.';