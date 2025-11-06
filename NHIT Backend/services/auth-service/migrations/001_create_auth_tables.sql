-- Auth Service Database Tables

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    session_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    INDEX idx_sessions_user_id (user_id),
    INDEX idx_sessions_token (session_token),
    INDEX idx_sessions_expires_at (expires_at)
);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_refresh_tokens_user_id (user_id),
    INDEX idx_refresh_tokens_expires_at (expires_at)
);

-- Password resets table
CREATE TABLE IF NOT EXISTS password_resets (
    token UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_password_resets_user_id (user_id),
    INDEX idx_password_resets_expires_at (expires_at)
);

-- Email verification tokens table
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    token UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_email_verification_user_id (user_id),
    INDEX idx_email_verification_expires_at (expires_at)
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_sessions_user_expires ON sessions(user_id, expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_expires ON refresh_tokens(user_id, expires_at);

-- Comments for documentation
COMMENT ON TABLE sessions IS 'Stores active user sessions with JWT tokens';
COMMENT ON TABLE refresh_tokens IS 'Stores refresh tokens for token rotation';
COMMENT ON TABLE password_resets IS 'Stores password reset tokens';
COMMENT ON TABLE email_verification_tokens IS 'Stores email verification tokens';
