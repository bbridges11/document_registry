-- ============================================
-- Document Registry - Rollback Initial Schema
-- ============================================
-- This migration drops all tables in reverse dependency order
-- ============================================

-- Drop publications table
DROP TABLE IF EXISTS publications CASCADE;

-- Drop deprecations table
DROP TABLE IF EXISTS deprecations CASCADE;

-- Drop users table
DROP TABLE IF EXISTS users CASCADE;

-- Drop approvals table
DROP TABLE IF EXISTS approvals CASCADE;

-- Drop stakeholders table
DROP TABLE IF EXISTS stakeholders CASCADE;

-- Drop versions table
DROP TABLE IF EXISTS versions CASCADE;

-- Drop document_tags table
DROP TABLE IF EXISTS document_tags CASCADE;

-- Drop documents table
DROP TABLE IF EXISTS documents CASCADE;