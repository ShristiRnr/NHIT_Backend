-- Migration: Create organizations and user_organizations tables
-- Description: Creates the database schema for the organization service

-- ================================================
-- Organizations Table
-- ================================================
CREATE TABLE IF NOT EXISTS organizations (
    org_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL UNIQUE,
    database_name VARCHAR(64) NOT NULL UNIQUE,
    description TEXT,
    logo VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_organizations_tenant_id ON organizations(tenant_id);
CREATE INDEX idx_organizations_code ON organizations(code);
CREATE INDEX idx_organizations_is_active ON organizations(is_active);
CREATE INDEX idx_organizations_created_by ON organizations(created_by);
CREATE INDEX idx_organizations_created_at ON organizations(created_at DESC);

-- ================================================
-- User Organizations Table (Junction Table)
-- ================================================
CREATE TABLE IF NOT EXISTS user_organizations (
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    role_id UUID NOT NULL,
    is_current_context BOOLEAN NOT NULL DEFAULT false,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, org_id),
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_user_organizations_user_id ON user_organizations(user_id);
CREATE INDEX idx_user_organizations_org_id ON user_organizations(org_id);
CREATE INDEX idx_user_organizations_role_id ON user_organizations(role_id);
CREATE INDEX idx_user_organizations_current_context ON user_organizations(user_id, is_current_context) WHERE is_current_context = true;

-- ================================================
-- Trigger to update updated_at timestamp
-- ================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_organizations_updated_at
    BEFORE UPDATE ON user_organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ================================================
-- Comments for documentation
-- ================================================
COMMENT ON TABLE organizations IS 'Stores organization information for multi-tenancy';
COMMENT ON COLUMN organizations.org_id IS 'Unique identifier for the organization';
COMMENT ON COLUMN organizations.tenant_id IS 'Reference to the tenant this organization belongs to';
COMMENT ON COLUMN organizations.code IS 'Unique organization code (e.g., NHIT, ABC)';
COMMENT ON COLUMN organizations.database_name IS 'Database name for this organization (multi-tenant isolation)';
COMMENT ON COLUMN organizations.is_active IS 'Whether the organization is active and accessible';
COMMENT ON COLUMN organizations.created_by IS 'User ID who created this organization';

COMMENT ON TABLE user_organizations IS 'Junction table for user-organization relationships';
COMMENT ON COLUMN user_organizations.user_id IS 'Reference to the user';
COMMENT ON COLUMN user_organizations.org_id IS 'Reference to the organization';
COMMENT ON COLUMN user_organizations.role_id IS 'Role the user has within this organization';
COMMENT ON COLUMN user_organizations.is_current_context IS 'Whether this is the users current active organization context';
