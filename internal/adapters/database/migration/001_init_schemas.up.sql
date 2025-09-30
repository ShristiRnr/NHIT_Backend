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
-- Departments
-- --------------------
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);


CREATE TABLE designations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- vendors table
CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    s_no VARCHAR(255),
    from_account_type VARCHAR(255),
    status VARCHAR(255),
    project VARCHAR(255),
    account_name VARCHAR(255),
    short_name VARCHAR(255),
    parent VARCHAR(255),
    account_number VARCHAR(255) NOT NULL,
    name_of_bank VARCHAR(255) NOT NULL,
    ifsc_code_id VARCHAR(255),
    ifsc_code VARCHAR(11) NOT NULL CHECK (ifsc_code ~ '^[A-Z]{4}0[A-Z0-9]{6}$'),
    vendor_type VARCHAR(255),
    vendor_code VARCHAR(255) NOT NULL UNIQUE,
    vendor_name VARCHAR(255) NOT NULL UNIQUE,
    vendor_email VARCHAR(255) NOT NULL UNIQUE,
    vendor_mobile VARCHAR(10) UNIQUE,
    activity_type VARCHAR(255),
    vendor_nick_name VARCHAR(255),
    email VARCHAR(255),
    mobile VARCHAR(10),
    gstin VARCHAR(255),
    pan VARCHAR(20) NOT NULL,
    pin VARCHAR(20),
    country_id VARCHAR(255),
    state_id VARCHAR(255),
    city_id VARCHAR(255),
    country_name VARCHAR(255),
    state_name VARCHAR(255),
    city_name VARCHAR(255),
    msme_classification VARCHAR(255),
    msme VARCHAR(255),
    msme_registration_number VARCHAR(255),
    msme_start_date DATE,
    msme_end_date DATE,
    material_nature VARCHAR(255),
    gst_defaulted VARCHAR(255),
    section_206AB_verified VARCHAR(255),
    benificiary_name VARCHAR(255) NOT NULL,
    remarks_address VARCHAR(255),
    common_bank_details VARCHAR(255),
    income_tax_type VARCHAR(255),
    file_path JSONB DEFAULT '[]'::jsonb,
    active CHAR(1) NOT NULL DEFAULT 'Y',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- auto update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_vendors_updated_at
BEFORE UPDATE ON vendors
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

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
CREATE INDEX idx_departments_name ON departments(name);
CREATE INDEX idx_departments_created_at ON departments(created_at);
CREATE INDEX idx_departments_updated_at ON departments(updated_at);
CREATE INDEX idx_designations_name ON designations(name);
CREATE INDEX idx_designations_created_at ON designations(created_at);
CREATE INDEX idx_designations_updated_at ON designations(updated_at);
CREATE INDEX idx_vendors_country_id ON vendors(country_id);
CREATE INDEX idx_vendors_state_id ON vendors(state_id);
CREATE INDEX idx_vendors_city_id ON vendors(city_id);
CREATE INDEX idx_vendors_vendor_type ON vendors(vendor_type);
CREATE INDEX idx_vendors_status ON vendors(status);
CREATE INDEX idx_vendors_active ON vendors(active);
CREATE INDEX idx_vendors_vendor_name_trgm ON vendors USING gin (vendor_name gin_trgm_ops);
CREATE INDEX idx_vendors_vendor_code_trgm ON vendors USING gin (vendor_code gin_trgm_ops);
CREATE INDEX idx_vendors_file_path ON vendors USING gin (file_path);
CREATE INDEX idx_vendors_created_at ON vendors(created_at DESC);