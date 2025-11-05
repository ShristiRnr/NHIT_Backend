-- ============================================
-- 003_INSERT_SEED_DATA.up.sql
-- Seed initial multi-tenant data: tenant, org, roles, permissions, users, mappings
-- ============================================

-- IDs used in this seed for easy rollback
-- tenant:        11111111-1111-1111-1111-111111111111
-- org:           55555555-5555-5555-5555-555555555555
-- roles:         SUPER_ADMIN aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa
--                ADMIN       bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb
-- users:         superadmin  11111111-1111-1111-1111-111111111111
--                admin       22222222-2222-2222-2222-222222222222

-- Tenant
INSERT INTO tenants (tenant_id, name, super_admin_user_id)
VALUES ('11111111-1111-1111-1111-111111111111', 'Demo Tenant', NULL)
ON CONFLICT (tenant_id) DO NOTHING;

-- Organization
INSERT INTO organizations (org_id, tenant_id, name)
VALUES ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'Demo Org')
ON CONFLICT (org_id) DO NOTHING;

-- Roles
INSERT INTO roles (role_id, tenant_id, name)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'SUPER_ADMIN'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'ADMIN')
ON CONFLICT (role_id) DO NOTHING;

-- Minimal Permissions
INSERT INTO permissions (permission_id, name, description)
VALUES
  (gen_random_uuid(), 'manage-users', 'Create, update, delete users'),
  (gen_random_uuid(), 'view-reports', 'View reports and analytics')
ON CONFLICT (name) DO NOTHING;

-- Map permissions to roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', p.permission_id
FROM permissions p WHERE p.name IN ('manage-users','view-reports')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', p.permission_id
FROM permissions p WHERE p.name IN ('view-reports')
ON CONFLICT DO NOTHING;

-- Users (passwords are placeholders; replace with real hashes in prod)
INSERT INTO users (user_id, tenant_id, name, email, password, email_verified_at)
VALUES
  ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'Super Admin', 'superadmin@example.com', 'changeme', NOW()),
  ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Admin User', 'admin@example.com', 'changeme', NULL)
ON CONFLICT (user_id) DO NOTHING;

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
ON CONFLICT DO NOTHING;

-- Link users to organization
INSERT INTO user_organizations (user_id, org_id, role_id)
VALUES
  ('11111111-1111-1111-1111-111111111111', '55555555-5555-5555-5555-555555555555', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('22222222-2222-2222-2222-222222222222', '55555555-5555-5555-5555-555555555555', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
ON CONFLICT DO NOTHING;

-- Update tenant.super_admin_user_id now that user exists
UPDATE tenants
SET super_admin_user_id = '11111111-1111-1111-1111-111111111111'
WHERE tenant_id = '11111111-1111-1111-1111-111111111111';
