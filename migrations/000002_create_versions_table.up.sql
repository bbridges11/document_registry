-- Create versions table
CREATE TABLE IF NOT EXISTS versions (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    content_s3_key VARCHAR(500) NOT NULL,
    content_hash VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, version)
);

-- Create indexes for faster lookups
CREATE INDEX idx_versions_document_id ON versions(document_id);
CREATE INDEX idx_versions_status ON versions(status);
CREATE INDEX idx_versions_created_by ON versions(created_by);
CREATE INDEX idx_versions_created_at ON versions(created_at DESC);
