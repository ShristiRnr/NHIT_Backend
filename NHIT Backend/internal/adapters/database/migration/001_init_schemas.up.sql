-- ============================================
-- 001_init_schemas.up.sql
-- Updated User Management System Schema
-- ============================================

-- --------------------
-- Tenants
-- --------------------
CREATE TABLE tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    super_admin_user_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Organizations
-- --------------------
CREATE TABLE organizations (
    org_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Roles
-- --------------------
CREATE TABLE roles (
    role_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Permissions
-- --------------------
CREATE TABLE permissions (
    permission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Role-Permission mapping
-- --------------------
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY(role_id, permission_id)
);

-- --------------------
-- Users
-- --------------------
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_logout_at TIMESTAMPTZ,
    last_login_ip VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- User Roles (Many-to-Many)
-- --------------------
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, role_id)
);

-- --------------------
-- User-Organization Links
-- --------------------
CREATE TABLE user_organizations (
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, org_id)
);

-- --------------------
-- Sessions
-- --------------------
CREATE TABLE sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    session_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- --------------------
-- User Login History
-- --------------------
CREATE TABLE user_login_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    ip_address VARCHAR(50),
    user_agent TEXT,
    login_time TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Password Reset Tokens
-- --------------------
CREATE TABLE password_resets (
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Email Verification Tokens
-- --------------------
CREATE TABLE email_verification_tokens (
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Refresh Tokens
-- --------------------
CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------
-- Departments
-- --------------------
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add department_id to users table
ALTER TABLE users ADD COLUMN department_id UUID REFERENCES departments(id) ON DELETE SET NULL;

-- --------------------
-- Designations
-- --------------------
CREATE TABLE designations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(250) NOT NULL,
    description VARCHAR(500) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    parent_id UUID REFERENCES designations(id) ON DELETE SET NULL,
    level INT DEFAULT 0,
    user_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create unique index for case-insensitive name
CREATE UNIQUE INDEX idx_designations_name_lower ON designations (LOWER(name));

-- Add designation_id to users table
ALTER TABLE users ADD COLUMN designation_id UUID REFERENCES designations(id) ON DELETE SET NULL;

-- --------------------
-- Indexes
-- --------------------
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_organizations_org_id ON user_organizations(org_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_user_login_history_user_id ON user_login_history(user_id);
CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_departments_name ON departments(name);
CREATE INDEX idx_users_department_id ON users(department_id);
CREATE INDEX idx_designations_name ON designations(name);
CREATE INDEX idx_designations_slug ON designations(slug);
CREATE INDEX idx_designations_parent_id ON designations(parent_id);
CREATE INDEX idx_designations_is_active ON designations(is_active);
CREATE INDEX idx_users_designation_id ON users(designation_id);
