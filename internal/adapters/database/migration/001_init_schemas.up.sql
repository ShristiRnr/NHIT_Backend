-- ============================================
-- 001_init_schemas.up.sql
-- User Management System Schema
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

-- Roles table
CREATE TABLE roles (
    role_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Permissions table
CREATE TABLE permissions (
    permission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);


-- Role-Permission mapping table
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY(role_id, permission_id)
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
-- Users
-- --------------------
-- Users table
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
    name VARCHAR(255) NOT NULL,
    emp_id VARCHAR(255) UNIQUE,
    number VARCHAR(15),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    active CHAR(1) DEFAULT 'Y',
    account_holder VARCHAR(255),
    bank_name VARCHAR(255),
    bank_account VARCHAR(255),
    ifsc_code VARCHAR(255),
    designation_id UUID,
    department_id UUID,
    email_verified_at TIMESTAMP,
    last_login_at TIMESTAMP,
    last_logout_at TIMESTAMP,
    last_login_ip VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- --------------------
-- Organization-User Links
-- --------------------
CREATE TABLE IF NOT EXISTS user_organizations (
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY (user_id, org_id)
);

-- --------------------
-- Sessions
-- --------------------
CREATE TABLE sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    token TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    expires_at TIMESTAMP
);

-- --------------------
-- User Login History
-- --------------------
CREATE TABLE user_login_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    ip_address VARCHAR(50),
    user_agent TEXT,
    login_time TIMESTAMP DEFAULT now()
);

-- --------------------
-- Password Reset Tokens
-- --------------------
CREATE TABLE password_resets (
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now()
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
-- Email Verifications
-- --------------------
CREATE TABLE email_verifications (
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT now()
);

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
CREATE INDEX idx_password_resets_email ON password_resets(email);
CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);

