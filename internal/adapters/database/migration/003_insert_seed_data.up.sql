-- ============================================
-- 003_INSERT_SEED_DATA.up.sql
-- Seed data for User + Auth Management
-- ============================================

-- --------------------
-- Insert Tenants
-- --------------------
INSERT INTO tenants (tenant_id, name, created_at, updated_at)
VALUES ('11111111-1111-1111-1111-111111111111', 'Default Tenant', NOW(), NOW());

-- --------------------
-- Insert Roles
-- --------------------
INSERT INTO roles (role_id, name, tenant_id, created_at, updated_at)
VALUES 
('22222222-2222-2222-2222-222222222222', 'Admin', '11111111-1111-1111-1111-111111111111', NOW(), NOW()),
('33333333-3333-3333-3333-333333333333', 'User', '11111111-1111-1111-1111-111111111111', NOW(), NOW());

-- --------------------
-- Insert Organizations
-- --------------------
INSERT INTO organizations (org_id, name, tenant_id, created_at, updated_at)
VALUES ('55555555-5555-5555-5555-555555555555', 'Default Organization', '11111111-1111-1111-1111-111111111111', NOW(), NOW());

-- --------------------
-- Insert Users
-- --------------------
INSERT INTO users (user_id, tenant_id, name, emp_id, email, number, password, active, account_holder, bank_name, bank_account, ifsc_code, designation_id, department_id, created_at, updated_at)
VALUES ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'John Doe', 'EMP001', 'john.doe@example.com', '9999999999', '$2y$10$hashedpassword', 'Y', 'John Doe', 'Demo Bank', '1234567890', 'IFSC001', NULL, NULL, NOW(), NOW());

-- --------------------
-- Assign Roles to User
-- --------------------
INSERT INTO user_roles (user_id, role_id)
VALUES ('44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222');

-- --------------------
-- Assign User to Organization
-- --------------------
INSERT INTO user_organizations (user_id, org_id, role_id)
VALUES ('44444444-4444-4444-4444-444444444444', '55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222');

-- --------------------
-- Insert Email Verification Token
-- --------------------
INSERT INTO email_verifications (user_id, token, created_at)
VALUES ('44444444-4444-4444-4444-444444444444', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NOW());

-- --------------------
-- Insert Refresh Token
-- --------------------
INSERT INTO refresh_tokens (token, user_id, expires_at, created_at)
VALUES ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444444', NOW() + INTERVAL '30 days', NOW());