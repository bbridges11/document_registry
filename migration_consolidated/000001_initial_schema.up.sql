-- ============================================
-- Document Registry - Initial Schema
-- ============================================
-- This migration creates the complete initial database schema
-- for the Document Registry service.
-- ============================================

-- ============================================
-- DOCUMENTS TABLE
-- ============================================
-- Core document metadata table
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

-- Indexes for documents
CREATE INDEX idx_documents_created_by ON documents(created_by);
CREATE INDEX idx_documents_type ON documents(document_type);
CREATE INDEX idx_documents_tags_gin ON documents USING GIN(tags);
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);
CREATE INDEX idx_documents_updated_at ON documents(updated_at DESC);

-- Document tags table (many-to-many)
CREATE TABLE IF NOT EXISTS document_tags (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL,
    PRIMARY KEY (document_id, tag)
);

CREATE INDEX idx_document_tags_tag ON document_tags(tag);

-- ============================================
-- VERSIONS TABLE
-- ============================================
-- Version history and content references
CREATE TABLE IF NOT EXISTS versions (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'PUBLISHED', 'DEPRECATING', 'DEPRECATED')),
    content_backend VARCHAR(50) NOT NULL,
    content_location VARCHAR(500) NOT NULL,
    content_hash VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, version)
);

-- Indexes for versions
CREATE INDEX idx_versions_document_id ON versions(document_id);
CREATE INDEX idx_versions_status ON versions(status);
CREATE INDEX idx_versions_created_by ON versions(created_by);
CREATE INDEX idx_versions_created_at ON versions(created_at DESC);
CREATE INDEX idx_versions_content_backend ON versions(content_backend);
CREATE INDEX idx_versions_updated_at_published ON versions(updated_at DESC) WHERE status = 'PUBLISHED';

-- ============================================
-- STAKEHOLDERS TABLE
-- ============================================
-- Document stakeholder relationships
CREATE TABLE IF NOT EXISTS stakeholders (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, user_id)
);

-- Indexes for stakeholders
CREATE INDEX idx_stakeholders_document_id ON stakeholders(document_id);
CREATE INDEX idx_stakeholders_user_id ON stakeholders(user_id);
CREATE INDEX idx_stakeholders_role ON stakeholders(role);
CREATE INDEX idx_stakeholders_document_user ON stakeholders(document_id, user_id);

-- ============================================
-- APPROVALS TABLE
-- ============================================
-- Version approval workflow
CREATE TABLE IF NOT EXISTS approvals (
    id UUID PRIMARY KEY,
    version_id UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(version_id, user_id, role)
);

-- Indexes for approvals
CREATE INDEX idx_approvals_version_id ON approvals(version_id);
CREATE INDEX idx_approvals_user_id ON approvals(user_id);
CREATE INDEX idx_approvals_approved ON approvals(approved);
CREATE INDEX idx_approvals_role ON approvals(role);

-- ============================================
-- USERS TABLE
-- ============================================
-- User registry
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(active);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- ============================================
-- DEPRECATIONS TABLE
-- ============================================
-- Version deprecation workflow
CREATE TABLE IF NOT EXISTS deprecations (
    id UUID PRIMARY KEY,
    version_id UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    
    -- Request metadata
    requested_by VARCHAR(255) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason TEXT NOT NULL,
    deprecation_note TEXT,
    
    -- Type and status
    auto_deprecated BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELED')),
    
    -- Completion metadata
    deprecated_by VARCHAR(255),
    deprecated_at TIMESTAMP,
    superseded_by UUID REFERENCES versions(id),
    
    -- Rollback support
    previous_status VARCHAR(50) NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for deprecations
CREATE INDEX idx_deprecations_version_id ON deprecations(version_id);
CREATE INDEX idx_deprecations_document_id ON deprecations(document_id);
CREATE INDEX idx_deprecations_status ON deprecations(status);
CREATE INDEX idx_deprecations_requested_by ON deprecations(requested_by);

-- Unique constraint: only one PENDING deprecation per version at a time
CREATE UNIQUE INDEX idx_deprecations_version_pending 
ON deprecations(version_id) 
WHERE status = 'PENDING';

-- ============================================
-- PUBLICATIONS TABLE
-- ============================================
-- Tracks where and when versions are published
CREATE TABLE IF NOT EXISTS publications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    published_to TEXT NOT NULL,
    published_by TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL CHECK (status IN ('success', 'failed')),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for publications
CREATE INDEX idx_publications_version_id ON publications(version_id);
CREATE INDEX idx_publications_document_id ON publications(document_id);
CREATE INDEX idx_publications_published_to ON publications(published_to);
CREATE INDEX idx_publications_status ON publications(status);
CREATE INDEX idx_publications_published_at ON publications(published_at DESC);

-- Comments for documentation
COMMENT ON TABLE publications IS 'Tracks where and when document versions are published';
COMMENT ON COLUMN publications.published_to IS 'Destination identifier (e.g., dev-portal, prod-gateway)';
COMMENT ON COLUMN publications.status IS 'Publication result: success or failed';
COMMENT ON COLUMN publications.error_message IS 'Error details if publication failed';