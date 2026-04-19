-- Create documents table
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tags TEXT[] DEFAULT '{}',
    document_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL
);

-- Create index on created_by for faster lookups
CREATE INDEX idx_documents_created_by ON documents(created_by);

-- Create index on document_type for filtering
CREATE INDEX idx_documents_type ON documents(document_type);

-- Create document_tags table for many-to-many relationship
CREATE TABLE IF NOT EXISTS document_tags (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL,
    PRIMARY KEY (document_id, tag)
);

-- Create index on tag for faster tag-based searches
CREATE INDEX idx_document_tags_tag ON document_tags(tag);
