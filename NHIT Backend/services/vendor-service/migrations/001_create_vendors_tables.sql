-- Create vendors table with all fields from PHP model
CREATE TABLE IF NOT EXISTS vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    vendor_code VARCHAR(100) NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_email VARCHAR(255) NOT NULL,
    vendor_mobile VARCHAR(20),
    vendor_type VARCHAR(50),
    vendor_nick_name VARCHAR(255),
    activity_type VARCHAR(255),
    email VARCHAR(255),
    mobile VARCHAR(20),
    gstin VARCHAR(50),
    pan VARCHAR(20) NOT NULL,
    pin VARCHAR(20),
    country_id VARCHAR(50),
    state_id VARCHAR(50),
    city_id VARCHAR(50),
    country_name VARCHAR(255),
    state_name VARCHAR(255),
    city_name VARCHAR(255),
    msme_classification VARCHAR(100),
    msme VARCHAR(100),
    msme_registration_number VARCHAR(100),
    msme_start_date DATE,
    msme_end_date DATE,
    material_nature VARCHAR(255),
    gst_defaulted VARCHAR(10),
    section_206ab_verified VARCHAR(10),
    beneficiary_name VARCHAR(255) NOT NULL,
    remarks_address TEXT,
    common_bank_details TEXT,
    income_tax_type VARCHAR(100),
    project VARCHAR(255),
    status VARCHAR(50),
    from_account_type VARCHAR(100),
    account_name VARCHAR(255),
    short_name VARCHAR(100),
    parent VARCHAR(255),
    file_paths JSONB,
    code_auto_generated BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Backward compatibility banking fields
    account_number VARCHAR(50),
    name_of_bank VARCHAR(255),
    ifsc_code VARCHAR(20),
    ifsc_code_id VARCHAR(50),
    
    -- Constraints
    CONSTRAINT vendors_tenant_vendor_code_unique UNIQUE (tenant_id, vendor_code),
    CONSTRAINT vendors_tenant_vendor_email_unique UNIQUE (tenant_id, vendor_email),
    CONSTRAINT vendors_pan_check CHECK (pan ~ '^[A-Z]{5}[0-9]{4}[A-Z]{1}$'),
    CONSTRAINT vendors_ifsc_check CHECK (ifsc_code IS NULL OR ifsc_code ~ '^[A-Z]{4}0[A-Z0-9]{6}$')
);

-- Create vendor_accounts table
CREATE TABLE IF NOT EXISTS vendor_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    account_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_type VARCHAR(50),
    name_of_bank VARCHAR(255) NOT NULL,
    branch_name VARCHAR(255),
    ifsc_code VARCHAR(20) NOT NULL,
    swift_code VARCHAR(20),
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    remarks TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT vendor_accounts_ifsc_check CHECK (ifsc_code ~ '^[A-Z]{4}0[A-Z0-9]{6}$'),
    CONSTRAINT vendor_accounts_account_number_check CHECK (account_number ~ '^[0-9]{9,18}$')
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_vendors_tenant_id ON vendors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vendors_vendor_code ON vendors(vendor_code);
CREATE INDEX IF NOT EXISTS idx_vendors_vendor_email ON vendors(vendor_email);
CREATE INDEX IF NOT EXISTS idx_vendors_is_active ON vendors(is_active);
CREATE INDEX IF NOT EXISTS idx_vendors_project ON vendors(project);
CREATE INDEX IF NOT EXISTS idx_vendors_vendor_type ON vendors(vendor_type);
CREATE INDEX IF NOT EXISTS idx_vendors_created_at ON vendors(created_at);

CREATE INDEX IF NOT EXISTS idx_vendor_accounts_vendor_id ON vendor_accounts(vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_is_primary ON vendor_accounts(is_primary);
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_is_active ON vendor_accounts(is_active);
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_created_at ON vendor_accounts(created_at);

-- Trigger to ensure only one primary account per vendor
CREATE OR REPLACE FUNCTION ensure_single_primary_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE vendor_accounts 
        SET is_primary = false, updated_at = NOW()
        WHERE vendor_id = NEW.vendor_id 
        AND id != NEW.id 
        AND is_primary = true;
    END IF;
    
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_primary_account
    BEFORE INSERT OR UPDATE ON vendor_accounts
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_primary_account();

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_vendors_updated_at
    BEFORE UPDATE ON vendors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_vendor_accounts_updated_at
    BEFORE UPDATE ON vendor_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
