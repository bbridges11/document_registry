-- Rollback content reference refactoring
-- Restore content_s3_key from content_backend and content_location

-- Step 1: Add back the old column
ALTER TABLE versions ADD COLUMN content_s3_key VARCHAR(500);

-- Step 2: Migrate data back (concatenate or just use location for S3)
UPDATE versions 
SET content_s3_key = content_location
WHERE content_backend = 's3';

-- Step 3: Make old column NOT NULL
ALTER TABLE versions ALTER COLUMN content_s3_key SET NOT NULL;

-- Step 4: Drop new columns and index
DROP INDEX IF EXISTS idx_versions_content_backend;
ALTER TABLE versions 
  DROP COLUMN content_backend,
  DROP COLUMN content_location;