-- Create stakeholders table
CREATE TABLE IF NOT EXISTS stakeholders (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, user_id)
);

-- Create indexes for faster lookups
CREATE INDEX idx_stakeholders_document_id ON stakeholders(document_id);
CREATE INDEX idx_stakeholders_user_id ON stakeholders(user_id);
CREATE INDEX idx_stakeholders_role ON stakeholders(role);
