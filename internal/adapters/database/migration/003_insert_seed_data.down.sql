-- ============================================
-- 003_INSERT_SEED_DATA.down.sql
-- Rollback initial seed data for User + Auth Management
-- ============================================

-- --------------------
-- Delete Email Verifications
-- --------------------
DELETE FROM email_verifications
WHERE user_id IN (
    '11111111-1111-1111-1111-111111111111', -- Super Admin
    '22222222-2222-2222-2222-222222222222', -- Admin
    '33333333-3333-3333-3333-333333333333', -- Manager
    '44444444-4444-4444-4444-444444444444'  -- Regular User
);

-- --------------------
-- Delete Refresh Tokens
-- --------------------
DELETE FROM refresh_tokens
WHERE user_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444'
);

-- --------------------
-- Delete User-Organization Links
-- --------------------
DELETE FROM user_organizations
WHERE user_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444'
);

-- --------------------
-- Delete Users
-- --------------------
DELETE FROM users
WHERE user_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444'
);

-- --------------------
-- Delete User Roles Mapping
-- --------------------
DELETE FROM user_roles
WHERE user_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444'
);

-- --------------------
-- Delete Role-Permission Mapping
-- --------------------
DELETE FROM role_permissions
WHERE role_id IN (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', -- Super Admin Role
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', -- Admin Role
    'cccccccc-cccc-cccc-cccc-cccccccccccc', -- Manager Role
    'dddddddd-dddd-dddd-dddd-dddddddddddd'  -- User Role
);

-- --------------------
-- Delete Roles
-- --------------------
DELETE FROM roles
WHERE role_id IN (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'dddddddd-dddd-dddd-dddd-dddddddddddd'
);

-- --------------------
-- Delete Permissions
-- --------------------
DELETE FROM permissions
WHERE name IN (
    'view-dashboard',
    'manage-users',
    'manage-roles',
    'manage-permissions',
    'view-reports',
    'manage-settings',
    'view-payments',
    'manage-payments',
    'view-notes',
    'manage-notes',
    'view-tickets',
    'manage-tickets',
    'view-accounts',
    'manage-accounts',
    'view-vendors',
    'manage-vendors',
    'view-beneficiaries',
    'manage-beneficiaries',
    'view-approvals',
    'manage-approvals',
    'create-user',
    'edit-user',
    'delete-user',
    'create-product',
    'edit-product',
    'delete-product',
    'import-payment-excel',
    'view-product'
);

-- --------------------
-- Delete Organizations
-- --------------------
DELETE FROM organizations
WHERE org_id IN (
    '55555555-5555-5555-5555-555555555555' -- Example seeded org
);

-- --------------------
-- Delete Tenants
-- --------------------
DELETE FROM tenants
WHERE tenant_id IN (
    '99999999-9999-9999-9999-999999999999' -- Example seeded tenant
);
