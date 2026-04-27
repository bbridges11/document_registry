-- Migrate existing contributor users to engineer role
UPDATE users 
SET role = 'engineer' 
WHERE role = 'contributor';

-- No down migration - this is a one-way data migration