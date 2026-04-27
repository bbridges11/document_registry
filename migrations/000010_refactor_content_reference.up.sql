-- Refactor content reference to support multiple storage backends
-- Split content_s3_key into content_backend (enum) and content_location (path)

-- Step 1: Add new columns
ALTER TABLE versions 
  ADD COLUMN content_backend VARCHAR(50),
  ADD COLUMN content_location VARCHAR(500);

-- Step 2: Migrate existing data (all existing data uses S3)
UPDATE versions 
SET 
  content_backend = 's3',
  content_location = content_s3_key
WHERE content_s3_key IS NOT NULL;

-- Step 3: Make new columns NOT NULL
ALTER TABLE versions 
  ALTER COLUMN content_backend SET NOT NULL,
  ALTER COLUMN content_location SET NOT NULL;

-- Step 4: Drop old column
ALTER TABLE versions DROP COLUMN content_s3_key;

-- Step 5: Add index on content_backend for filtering by storage type
CREATE INDEX idx_versions_content_backend ON versions(content_backend);