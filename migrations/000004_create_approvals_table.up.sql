-- Create approvals table
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

-- Create indexes for faster lookups
CREATE INDEX idx_approvals_version_id ON approvals(version_id);
CREATE INDEX idx_approvals_user_id ON approvals(user_id);
CREATE INDEX idx_approvals_approved ON approvals(approved);
CREATE INDEX idx_approvals_role ON approvals(role);