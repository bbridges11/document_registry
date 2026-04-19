-- ============================================
-- Migration 006 Rollback: Remove Search Indexes
-- ============================================

DROP INDEX IF EXISTS idx_documents_tags_gin;
DROP INDEX IF EXISTS idx_documents_created_at;
DROP INDEX IF EXISTS idx_documents_updated_at;
DROP INDEX IF EXISTS idx_stakeholders_document_user;
DROP INDEX IF EXISTS idx_versions_updated_at_published;

-- DROP INDEX IF EXISTS idx_documents_fulltext;