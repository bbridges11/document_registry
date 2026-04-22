-- ============================================
-- Migration 006: Add Search Optimization Indexes
-- ============================================

-- CRITICAL: GIN index for tag array searches
-- Without this, tag searches will do full table scans
CREATE INDEX IF NOT EXISTS idx_documents_tags_gin ON documents USING GIN(tags);

-- Date indexes for filtering and sorting
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_documents_updated_at ON documents(updated_at DESC);

-- Composite index for stakeholder searches (my_stakeholder_docs filter)
CREATE INDEX IF NOT EXISTS idx_stakeholders_document_user ON stakeholders(document_id, user_id);

-- Partial index for recently published searches
CREATE INDEX IF NOT EXISTS idx_versions_updated_at_published 
ON versions(updated_at DESC) 
WHERE status = 'PUBLISHED';

-- OPTIONAL: Full-text search index (commented out for future use)
-- Uncomment when implementing full-text search feature
-- CREATE INDEX IF NOT EXISTS idx_documents_fulltext 
-- ON documents USING GIN(to_tsvector('english', name || ' ' || description));