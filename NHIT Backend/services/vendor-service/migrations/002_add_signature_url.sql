-- Add signature_url column to vendors table
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS signature_url VARCHAR(500);

-- Add comment
COMMENT ON COLUMN vendors.signature_url IS 'MinIO URL/path to vendor signature image';
