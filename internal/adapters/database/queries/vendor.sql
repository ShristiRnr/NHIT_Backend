-- name: CreateVendor :one
INSERT INTO vendors (
    s_no, from_account_type, status, project, account_name, short_name, parent,
    account_number, name_of_bank, ifsc_code_id, ifsc_code, vendor_type,
    vendor_code, vendor_name, vendor_email, vendor_mobile, activity_type,
    vendor_nick_name, email, mobile, gstin, pan, pin, country_id, state_id, city_id,
    country_name, state_name, city_name, msme_classification, msme,
    msme_registration_number, msme_start_date, msme_end_date, material_nature,
    gst_defaulted, section_206AB_verified, benificiary_name, remarks_address,
    common_bank_details, income_tax_type, file_path, active
) VALUES (
    @s_no, @from_account_type, @status, @project, @account_name, @short_name, @parent,
    @account_number, @name_of_bank, @ifsc_code_id, @ifsc_code, @vendor_type,
    @vendor_code, @vendor_name, @vendor_email, @vendor_mobile, @activity_type,
    @vendor_nick_name, @email, @mobile, @gstin, @pan, @pin, @country_id, @state_id, @city_id,
    @country_name, @state_name, @city_name, @msme_classification, @msme,
    @msme_registration_number, @msme_start_date, @msme_end_date, @material_nature,
    @gst_defaulted, @section_206AB_verified, @benificiary_name, @remarks_address,
    @common_bank_details, @income_tax_type, @file_path, @active
) RETURNING *;

-- name: GetVendor :one
SELECT * FROM vendors
WHERE id = $1 LIMIT 1;

-- name: ListVendors :many
SELECT * FROM vendors
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateVendor :one
UPDATE vendors SET
    s_no = @s_no,
    from_account_type = @from_account_type,
    status = @status,
    project = @project,
    account_name = @account_name,
    short_name = @short_name,
    parent = @parent,
    account_number = @account_number,
    name_of_bank = @name_of_bank,
    ifsc_code_id = @ifsc_code_id,
    ifsc_code = @ifsc_code,
    vendor_type = @vendor_type,
    vendor_code = @vendor_code,
    vendor_name = @vendor_name,
    vendor_email = @vendor_email,
    vendor_mobile = @vendor_mobile,
    activity_type = @activity_type,
    vendor_nick_name = @vendor_nick_name,
    email = @email,
    mobile = @mobile,
    gstin = @gstin,
    pan = @pan,
    pin = @pin,
    country_id = @country_id,
    state_id = @state_id,
    city_id = @city_id,
    country_name = @country_name,
    state_name = @state_name,
    city_name = @city_name,
    msme_classification = @msme_classification,
    msme = @msme,
    msme_registration_number = @msme_registration_number,
    msme_start_date = @msme_start_date,
    msme_end_date = @msme_end_date,
    material_nature = @material_nature,
    gst_defaulted = @gst_defaulted,
    section_206AB_verified = @section_206AB_verified,
    benificiary_name = @benificiary_name,
    remarks_address = @remarks_address,
    common_bank_details = @common_bank_details,
    income_tax_type = @income_tax_type,
    file_path = @file_path,
    active = @active
WHERE id = @id
RETURNING *;

-- name: DeleteVendor :exec
DELETE FROM vendors WHERE id = $1;

-- name: SearchVendors :many
SELECT * FROM vendors
WHERE vendor_name ILIKE '%' || $1 || '%'
   OR vendor_code ILIKE '%' || $1 || '%'
   OR vendor_email ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;