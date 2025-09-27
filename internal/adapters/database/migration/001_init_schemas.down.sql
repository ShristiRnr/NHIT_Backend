-- ============================================
-- 001_INIT_SCHEMAS.DOWN.SQL
-- Drop all tables in reverse order of dependencies
-- ============================================

-- --------------------
-- Drop Email Verifications
-- --------------------
DROP TABLE IF EXISTS email_verifications CASCADE;

-- --------------------
-- Drop Password Reset Tokens
-- --------------------
DROP TABLE IF EXISTS password_resets CASCADE;

-- --------------------
-- Drop Refresh Tokens
-- --------------------
DROP TABLE IF EXISTS refresh_tokens CASCADE;

-- --------------------
-- Drop Sessions
-- --------------------
DROP TABLE IF EXISTS sessions CASCADE;

-- --------------------
-- Drop User Login History
-- --------------------
DROP TABLE IF EXISTS user_login_history CASCADE;

-- --------------------
-- Drop User-Organization Links
-- --------------------
DROP TABLE IF EXISTS user_organizations CASCADE;

-- --------------------
-- Drop User-Roles Mapping
-- --------------------
DROP TABLE IF EXISTS user_roles CASCADE;

-- --------------------
-- Drop Role-Permissions Mapping
-- --------------------
DROP TABLE IF EXISTS role_permissions CASCADE;

-- --------------------
-- Drop Permissions
-- --------------------
DROP TABLE IF EXISTS permissions CASCADE;

-- --------------------
-- Drop Users
-- --------------------
DROP TABLE IF EXISTS users CASCADE;

-- --------------------
-- Drop Roles
-- --------------------
DROP TABLE IF EXISTS roles CASCADE;

-- --------------------
-- Drop Organizations
-- --------------------
DROP TABLE IF EXISTS organizations CASCADE;

-- --------------------
-- Drop Tenants
-- --------------------
DROP TABLE IF EXISTS tenants CASCADE;
