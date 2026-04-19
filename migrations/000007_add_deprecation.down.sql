-- ============================================
-- Migration 007 Rollback: Remove Deprecation Support
-- ============================================

-- Drop deprecations table (this will also drop all indexes and constraints)
DROP TABLE IF EXISTS deprecations;

-- Restore version status constraint to original values
ALTER TABLE versions DROP CONSTRAINT IF EXISTS versions_status_check;
ALTER TABLE versions ADD CONSTRAINT versions_status_check 
CHECK (status IN ('DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'PUBLISHED', 'DEPRECATED'));

-- Note: DEPRECATING status is removed, DEPRECATED is kept for backward compatibility
-- Any versions in DEPRECATING status would need manual cleanup before running this migration