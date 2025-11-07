-- Designation Service Database Schema

-- Designations table
CREATE TABLE IF NOT EXISTS designations (
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
CREATE UNIQUE INDEX IF NOT EXISTS idx_designations_name_lower ON designations (LOWER(name));

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_designations_name ON designations(name);
CREATE INDEX IF NOT EXISTS idx_designations_slug ON designations(slug);
CREATE INDEX IF NOT EXISTS idx_designations_parent_id ON designations(parent_id);
CREATE INDEX IF NOT EXISTS idx_designations_is_active ON designations(is_active);
CREATE INDEX IF NOT EXISTS idx_designations_level ON designations(level);

-- Comments for documentation
COMMENT ON TABLE designations IS 'Stores designation/position information with hierarchical structure';
COMMENT ON COLUMN designations.id IS 'Unique identifier for the designation';
COMMENT ON COLUMN designations.name IS 'Designation name';
COMMENT ON COLUMN designations.slug IS 'URL-friendly slug (unique)';
COMMENT ON COLUMN designations.parent_id IS 'Parent designation for hierarchy';
COMMENT ON COLUMN designations.level IS 'Hierarchy level (0 for root)';
COMMENT ON COLUMN designations.user_count IS 'Number of users with this designation';
