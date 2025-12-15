package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Vendor CRUD operations
		"/vendor.v1.VendorService/CreateVendor":       {"create-vendor"},
		"/vendor.v1.VendorService/GetVendor":          {"view-vendors"},
		"/vendor.v1.VendorService/GetVendorByCode":    {"view-vendors"},
		"/vendor.v1.VendorService/ListVendors":        {"view-vendors"},
		"/vendor.v1.VendorService/UpdateVendor":       {"edit-vendor"},
		"/vendor.v1.VendorService/DeleteVendor":       {"delete-vendor"},
		
		// Vendor Code operations
		"/vendor.v1.VendorService/GenerateVendorCode":   {"create-vendor"},
		"/vendor.v1.VendorService/UpdateVendorCode":     {"edit-vendor"},
		"/vendor.v1.VendorService/RegenerateVendorCode": {"edit-vendor"},
		
		// Vendor Account operations
		"/vendor.v1.VendorService/CreateVendorAccount":     {"manage-vendor-accounts"},
		"/vendor.v1.VendorService/GetVendorAccounts":       {"view-vendors"},
		"/vendor.v1.VendorService/GetVendorBankingDetails": {"view-vendors"},
		"/vendor.v1.VendorService/UpdateVendorAccount":     {"manage-vendor-accounts"},
		"/vendor.v1.VendorService/DeleteVendorAccount":     {"manage-vendor-accounts"},
		"/vendor.v1.VendorService/ToggleAccountStatus":     {"manage-vendor-accounts"},
		
		// Dropdown operations
		"/vendor.v1.VendorService/GetProjectsDropdown":     {"view-vendors"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Vendor code generation can be public for UI
		"/vendor.v1.VendorService/GenerateVendorCode",
	}
}
