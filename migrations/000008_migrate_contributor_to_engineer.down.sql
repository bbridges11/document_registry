-- Reverse migration (if needed for rollback)
UPDATE users 
SET role = 'contributor' 
WHERE role = 'engineer';