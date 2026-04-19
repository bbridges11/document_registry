-- ============================================
-- Migration 007: Add Deprecation Support
-- ============================================

-- Add new statuses to versions table
-- Note: We need to drop and recreate the constraint to add new values
ALTER TABLE versions DROP CONSTRAINT IF EXISTS versions_status_check;
ALTER TABLE versions ADD CONSTRAINT versions_status_check 
CHECK (status IN ('DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'PUBLISHED', 'DEPRECATING', 'DEPRECATED'));

-- Create deprecations table
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

-- Create indexes for efficient queries
CREATE INDEX idx_deprecations_version_id ON deprecations(version_id);
CREATE INDEX idx_deprecations_document_id ON deprecations(document_id);
CREATE INDEX idx_deprecations_status ON deprecations(status);
CREATE INDEX idx_deprecations_requested_by ON deprecations(requested_by);

-- Unique constraint: only one PENDING deprecation per version at a time
-- This prevents multiple concurrent deprecation requests for the same version
CREATE UNIQUE INDEX idx_deprecations_version_pending 
ON deprecations(version_id) 
WHERE status = 'PENDING';