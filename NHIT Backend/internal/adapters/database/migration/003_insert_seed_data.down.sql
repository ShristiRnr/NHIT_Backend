-- ============================================
-- 003_INSERT_SEED_DATA.down.sql
-- Rollback seed data created by 003_insert_seed_data.up.sql
-- ============================================

-- Undo tenant.super_admin_user_id update
UPDATE tenants
SET super_admin_user_id = NULL
WHERE tenant_id = '11111111-1111-1111-1111-111111111111';

-- Delete user-organization links
DELETE FROM user_organizations
WHERE (user_id, org_id) IN (
  ('11111111-1111-1111-1111-111111111111', '55555555-5555-5555-5555-555555555555'),
  ('22222222-2222-2222-2222-222222222222', '55555555-5555-5555-5555-555555555555')
);

-- Delete user-roles mappings
DELETE FROM user_roles
WHERE (user_id, role_id) IN (
  ('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
);

-- Delete users
DELETE FROM users
WHERE user_id IN (
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222'
);

-- Delete role-permission mappings for seeded roles
DELETE FROM role_permissions
WHERE role_id IN (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'
);

-- Optionally delete the minimal permissions if not used by others
DELETE FROM permissions
WHERE name IN ('manage-users','view-reports');

-- Delete roles
DELETE FROM roles
WHERE role_id IN (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'
);

-- Delete organization
DELETE FROM organizations
WHERE org_id = '55555555-5555-5555-5555-555555555555';

-- Delete tenant
DELETE FROM tenants
WHERE tenant_id = '11111111-1111-1111-1111-111111111111';
