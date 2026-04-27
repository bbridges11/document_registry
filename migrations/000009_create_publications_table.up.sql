-- ============================================
-- Migration 009: Create Publications Table
-- ============================================

CREATE TABLE IF NOT EXISTS publications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    published_to TEXT NOT NULL,
    published_by TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,
    error_message TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT publications_status_check CHECK (status IN ('success', 'failed'))
);

-- Indexes for efficient querying
CREATE INDEX idx_publications_version_id ON publications(version_id);
CREATE INDEX idx_publications_document_id ON publications(document_id);
CREATE INDEX idx_publications_published_to ON publications(published_to);
CREATE INDEX idx_publications_status ON publications(status);
CREATE INDEX idx_publications_published_at ON publications(published_at DESC);

-- Comment
COMMENT ON TABLE publications IS 'Tracks where and when document versions are published';
COMMENT ON COLUMN publications.published_to IS 'Destination identifier (e.g., dev-portal, prod-gateway)';
COMMENT ON COLUMN publications.status IS 'Publication result: success or failed';
COMMENT ON COLUMN publications.error_message IS 'Error details if publication failed';